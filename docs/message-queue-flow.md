# 订单服务消息队列执行流程详解

## 📋 消息队列概览

订单服务涉及以下消息队列：

1. **延迟消息队列** (`order.timeout.queue`) - 订单超时检查
2. **Outbox 事件队列** (`order.outbox`) - Outbox 模式派发
3. **购物车同步队列** (`cart-sync`) - 删除购物车事件
4. **支付成功队列** (`payment.success.notify`) - 支付成功事件
5. **死信队列** (`order.timeout.dlq`) - 订单超时死信

---

## 🔄 完整流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                   订单创建流程                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────┐
        │  CreateOrder() 创建订单               │
        │  - 扣减库存                          │
        │  - 创建订单（事务）                  │
        │  - 写入 Outbox 事件（事务）          │
        └─────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
                ▼                           ▼
    ┌───────────────────────┐   ┌───────────────────────┐
    │ 发送延迟消息           │   │ 写入 Outbox 表         │
    │ (立即发送)             │   │ (事务中写入)          │
    │                       │   │                       │
    │ Exchange:             │   │ order_outbox 表:     │
    │ order.timeout.delayed │   │ - cart.items.remove  │
    │ Queue:                │   │ - Status: Pending    │
    │ order.timeout.queue    │   └───────────────────────┘
    │ Delay: 30分钟          │              │
    └───────────────────────┘              │
                │                          │
                │                          ▼
                │              ┌───────────────────────┐
                │              │ Outbox Dispatcher     │
                │              │ (每5秒执行一次)        │
                │              │                       │
                │              │ 1. 查询 Pending 事件  │
                │              │ 2. 发送到 MQ         │
                │              │ 3. 标记为 Sent       │
                │              └───────────────────────┘
                │                          │
                │                          ▼
                │              ┌───────────────────────┐
                │              │ cart-sync 队列        │
                │              │ (购物车服务消费)       │
                │              └───────────────────────┘
                │
                │ 等待30分钟...
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   订单超时处理流程                                 │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ 延迟消息到期           │
    │ order.timeout.queue    │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ StartOrderTimeoutConsumer│
    │ 消费延迟消息            │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ HandleOrderTimeout()  │
    │ - 查询订单状态         │
    │ - 更新为已关闭         │
    │ - 创建 Outbox 事件    │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ 写入 Outbox 表         │
    │ order.timeout          │
    │ Status: Pending        │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ Outbox Dispatcher     │
    │ (每5秒执行一次)        │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ order.timeout.dlq     │
    │ (死信队列)             │
    └───────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   支付成功处理流程                                 │
└─────────────────────────────────────────────────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ payment.success.notify│
    │ (支付服务发送)         │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ StartPaymentEventConsumer│
    │ 消费支付成功消息        │
    └───────────────────────┘
                │
                ▼
    ┌───────────────────────┐
    │ HandlePaymentSucceededEvent()│
    │ - 更新订单状态为已支付  │
    └───────────────────────┘
