package repository

import (
	"context"
	"errors"
	"time"
	"zjMall/internal/payment-service/model"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	// CreatePayment 创建支付单
	CreatePayment(ctx context.Context, payment *model.Payment) error
	// GetPaymentByPaymentNo 根据支付单号查询支付单
	GetPaymentByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error)
	// GetPaymentByOrderNo 根据订单号查询支付单
	GetPaymentByOrderNo(ctx context.Context, orderNo string) (*model.Payment, error)
	// UpdatePayment 更新支付单（使用乐观锁）
	UpdatePayment(ctx context.Context, payment *model.Payment) error
	// GetExpiredPayments 查询超时的待支付支付单（用于定时任务）
	GetExpiredPayments(ctx context.Context, limit int) ([]*model.Payment, error)
	// GetPaymentByTradeNo 根据交易号查询支付单
	GetPaymentByTradeNo(ctx context.Context, tradeNo string) (*model.Payment, error)
	// WithTransaction 在事务中执行回调，提供事务内的 PaymentRepository
	WithTransaction(ctx context.Context, fn func(txCtx context.Context, txRepo PaymentRepository) error) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreatePayment(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetPaymentByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetPaymentByOrderNo(ctx context.Context, orderNo string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) UpdatePayment(ctx context.Context, payment *model.Payment) error {
	updateFields := map[string]interface{}{
		"status":  payment.Status,
		"version": gorm.Expr("version + 1"),
	}

	// 只更新非空字段
	if payment.TradeNo != "" {
		updateFields["trade_no"] = payment.TradeNo
	}
	if payment.PaidAt != nil {
		updateFields["paid_at"] = payment.PaidAt
	}
	if payment.ExpiredAt != nil {
		updateFields["expired_at"] = payment.ExpiredAt
	}
	if payment.NotifyURL != "" {
		updateFields["notify_url"] = payment.NotifyURL
	}
	if payment.ReturnURL != "" {
		updateFields["return_url"] = payment.ReturnURL
	}

	return r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("id = ? AND version = ?", payment.ID, payment.Version).
		Updates(updateFields).Error
}

func (r *paymentRepository) GetExpiredPayments(ctx context.Context, limit int) ([]*model.Payment, error) {
	// 直接使用指针切片，让 GORM 直接填充指针，避免值类型到指针类型的转换
	// 这样更高效，也更清晰
	var payments []*model.Payment
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status = ? AND expired_at < ?", model.PaymentStatusPending, now).
		Limit(limit).
		Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) GetPaymentByTradeNo(ctx context.Context, tradeNo string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("trade_no = ?", tradeNo).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// WithTransaction 在事务中执行回调
func (r *paymentRepository) WithTransaction(ctx context.Context, fn func(txCtx context.Context, txRepo PaymentRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &paymentRepository{db: tx}
		return fn(ctx, txRepo)
	})
}
