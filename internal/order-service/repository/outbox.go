package repository

import (
	"context"
	"time"
	"zjMall/internal/order-service/model"

	"gorm.io/gorm"
)

// OutboxStatus 定义 Outbox 状态
const (
	OutboxStatusPending = 0
	OutboxStatusSent    = 1
	OutboxStatusFailed  = 2
)

// OrderOutboxRepository Outbox 仓库接口
type OrderOutboxRepository interface {
	// Create 创建一条 Outbox 记录
	Create(ctx context.Context, event *model.OrderOutbox) error
	// FetchPending 获取待发送的 Outbox 记录
	FetchPending(ctx context.Context, limit int) ([]*model.OrderOutbox, error)
	// MarkSent 标记为已发送
	MarkSent(ctx context.Context, id uint64) error
	// MarkFailed 标记为发送失败并增加重试次数
	MarkFailed(ctx context.Context, id uint64, errMsg string) error
}

type orderOutboxRepository struct {
	db *gorm.DB
}

// NewOrderOutboxRepository 创建 Outbox 仓库
func NewOrderOutboxRepository(db *gorm.DB) OrderOutboxRepository {
	return &orderOutboxRepository{db: db}
}

func (r *orderOutboxRepository) Create(ctx context.Context, event *model.OrderOutbox) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *orderOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*model.OrderOutbox, error) {
	var events []*model.OrderOutbox
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

func (r *orderOutboxRepository) MarkSent(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.OrderOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     OutboxStatusSent,
			"updated_at": time.Now(),
		}).Error
}

func (r *orderOutboxRepository) MarkFailed(ctx context.Context, id uint64, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&model.OrderOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      OutboxStatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"error_msg":   errMsg,
			"updated_at":  time.Now(),
		}).Error
}