```

---

## 📝 详细代码流程

### 1️⃣ 订单创建时发送延迟消息

**代码位置**: `internal/order-service/service/order-service.go:363-378`

```go
// 发送延迟消息，用于订单超时检查（使用 RabbitMQ 延迟消息插件）
if s.delayedProducer != nil {
    timeoutPayload := map[string]interface{}{
        "order_no":   orderNo,
        "user_id":    userID,
        "pay_amount": payAmount,
        "created_at": time.Now().Format(time.RFC3339),
    }
    delayMs := int64(s.orderTimeoutDelay.Milliseconds()) // 1800000ms = 30分钟
    if err := s.delayedProducer.SendDelayedMessage(
        ctx, 
        "order.timeout.delayed",  // Exchange
        "order.timeout.queue",    // Queue
        timeoutPayload, 
        delayMs
    ); err != nil {
        // 发送失败不影响订单创建成功
        log.Printf("⚠️ 发送订单超时延迟消息失败: %v", err)
    }
}
```

**执行时机**: 订单创建成功后立即发送

**消息内容**:
```json
{
  "order_no": "0100000000000001",
  "user_id": "user123",
  "pay_amount": 100.00,
  "created_at": "2024-01-01T10:00:00Z"
}
```

**延迟时间**: 30分钟（1800000毫秒）

---

### 2️⃣ Outbox 模式 - 删除购物车事件

**代码位置**: `internal/order-service/service/order-service.go:303-324`

#### 2.1 创建订单时写入 Outbox

```go
// 生成outbox事件（删除购物车的消息），采用outbox模式保证可靠性
var outboxEvent *model.OrderOutbox
if len(cartItemIDs) > 0 {
    cartRemovePayload := map[string]interface{}{
        "user_id":       userID,
        "cart_item_ids": cartItemIDs,
        "order_no":      orderNo,
    }
    payloadJSON, _ := json.Marshal(cartRemovePayload)
    
    outboxEvent = &model.OrderOutbox{
        EventType:   "cart.items.remove",
        AggregateID: orderNo,
        Payload:     string(payloadJSON),
        Status:      repository.OutboxStatusPending, // 0 - 待发送
    }
}

