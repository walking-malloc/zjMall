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

	// SendDelayedMessage 发送延迟消息（需要 rabbitmq_delayed_message_exchange 插件）
	// delayMs 延迟毫秒数
	SendDelayedMessage(ctx context.Context, exchange, routingKey string, data interface{}, delayMs int64) error

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
	// 检查 channel 是否有效
	if m.channel == nil {
		return fmt.Errorf("发送消息失败: RabbitMQ channel 为 nil")
	}

	// 检查连接是否关闭（通过检查 channel 的 connection 状态）
	if m.channel.IsClosed() {
		return fmt.Errorf("发送消息失败: RabbitMQ channel/connection 已关闭")
	}

	// 序列化消息体
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 使用 topic 作为队列名（如果提供了 topic），否则使用默认队列名
	queueName := m.queue
	if topic != "" {
		queueName = topic
	}

	err = m.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key (queue 名)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	log.Printf("✅ RabbitMQ 消息发送成功: Queue=%s, Topic=%s", queueName, topic)
	return nil
}

// SendOrderedMessage 发送顺序消息（按 key 分区，保证同一 key 的消息有序）
// RocketMQ 5.x 通过设置 MessageGroup 来实现顺序消息
func (m *messageProducer) SendOrderedMessage(ctx context.Context, topic string, key string, data interface{}) error {
	// 检查 channel 是否有效
	if m.channel == nil {
		return fmt.Errorf("发送顺序消息失败: RabbitMQ channel 为 nil")
	}

	// 检查连接是否关闭
	if m.channel.IsClosed() {
		return fmt.Errorf("发送顺序消息失败: RabbitMQ channel/connection 已关闭")
	}

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

// SendDelayedMessage 发送延迟消息（需要 rabbitmq_delayed_message_exchange 插件）
// exchange: 延迟消息 exchange 名称
// routingKey: 路由键（通常是队列名）
// data: 消息数据
// delayMs: 延迟毫秒数
func (m *messageProducer) SendDelayedMessage(ctx context.Context, exchange, routingKey string, data interface{}, delayMs int64) error {
	// 检查 channel 是否有效
	if m.channel == nil {
		return fmt.Errorf("发送延迟消息失败: RabbitMQ channel 为 nil")
	}

	// 检查连接是否关闭
	if m.channel.IsClosed() {
		return fmt.Errorf("发送延迟消息失败: RabbitMQ channel/connection 已关闭")
	}

	// 序列化消息体
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送延迟消息，需要在 headers 中设置 x-delay（毫秒）
	err = m.channel.PublishWithContext(ctx,
		exchange,   // delayed exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Headers: amqp.Table{
				"x-delay": delayMs, // 延迟毫秒数
			},
		},
	)
	if err != nil {
		return fmt.Errorf("发送延迟消息失败: %w", err)
	}

	log.Printf("✅ RabbitMQ 延迟消息发送成功: Exchange=%s, RoutingKey=%s, Delay=%dms", exchange, routingKey, delayMs)
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
