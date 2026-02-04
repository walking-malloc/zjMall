package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// DispatchOutboxEvents 从 Outbox 表中拉取待发送事件并发送到 MQ
// 建议在支付服务启动时以后台 goroutine 的方式周期性调用
func (s *PaymentService) DispatchOutboxEvents(ctx context.Context, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	events, err := s.outboxRepo.FetchPending(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("获取待发送 Outbox 事件失败: %w", err)
	}
	if len(events) == 0 {
		return nil
	}

	for _, evt := range events {
		// 根据事件类型选择路由键/队列，这里先简单写死
		// 当前只支持 payment.succeeded 事件，其他类型打日志并标记失败
		const routingKey = "payment.success.notify"
		if evt.EventType != "payment.succeeded" {
			log.Printf("⚠️ 未知的 Outbox 事件类型: %s, id=%d", evt.EventType, evt.ID)
			_ = s.outboxRepo.MarkFailed(ctx, evt.ID, "unknown event type")
			continue
		}

		// payload 已经是 JSON 字符串，但 MQ 封装会再做一次 json.Marshal，
		// 为了避免双重 JSON，这里先反序列化成 map 再传给 MQ
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(evt.Payload), &payload); err != nil {
			log.Printf("⚠️ 解析 Outbox payload 失败: id=%d, err=%v", evt.ID, err)
			_ = s.outboxRepo.MarkFailed(ctx, evt.ID, fmt.Sprintf("unmarshal payload: %v", err))
			continue
		}

		sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = s.paymentMQ.SendMessage(sendCtx, routingKey, payload)
		cancel()

		if err != nil {
			log.Printf("⚠️ 发送 Outbox 事件到 MQ 失败: id=%d, err=%v", evt.ID, err)
			_ = s.outboxRepo.MarkFailed(ctx, evt.ID, fmt.Sprintf("send to mq: %v", err))
			continue
		}

		if err := s.outboxRepo.MarkSent(ctx, evt.ID); err != nil {
			log.Printf("⚠️ 标记 Outbox 事件已发送失败: id=%d, err=%v", evt.ID, err)
			// 这里不再回滚 MQ 消息，由消费方通过幂等保证安全
			continue
		}
	}

	return nil
}
