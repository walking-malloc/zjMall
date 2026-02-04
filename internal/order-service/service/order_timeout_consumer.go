package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	inventoryv1 "zjMall/gen/go/api/proto/inventory"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartOrderTimeoutConsumer 启动订单超时消息消费者（处理延迟消息）
func StartOrderTimeoutConsumer(ctx context.Context, orderService *OrderService, ch *amqp.Channel, queueName string) {
	if ch == nil {
		log.Println("⚠️ [OrderTimeoutConsumer] RabbitMQ Channel 为 nil，跳过消费者启动")
		return
	}

	log.Printf("✅ [OrderTimeoutConsumer] 启动订单超时消息消费者，队列=%s", queueName)

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // autoAck（手动确认）
		false,     // exclusive
		false,     // noLocal
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		log.Printf("❌ [OrderTimeoutConsumer] 注册消费者失败: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("ℹ️ [OrderTimeoutConsumer] 订单超时消费者退出")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("⚠️ [OrderTimeoutConsumer] 消息通道已关闭")
				return
			}

			// 处理订单超时消息
			if err := handleOrderTimeoutMessage(ctx, orderService, msg); err != nil {
				log.Printf("❌ [OrderTimeoutConsumer] 处理订单超时消息失败: %v", err)
				// 消息处理失败，拒绝消息并重新入队（可以设置重试次数限制）
				_ = msg.Nack(false, true)
			} else {
				// 处理成功，确认消息
				_ = msg.Ack(false)
			}
		}
	}
}

// handleOrderTimeoutMessage 处理订单超时消息
func handleOrderTimeoutMessage(ctx context.Context, orderService *OrderService, msg amqp.Delivery) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	orderNo, ok := payload["order_no"].(string)
	if !ok || orderNo == "" {
		return fmt.Errorf("消息中缺少 order_no 字段")
	}

	log.Printf("ℹ️ [OrderTimeoutConsumer] 收到订单超时消息: orderNo=%s", orderNo)

	// 处理订单超时
	if err := orderService.HandleOrderTimeout(ctx, orderNo); err != nil {
		return fmt.Errorf("处理订单超时失败: %w", err)
	}

	return nil
}

// HandleOrderTimeout 处理订单超时（回滚库存并关闭订单）
func (s *OrderService) HandleOrderTimeout(ctx context.Context, orderNo string) error {
	// 查询订单（不校验用户ID，因为可能是超时自动处理）
	order, orderItems, err := s.orderRepo.GetOrderByNoNoUser(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 检查订单状态，只有待支付状态才处理超时
	if order.Status != OrderStatusPendingPay {
		log.Printf("ℹ️ [OrderService] HandleOrderTimeout: 订单状态已变更，跳过处理: orderNo=%s, status=%d", orderNo, order.Status)
		return nil
	}

	// 更新订单状态为已关闭（使用乐观锁）
	err = s.orderRepo.UpdateOrderStatus(ctx, orderNo, OrderStatusPendingPay, OrderStatusClosed)
	if err != nil {
		// 如果更新失败（可能是订单已被支付或取消），跳过
		log.Printf("⚠️ [OrderService] HandleOrderTimeout: 更新订单状态失败: orderNo=%s, err=%v", orderNo, err)
		return nil // 不返回错误，避免消息重复处理
	}

	// 回滚库存（订单超时时释放库存，与用户取消订单逻辑一致）
	if len(orderItems) > 0 {
		var rollbackItems []*inventoryv1.SkuQuantity
		for _, item := range orderItems {
			rollbackItems = append(rollbackItems, &inventoryv1.SkuQuantity{
				SkuId:    item.SKUID,
				Quantity: int64(item.Quantity),
			})
		}
		if len(rollbackItems) > 0 {
			if rollbackErr := s.inventoryClient.RollbackStock(ctx, orderNo, rollbackItems); rollbackErr != nil {
				log.Printf("❌ [OrderService] HandleOrderTimeout: 回滚库存失败: orderNo=%s, err=%v", orderNo, rollbackErr)
				// 记录告警，但不影响订单超时处理流程（库存服务应该是幂等的，可以后续补偿）
			} else {
				log.Printf("✅ [OrderService] HandleOrderTimeout: 库存回滚成功: orderNo=%s", orderNo)
			}
		}
	}

	log.Printf("✅ [OrderService] HandleOrderTimeout: 订单超时处理成功: orderNo=%s", orderNo)
	return nil
}
