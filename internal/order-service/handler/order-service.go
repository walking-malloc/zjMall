package handler

import (
	"context"
	orderv1 "zjMall/gen/go/api/proto/order"
	"zjMall/internal/order-service/service"
)

type OrderServiceHandler struct {
	orderv1.UnimplementedOrderServiceServer
	orderService *service.OrderService
}

func NewOrderServiceHandler(orderService *service.OrderService) *OrderServiceHandler {
	return &OrderServiceHandler{
		orderService: orderService,
	}
}

// 创建订单
func (h *OrderServiceHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	// 简单参数校验
	if len(req.Items) == 0 {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "订单商品不能为空",
		}, nil
	}
	if req.AddressId == "" {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: "收货地址不能为空",
		}, nil
	}
	return h.orderService.CreateOrder(ctx, req)
}

// 获取订单详情
func (h *OrderServiceHandler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	if req.OrderNo == "" {
		return &orderv1.GetOrderResponse{
			Code:    1,
			Message: "订单号不能为空",
		}, nil
	}
	return h.orderService.GetOrder(ctx, req)
}

// 获取用户订单
func (h *OrderServiceHandler) ListUserOrders(ctx context.Context, req *orderv1.ListUserOrdersRequest) (*orderv1.ListUserOrdersResponse, error) {
	return h.orderService.ListUserOrders(ctx, req)
}

// 取消订单
func (h *OrderServiceHandler) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	if req.OrderNo == "" {
		return &orderv1.CancelOrderResponse{
			Code:    1,
			Message: "订单号不能为空",
		}, nil
	}
	return h.orderService.CancelOrder(ctx, req)
}

// 标记订单已支付
func (h *OrderServiceHandler) MarkOrderPaid(ctx context.Context, req *orderv1.MarkOrderPaidRequest) (*orderv1.MarkOrderPaidResponse, error) {
	if req.OrderNo == "" {
		return &orderv1.MarkOrderPaidResponse{
			Code:    1,
			Message: "订单号不能为空",
		}, nil
	}
	if req.PayChannel == "" {
		return &orderv1.MarkOrderPaidResponse{
			Code:    1,
			Message: "支付渠道不能为空",
		}, nil
	}
	if req.PayTradeNo == "" {
		return &orderv1.MarkOrderPaidResponse{
			Code:    1,
			Message: "支付流水号不能为空",
		}, nil
	}
	return h.orderService.MarkOrderPaid(ctx, req)
}
