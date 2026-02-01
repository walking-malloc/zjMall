package repository

import (
	"context"
	"zjMall/internal/payment-service/model"

	"gorm.io/gorm"
)

type PaymentLogRepository interface {
	// CreatePaymentLog 创建支付日志
	CreatePaymentLog(ctx context.Context, paymentLog *model.PaymentLog) error
}

type paymentLogRepository struct {
	db *gorm.DB
}

func NewPaymentLogRepository(db *gorm.DB) PaymentLogRepository {
	return &paymentLogRepository{db: db}
}

func (r *paymentLogRepository) CreatePaymentLog(ctx context.Context, paymentLog *model.PaymentLog) error {
	return r.db.WithContext(ctx).Create(paymentLog).Error
}
