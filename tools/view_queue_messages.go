package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run view_queue_messages.go <queue_name>")
		fmt.Println("Example: go run view_queue_messages.go order.timeout.queue")
		os.Exit(1)
	}

	queueName := os.Args[1]

	// 连接 RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5673/")
	if err != nil {
		log.Fatalf("连接 RabbitMQ 失败: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("创建 Channel 失败: %v", err)
	}
	defer ch.Close()

	// 获取队列信息
	q, err := ch.QueueInspect(queueName)
	if err != nil {
		log.Fatalf("查询队列失败: %v", err)
	}

	fmt.Printf("队列名称: %s\n", q.Name)
	fmt.Printf("消息总数: %d\n", q.Messages)
	fmt.Printf("准备消费的消息: %d\n", q.MessagesReady)
	fmt.Printf("未确认的消息: %d\n", q.MessagesUnacked)
	fmt.Printf("消费者数量: %d\n", q.Consumers)
	fmt.Println()

	// 消费消息（不自动确认，查看后拒绝以重新入队）
	msgs, err := ch.Consume(
		queueName,
		"viewer-consumer",
		false, // 手动确认
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("消费消息失败: %v", err)
	}

	fmt.Printf("正在监听队列 %s 的消息...\n", queueName)
	fmt.Println("按 Ctrl+C 退出")
	fmt.Println()

	count := 0
	for msg := range msgs {
		count++
		fmt.Printf("========== 消息 #%d ==========\n", count)
		fmt.Printf("DeliveryTag: %d\n", msg.DeliveryTag)
		fmt.Printf("Exchange: %s\n", msg.Exchange)
		fmt.Printf("RoutingKey: %s\n", msg.RoutingKey)
		fmt.Printf("ContentType: %s\n", msg.ContentType)
		fmt.Printf("Timestamp: %s\n", msg.Timestamp)
		fmt.Printf("Headers: %v\n", msg.Headers)
		fmt.Printf("Body:\n")

		// 尝试格式化 JSON
		var jsonData interface{}
		if err := json.Unmarshal(msg.Body, &jsonData); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
			fmt.Println(string(prettyJSON))
		} else {
			fmt.Println(string(msg.Body))
		}

		fmt.Println()

		// 拒绝消息并重新入队（这样消息不会被删除）
		_ = msg.Nack(false, true)

		// 只查看前 10 条消息
		if count >= 10 {
			fmt.Println("已查看 10 条消息，退出...")
			break
		}
	}
}