// 在事务中创建订单和 Outbox 事件
s.orderRepo.CreateOrder(ctx, order, items, outboxEvent)
```

**关键点**: 
- ✅ 订单和 Outbox 事件在**同一事务**中写入
- ✅ 保证数据一致性

#### 2.2 Outbox Dispatcher 派发事件

**代码位置**: `internal/order-service/service/outbox_dispatcher.go:14-100`

**启动位置**: `cmd/order-service/main.go:186-199`

```go
// 每5秒执行一次
go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            // 派发 Outbox 事件
            orderService.DispatchOutboxEvents(ctx, outboxProducer, 100)
        }
    }
}()
```

**派发逻辑**:

```go
func (s *OrderService) DispatchOutboxEvents(ctx, producer, batchSize) {
    // 1. 从数据库查询待发送事件
    events, _ := s.outboxRepo.FetchPending(ctx, batchSize)
    
    for _, evt := range events {
        switch evt.EventType {
        case "cart.items.remove":
            // 2. 解析 payload
            var cartPayload map[string]interface{}
            json.Unmarshal([]byte(evt.Payload), &cartPayload)
            
            // 3. 转换为 CartEvent 格式
            userID := cartPayload["user_id"].(string)
            itemID := cartItemIDs[0].(string)
            cartEvent := mq.NewCartItemRemovedEvent(userID, itemID)
            
            // 4. 发送到购物车同步队列
            producer.SendMessage(ctx, "cart-sync", cartEvent)
            
            // 5. 标记为已发送
            s.outboxRepo.MarkSent(ctx, evt.ID)
            
        case "order.timeout":
            // 发送到死信队列
            producer.SendMessage(ctx, "order.timeout.dlq", timeoutPayload)
            s.outboxRepo.MarkSent(ctx, evt.ID)
        }
    }
}
```

**执行流程**:
1. 每5秒查询一次 `order_outbox` 表中 `status = 0` (Pending) 的事件
2. 根据事件类型发送到不同的队列
3. 发送成功后标记为 `status = 1` (Sent)
4. 发送失败标记为 `status = 2` (Failed)

---

### 3️⃣ 订单超时处理 - 延迟消息消费

**代码位置**: `internal/order-service/service/order_timeout_consumer.go`

**启动位置**: `cmd/order-service/main.go:152-156`

```go
// 启动订单超时消息消费者
go service.StartOrderTimeoutConsumer(ctx, orderService, delayedCh, "order.timeout.queue")
```

#### 3.1 消费者启动

```go
func StartOrderTimeoutConsumer(ctx, orderService, ch, queueName) {
    // 1. 注册消费者
    msgs, _ := ch.Consume(
        queueName, // "order.timeout.queue"
        "",        // consumer tag
        false,     // autoAck = false (手动确认)
        false,     // exclusive
        false,     // noLocal
        false,     // noWait
        nil,
    )
    
    // 2. 循环消费消息
    for {
        select {
        case msg := <-msgs:
            // 3. 处理消息
            if err := handleOrderTimeoutMessage(ctx, orderService, msg); err != nil {
                msg.Nack(false, true) // 失败，重新入队
            } else {
                msg.Ack(false) // 成功，确认消息
            }
        }
    }
}
```

#### 3.2 处理订单超时

```go
func handleOrderTimeoutMessage(ctx, orderService, msg) error {
    // 1. 解析消息
    var payload map[string]interface{}
    json.Unmarshal(msg.Body, &payload)
    orderNo := payload["order_no"].(string)
    
    // 2. 调用处理逻辑
    return orderService.HandleOrderTimeout(ctx, orderNo)
}
```

#### 3.3 订单超时业务逻辑

```go
func (s *OrderService) HandleOrderTimeout(ctx, orderNo) error {
    // 1. 查询订单
    order, _ := s.orderRepo.GetOrderByNoNoUser(ctx, orderNo)
    
    // 2. 检查订单状态（只有待支付才处理）
    if order.Status != OrderStatusPendingPay {
        return nil // 已支付或已取消，跳过
    }
    
    // 3. 更新订单状态为已关闭（乐观锁）
    s.orderRepo.UpdateOrderStatus(ctx, orderNo, 
        OrderStatusPendingPay, 
        OrderStatusClosed)
    
    // 4. 创建 Outbox 事件（发送到死信队列）
    timeoutPayload := map[string]interface{}{
        "order_no":   order.OrderNo,
        "user_id":    order.UserID,
        "pay_amount": order.PayAmount,
        "created_at": order.CreatedAt.Format(time.RFC3339),
        "timeout_at": time.Now().Format(time.RFC3339),
        "reason":     "订单超时未支付",
    }
    
    outboxEvent := &model.OrderOutbox{
        EventType:   "order.timeout",
        AggregateID: orderNo,
        Payload:     string(payloadJSON),
        Status:      repository.OutboxStatusPending,
    }
    
    // 5. 写入 Outbox 表
    s.outboxRepo.Create(ctx, outboxEvent)
    
    return nil
}
```

**执行时机**: 延迟消息到期后（30分钟后）

**后续流程**: Outbox Dispatcher 会将 `order.timeout` 事件发送到死信队列

---

### 4️⃣ 支付成功事件处理

**代码位置**: `internal/order-service/service/payment_event_consumer.go`

**启动位置**: `cmd/order-service/main.go:171`

```go
go service.StartPaymentEventConsumer(ctx, orderService, ch, "payment.success.notify")
```

#### 4.1 消费者启动

```go
func StartPaymentEventConsumer(ctx, svc, ch, queue) {
    // 1. 声明队列
    ch.QueueDeclare(queue, true, false, false, false, nil)
    
    // 2. 设置 Qos（公平分发）
    ch.Qos(1, 0, false)
    
    // 3. 注册消费者
    msgs, _ := ch.Consume(
        queue,
        "order-service-payment-consumer",
        false, // 手动确认
        false, false, false, nil,
    )
    
    // 4. 消费消息
    for {
        select {
        case msg := <-msgs:
            var evt PaymentSucceededEvent
            json.Unmarshal(msg.Body, &evt)
            
            // 5. 处理支付成功事件
            if err := svc.HandlePaymentSucceededEvent(ctx, &evt); err != nil {
                msg.Nack(false, true) // 失败，重新入队
            } else {
                msg.Ack(false) // 成功，确认消息
            }
        }
    }
}
```

#### 4.2 处理支付成功

**代码位置**: `internal/order-service/service/order-service.go:530-567`

```go
func (s *OrderService) HandlePaymentSucceededEvent(ctx, evt *PaymentSucceededEvent) error {
    // 1. 解析支付时间
    paidAt := time.Now()
    if evt.PaidAt != "" {
        paidAt, _ = time.Parse(time.RFC3339, evt.PaidAt)
    }
    
    // 2. 更新订单状态（使用乐观锁，保证幂等）
    err := s.orderRepo.UpdateOrderPaid(
        ctx,
        evt.OrderNo,
        OrderStatusPendingPay,  // 从：待支付
        OrderStatusPaid,         // 到：已支付
        evt.Channel,
        evt.TradeNo,
        paidAt,
    )
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 订单状态已变更（可能已被其他流程处理），幂等返回成功
        return nil
    }
    
    return err
}
```

**消息来源**: 支付服务通过 Outbox 模式发送

**关键点**: 
- ✅ 使用乐观锁保证幂等性
- ✅ 如果订单状态已变更，幂等返回成功

---

## 🔑 关键设计点

### 1. Outbox 模式保证可靠性

```
订单创建（事务）
    ├─> 插入订单表
    ├─> 插入订单明细表
    └─> 插入 Outbox 表 ← 关键！
            │
            └─> Outbox Dispatcher（异步）
                └─> 发送到 MQ
