# 订单支付流程时序图

## 完整流程：从提交订单到支付完成

```mermaid
sequenceDiagram
    autonumber
    participant User as 用户/前端
    participant GW as 网关(Gateway)
    participant OrderSvc as 订单服务
    participant ProductSvc as 商品服务
    participant PromoSvc as 促销服务
    participant InvSvc as 库存服务
    participant PaySvc as 支付服务
    participant ThirdPay as 第三方支付<br/>(微信/支付宝)
    participant MQ as 消息队列(MQ)

    Note over User,MQ: ========== 阶段一：提交订单 ==========
    
    User->>GW: 1. 提交订单请求<br/>(商品列表、地址、优惠券)
    GW->>OrderSvc: 2. CreateOrder(orderNo, items, address)
    
    Note over OrderSvc: 订单服务开始处理
    
    OrderSvc->>ProductSvc: 3. 校验商品状态和价格
    ProductSvc-->>OrderSvc: 3.1 返回商品信息
    
    OrderSvc->>PromoSvc: 4. 计算优惠金额
    PromoSvc-->>OrderSvc: 4.1 返回优惠金额
    
    OrderSvc->>InvSvc: 5. TryDeductStocks(orderNo, items)<br/>预占库存
    Note over InvSvc: 使用乐观锁防止超卖
    InvSvc-->>OrderSvc: 5.1 预占结果(成功/失败)
    
    alt 库存预占失败
        OrderSvc-->>GW: 返回错误：库存不足
        GW-->>User: 提示库存不足
    else 库存预占成功
        OrderSvc->>OrderSvc: 6. 创建订单记录<br/>(状态=待支付)
        OrderSvc->>MQ: 7. 发送「订单创建」事件
        OrderSvc-->>GW: 8. 返回订单号和应付金额
        GW-->>User: 9. 显示订单确认页
    end

    Note over User,MQ: ========== 阶段二：创建支付单 ==========
    
    User->>GW: 10. 选择支付方式，点击「去支付」
    GW->>PaySvc: 11. CreatePayment(orderNo, payChannel)
    
    Note over PaySvc: 支付服务开始处理
    
    PaySvc->>OrderSvc: 12. 查询订单信息(可选)
    OrderSvc-->>PaySvc: 12.1 返回订单详情
    
    PaySvc->>PaySvc: 13. 校验订单状态和金额
    PaySvc->>PaySvc: 14. 生成支付单号<br/>创建支付单记录(状态=待支付)
    PaySvc->>PaySvc: 15. 记录支付日志
    
    PaySvc->>ThirdPay: 16. 调用统一下单接口<br/>(创建支付订单)
    ThirdPay-->>PaySvc: 16.1 返回支付URL/参数
    
    PaySvc-->>GW: 17. 返回支付信息<br/>(pay_url, qr_code, pay_params)
    GW-->>User: 18. 跳转到支付页面

    Note over User,MQ: ========== 阶段三：用户支付 ==========
    
    User->>ThirdPay: 19. 在第三方平台完成支付<br/>(输入密码、确认支付)
    ThirdPay->>ThirdPay: 20. 处理支付请求

    Note over User,MQ: ========== 阶段四：支付回调处理 ==========
    
    ThirdPay->>PaySvc: 21. PaymentCallback回调<br/>(trade_no, amount, status, sign)
    
    Note over PaySvc: 支付服务处理回调
    
    PaySvc->>PaySvc: 22. 签名校验(防止伪造)
    alt 签名校验失败
        PaySvc-->>ThirdPay: 返回错误
        Note over ThirdPay: 第三方会重试
    else 签名校验成功
        PaySvc->>PaySvc: 23. 查询支付单记录
        PaySvc->>PaySvc: 24. 幂等性校验<br/>(检查是否已处理)
        
        alt 已处理过(幂等)
            PaySvc-->>ThirdPay: 返回成功(幂等处理)
        else 未处理过
            PaySvc->>PaySvc: 25. 金额校验<br/>(回调金额 vs 订单金额)
            
            alt 金额不一致
                PaySvc->>PaySvc: 记录告警日志
                PaySvc-->>ThirdPay: 返回错误(需人工对账)
            else 金额一致
                PaySvc->>PaySvc: 26. 更新支付单状态<br/>(使用乐观锁)<br/>status=支付成功<br/>记录trade_no和paid_at
                PaySvc->>PaySvc: 27. 记录支付日志
                
                PaySvc->>OrderSvc: 28. MarkOrderPaid(orderNo, payChannel, tradeNo)<br/>通知订单服务支付成功
                PaySvc->>MQ: 29. 发送「支付成功」事件
                PaySvc-->>ThirdPay: 30. 返回"success"
            end
        end
    end

    Note over User,MQ: ========== 阶段五：订单服务处理支付成功 ==========
    
    OrderSvc->>OrderSvc: 31. 查询订单记录
    OrderSvc->>OrderSvc: 32. 校验订单状态(幂等性)
    
    alt 订单状态不是待支付
        Note over OrderSvc: 已处理过，幂等跳过
    else 订单状态是待支付
        OrderSvc->>OrderSvc: 33. 更新订单状态<br/>(使用乐观锁)<br/>status=待发货<br/>记录支付信息
        OrderSvc->>InvSvc: 34. 确认库存扣减(可选)<br/>如果设计为预占+支付扣减
        InvSvc-->>OrderSvc: 34.1 扣减结果(幂等)
        OrderSvc->>MQ: 35. 发送「订单支付成功」事件
    end

    Note over User,MQ: ========== 阶段六：支付完成 ==========
    
    User->>GW: 36. 查询支付状态(轮询)
    GW->>PaySvc: 37. QueryPaymentStatus(paymentNo)
    PaySvc-->>GW: 38. 返回支付状态(支付成功)
    GW-->>User: 39. 显示「支付成功」页面
    
    Note over MQ: 其他服务订阅事件
    MQ->>MQ: 履约服务：触发出库流程
    MQ->>MQ: 营销服务：发放积分/优惠券
    MQ->>MQ: 风控服务：风控分析
```

