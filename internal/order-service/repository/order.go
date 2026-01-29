package repository

import (
	"context"
	"time"
	"zjMall/internal/order-service/model"

	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order, items []*model.OrderItem) error
	GetOrderByNo(ctx context.Context, userID, orderNo string) (*model.Order, []*model.OrderItem, error)
	GetOrderByNoNoUser(ctx context.Context, orderNo string) (*model.Order, []*model.OrderItem, error) // 不校验用户ID，用于支付回调等场景
	ListUserOrders(ctx context.Context, userID string, status int32, offset, limit int) ([]*model.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus int32) error
	UpdateOrderPaid(ctx context.Context, orderNo string, fromStatus, toStatus int32, payChannel, payTradeNo string, paidAt time.Time) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CreateOrder 在事务中创建订单主表和明细
func (r *orderRepository) CreateOrder(ctx context.Context, order *model.Order, items []*model.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) GetOrderByNo(ctx context.Context, userID, orderNo string) (*model.Order, []*model.OrderItem, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).
		Where("order_no = ? AND user_id = ?", orderNo, userID).
		First(&order).Error; err != nil {
		return nil, nil, err
	}

	var items []*model.OrderItem
	if err := r.db.WithContext(ctx).
		Where("order_no = ? AND user_id = ?", orderNo, userID).
		Find(&items).Error; err != nil {
		return nil, nil, err
	}

	return &order, items, nil
}

// GetOrderByNoNoUser 根据订单号查询订单（不校验用户ID）
func (r *orderRepository) GetOrderByNoNoUser(ctx context.Context, orderNo string) (*model.Order, []*model.OrderItem, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).
		Where("order_no = ?", orderNo).
		First(&order).Error; err != nil {
		return nil, nil, err
	}

	var items []*model.OrderItem
	if err := r.db.WithContext(ctx).
		Where("order_no = ?", orderNo).
		Find(&items).Error; err != nil {
		return nil, nil, err
	}

	return &order, items, nil
}

func (r *orderRepository) ListUserOrders(ctx context.Context, userID string, status int32, offset, limit int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	tx := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)
	if status > 0 {
		tx = tx.Where("status = ?", status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus int32) error {
	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, fromStatus).
		Update("status", toStatus).Error
}

// UpdateOrderPaid 更新订单支付信息和状态
func (r *orderRepository) UpdateOrderPaid(ctx context.Context, orderNo string, fromStatus, toStatus int32, payChannel, payTradeNo string, paidAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, fromStatus).
		Updates(map[string]interface{}{
			"status":       toStatus,
			"pay_channel":  payChannel,
			"pay_trade_no": payTradeNo,
			"paid_at":      paidAt,
		}).Error
}
