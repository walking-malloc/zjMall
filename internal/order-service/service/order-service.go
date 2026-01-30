package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	inventoryv1 "zjMall/gen/go/api/proto/inventory"
	orderv1 "zjMall/gen/go/api/proto/order"
	"zjMall/internal/common/client"
	"zjMall/internal/common/lock"
	"zjMall/internal/common/middleware"
	"zjMall/internal/order-service/model"
	"zjMall/internal/order-service/repository"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	// 正向流程
	OrderStatusPendingPay = int8(orderv1.OrderStatus_ORDER_STATUS_PENDING_PAY) // 待支付（可取消）
	OrderStatusPaid       = int8(orderv1.OrderStatus_ORDER_STATUS_PAID)        // 已支付（可退款）
	OrderStatusShipped    = int8(orderv1.OrderStatus_ORDER_STATUS_SHIPPED)     // 已发货（可收货/退货）
	OrderStatusCompleted  = int8(orderv1.OrderStatus_ORDER_STATUS_COMPLETED)   // 已完成（不可修改）

	// 逆向流程
	OrderStatusCancelled = int8(orderv1.OrderStatus_ORDER_STATUS_CANCELLED) // 已取消（用户主动）
	OrderStatusRefunding = int8(orderv1.OrderStatus_ORDER_STATUS_REFUNDING) // 退款中
	OrderStatusRefunded  = int8(orderv1.OrderStatus_ORDER_STATUS_REFUNDED)  // 已退款
	OrderStatusClosed    = int8(orderv1.OrderStatus_ORDER_STATUS_CLOSED)    // 已关闭（超时自动）

	OrderTokenCacheKeyPrefix      = "order:token"
	OrderTokenCacheExpireSeconds  = 300
	OrderIdempotentCacheKeyPrefix = "order:idempotent"
)

// OrderService 订单服务（业务逻辑层）
type OrderService struct {
	orderRepo       repository.OrderRepository
	productClient   client.ProductClient
	inventoryClient client.InventoryClient
	userClient      client.UserClient
	cartClient      client.CartClient
	redisClient     *redis.Client
}

func NewOrderService(orderRepo repository.OrderRepository, productClient client.ProductClient, inventoryClient client.InventoryClient, userClient client.UserClient, cartClient client.CartClient, redisClient *redis.Client) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		productClient:   productClient,
		inventoryClient: inventoryClient,
		userClient:      userClient,
		cartClient:      cartClient,
		redisClient:     redisClient,
	}
}

