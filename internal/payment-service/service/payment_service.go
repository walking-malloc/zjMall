package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	paymentv1 "zjMall/gen/go/api/proto/payment"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/client"
	"zjMall/internal/common/lock"
	"zjMall/internal/common/middleware"
	"zjMall/internal/common/mq"
	"zjMall/internal/payment-service/model"
	"zjMall/internal/payment-service/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	PaymentTokenCacheKeyPrefix       = "payment:token"
	PaymentTokenCacheExpireSeconds   = 300                  // Token有效期5分钟
	PaymentIdempotencyKeyPrefix      = "payment:idempotent" // 幂等性key前缀
	PaymentIdempotencyExpireSeconds  = 300                  // 幂等性key有效期5分钟
	PaymentLockKeyPrefix             = "payment:lock"
	PaymentLockExpireSeconds         = 300                           // 锁有效期5分钟
	CallBackIdempotencyKeyPrefix     = "payment:callback:idempotent" // 回调幂等性key前缀
	CallBackIdempotencyExpireSeconds = 300                           // 回调幂等性key有效期5分钟
	PaySuccessNotifyTopic            = "payment:success:notify"
)

// PaymentService 支付服务
type PaymentService struct {
	paymentRepo        repository.PaymentRepository
	paymentLogRepo     repository.PaymentLogRepository
	paymentChannelRepo repository.PaymentChannelRepository
	paymentTimeout     time.Duration // 支付超时时间，默认30分钟
	orderClient        client.OrderClient
	cacheRepo          cache.CacheRepository
	lockService        lock.DistributedLockService
	paymentMQ          mq.MessageProducer
	outboxRepo         repository.PaymentOutboxRepository
}

// NewPaymentService 创建支付服务
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	paymentLogRepo repository.PaymentLogRepository,
	paymentChannelRepo repository.PaymentChannelRepository,
	orderClient client.OrderClient,
	cacheRepo cache.CacheRepository,
	paymentTimeout time.Duration,
	lockService lock.DistributedLockService,
	paymentMQ mq.MessageProducer,
	outboxRepo repository.PaymentOutboxRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo:        paymentRepo,
		paymentLogRepo:     paymentLogRepo,
		paymentChannelRepo: paymentChannelRepo,
		paymentTimeout:     paymentTimeout,
		orderClient:        orderClient,
		cacheRepo:          cacheRepo,
		lockService:        lockService,
		paymentMQ:          paymentMQ,
		outboxRepo:         outboxRepo,
	}
}

// CreatePaymentRequest 创建支付单请求
type CreatePaymentRequest struct {
	OrderNo    string
	UserID     string
	Amount     float64
	PayChannel string
	ReturnURL  string
}

