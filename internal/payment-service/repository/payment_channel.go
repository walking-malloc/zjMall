package repository

import (
	"context"
	"errors"
	"fmt"
	"zjMall/internal/payment-service/model"

	"gorm.io/gorm"
)

type PaymentChannelRepository interface {
	GetPaymentChannelByChannelCode(ctx context.Context, channelCode string) (*model.PaymentChannel, error)
	CreatePaymentChannel(ctx context.Context, paymentChannel *model.PaymentChannel) error
	UpdatePaymentChannel(ctx context.Context, paymentChannel *model.PaymentChannel) error
}

type paymentChannelRepository struct {
	db *gorm.DB
}

func NewPaymentChannelRepository(db *gorm.DB) PaymentChannelRepository {
	return &paymentChannelRepository{db: db}
}

func (r *paymentChannelRepository) GetPaymentChannelByChannelCode(ctx context.Context, channelCode string) (*model.PaymentChannel, error) {
	var paymentChannel model.PaymentChannel
	if err := r.db.WithContext(ctx).Where("channel_code = ?", channelCode).First(&paymentChannel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("支付渠道不存在: %s", channelCode)
		}
		return nil, err
	}
	return &paymentChannel, nil
}

func (r *paymentChannelRepository) CreatePaymentChannel(ctx context.Context, paymentChannel *model.PaymentChannel) error {
	return r.db.WithContext(ctx).Create(paymentChannel).Error
}

func (r *paymentChannelRepository) UpdatePaymentChannel(ctx context.Context, paymentChannel *model.PaymentChannel) error {
	return r.db.WithContext(ctx).Model(&model.PaymentChannel{}).Where("id = ?", paymentChannel.ID).Updates(paymentChannel).Error
}