// 防止重复生成订单，前端提交token，后端消费并删除，然后获取分布式锁
func (s *OrderService) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	if len(req.Items) == 0 {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "订单商品不能为空",
		}, nil
	}
	//检查token
	if req.Token == "" {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "Token不能为空",
		}, nil
	}

	//获取分布式锁
	lockKey := fmt.Sprintf("%s:%s:%s", OrderIdempotentCacheKeyPrefix, userID, req.Token)
	lockService := lock.NewRedisLockService(s.redisClient)
	lockAcquired, err := lockService.AcquireLock(ctx, lockKey, 10*time.Second)
	if err != nil || !lockAcquired {
		log.Printf("❌ [OrderService] CreateOrder: 获取分布式锁失败: %v", err)
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "系统繁忙，请稍后重试",
		}, nil
	}

	defer func() {
		if err := lockService.ReleaseLock(ctx, lockKey); err != nil {
			log.Printf("⚠️ [OrderService] CreateOrder: 释放锁失败: %v", err)
		}
	}()

	// 检查并消费Token
	tokenValid, err := s.checkAndConsumeToken(ctx, userID, req.Token)
	if err != nil {
		log.Printf("❌ [OrderService] CreateOrder: 检查Token失败: %v", err)
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "系统繁忙，请稍后重试",
		}, nil
	}
	if !tokenValid {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "Token已失效或已使用",
		}, nil
	}
	// 生成订单号（依赖数据库唯一索引保证唯一性）
	// 注意：不再提前检查订单号是否存在，因为：
	// 1. 检查和使用之间存在时间窗口，无法避免并发冲突
	// 2. 数据库唯一索引会保证订单号唯一性
	// 3. 如果订单号冲突，创建订单时会失败，然后回滚库存
	var orderNo string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		orderNo = orderNoGenerator(model.OrderTypeNormal)
		// 直接使用订单号，如果冲突会在创建订单时被数据库唯一索引捕获
		if orderNo != "" {
			break
		}
	}
	if orderNo == "" {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "订单号生成失败，请重试",
		}, nil
	}
	var totalAmount float64
	// 存储商品和SKU信息，用于填充订单明细快照
	type itemSnapshot struct {
		productTitle string
		productImage string
		skuName      string
		price        float64
	}
	itemSnapshots := make(map[string]*itemSnapshot) // key: skuId

	for _, item := range req.Items {
		// 检查商品是否存在
		product, skus, err := s.productClient.GetProduct(ctx, item.ProductId)
		if err != nil || product == nil || len(skus) == 0 {
			return &orderv1.CreateOrderResponse{
				Code:    1,
				Message: fmt.Sprintf("商品%s不存在或SKU不存在", item.ProductId),
			}, nil
		}
		// 注意：不再提前检查库存，因为 DeductStock 会使用乐观锁和 WHERE 条件检查库存
		// 这样可以避免时间窗口问题，并且减少一次网络调用
		// 查找对应的SKU并保存快照信息
		found := false
		for _, sku := range skus {
			if sku.Id == item.SkuId {
				totalAmount += float64(item.Quantity) * sku.Price
				itemSnapshots[item.SkuId] = &itemSnapshot{
					productTitle: product.Title,
					productImage: product.MainImage,
					skuName:      sku.Name,
					price:        sku.Price,
				}
				found = true
				break
			}
		}
		if !found {
			return &orderv1.CreateOrderResponse{
				Code:    1,
				Message: fmt.Sprintf("SKU %s 不存在", item.SkuId),
			}, nil
		}
	}

	// 计算订单金额
	discountAmount := 0.0 // TODO:待完成优惠计算
	shippingAmount := 0.0 // TODO:需要根据订单类型和收货地址计算运费
	payAmount := totalAmount - discountAmount + shippingAmount
	if payAmount < 0 {
		payAmount = 0
	}

	// 获取用户地址
	log.Printf("req.AddressId: %s", req.AddressId)
	userAddress, err := s.userClient.GetUserAddress(ctx, req.AddressId)
	log.Printf("userAddress: %+v, err: %v", userAddress, err)
	if err != nil || userAddress == nil {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: fmt.Sprintf("获取用户地址失败: %v", err),
		}, nil
	}

	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		Status:          int8(OrderStatusPendingPay),
		TotalAmount:     totalAmount,
		DiscountAmount:  discountAmount,
		ShippingAmount:  shippingAmount,
		PayAmount:       payAmount,
		BuyerRemark:     req.BuyerRemark,
		ReceiverName:    userAddress.ReceiverName,
		ReceiverPhone:   userAddress.ReceiverPhone,
		ReceiverAddress: fmt.Sprintf("%s%s%s%s", userAddress.Province, userAddress.City, userAddress.District, userAddress.Detail),
		Version:         1, // 初始化版本号为1
	}

	// 创建订单明细（填充商品快照信息）
	var items []*model.OrderItem
	var deductItems []*inventoryv1.SkuQuantity // 用于库存扣减
	for _, it := range req.Items {
		snapshot := itemSnapshots[it.SkuId]
		if snapshot == nil {
			return &orderv1.CreateOrderResponse{
				Code:    1,
				Message: fmt.Sprintf("SKU %s 快照信息缺失", it.SkuId),
			}, nil
		}
		item := &model.OrderItem{
			OrderNo:      orderNo,
			UserID:       userID,
			ProductID:    it.ProductId,
			SKUID:        it.SkuId,
			ProductTitle: snapshot.productTitle,
			ProductImage: snapshot.productImage,
			SKUName:      snapshot.skuName,
			Price:        snapshot.price,
			Quantity:     it.Quantity,
			Subtotal:     float64(it.Quantity) * snapshot.price,
		}
		items = append(items, item)
		deductItems = append(deductItems, &inventoryv1.SkuQuantity{
			SkuId:    it.SkuId,
			Quantity: int64(it.Quantity),
		})
	}

	// 先扣减库存（在创建订单之前，防止超卖）
	// 注意：这里使用订单号作为幂等键，如果订单创建失败，会回滚库存
	if err := s.inventoryClient.DeductStock(ctx, orderNo, deductItems); err != nil {
		log.Printf("❌ [OrderService] CreateOrder: 扣减库存失败: %v", err)
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: fmt.Sprintf("库存扣减失败: %v", err),
		}, nil
	}

	// 创建订单（在事务中）
	if err := s.orderRepo.CreateOrder(ctx, order, items); err != nil {
		log.Printf("❌ [OrderService] CreateOrder: 创建订单失败，回滚库存: %v", err)

		// 检查是否是订单号冲突错误（唯一索引冲突）
		isDuplicateOrderNo := false
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			isDuplicateOrderNo = true
		} else if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") ||
			strings.Contains(err.Error(), "duplicate key") {
			isDuplicateOrderNo = true
		}

		// 订单创建失败，回滚库存
		// 注意：如果是订单号冲突，库存服务会幂等返回（因为订单号已存在），但为了安全还是尝试回滚
		if rollbackErr := s.inventoryClient.RollbackStock(ctx, orderNo, deductItems); rollbackErr != nil {
			log.Printf("❌ [OrderService] CreateOrder: 回滚库存失败: %v", rollbackErr)
			// 记录告警，需要人工介入
		}

		if isDuplicateOrderNo {
			// 订单号冲突，建议用户重试
			return &orderv1.CreateOrderResponse{
				Code:    1,
				Message: "订单号冲突，请重试",
			}, nil
		}

		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: fmt.Sprintf("创建订单失败: %v", err),
		}, nil
	}

	// 订单创建成功，删除购物车中对应的商品
	// 收集购物车项ID（如果有的话）
	var cartItemIDs []string
	for _, it := range req.Items {
		if it.CartItemId != "" {
			cartItemIDs = append(cartItemIDs, it.CartItemId)
		}
	}

	// 如果有购物车项ID，调用购物车服务删除
	if len(cartItemIDs) > 0 && s.cartClient != nil {
		if err := s.cartClient.RemoveItems(ctx, cartItemIDs); err != nil {
			// 购物车删除失败不影响订单创建成功，只记录日志
			log.Printf("⚠️ [OrderService] CreateOrder: 删除购物车项失败（订单已创建成功）: %v", err)
		} else {
			log.Printf("✅ [OrderService] CreateOrder: 成功删除 %d 个购物车项", len(cartItemIDs))
		}
	}

	return &orderv1.CreateOrderResponse{
		Code:      0,
		Message:   "创建成功",
		OrderNo:   orderNo,
		PayAmount: fmt.Sprintf("%.2f", payAmount),
	}, nil
}