// CreatePayment 创建支付单
// 注意：参数校验应该在 handler 层完成，service 层只做业务逻辑校验
func (s *PaymentService) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	//获取userId
	userId := middleware.GetUserIDFromContext(ctx)
	if userId == "" {
		return nil, fmt.Errorf("用户未登录")
	}
	// 幂等性key：基于用户ID和订单号（不依赖token，因为token可能为空）
	idempotencyKey := fmt.Sprintf("%s:%s:%s", PaymentIdempotencyKeyPrefix, userId, req.Token)

	// 先检查幂等性key是否存在（快速路径）
	existingPaymentNo, err := s.cacheRepo.Get(ctx, idempotencyKey)
	if err == nil && existingPaymentNo != "" {
		// 幂等性key存在，说明已经创建过支付单，直接查询并返回
		existingPayment, err := s.paymentRepo.GetPaymentByPaymentNo(ctx, existingPaymentNo)
		if err == nil && existingPayment != nil {
			return &paymentv1.CreatePaymentResponse{
				Code:    0,
				Message: "success",
				Payment: s.convertPaymentToProto(existingPayment),
			}, nil
		}
	}

	// 获取分布式锁（基于订单号，防止同一订单并发创建）
	lockKey := fmt.Sprintf("%s:%s", PaymentLockKeyPrefix, req.OrderNo)
	acquired, err := s.lockService.AcquireLock(ctx, lockKey, time.Duration(PaymentLockExpireSeconds)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("获取分布式锁失败: %w", err)
	}
	if !acquired {
		return nil, fmt.Errorf("系统繁忙，请稍后重试")
	}
	defer s.lockService.ReleaseLock(ctx, lockKey)

	// 获取锁后再次检查幂等性key（双重检查，防止并发）
	existingPaymentNo, err = s.cacheRepo.Get(ctx, idempotencyKey)
	if err == nil && existingPaymentNo != "" {
		existingPayment, err := s.paymentRepo.GetPaymentByPaymentNo(ctx, existingPaymentNo)
		if err == nil && existingPayment != nil {
			return &paymentv1.CreatePaymentResponse{
				Code:    0,
				Message: "success",
				Payment: s.convertPaymentToProto(existingPayment),
			}, nil
		}
	}

	//首先查看订单
	order, err := s.orderClient.GetOrderByNo(ctx, req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("订单不存在: %s", req.OrderNo)
	}
	//检查订单是否为待支付状态
	if int8(order.Status) != model.PaymentStatusPending {
		return nil, fmt.Errorf("订单状态不正确: %s", req.OrderNo)
	}
	//检查订单支付金额是否大于0
	payAmount, err := strconv.ParseFloat(order.PayAmount, 64)
	if err != nil {
		return nil, fmt.Errorf("订单支付金额格式错误: %w", err)
	}
	if payAmount <= 0 {
		return nil, fmt.Errorf("订单支付金额不能小于0")
	}
	// 检查是否已存在支付单（幂等性保障）
	existingPayment, err := s.paymentRepo.GetPaymentByOrderNo(ctx, req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("查询支付单失败: %w", err)
	}
	if existingPayment != nil {
		// 如果已存在且状态为待支付，直接返回（幂等处理）
		if existingPayment.Status == model.PaymentStatusPending {
			return &paymentv1.CreatePaymentResponse{
				Code:    0,
				Message: "success",
				Payment: s.convertPaymentToProto(existingPayment),
			}, nil
		}
		// 如果已存在但状态不是待支付，返回错误
		return nil, fmt.Errorf("订单已存在支付单，状态为: %d", existingPayment.Status)
	}
	// 检验支付渠道是否有效（req.PayChannel 是字符串，直接使用）
	paymentChannel, err := s.paymentChannelRepo.GetPaymentChannelByChannelCode(ctx, req.PayChannel)
	if err != nil {
		return nil, fmt.Errorf("查询支付渠道失败: %w", err)
	}
	if paymentChannel == nil {
		return nil, fmt.Errorf("支付渠道不存在: %s", req.PayChannel)
	}
	//生成支付单号
	paymentNo := s.generatePaymentNo()
	expiredAt := time.Now().Add(s.paymentTimeout)
	// 创建支付单
	payment := &model.Payment{
		PaymentNo:  paymentNo,
		OrderNo:    req.OrderNo,
		UserID:     userId,
		Amount:     payAmount,
		PayChannel: paymentChannel.ChannelCode, // 使用渠道代码，不是名称
		Status:     model.PaymentStatusPending,
		NotifyURL:  paymentChannel.NotifyURL,
		ReturnURL:  req.ReturnUrl, // 使用请求中的返回地址，如果没有则使用渠道配置的
		ExpiredAt:  &expiredAt,    //TODO:定时任务处理吗？
		Version:    1,
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		// 检查是否是唯一索引冲突（并发创建）
		if isDuplicateKeyError(err) {
			// 重新查询已存在的支付单
			existingPayment, _ := s.paymentRepo.GetPaymentByOrderNo(ctx, req.OrderNo)
			if existingPayment != nil {
				return &paymentv1.CreatePaymentResponse{
					Code:    0,
					Message: "success",
					Payment: s.convertPaymentToProto(existingPayment),
				}, nil
			}
		}
		return nil, fmt.Errorf("创建支付单失败: %w", err)
	}

	// 记录支付日志
	log := &model.PaymentLog{
		PaymentNo: paymentNo,
		OrderNo:   req.OrderNo,
		UserID:    userId,
		Action:    model.PaymentLogActionCreate,
		Channel:   paymentChannel.ChannelCode,
		Amount:    payAmount,
	}
	if err := s.paymentLogRepo.CreatePaymentLog(ctx, log); err != nil {
		fmt.Printf("⚠️ 记录支付日志失败: %v\n", err)
	}

	// 生成支付参数（学习模式：模拟支付参数，便于开发和测试）
	// 在实际生产环境中，这里应该调用第三方支付SDK（如支付宝、微信支付）
	payURL, qrCode, payParams := s.generatePaymentParamsForLearning(payment, paymentChannel)

	// 设置幂等性key（存储支付单号，有效期5分钟）
	// 这样后续相同请求可以直接返回已创建的支付单
	if err := s.cacheRepo.Set(ctx, idempotencyKey, paymentNo, time.Duration(PaymentIdempotencyExpireSeconds)*time.Second); err != nil {
		fmt.Printf("⚠️ 设置幂等性key失败: %v\n", err)
		// 不影响主流程，继续返回
	}
	go func() {
		//TODO:异步调用支付网关

	}()

	return &paymentv1.CreatePaymentResponse{
		Code:      0,
		Message:   "success",
		Payment:   s.convertPaymentToProto(payment),
		PayUrl:    payURL,
		QrCode:    qrCode,
		PayParams: payParams,
	}, nil
}

