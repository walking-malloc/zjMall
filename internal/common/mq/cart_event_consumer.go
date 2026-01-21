package mq

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"zjMall/internal/cart-service/model"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StartCartEventConsumer 启动购物车事件消费者，从 MQ 中消费 CartEvent 并同步到 MySQL
func StartCartEventConsumer(ctx context.Context, db *gorm.DB, ch *amqp.Channel, queue string) {
	if ch == nil {
		log.Println("⚠️ [CartConsumer] RabbitMQ Channel 为 nil，跳过消费者启动")
		return
	}

	if db == nil {
		log.Println("⚠️ [CartConsumer] DB 为 nil，跳过消费者启动")
		return
	}

	// 确保队列存在（与生产端队列名保持一致）
	_, err := ch.QueueDeclare(
		queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		log.Printf("❌ [CartConsumer] 声明队列失败: %v", err)
		return
	}

	// 公平分发，一次只投递一条未确认的消息给当前消费者
	if err := ch.Qos(1, 0, false); err != nil {
		log.Printf("⚠️ [CartConsumer] 设置 Qos 失败: %v", err)
	}

	msgs, err := ch.Consume(
		queue,
		"cart-service-consumer", // consumer
		false,                   // autoAck
		false,                   // exclusive
		false,                   // noLocal
		false,                   // noWait
		nil,                     // args
	)
	if err != nil {
		log.Printf("❌ [CartConsumer] 启动消费者失败: %v", err)
		return
	}

	log.Printf("✅ [CartConsumer] 已启动，正在消费队列: %s", queue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("ℹ️ [CartConsumer] 上下文已取消，退出消费者循环")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("⚠️ [CartConsumer] 消息通道已关闭，退出消费者循环")
					return
				}

				start := time.Now()
				if err := handleCartEventMessage(ctx, db, &msg); err != nil {
					log.Printf("❌ [CartConsumer] 处理消息失败，将重回队列: %v", err)
					// 处理失败，短暂休眠后重回队列，避免空转重试过快
					_ = msg.Nack(false, true)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				_ = msg.Ack(false)
				log.Printf("✅ [CartConsumer] 消息处理完成，耗时=%s", time.Since(start))
			}
		}
	}()
}

// handleCartEventMessage 解析并处理单条 MQ 消息
func handleCartEventMessage(ctx context.Context, db *gorm.DB, msg *amqp.Delivery) error {
	var event CartEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("❌ [CartConsumer] 解析 CartEvent 失败，丢弃消息: %v, body=%s", err, string(msg.Body))
		// 反序列化失败通常是不可恢复错误，不重回队列
		_ = msg.Nack(false, false)
		return nil
	}

	log.Printf("🔍 [CartConsumer] 收到购物车事件: type=%s, user_id=%s, item_id=%s", event.EventType, event.UserID, event.ItemID)

	return syncCartEventToMySQL(ctx, db, &event)
}

// syncCartEventToMySQL 根据 CartEvent 同步到 MySQL
func syncCartEventToMySQL(ctx context.Context, db *gorm.DB, e *CartEvent) error {
	switch e.EventType {
	case CartEventItemAdded:
		return handleItemAdded(ctx, db, e)
	case CartEventItemUpdated:
		return handleItemUpdated(ctx, db, e)
	case CartEventItemRemoved:
		return handleItemRemoved(ctx, db, e)
	case CartEventCleared:
		return handleCartCleared(ctx, db, e)
	default:
		log.Printf("⚠️ [CartConsumer] 未知的事件类型: %s", e.EventType)
		return nil
	}
}

func handleItemAdded(ctx context.Context, db *gorm.DB, e *CartEvent) error {
	// 使用 CartItem 模型，将事件数据落库
	item := &model.CartItem{
		// ID 来自事件 data 中的 id 字段
	}

	item.ID = e.ItemID
	item.UserID = e.UserID

	if v, ok := e.Data["product_id"].(string); ok {
		item.ProductID = v
	}
	if v, ok := e.Data["sku_id"].(string); ok {
		item.SKUID = v
	}
	if v, ok := e.Data["product_title"].(string); ok {
		item.ProductTitle = v
	}
	if v, ok := e.Data["product_image"].(string); ok {
		item.ProductImage = v
	}
	if v, ok := e.Data["sku_name"].(string); ok {
		item.SKUName = v
	}
	if v, ok := e.Data["price"].(float64); ok {
		item.Price = v
	}
	if v, ok := e.Data["current_price"].(float64); ok {
		item.CurrentPrice = v
	}
	if v, ok := e.Data["quantity"].(float64); ok {
		item.Quantity = int32(v)
	}
	if v, ok := e.Data["stock"].(float64); ok {
		item.Stock = int32(v)
	}
	if v, ok := e.Data["is_valid"].(bool); ok {
		item.IsValid = v
	}
	if v, ok := e.Data["invalid_reason"].(string); ok {
		item.InvalidReason = v
	}

	// 使用 OnConflict 保证幂等性：如果已存在则更新
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"user_id",
				"product_id",
				"sku_id",
				"product_title",
				"product_image",
				"sku_name",
				"price",
				"current_price",
				"quantity",
				"stock",
				"is_valid",
				"invalid_reason",
				"updated_at",
			}),
		}).
		Create(item).Error
}

func handleItemUpdated(ctx context.Context, db *gorm.DB, e *CartEvent) error {
	// 数量在 Data 中，以 JSON number 形式存在，反序列化后是 float64
	var quantity *int32
	if v, ok := e.Data["quantity"].(float64); ok {
		q := int32(v)
		quantity = &q
	}

	if quantity == nil {
		log.Printf("⚠️ [CartConsumer] 更新事件缺少 quantity 字段，忽略: user_id=%s, item_id=%s", e.UserID, e.ItemID)
		return nil
	}

	return db.WithContext(ctx).
		Model(&model.CartItem{}).
		Where("id = ? AND user_id = ?", e.ItemID, e.UserID).
		Update("quantity", *quantity).Error
}

func handleItemRemoved(ctx context.Context, db *gorm.DB, e *CartEvent) error {
	return db.WithContext(ctx).
		Where("id = ? AND user_id = ?", e.ItemID, e.UserID).
		Delete(&model.CartItem{}).Error
}

func handleCartCleared(ctx context.Context, db *gorm.DB, e *CartEvent) error {
	return db.WithContext(ctx).
		Where("user_id = ?", e.UserID).
		Delete(&model.CartItem{}).Error
}