// GetOrder 获取订单详情（根据订单号查询订单主表和明细表）
func (s *OrderService) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	// 1. 校验用户登录状态
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		log.Printf("⚠️ [OrderService] GetOrder: 用户未登录")
		return &orderv1.GetOrderResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	// 2. 校验订单号
	if req.OrderNo == "" {
		log.Printf("⚠️ [OrderService] GetOrder: 订单号为空")
		return &orderv1.GetOrderResponse{
			Code:    1,
			Message: "订单号不能为空",
		}, nil
	}

	// 3. 查询订单主表和明细表
	order, items, err := s.orderRepo.GetOrderByNo(ctx, userID, req.OrderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("⚠️ [OrderService] GetOrder: 订单不存在, orderNo=%s, userID=%s", req.OrderNo, userID)
			return &orderv1.GetOrderResponse{
				Code:    1,
				Message: "订单不存在或无权访问",
			}, nil
		}
		log.Printf("❌ [OrderService] GetOrder: 查询订单失败, orderNo=%s, userID=%s, error=%v", req.OrderNo, userID, err)
		return &orderv1.GetOrderResponse{
			Code:    1,
			Message: "查询订单失败，请稍后重试",
		}, nil
	}

	// 4. 转换并返回数据
	log.Printf("✅ [OrderService] GetOrder: 查询成功, orderNo=%s, userID=%s, itemCount=%d", req.OrderNo, userID, len(items))
	return &orderv1.GetOrderResponse{
		Code:    0,
		Message: "查询成功",
		Order:   convertOrderToProto(order),
		Items:   convertOrderItemsToProto(items),
	}, nil
}

