package mq

import (
	"context"
	"time"
)

// CartEvent 购物车事件
type CartEvent struct {
	EventType string                 `json:"event_type"` // 事件类型：item.added, item.updated, item.removed, cart.cleared
	UserID    string                 `json:"user_id"`
	ItemID    string                 `json:"item_id,omitempty"`    // 购物车项ID（可选）
	Data      map[string]interface{} `json:"data"`                 // 事件数据
	Timestamp time.Time              `json:"timestamp"`
}

// CartEventType 购物车事件类型常量
const (
	CartEventItemAdded   = "cart.item.added"   // 添加商品
	CartEventItemUpdated = "cart.item.updated" // 更新商品数量
	CartEventItemRemoved = "cart.item.removed" // 删除商品
	CartEventCleared     = "cart.cleared"      // 清空购物车
)

// CartTopic 购物车相关 Topic
const (
	CartSyncTopic = "cart-sync" // 购物车同步 Topic
)

// SendCartEvent 发送购物车事件（顺序消息，按 user_id 分区）
func SendCartEvent(ctx context.Context, producer MessageProducer, event *CartEvent) error {
	// 使用顺序消息，保证同一用户的操作有序
	return producer.SendOrderedMessage(ctx, CartSyncTopic, event.UserID, event)
}

// NewCartItemAddedEvent 创建"添加商品"事件
func NewCartItemAddedEvent(userID, itemID string, itemData map[string]interface{}) *CartEvent {
	return &CartEvent{
		EventType: CartEventItemAdded,
		UserID:    userID,
		ItemID:    itemID,
		Data:      itemData,
		Timestamp: time.Now(),
	}
}

// NewCartItemUpdatedEvent 创建"更新商品"事件
func NewCartItemUpdatedEvent(userID, itemID string, itemData map[string]interface{}) *CartEvent {
	return &CartEvent{
		EventType: CartEventItemUpdated,
		UserID:    userID,
		ItemID:    itemID,
		Data:      itemData,
		Timestamp: time.Now(),
	}
}

// NewCartItemRemovedEvent 创建"删除商品"事件
func NewCartItemRemovedEvent(userID, itemID string) *CartEvent {
	return &CartEvent{
		EventType: CartEventItemRemoved,
		UserID:    userID,
		ItemID:    itemID,
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
}

// NewCartClearedEvent 创建"清空购物车"事件
func NewCartClearedEvent(userID string) *CartEvent {
	return &CartEvent{
		EventType: CartEventCleared,
		UserID:    userID,
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
}

