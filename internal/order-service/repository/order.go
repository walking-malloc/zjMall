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
	ListUserOrders(ctx context.Context, userID string, status int8, offset, limit int) ([]*model.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus int8) error
	UpdateOrderPaid(ctx context.Context, orderNo string, fromStatus, toStatus int8, payChannel, payTradeNo string, paidAt time.Time) error
	// GetTimeoutOrders 查询超时的订单（待支付状态，创建时间超过指定时间）
	GetTimeoutOrders(ctx context.Context, status int8, timeoutDuration time.Duration, limit int) ([]*model.Order, error)
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

func (r *orderRepository) ListUserOrders(ctx context.Context, userID string, status int8, offset, limit int) ([]*model.Order, int64, error) {
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

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus int8) error {
	// 先查询订单获取当前version
	var order model.Order
	if err := r.db.WithContext(ctx).
		Where("order_no = ? AND status = ?", orderNo, fromStatus).
		First(&order).Error; err != nil {
		return err
	}

	// 使用乐观锁更新：WHERE条件包含version，更新时version+1
	result := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("order_no = ? AND status = ? AND version = ?", orderNo, fromStatus, order.Version).
		Updates(map[string]interface{}{
			"status":  toStatus,
			"version": gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	// 检查是否更新成功（RowsAffected=0表示version不匹配，可能是并发修改）
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 或者返回自定义错误：订单已被其他请求修改
	}

	return nil
}

// GetTimeoutOrders 查询超时的订单（待支付状态，创建时间超过指定时间）
func (r *orderRepository) GetTimeoutOrders(ctx context.Context, status int8, timeoutDuration time.Duration, limit int) ([]*model.Order, error) {
	var orders []*model.Order
	timeoutTime := time.Now().Add(-timeoutDuration)
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", status, timeoutTime).
		Order("created_at ASC").
		Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateOrderPaid 更新订单支付信息和状态（使用乐观锁）
func (r *orderRepository) UpdateOrderPaid(ctx context.Context, orderNo string, fromStatus, toStatus int8, payChannel, payTradeNo string, paidAt time.Time) error {
	// 先查询订单获取当前version
	var order model.Order
	if err := r.db.WithContext(ctx).
		Where("order_no = ? AND status = ?", orderNo, fromStatus).
		First(&order).Error; err != nil {
		return err
	}

	// 使用乐观锁更新：WHERE条件包含version，更新时version+1
	// 将 paidAt 转换为指针类型（因为模型中使用的是 *time.Time）
	paidAtPtr := &paidAt
	result := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("order_no = ? AND status = ? AND version = ?", orderNo, fromStatus, order.Version).
		Updates(map[string]interface{}{
			"status":       toStatus,
			"pay_channel":  payChannel,
			"pay_trade_no": payTradeNo,
			"paid_at":      paidAtPtr,
			"version":      gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	// 检查是否更新成功（RowsAffected=0表示version不匹配，可能是并发修改）
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 或者返回自定义错误：订单已被其他请求修改
	}

	return nil
}