// ListUserOrders 我的订单列表
func (s *OrderService) ListUserOrders(ctx context.Context, req *orderv1.ListUserOrdersRequest) (*orderv1.ListUserOrdersResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &orderv1.ListUserOrdersResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Page > 100 {
		req.Page = 100
	}
	offset := int((req.Page - 1) * req.PageSize)
	limit := int(req.PageSize)

	orders, total, err := s.orderRepo.ListUserOrders(ctx, userID, int8(req.Status), offset, limit)
	if err != nil {
		log.Printf("❌ [OrderService] ListUserOrders: 查询失败: %v", err)
		return &orderv1.ListUserOrdersResponse{
			Code:    1,
			Message: "查询失败",
		}, nil
	}

	var protoOrders []*orderv1.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, convertOrderToProto(o))
	}

	return &orderv1.ListUserOrdersResponse{
		Code:    0,
		Message: "查询成功",
		Orders:  protoOrders,
		Total:   int32(total),
	}, nil
}

// CancelOrder 用户取消订单（仅待支付）
func (s *OrderService) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &orderv1.CancelOrderResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	// 更新订单状态（使用乐观锁）
	if err := s.orderRepo.UpdateOrderStatus(ctx, req.OrderNo,
		int8(OrderStatusPendingPay),
		int8(OrderStatusCancelled)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("⚠️ [OrderService] CancelOrder: 订单状态已被其他请求修改: %v", err)
			return &orderv1.CancelOrderResponse{
				Code:    1,
				Message: "订单状态已变更，请刷新后重试",
			}, nil
		}
		log.Printf("❌ [OrderService] CancelOrder: 取消订单失败: %v", err)
		return &orderv1.CancelOrderResponse{
			Code:    1,
			Message: "取消订单失败",
		}, nil
	}

	// 回滚库存（订单取消时释放库存）
	_, orderItems, err := s.orderRepo.GetOrderByNo(ctx, userID, req.OrderNo)
	if err == nil {
		var rollbackItems []*inventoryv1.SkuQuantity
		for _, item := range orderItems {
			rollbackItems = append(rollbackItems, &inventoryv1.SkuQuantity{
				SkuId:    item.SKUID,
				Quantity: int64(item.Quantity),
			})
		}
		if len(rollbackItems) > 0 {
			if rollbackErr := s.inventoryClient.RollbackStock(ctx, req.OrderNo, rollbackItems); rollbackErr != nil {
				log.Printf("❌ [OrderService] CancelOrder: 回滚库存失败: %v", rollbackErr)
				// 记录告警，但不影响订单取消流程
			}
		}
	}

	return &orderv1.CancelOrderResponse{
		Code:    0,
		Message: "取消成功",
	}, nil
}

// MarkOrderPaid 标记订单已支付（简化版）
func (s *OrderService) MarkOrderPaid(ctx context.Context, req *orderv1.MarkOrderPaidRequest) (*orderv1.MarkOrderPaidResponse, error) {

	now := time.Now()
	if err := s.orderRepo.UpdateOrderPaid(ctx, req.OrderNo, OrderStatusPendingPay, OrderStatusPaid, req.PayChannel, req.PayTradeNo, now); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("⚠️ [OrderService] MarkOrderPaid: 订单状态已被其他请求修改（可能是重复回调）: %v", err)
			return &orderv1.MarkOrderPaidResponse{
				Code:    0, // 幂等返回成功
				Message: "订单已处理",
			}, nil
		}
		log.Printf("❌ [OrderService] MarkOrderPaid: 更新订单支付状态失败: %v", err)
		return &orderv1.MarkOrderPaidResponse{
			Code:    1,
			Message: "更新订单状态失败",
		}, nil
	}

	return &orderv1.MarkOrderPaidResponse{
		Code:    0,
		Message: "更新成功",
	}, nil
}

// ======== 辅助转换函数 ========

