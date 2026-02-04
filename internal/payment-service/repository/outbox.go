package repository

import (
	"context"
	"time"
	"zjMall/internal/payment-service/model"

	"gorm.io/gorm"
)

// OutboxStatus 定义 Outbox 状态
const (
	OutboxStatusPending = 0
	OutboxStatusSent    = 1
	OutboxStatusFailed  = 2
)

// PaymentOutboxRepository Outbox 仓库接口
type PaymentOutboxRepository interface {
	// Create 创建一条 Outbox 记录
	Create(ctx context.Context, event *model.PaymentOutbox) error
	// FetchPending 获取待发送的 Outbox 记录
	FetchPending(ctx context.Context, limit int) ([]*model.PaymentOutbox, error)
	// MarkSent 标记为已发送
	MarkSent(ctx context.Context, id uint64) error
	// MarkFailed 标记为发送失败并增加重试次数
	MarkFailed(ctx context.Context, id uint64, errMsg string) error
}

type paymentOutboxRepository struct {
	db *gorm.DB
}

// NewPaymentOutboxRepository 创建 Outbox 仓库
func NewPaymentOutboxRepository(db *gorm.DB) PaymentOutboxRepository {
	return &paymentOutboxRepository{db: db}
}

func (r *paymentOutboxRepository) Create(ctx context.Context, event *model.PaymentOutbox) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *paymentOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*model.PaymentOutbox, error) {
	var events []*model.PaymentOutbox
	err := r.db.WithContext(ctx).
		Where("status = ?", OutboxStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *paymentOutboxRepository) MarkSent(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.PaymentOutbox{}).
		Where("id = ? ", id).
		Updates(map[string]interface{}{
			"status":     OutboxStatusSent,
			"updated_at": time.Now(),
		}).Error
}

func (r *paymentOutboxRepository) MarkFailed(ctx context.Context, id uint64, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&model.PaymentOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      OutboxStatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"error_msg":   errMsg,
			"updated_at":  time.Now(),
		}).Error
}
