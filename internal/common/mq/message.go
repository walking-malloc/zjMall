package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageProducer 消息生产者接口
type MessageProducer interface {
	// SendMessage 发送普通消息
	SendMessage(ctx context.Context, topic string, data interface{}) error

	// SendOrderedMessage 发送顺序消息（按 key 分区，保证同一 key 的消息有序）
	SendOrderedMessage(ctx context.Context, topic string, key string, data interface{}) error

	// SendTransactionMessage 发送事务消息（暂不支持，返回错误）
	SendTransactionMessage(ctx context.Context, topic string, data interface{},
		localTransactionFunc interface{}) error
}

type messageProducer struct {
	channel *amqp.Channel
	queue   string
}

// NewMessageProducer 创建消息生产者（RabbitMQ）
// 这里的参数是 RabbitMQ 的 Channel 和队列名
func NewMessageProducer(ch *amqp.Channel, queue string) MessageProducer {
	return &messageProducer{
		channel: ch,
		queue:   queue,
	}
}

// SendMessage 发送普通消息
func (m *messageProducer) SendMessage(ctx context.Context, topic string, data interface{}) error {
	// 序列化消息体
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// RabbitMQ 中我们使用队列名作为目标（这里忽略 topic，或将 topic 写入 header）
	err = m.channel.PublishWithContext(ctx,
		"",      // exchange
		m.queue, // routing key (queue 名)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	log.Printf("✅ RabbitMQ 消息发送成功: Queue=%s, Topic=%s", m.queue, topic)
	return nil
}

// SendOrderedMessage 发送顺序消息（按 key 分区，保证同一 key 的消息有序）
// RocketMQ 5.x 通过设置 MessageGroup 来实现顺序消息
func (m *messageProducer) SendOrderedMessage(ctx context.Context, topic string, key string, data interface{}) error {
	// 序列化消息体
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// RabbitMQ 简单实现：同样发送到队列（如需严格有序可根据 key 使用不同队列或 exchange+routingKey）
	err = m.channel.PublishWithContext(ctx,
		"",
		m.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("发送顺序消息失败: %w", err)
	}

	log.Printf("✅ 顺序消息发送成功（按普通消息发送）: Topic=%s, Key=%s", topic, key)
	return nil
}

// SendTransactionMessage 发送事务消息
// 注意：RocketMQ 5.x 的事务消息 API 与 4.x 不同，需要单独实现
func (m *messageProducer) SendTransactionMessage(ctx context.Context, topic string, data interface{},
	localTransactionFunc interface{}) error {
	// RocketMQ 5.x 的事务消息需要专门的 TransactionProducer
	// 这里暂时返回错误，提示需要使用专门的 TransactionProducer
	_ = topic
	_ = data
	_ = localTransactionFunc
	return fmt.Errorf("事务消息需要使用 TransactionProducer，请参考 RocketMQ 5.x 文档实现")
}