func convertOrderToProto(o *model.Order) *orderv1.Order {
	res := &orderv1.Order{
		OrderNo:         o.OrderNo,
		UserId:          o.UserID,
		Status:          orderv1.OrderStatus(o.Status),
		TotalAmount:     fmt.Sprintf("%.2f", o.TotalAmount),
		DiscountAmount:  fmt.Sprintf("%.2f", o.DiscountAmount),
		ShippingAmount:  fmt.Sprintf("%.2f", o.ShippingAmount),
		PayAmount:       fmt.Sprintf("%.2f", o.PayAmount),
		PayChannel:      o.PayChannel,
		ReceiverName:    o.ReceiverName,
		ReceiverPhone:   o.ReceiverPhone,
		ReceiverAddress: o.ReceiverAddress,
		BuyerRemark:     o.BuyerRemark,
		CreatedAt:       timestamppb.New(o.CreatedAt),
		PaidAt:          nil,
		ShippedAt:       nil,
		CompletedAt:     nil,
	}
	// 只有当时间字段不为 nil 时才设置
	if o.PaidAt != nil {
		res.PaidAt = timestamppb.New(*o.PaidAt)
	}
	if o.ShippedAt != nil {
		res.ShippedAt = timestamppb.New(*o.ShippedAt)
	}
	if o.CompletedAt != nil {
		res.CompletedAt = timestamppb.New(*o.CompletedAt)
	}
	return res
}

func convertOrderItemsToProto(items []*model.OrderItem) []*orderv1.OrderItem {
	var res []*orderv1.OrderItem
	for _, it := range items {
		res = append(res, &orderv1.OrderItem{
			Id:             it.ID,
			OrderNo:        it.OrderNo,
			ProductId:      it.ProductID,
			SkuId:          it.SKUID,
			ProductTitle:   it.ProductTitle,
			ProductImage:   it.ProductImage,
			SkuName:        it.SKUName,
			Price:          fmt.Sprintf("%.2f", it.Price),
			Quantity:       it.Quantity,
			SubtotalAmount: fmt.Sprintf("%.2f", it.Subtotal),
		})
	}
	return res
}
func orderNoGenerator(orderType string) string {
	var orderNo string
	date := time.Now().Format("200601021504") //12位
	if orderType == model.OrderTypeNormal {
		orderNo = fmt.Sprintf("%s%s%06d", model.OrderTypeNormalPrefix, date, rand.Intn(1000000))
	} else if orderType == model.OrderTypeSeckill {
		orderNo = fmt.Sprintf("%s%s%06d", model.OrderTypeSeckillPrefix, date, rand.Intn(1000000))
	}
	orderNo += "00" //留作扩展用
	return orderNo
}

// GenerateOrderToken 生成订单幂等性Token
func (s *OrderService) GenerateOrderToken(ctx context.Context, req *orderv1.GenerateOrderTokenRequest) (*orderv1.GenerateOrderTokenResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &orderv1.GenerateOrderTokenResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}
	// 生成UUID作为Token
	token := uuid.New().String()

	// Token有效期5分钟（300秒）
	cacheKey := fmt.Sprintf("%s:%s:%s", OrderTokenCacheKeyPrefix, userID, token)
	set, err := s.redisClient.SetNX(ctx, cacheKey, "1", time.Duration(OrderTokenCacheExpireSeconds)*time.Second).Result()
	if err != nil {
		log.Printf("❌ [OrderService] GenerateOrderToken: 设置Token失败: %v", err)
		return &orderv1.GenerateOrderTokenResponse{
			Code:    1,
			Message: "系统繁忙，请稍后重试",
		}, nil
	}
	if !set {
		// 极低概率的UUID碰撞，重新生成
		log.Printf("⚠️ [OrderService] GenerateOrderToken: Token已存在，重新生成")
		return s.GenerateOrderToken(ctx, req)
	}

	return &orderv1.GenerateOrderTokenResponse{
		Code:          0,
		Message:       "生成成功",
		Token:         token,
		ExpireSeconds: OrderTokenCacheExpireSeconds,
	}, nil
}
func (s *OrderService) checkAndConsumeToken(ctx context.Context, userID, token string) (bool, error) {
	if userID == "" || token == "" {
		return false, errors.New("参数错误")
	}

	tokenKey := fmt.Sprintf("%s:%s:%s", OrderTokenCacheKeyPrefix, userID, token)

	// 使用Lua脚本保证原子性：检查并删除
	luaScript := `
        local tokenKey = KEYS[1]
        local value = redis.call('GET', tokenKey)
        if value == '1' then
            redis.call('DEL', tokenKey)
            return 1
        else
            return 0
        end
    `

	result, err := s.redisClient.Eval(ctx, luaScript, []string{tokenKey}).Int64()
	if err != nil {
		return false, fmt.Errorf("检查Token失败: %w", err)
	}

	return result == 1, nil
}