// GetPayment 查询支付单
func (s *PaymentService) GetPayment(ctx context.Context, paymentNo string) (*model.Payment, error) {
	if paymentNo == "" {
		return nil, fmt.Errorf("支付单号不能为空")
	}

	payment, err := s.paymentRepo.GetPaymentByPaymentNo(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("查询支付单失败: %w", err)
	}

	return payment, nil
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	PayChannel  string
	PaymentNo   string
	TradeNo     string
	Amount      string
	Status      string
	Sign        string
	ExtraParams map[string]string
}

// HandlePaymentCallback 处理支付回调
func (s *PaymentService) HandlePaymentCallback(ctx context.Context, req *PaymentCallbackRequest) error {
	// 1. 参数校验
	if req.PaymentNo == "" {
		return fmt.Errorf("支付单号不能为空")
	}
	if req.TradeNo == "" {
		return fmt.Errorf("第三方交易号不能为空")
	}
	if req.Amount == "" {
		return fmt.Errorf("支付金额不能为空")
	}
	//幂等性key：基于支付单号和交易号
	idempotencyKey := fmt.Sprintf("%s:%s:%s", CallBackIdempotencyKeyPrefix, req.PaymentNo, req.TradeNo)
	result, err := s.cacheRepo.Get(ctx, idempotencyKey)
	if err == nil && result != "" {
		if result == "SUCCESS" {
			return nil // 已成功处理
		} else if result == "PROCESSING" {
			return fmt.Errorf("支付回调正在处理中，请勿重复提交")
		} else {
			return fmt.Errorf("上次处理失败: %s", result)
		}
	}
	if ok, err := s.cacheRepo.SetNX(ctx, idempotencyKey, "PROCESSING", time.Duration(CallBackIdempotencyExpireSeconds)*time.Second); err != nil || !ok {
		log.Printf("⚠️ 设置幂等性key失败: %v\n", err)
		return fmt.Errorf("系统繁忙，请稍后重试")
	}
	//获取分布式锁，避免重复回调
	lockKey := fmt.Sprintf("%s:%s", PaymentLockKeyPrefix, req.PaymentNo)
	acquired, err := s.lockService.AcquireLock(ctx, lockKey, time.Duration(PaymentLockExpireSeconds)*time.Second)
	if err != nil {
		return fmt.Errorf("获取分布式锁失败: %w", err)
	}
	if !acquired {
		return fmt.Errorf("系统繁忙，请稍后重试")
	}
	defer s.lockService.ReleaseLock(ctx, lockKey)

	// 2. 查询支付单
	payment, err := s.paymentRepo.GetPaymentByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		return fmt.Errorf("查询支付单失败: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("支付单不存在: %s", req.PaymentNo)
	}

	// 3. 幂等性校验：如果已经是支付成功状态，直接返回成功
	if payment.Status == model.PaymentStatusSuccess {
		return nil // 幂等处理
	}
	//检查交易号是否被其他订单使用（防止重复入账）
	otherPayment, err := s.paymentRepo.GetPaymentByTradeNo(ctx, req.TradeNo)
	if err != nil {
		return fmt.Errorf("查询支付单失败: %w", err)
	}
	if otherPayment != nil && otherPayment.PaymentNo != payment.PaymentNo {
		return fmt.Errorf("交易号已存在: %s, 支付单号: %s", req.TradeNo, otherPayment.PaymentNo)
	}
	// 4. 签名校验（TODO: 后续实现），校验是否是平台发来的回调，防止伪造回调 采用支付宝公钥对签名进行校验
	// if !s.verifySign(req) {
	//     return fmt.Errorf("签名校验失败")
	// }

	// 5. 金额校验
	callbackAmount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return fmt.Errorf("支付金额格式错误: %w", err)
	}

	if callbackAmount != payment.Amount {
		// 记录告警日志
		log.Printf("⚠️ 支付金额不一致: payment_no=%s, 订单金额=%.2f, 回调金额=%.2f\n",
			req.PaymentNo, payment.Amount, callbackAmount)
		return fmt.Errorf("支付金额不一致: 订单金额=%.2f, 回调金额=%.2f", payment.Amount, callbackAmount)
	}

	// 6. 判断支付状态
	var newStatus int8
	if req.Status == "success" || req.Status == "SUCCESS" {
		newStatus = model.PaymentStatusSuccess
	} else {
		newStatus = model.PaymentStatusFailed
	}

	// 7. 更新支付单状态（使用乐观锁）
	oldStatus := payment.Status
	payment.Status = newStatus
	payment.TradeNo = req.TradeNo
	if newStatus == model.PaymentStatusSuccess {
		now := time.Now()
		payment.PaidAt = &now
	}

	// 使用 Outbox 模式：在一个本地事务中更新支付单、记录日志、写入 Outbox 事件
	if err := s.paymentRepo.WithTransaction(ctx, func(txCtx context.Context, txRepo repository.PaymentRepository) error {
		// 7.1 更新支付单状态
		if err := txRepo.UpdatePayment(txCtx, payment); err != nil {
			return fmt.Errorf("更新支付单状态失败: %w", err)
		}

		// 7.2 记录支付日志
		paymentLog := &model.PaymentLog{
			PaymentNo:  req.PaymentNo,
			OrderNo:    payment.OrderNo,
			UserID:     payment.UserID,
			Action:     model.PaymentLogActionCallback,
			FromStatus: &oldStatus,
			ToStatus:   &newStatus,
			Channel:    req.PayChannel,
			Amount:     payment.Amount,
			TradeNo:    req.TradeNo,
		}
		if err := s.paymentLogRepo.CreatePaymentLog(txCtx, paymentLog); err != nil {
			return fmt.Errorf("记录支付日志失败: %w", err)
		}

		// 7.3 仅在支付成功时写入 Outbox 事件
		if newStatus == model.PaymentStatusSuccess {
			payload := map[string]interface{}{
				"payment_no": payment.PaymentNo,
				"order_no":   payment.OrderNo,
				"user_id":    payment.UserID,
				"amount":     payment.Amount,
				"trade_no":   payment.TradeNo,
				"paid_at":    payment.PaidAt,
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("序列化支付成功事件失败: %w", err)
			}

			event := &model.PaymentOutbox{
				EventType:   "payment.succeeded",
				AggregateID: payment.PaymentNo,
				Payload:     string(payloadBytes),
				Status:      repository.OutboxStatusPending,
				RetryCount:  0,
			}

			if err := s.outboxRepo.Create(txCtx, event); err != nil {
				return fmt.Errorf("写入支付 Outbox 事件失败: %w", err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// 设置幂等性key为成功（放在事务之外，失败不影响主流程）
	if err := s.cacheRepo.Set(ctx, idempotencyKey, "SUCCESS", 24*time.Hour); err != nil {
		log.Printf("⚠️ 设置回调幂等性key失败: %v\n", err)
	}

	return nil
}

// QueryPaymentStatus 查询支付状态
func (s *PaymentService) QueryPaymentStatus(ctx context.Context, paymentNo string) (int8, string, error) {
	if paymentNo == "" {
		return 0, "", fmt.Errorf("支付单号不能为空")
	}

	payment, err := s.paymentRepo.GetPaymentByPaymentNo(ctx, paymentNo)
	if err != nil {
		return 0, "", fmt.Errorf("查询支付单失败: %w", err)
	}
	if payment == nil {
		return 0, "", fmt.Errorf("支付单不存在: %s", paymentNo)
	}
	//TODO:超过一定时间需要对账和异常处理
	return payment.Status, payment.TradeNo, nil
}

// CloseExpiredPayments 关闭超时的支付单（定时任务调用）
func (s *PaymentService) CloseExpiredPayments(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = 100 // 默认每次处理100条
	}

	// 查询超时的支付单
	expiredPayments, err := s.paymentRepo.GetExpiredPayments(ctx, limit)
	if err != nil {
		return fmt.Errorf("查询超时支付单失败: %w", err)
	}

	for _, payment := range expiredPayments {
		// 更新支付单状态为已关闭
		oldStatus := payment.Status
		payment.Status = model.PaymentStatusClosed

		if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
			fmt.Printf("⚠️ 关闭支付单失败: payment_no=%s, err=%v\n", payment.PaymentNo, err)
			continue
		}

		// 记录支付日志
		log := &model.PaymentLog{
			PaymentNo:  payment.PaymentNo,
			OrderNo:    payment.OrderNo,
			UserID:     payment.UserID,
			Action:     model.PaymentLogActionClose,
			FromStatus: &oldStatus,
			ToStatus:   &payment.Status,
			Channel:    payment.PayChannel,
			Amount:     payment.Amount,
		}
		if err := s.paymentLogRepo.CreatePaymentLog(ctx, log); err != nil {
			fmt.Printf("⚠️ 记录支付日志失败: %v\n", err)
		}
	}

	return nil
}

// generatePaymentNo 生成支付单号
// 格式：{前缀(2位)}{日期时间(12位)}{随机数(6位)}{扩展位(2位)} = 总共22位
func (s *PaymentService) generatePaymentNo() string {
	// 生成日期时间字符串（12位：YYYYMMDDHHmm）
	dateTime := time.Now().Format("200601021504")

	// 生成6位随机数（000000-999999）
	// 使用 math/rand 与订单号生成方式保持一致
	randomNum := rand.Intn(1000000)

	// 组合：前缀(2) + 日期时间(12) + 随机数(6) + 扩展位(2) = 22位
	paymentNo := fmt.Sprintf("%s%s%06d00",
		model.PaymentNoPrefix, // 10
		dateTime,              // 202401301430
		randomNum,             // 123456
		// 00 扩展位已包含在格式字符串中
	)

	return paymentNo
}

// GeneratePaymentToken 生成支付幂等性Token
func (s *PaymentService) GeneratePaymentToken(ctx context.Context, req *paymentv1.GeneratePaymentTokenRequest) (*paymentv1.GeneratePaymentTokenResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &paymentv1.GeneratePaymentTokenResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	token := fmt.Sprintf("%s%s%s", userID, req.OrderNo, time.Now().Format("20060102150405"))

	// Token有效期5分钟（300秒）
	cacheKey := fmt.Sprintf("%s:%s:%s", PaymentTokenCacheKeyPrefix, userID, token)
	set, err := s.cacheRepo.SetNX(ctx, cacheKey, "1", time.Duration(PaymentTokenCacheExpireSeconds)*time.Second)
	if err != nil {
		fmt.Printf("❌ [PaymentService] GeneratePaymentToken: 设置Token失败: %v\n", err)
		return &paymentv1.GeneratePaymentTokenResponse{
			Code:    1,
			Message: "系统繁忙，请稍后重试",
		}, nil
	}
	if !set {
		return &paymentv1.GeneratePaymentTokenResponse{
			Code:    1,
			Message: "Token已存在，请勿重复生成",
		}, nil
	}

	return &paymentv1.GeneratePaymentTokenResponse{
		Code:          0,
		Message:       "success",
		Token:         token,
		ExpireSeconds: PaymentTokenCacheExpireSeconds,
	}, nil
}

// convertPaymentToProto 转换 Payment 模型为 proto 消息
func (s *PaymentService) convertPaymentToProto(payment *model.Payment) *paymentv1.Payment {
	if payment == nil {
		return nil
	}

	protoPayment := &paymentv1.Payment{
		Id:         payment.ID,
		PaymentNo:  payment.PaymentNo,
		OrderNo:    payment.OrderNo,
		UserId:     payment.UserID,
		Amount:     fmt.Sprintf("%.2f", payment.Amount),
		PayChannel: payment.PayChannel, // PayChannel 在 proto 中是 string 类型，直接使用
		Status:     s.convertStatusToProto(payment.Status),
		TradeNo:    payment.TradeNo,
		NotifyUrl:  payment.NotifyURL,
		ReturnUrl:  payment.ReturnURL,
		CreatedAt:  timestamppb.New(payment.CreatedAt),
	}

	if payment.PaidAt != nil {
		protoPayment.PaidAt = timestamppb.New(*payment.PaidAt)
	}
	if payment.ExpiredAt != nil {
		protoPayment.ExpiredAt = timestamppb.New(*payment.ExpiredAt)
	}

	return protoPayment
}

// convertStatusToProto 转换支付状态为 proto PaymentStatus
func (s *PaymentService) convertStatusToProto(status int8) paymentv1.PaymentStatus {
	switch status {
	case model.PaymentStatusPending:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING
	case model.PaymentStatusProcessing:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PROCESSING
	case model.PaymentStatusSuccess:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS
	case model.PaymentStatusFailed:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
	case model.PaymentStatusClosed:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_CLOSED
	case model.PaymentStatusRefunded:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_REFUNDED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

// isDuplicateKeyError 检查是否是唯一索引冲突错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate entry") ||
		strings.Contains(errStr, "UNIQUE constraint") ||
		strings.Contains(errStr, "duplicate key")
}

// generatePaymentParamsForLearning 生成支付参数（学习模式：模拟支付参数）
// 在实际生产环境中，这里应该调用第三方支付SDK（如支付宝、微信支付）
// 学习模式下返回模拟的支付URL、二维码和支付参数，便于前端开发和测试
func (s *PaymentService) generatePaymentParamsForLearning(payment *model.Payment, channel *model.PaymentChannel) (payURL, qrCode string, payParams map[string]string) {
	payParams = make(map[string]string)

	// 根据支付渠道生成不同的模拟参数
	switch channel.ChannelCode {
	case model.PayChannelWeChat:
		// 微信支付 - 学习模式模拟
		// H5/PC支付URL（模拟）
		payURL = fmt.Sprintf("https://pay.weixin.qq.com/learning?payment_no=%s&amount=%.2f&order_no=%s",
			payment.PaymentNo, payment.Amount, payment.OrderNo)

		// 二维码（模拟微信支付码串，前端可以生成二维码）
		qrCode = fmt.Sprintf("weixin://wxpay/bizpayurl?pr=LEARNING_%s", payment.PaymentNo)

		// APP支付参数（模拟）
		payParams = map[string]string{
			"appId":     channel.AppID,
			"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
			"nonceStr":  fmt.Sprintf("learning_%s", payment.PaymentNo),
			"package":   fmt.Sprintf("prepay_id=LEARNING_%s", payment.PaymentNo),
			"signType":  "RSA",
			"paySign":   "LEARNING_MODE_SIGNATURE", // 学习模式签名（非真实签名）
		}

	case model.PayChannelAlipay:
		// 支付宝支付 - 学习模式模拟
		// H5/PC支付URL（模拟支付宝沙箱环境）
		payURL = fmt.Sprintf("https://openapi.alipaydev.com/gateway.do?payment_no=%s&amount=%.2f&order_no=%s",
			payment.PaymentNo, payment.Amount, payment.OrderNo)

		// 二维码（模拟支付宝二维码URL）
		qrCode = fmt.Sprintf("https://qr.alipay.com/learning_%s", payment.PaymentNo)

		// APP支付参数（模拟）
		payParams = map[string]string{
			"app_id":    channel.AppID,
			"method":    "alipay.trade.app.pay",
			"charset":   "utf-8",
			"sign_type": "RSA2",
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"version":   "1.0",
			"biz_content": fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"学习模式订单-%s"}`,
				payment.PaymentNo, payment.Amount, payment.OrderNo),
			"sign": "LEARNING_MODE_SIGNATURE", // 学习模式签名（非真实签名）
		}

	case model.PayChannelBalance:
		// 余额支付 - 不需要跳转，直接扣款
		payURL = ""
		qrCode = ""
		payParams = map[string]string{
			"payment_no": payment.PaymentNo,
			"amount":     fmt.Sprintf("%.2f", payment.Amount),
			"channel":    "balance",
			"note":       "余额支付，无需跳转",
		}

	default:
		// 未知渠道，返回空参数
		fmt.Printf("⚠️ 未知支付渠道: %s\n", channel.ChannelCode)
		return "", "", make(map[string]string)
	}

	// 学习模式提示日志
	fmt.Printf("📚 [学习模式] 支付参数已生成 - 支付单号: %s, 渠道: %s, 金额: %.2f\n",
		payment.PaymentNo, channel.ChannelCode, payment.Amount)
	fmt.Printf("   - PayURL: %s\n", payURL)
	fmt.Printf("   - QRCode: %s\n", qrCode)
	fmt.Printf("   - PayParams: %v\n", payParams)

	return payURL, qrCode, payParams
}