## 支付超时处理流程

```mermaid
sequenceDiagram
    autonumber
    participant Timer as 定时任务
    participant PaySvc as 支付服务
    participant OrderSvc as 订单服务
    participant InvSvc as 库存服务
    participant MQ as 消息队列

    Note over Timer,MQ: ========== 支付超时自动关闭 ==========
    
    Timer->>PaySvc: 1. 扫描待支付支付单<br/>(status=待支付, expired_at < 当前时间)
    
    PaySvc->>PaySvc: 2. 查询超时支付单列表
    
    loop 每个超时支付单
        PaySvc->>PaySvc: 3. 更新支付单状态<br/>status=已关闭
        PaySvc->>PaySvc: 4. 记录支付日志
        
        PaySvc->>OrderSvc: 5. 通知订单服务支付超时<br/>(orderNo)
        
        OrderSvc->>OrderSvc: 6. 更新订单状态<br/>status=已关闭
        OrderSvc->>InvSvc: 7. RollbackStocks(orderNo, items)<br/>释放预占库存
        InvSvc-->>OrderSvc: 7.1 回滚结果
        
        OrderSvc->>MQ: 8. 发送「订单关闭」事件
    end
```

## 用户主动取消订单流程

```mermaid
sequenceDiagram
    autonumber
    participant User as 用户/前端
    participant GW as 网关
    participant OrderSvc as 订单服务
    participant PaySvc as 支付服务
    participant InvSvc as 库存服务
    participant MQ as 消息队列

    Note over User,MQ: ========== 用户取消订单 ==========
    
    User->>GW: 1. 取消订单请求<br/>(orderNo, reason)
    GW->>OrderSvc: 2. CancelOrder(orderNo, reason)
    
    OrderSvc->>OrderSvc: 3. 查询订单记录
    OrderSvc->>OrderSvc: 4. 校验订单状态<br/>(必须是待支付)
    
    alt 订单状态不是待支付
        OrderSvc-->>GW: 返回错误：订单状态不允许取消
        GW-->>User: 提示无法取消
    else 订单状态是待支付
        OrderSvc->>OrderSvc: 5. 更新订单状态<br/>status=已取消
        OrderSvc->>PaySvc: 6. 关闭支付单(如果存在)
        PaySvc->>PaySvc: 6.1 更新支付单状态<br/>status=已关闭
        OrderSvc->>InvSvc: 7. RollbackStocks(orderNo, items)<br/>释放预占库存
        InvSvc-->>OrderSvc: 7.1 回滚结果
        OrderSvc->>MQ: 8. 发送「订单取消」事件
        OrderSvc-->>GW: 9. 返回取消成功
        GW-->>User: 10. 显示取消成功
    end
```

## 关键设计要点说明

### 1. 防超卖机制
- **下单时**：库存预占（短暂锁定，使用乐观锁）
- **支付成功**：确认扣减（如果设计为预占+支付扣减）
- **支付超时/取消**：释放预占库存

### 2. 幂等性保障
- **支付回调**：使用 `payment_no + status` 作为幂等键
- **库存操作**：使用 `order_no + sku_id + reason` 作为幂等键
- **订单状态更新**：使用乐观锁（version字段）

### 3. 安全性保障
- **签名校验**：所有支付回调必须校验签名
- **金额校验**：回调金额必须与订单金额一致
- **敏感信息**：加密存储（API密钥、私钥等）

### 4. 异常处理
- **支付超时**：定时任务自动扫描并关闭
- **回调丢失**：支持主动查询第三方支付状态
- **金额不一致**：记录告警日志，人工对账
- **服务异常**：重试机制、降级策略

### 5. 服务解耦
- **异步通信**：使用MQ发送事件，服务间解耦
- **同步通信**：关键流程使用gRPC同步调用
- **事件驱动**：便于扩展和监控
