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


