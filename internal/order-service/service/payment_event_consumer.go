package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartPaymentEventConsumer 启动支付成功事件消费者，从 MQ 中消费 PaymentSucceededEvent 并更新订单状态
func StartPaymentEventConsumer(ctx context.Context, svc *OrderService, ch *amqp.Channel, queue string) {
	if ch == nil {
		log.Println("⚠️ [OrderPaymentConsumer] RabbitMQ Channel 为 nil，跳过消费者启动")
		return
	}
	if svc == nil {
		log.Println("⚠️ [OrderPaymentConsumer] OrderService 为 nil，跳过消费者启动")
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
		log.Printf("❌ [OrderPaymentConsumer] 声明队列失败: %v", err)
		return
	}

	// 公平分发，一次只投递一条未确认的消息给当前消费者
	if err := ch.Qos(1, 0, false); err != nil {
		log.Printf("⚠️ [OrderPaymentConsumer] 设置 Qos 失败: %v", err)
	}

	msgs, err := ch.Consume(
		queue,
		"order-service-payment-consumer", // consumer
		false,                            // autoAck
		false,                            // exclusive
		false,                            // noLocal
		false,                            // noWait
		nil,                              // args
	)
	if err != nil {
		log.Printf("❌ [OrderPaymentConsumer] 启动消费者失败: %v", err)
		return
	}

	log.Printf("✅ [OrderPaymentConsumer] 已启动，正在消费支付成功事件队列: %s", queue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("ℹ️ [OrderPaymentConsumer] 上下文已取消，退出消费者循环")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("⚠️ [OrderPaymentConsumer] 消息通道已关闭，退出消费者循环")
					return
				}

				start := time.Now()
				var evt PaymentSucceededEvent
				if err := json.Unmarshal(msg.Body, &evt); err != nil {
					log.Printf("❌ [OrderPaymentConsumer] 解析 PaymentSucceededEvent 失败，丢弃消息: %v, body=%s", err, string(msg.Body))
					_ = msg.Nack(false, false)
					continue
				}

				if err := svc.HandlePaymentSucceededEvent(ctx, &evt); err != nil {
					log.Printf("❌ [OrderPaymentConsumer] 处理支付成功事件失败，将重回队列: %v", err)
					_ = msg.Nack(false, true)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				_ = msg.Ack(false)
				log.Printf("✅ [OrderPaymentConsumer] 支付成功事件处理完成，orderNo=%s，耗时=%s", evt.OrderNo, time.Since(start))
			}
		}
	}()
}
