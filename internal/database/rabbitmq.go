package database

import (
	"fmt"
	"log"

	"zjMall/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConnection 全局连接
var RabbitMQConnection *amqp.Connection

// RabbitMQChannel 全局通道
var RabbitMQChannel *amqp.Channel

// InitRabbitMQ 初始化 RabbitMQ 连接和通道
func InitRabbitMQ(cfg *config.RabbitMQConfig) (*amqp.Channel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("RabbitMQ 配置不能为空")
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("创建 RabbitMQ Channel 失败: %w", err)
	}

	// 声明一个持久化队列（如果已存在则复用）
	_, err = ch.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("声明 RabbitMQ 队列失败: %w", err)
	}

	RabbitMQConnection = conn
	RabbitMQChannel = ch

	log.Printf("✅ RabbitMQ 连接成功: %s, 队列=%s", url, cfg.Queue)
	return ch, nil
}

// EnablePublisherConfirm 开启 Publisher Confirm 模式，返回确认通道
// 调用者需在 Publish 后从返回的 channel 读取确认，确认顺序与发布顺序一致
func EnablePublisherConfirm(ch *amqp.Channel) (<-chan amqp.Confirmation, error) {
	if ch == nil {
		return nil, fmt.Errorf("RabbitMQ Channel 不能为空")
	}
	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("开启 Publisher Confirm 失败: %w", err)
	}
	// 使用缓冲避免阻塞 amqp 库
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 100))
	log.Println("✅ Publisher Confirm 已开启")
	return confirms, nil
}

// CloseRabbitMQ 关闭 RabbitMQ 连接
func CloseRabbitMQ() error {
	if RabbitMQChannel != nil {
		_ = RabbitMQChannel.Close()
	}
	if RabbitMQConnection != nil {
		return RabbitMQConnection.Close()
	}
	return nil
}

// InitDelayedExchange 初始化延迟消息 Exchange（需要 rabbitmq_delayed_message_exchange 插件）
// exchangeName: 延迟消息 exchange 名称
// queueName: 绑定的队列名称
func InitDelayedExchange(ch *amqp.Channel, exchangeName, queueName string) error {
	if ch == nil {
		return fmt.Errorf("RabbitMQ Channel 不能为空")
	}

	// 声明延迟消息 exchange（类型为 x-delayed-message）
	err := ch.ExchangeDeclare(
		exchangeName,        // name
		"x-delayed-message", // type（延迟消息插件类型）
		true,                // durable
		false,               // autoDelete
		false,               // internal
		false,               // noWait
		amqp.Table{
			"x-delayed-type": "direct", // 延迟消息的底层类型
		},
	)
	if err != nil {
		return fmt.Errorf("声明延迟消息 Exchange 失败: %w", err)
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %w", err)
	}

	// 绑定队列到 exchange
	err = ch.QueueBind(
		queueName,    // queue name
		queueName,    // routing key
		exchangeName, // exchange
		false,        // noWait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("绑定队列到 Exchange 失败: %w", err)
	}

	log.Printf("✅ 延迟消息 Exchange 初始化成功: Exchange=%s, Queue=%s", exchangeName, queueName)
	return nil
}