```

**优势**:
- ✅ 订单和事件在同一事务中，保证一致性
- ✅ MQ 不可用时，事件不会丢失
- ✅ Dispatcher 可以重试失败的事件

### 2. 延迟消息替代定时检查

```
传统方式: 定时任务 → 查询数据库 → 处理超时订单
          (每分钟执行，数据库压力大)

延迟消息: 创建订单 → 发送延迟消息 → 30分钟后自动触发
          (精确、无需轮询、利用 MQ 可靠性)
```

### 3. 消息确认机制

所有消费者都使用**手动确认**模式：
- ✅ 处理成功 → `msg.Ack(false)` - 确认消息
- ❌ 处理失败 → `msg.Nack(false, true)` - 拒绝并重新入队

### 4. 幂等性保证

- **订单创建**: Token 一次性消费 + 分布式锁
- **订单超时**: 乐观锁检查订单状态
- **支付成功**: 乐观锁 + 状态检查

---

## 📊 消息队列总结表

| 队列名称 | Exchange | 用途 | 生产者 | 消费者 | 触发时机 |
|---------|----------|------|--------|--------|----------|
| `order.timeout.queue` | `order.timeout.delayed` | 订单超时检查 | 订单创建时 | `StartOrderTimeoutConsumer` | 延迟30分钟 |
| `order.outbox` | (direct) | Outbox 派发 | - | `DispatchOutboxEvents` | 每5秒轮询 |
| `cart-sync` | (direct) | 删除购物车 | Outbox Dispatcher | 购物车服务 | Outbox 派发时 |
| `payment.success.notify` | (direct) | 支付成功通知 | 支付服务 | `StartPaymentEventConsumer` | 支付成功时 |
| `order.timeout.dlq` | (direct) | 订单超时死信 | Outbox Dispatcher | (待实现) | Outbox 派发时 |

---

## 🎯 完整时序图

```
订单创建:
  订单服务 → [事务] → 订单表 + Outbox表
              ↓
        延迟消息 → order.timeout.queue (30分钟后触发)
              ↓
        Outbox Dispatcher (每5秒) → cart-sync队列

30分钟后:
  延迟消息到期 → order.timeout.queue
              ↓
        StartOrderTimeoutConsumer
              ↓
        HandleOrderTimeout → 更新订单状态 + 创建 Outbox 事件
              ↓
        Outbox Dispatcher → order.timeout.dlq

支付成功:
  支付服务 → payment.success.notify
              ↓
        StartPaymentEventConsumer
              ↓
        HandlePaymentSucceededEvent → 更新订单状态为已支付
```
