package model

import "time"

// OrderOutbox 订单服务 Outbox 事件表模型
// 用于实现 Outbox 模式，保证事件可靠投递到 MQ
type OrderOutbox struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EventType   string    `gorm:"column:event_type;size:64;not null" json:"event_type"`
	AggregateID string    `gorm:"column:aggregate_id;size:64;not null" json:"aggregate_id"`
	Payload     string    `gorm:"column:payload;type:json;not null" json:"payload"`
	Status      int8      `gorm:"column:status;not null;default:0" json:"status"` // 0-待发送，1-已发送，2-发送失败
	RetryCount  int       `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	ErrorMsg    string    `gorm:"column:error_msg;size:500" json:"error_msg"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrderOutbox) TableName() string {
	return "order_outbox"
}
