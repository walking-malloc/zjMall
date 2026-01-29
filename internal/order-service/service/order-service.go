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
	"zjMall/internal/common/middleware"
	"zjMall/internal/order-service/model"
	"zjMall/internal/order-service/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	// 正向流程
	OrderStatusPendingPay = int32(orderv1.OrderStatus_ORDER_STATUS_PENDING_PAY) // 待支付（可取消）
	OrderStatusPaid       = int32(orderv1.OrderStatus_ORDER_STATUS_PAID)        // 已支付（可退款）
	OrderStatusShipped    = int32(orderv1.OrderStatus_ORDER_STATUS_SHIPPED)     // 已发货（可收货/退货）
	OrderStatusCompleted  = int32(orderv1.OrderStatus_ORDER_STATUS_COMPLETED)   // 已完成（不可修改）

	// 逆向流程
	OrderStatusCancelled = int32(orderv1.OrderStatus_ORDER_STATUS_CANCELLED) // 已取消（用户主动）
	OrderStatusRefunding = int32(orderv1.OrderStatus_ORDER_STATUS_REFUNDING) // 退款中
	OrderStatusRefunded  = int32(orderv1.OrderStatus_ORDER_STATUS_REFUNDED)  // 已退款
	OrderStatusClosed    = int32(orderv1.OrderStatus_ORDER_STATUS_CLOSED)    // 已关闭（超时自动）
)

// OrderService 订单服务（业务逻辑层）
type OrderService struct {
	orderRepo       repository.OrderRepository
	productClient   client.ProductClient
	inventoryClient client.InventoryClient
	userClient      client.UserClient
}

func NewOrderService(orderRepo repository.OrderRepository, productClient client.ProductClient, inventoryClient client.InventoryClient, userClient client.UserClient) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		productClient:   productClient,
		inventoryClient: inventoryClient,
		userClient:      userClient,
	}
}

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
	// 获取用户地址
	userAddress, err := s.userClient.GetUserAddress(ctx, req.AddressId)
	if err != nil || userAddress == nil {
		return &orderv1.CreateOrderResponse{
			Code:    1,
			Message: fmt.Sprintf("获取用户地址失败: %v", err),
		}, nil
	}

	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		Status:          OrderStatusPendingPay,
		TotalAmount:     totalAmount,
		DiscountAmount:  discountAmount,
		ShippingAmount:  shippingAmount,
		PayAmount:       payAmount,
		BuyerRemark:     req.BuyerRemark,
		ReceiverName:    userAddress.ReceiverName,
		ReceiverPhone:   userAddress.ReceiverPhone,
		ReceiverAddress: fmt.Sprintf("%s%s%s%s", userAddress.Province, userAddress.City, userAddress.District, userAddress.Detail),
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

	return &orderv1.CreateOrderResponse{
		Code:      0,
		Message:   "创建成功",
		OrderNo:   orderNo,
		PayAmount: fmt.Sprintf("%.2f", payAmount),
	}, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &orderv1.GetOrderResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	order, items, err := s.orderRepo.GetOrderByNo(ctx, userID, req.OrderNo)
	if err != nil {
		log.Printf("❌ [OrderService] GetOrder: 查询订单失败: %v", err)
		return &orderv1.GetOrderResponse{
			Code:    1,
			Message: "订单不存在",
		}, nil
	}

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

	orders, total, err := s.orderRepo.ListUserOrders(ctx, userID, int32(req.Status), offset, limit)
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

	// 先查询订单，校验归属和状态
	order, _, err := s.orderRepo.GetOrderByNo(ctx, userID, req.OrderNo)
	if err != nil {
		log.Printf("❌ [OrderService] CancelOrder: 查询订单失败: %v", err)
		return &orderv1.CancelOrderResponse{
			Code:    1,
			Message: "订单不存在",
		}, nil
	}

	// 校验订单状态：只有待支付订单可以取消
	if order.Status != OrderStatusPendingPay {
		return &orderv1.CancelOrderResponse{
			Code:    1,
			Message: fmt.Sprintf("订单状态为 %d，无法取消", order.Status),
		}, nil
	}

	// 更新订单状态
	if err := s.orderRepo.UpdateOrderStatus(ctx, req.OrderNo,
		OrderStatusPendingPay,
		OrderStatusCancelled); err != nil {
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
	// 先查询订单，校验状态
	order, _, err := s.orderRepo.GetOrderByNoNoUser(ctx, req.OrderNo)
	if err != nil {
		log.Printf("❌ [OrderService] MarkOrderPaid: 查询订单失败: %v", err)
		return &orderv1.MarkOrderPaidResponse{
			Code:    1,
			Message: "订单不存在",
		}, nil
	}

	// 校验订单状态：只有待支付订单可以标记为已支付
	if order.Status != OrderStatusPendingPay {
		return &orderv1.MarkOrderPaidResponse{
			Code:    1,
			Message: fmt.Sprintf("订单状态为 %d，无法标记为已支付", order.Status),
		}, nil
	}

	// 更新订单状态和支付信息
	now := time.Now()
	if err := s.orderRepo.UpdateOrderPaid(ctx, req.OrderNo, OrderStatusPendingPay, OrderStatusPaid, req.PayChannel, req.PayTradeNo, now); err != nil {
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
	return &orderv1.Order{
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
		PaidAt:          timestamppb.New(o.PaidAt),
		ShippedAt:       timestamppb.New(o.ShippedAt),
		CompletedAt:     timestamppb.New(o.CompletedAt),
	}
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
