package handler

import (
	"context"
	"fmt"
	"log"

	paymentv1 "zjMall/gen/go/api/proto/payment"
	"zjMall/internal/payment-service/model"
	"zjMall/internal/payment-service/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentHandler 支付服务 Handler
type PaymentHandler struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

// NewPaymentHandler 创建支付 Handler
func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		svc: svc,
	}
}

// CreatePayment 创建支付单
func (h *PaymentHandler) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	// 1. 参数校验（使用 validator）
	validator := service.NewCreatePaymentRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &paymentv1.CreatePaymentResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.svc.CreatePayment(ctx, req)
}

// GetPayment 查询支付单
func (h *PaymentHandler) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	// 1. 参数校验（使用 validator）
	if req.PaymentNo == "" {
		return &paymentv1.GetPaymentResponse{
			Code:    1,
			Message: "支付单号不能为空",
		}, nil
	}

	// 2. 调用服务层查询支付单
	payment, err := h.svc.GetPayment(ctx, req.PaymentNo)
	if err != nil {
		log.Printf("❌ [PaymentHandler] GetPayment: 查询支付单失败 payment_no=%s, err=%v", req.PaymentNo, err)
		return &paymentv1.GetPaymentResponse{
			Code:    1,
			Message: fmt.Sprintf("查询支付单失败: %v", err),
		}, nil
	}

	if payment == nil {
		return &paymentv1.GetPaymentResponse{
			Code:    1,
			Message: "支付单不存在",
		}, nil
	}

	// 3. 转换为 proto 响应
	return &paymentv1.GetPaymentResponse{
		Code:    0,
		Message: "success",
		Payment: h.convertPaymentToProto(payment),
	}, nil
}

// PaymentCallback 支付回调（第三方支付平台回调）
func (h *PaymentHandler) PaymentCallback(ctx context.Context, req *paymentv1.PaymentCallbackRequest) (*paymentv1.PaymentCallbackResponse, error) {
	// 1. 参数校验（使用 validator）
	validator := service.NewPaymentCallbackRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &paymentv1.PaymentCallbackResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	// 3. 转换 extra_params
	extraParams := make(map[string]string)
	if req.ExtraParams != nil {
		extraParams = req.ExtraParams
	}

	// 4. 调用服务层处理回调
	serviceReq := &service.PaymentCallbackRequest{
		PayChannel:  payChannel,
		PaymentNo:   req.PaymentNo,
		TradeNo:     req.TradeNo,
		Amount:      req.Amount,
		Status:      req.Status,
		Sign:        req.Sign,
		ExtraParams: extraParams,
	}

	if err := h.svc.HandlePaymentCallback(ctx, serviceReq); err != nil {
		log.Printf("❌ [PaymentHandler] PaymentCallback: 处理回调失败 payment_no=%s, err=%v", req.PaymentNo, err)
		return &paymentv1.PaymentCallbackResponse{
			Code:    1,
			Message: fmt.Sprintf("处理回调失败: %v", err),
		}, nil
	}

	// 5. 返回成功（第三方平台会重试直到收到成功响应）
	return &paymentv1.PaymentCallbackResponse{
		Code:    0,
		Message: "success",
	}, nil
}

// QueryPaymentStatus 查询支付状态
func (h *PaymentHandler) QueryPaymentStatus(ctx context.Context, req *paymentv1.QueryPaymentStatusRequest) (*paymentv1.QueryPaymentStatusResponse, error) {
	// 1. 参数校验（使用 validator）
	if req.PaymentNo == "" {
		return &paymentv1.QueryPaymentStatusResponse{
			Code:    1,
			Message: "支付单号不能为空",
		}, nil
	}

	// 2. 调用服务层查询支付状态
	status, tradeNo, err := h.svc.QueryPaymentStatus(ctx, req.PaymentNo)
	if err != nil {
		log.Printf("❌ [PaymentHandler] QueryPaymentStatus: 查询支付状态失败 payment_no=%s, err=%v", req.PaymentNo, err)
		return &paymentv1.QueryPaymentStatusResponse{
			Code:    1,
			Message: fmt.Sprintf("查询支付状态失败: %v", err),
		}, nil
	}

	// 3. 转换为 proto 响应
	return &paymentv1.QueryPaymentStatusResponse{
		Code:    0,
		Message: "success",
		Status:  h.convertStatusToProto(status),
		TradeNo: tradeNo,
	}, nil
}

// convertPaymentToProto 转换 Payment 模型为 proto 消息
func (h *PaymentHandler) convertPaymentToProto(payment *model.Payment) *paymentv1.Payment {
	if payment == nil {
		return nil
	}

	protoPayment := &paymentv1.Payment{
		Id:         payment.ID,
		PaymentNo:  payment.PaymentNo,
		OrderNo:    payment.OrderNo,
		UserId:     payment.UserID,
		Amount:     fmt.Sprintf("%.2f", payment.Amount),
		PayChannel: payment.PayChannel,
		Status:     h.convertStatusToProto(payment.Status),
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
func (h *PaymentHandler) convertStatusToProto(status int8) paymentv1.PaymentStatus {
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

// convertStatusFromProto 转换 proto PaymentStatus 为 int8
func (h *PaymentHandler) convertStatusFromProto(status paymentv1.PaymentStatus) int8 {
	switch status {
	case paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING:
		return model.PaymentStatusPending
	case paymentv1.PaymentStatus_PAYMENT_STATUS_PROCESSING:
		return model.PaymentStatusProcessing
	case paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS:
		return model.PaymentStatusSuccess
	case paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED:
		return model.PaymentStatusFailed
	case paymentv1.PaymentStatus_PAYMENT_STATUS_CLOSED:
		return model.PaymentStatusClosed
	case paymentv1.PaymentStatus_PAYMENT_STATUS_REFUNDED:
		return model.PaymentStatusRefunded
	default:
		return 0
	}
}

// GeneratePaymentToken 生成支付幂等性Token
func (h *PaymentHandler) GeneratePaymentToken(ctx context.Context, req *paymentv1.GeneratePaymentTokenRequest) (*paymentv1.GeneratePaymentTokenResponse, error) {
	if req.OrderNo == "" {
		return &paymentv1.GeneratePaymentTokenResponse{
			Code:    1,
			Message: "订单号不能为空",
		}, nil
	}
	return h.svc.GeneratePaymentToken(ctx, req)
}
