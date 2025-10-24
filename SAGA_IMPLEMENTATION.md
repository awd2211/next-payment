# Saga 模式实现文档

## 概述

本文档详细说明了支付平台的 Saga 分布式事务补偿机制实现，用于处理跨多个微服务的分布式事务场景。

## 背景和问题

### 分布式事务挑战

在支付流程中，涉及多个微服务的协作：

```
Payment Gateway → Order Service (创建订单)
                ↓
           Channel Adapter (调用支付渠道)
```

如果任一步骤失败，需要回滚之前已完成的操作，但传统的数据库事务无法跨越服务边界。

### 典型失败场景

1. **订单创建成功，但渠道调用失败** → 需要取消订单
2. **渠道调用成功，但后续步骤失败** → 需要取消渠道支付并取消订单
3. **网络超时导致状态不一致** → 需要补偿恢复一致性

## Saga 模式设计

### 核心概念

**Saga** 是一系列本地事务的序列：
- 每个本地事务更新数据库并发布消息/事件触发下一个步骤
- 如果某个步骤失败，Saga 执行**补偿事务**回滚之前的步骤

### 实现方式

我们采用 **Saga 编排器（Orchestration）** 模式：
- 由中央编排器负责协调所有步骤
- 编排器跟踪 Saga 状态并在失败时触发补偿
- 步骤定义包含正向执行函数和补偿函数

### 架构图

```
┌─────────────────────────────────────────────────────────┐
│                  Saga Orchestrator                      │
│  - 管理 Saga 生命周期                                    │
│  - 执行步骤并追踪状态                                     │
│  - 失败时触发补偿                                         │
└────────────────┬────────────────────────────────────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
┌───▼────┐  ┌───▼────┐  ┌───▼────┐
│ Step 1 │  │ Step 2 │  │ Step 3 │
│        │  │        │  │        │
│Execute │  │Execute │  │Execute │
│  ↕     │  │  ↕     │  │  ↕     │
│Compens │  │Compens │  │Compens │
└────────┘  └────────┘  └────────┘
```

## 代码实现

### 1. Saga 核心包 (`pkg/saga/saga.go`)

**关键组件**:

#### SagaOrchestrator - Saga 编排器

```go
type SagaOrchestrator struct {
    db    *gorm.DB    // 用于持久化 Saga 状态
    redis *redis.Client
}

// 创建编排器
orchestrator := saga.NewSagaOrchestrator(db, redis)
```

#### Saga 实例

```go
type Saga struct {
    ID            uuid.UUID     // Saga ID
    BusinessID    string        // 业务ID（如 payment_no）
    BusinessType  string        // 业务类型（payment, refund）
    Status        SagaStatus    // pending, in_progress, completed, compensated, failed
    Steps         []SagaStep    // 步骤列表
    CurrentStep   int           // 当前步骤索引
    Metadata      string        // JSON格式元数据
    ...
}
```

**状态机**:

```
pending → in_progress → completed ✅
                ↓
           (失败) → compensated ⚠️
```

#### SagaStep - Saga 步骤

```go
type SagaStep struct {
    ID              uuid.UUID
    SagaID          uuid.UUID
    StepOrder       int        // 步骤顺序
    StepName        string     // 步骤名称
    Status          StepStatus // pending, completed, compensated, failed
    ExecuteData     string     // 执行参数（JSON）
    CompensateData  string     // 补偿参数（JSON）
    Result          string     // 执行结果（JSON）
    RetryCount      int        // 重试次数
    MaxRetryCount   int        // 最大重试次数
    NextRetryAt     *time.Time // 下次重试时间
    ...
}
```

#### 步骤定义

```go
type StepDefinition struct {
    Name           string
    Execute        StepFunc       // 执行函数
    Compensate     CompensateFunc // 补偿函数
    MaxRetryCount  int            // 最大重试次数
}

type StepFunc func(ctx context.Context, executeData string) (result string, err error)
type CompensateFunc func(ctx context.Context, compensateData string, executeResult string) error
```

### 2. 支付 Saga 服务 (`services/payment-gateway/internal/service/saga_payment_service.go`)

#### 支付流程 Saga 定义

```go
func (s *SagaPaymentService) ExecutePaymentSaga(ctx context.Context, payment *model.Payment) error {
    // 1. 构建 Saga
    sagaBuilder := s.orchestrator.NewSagaBuilder(payment.PaymentNo, "payment")

    // 2. 定义步骤
    stepDefs := []saga.StepDefinition{
        {
            Name: "CreateOrder",
            Execute: func(ctx context.Context, executeData string) (string, error) {
                return s.executeCreateOrder(ctx, payment)
            },
            Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
                return s.compensateCreateOrder(ctx, payment)
            },
            MaxRetryCount: 3,
        },
        {
            Name: "CallPaymentChannel",
            Execute: func(ctx context.Context, executeData string) (string, error) {
                return s.executeCallPaymentChannel(ctx, payment)
            },
            Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
                return s.compensateCallPaymentChannel(ctx, payment, executeResult)
            },
            MaxRetryCount: 3,
        },
    }

    // 3. 添加步骤到构建器
    for _, stepDef := range stepDefs {
        sagaBuilder.AddStep(stepDef.Name, stepDef.Execute, stepDef.Compensate, stepDef.MaxRetryCount)
    }

    // 4. 构建并执行 Saga
    sagaInstance, err := sagaBuilder.Build(ctx)
    if err != nil {
        return fmt.Errorf("failed to build saga: %w", err)
    }

    return s.orchestrator.Execute(ctx, sagaInstance, stepDefs)
}
```

#### 步骤1: 创建订单

```go
func (s *SagaPaymentService) executeCreateOrder(ctx context.Context, payment *model.Payment) (string, error) {
    resp, err := s.orderClient.CreateOrder(ctx, &client.CreateOrderRequest{
        MerchantID:  payment.MerchantID,
        OrderNo:     payment.OrderNo,
        PaymentNo:   payment.PaymentNo,
        Amount:      payment.Amount,
        Currency:    payment.Currency,
        // ...
    })

    if err != nil {
        return "", fmt.Errorf("create order failed: %w", err)
    }

    resultBytes, _ := json.Marshal(resp)
    return string(resultBytes), nil
}
```

**补偿**:

```go
func (s *SagaPaymentService) compensateCreateOrder(ctx context.Context, payment *model.Payment) error {
    // 调用 Order Service 取消订单
    err := s.orderClient.CancelOrder(ctx, payment.OrderNo, "支付流程失败，自动取消")
    if err != nil {
        return fmt.Errorf("cancel order failed: %w", err)
    }
    return nil
}
```

#### 步骤2: 调用支付渠道

```go
func (s *SagaPaymentService) executeCallPaymentChannel(ctx context.Context, payment *model.Payment) (string, error) {
    resp, err := s.channelClient.CreatePayment(ctx, &client.CreatePaymentRequest{
        PaymentNo:     payment.PaymentNo,
        MerchantID:    payment.MerchantID.String(),
        Amount:        payment.Amount,
        Currency:      payment.Currency,
        Channel:       payment.Channel,
        // ...
    })

    if err != nil {
        return "", fmt.Errorf("call payment channel failed: %w", err)
    }

    // 更新支付记录
    payment.ChannelOrderNo = resp.ChannelTradeNo
    s.paymentRepo.Update(ctx, payment)

    resultBytes, _ := json.Marshal(resp)
    return string(resultBytes), nil
}
```

**补偿**:

```go
func (s *SagaPaymentService) compensateCallPaymentChannel(ctx context.Context, payment *model.Payment, executeResult string) error {
    // 取消渠道支付
    if payment.ChannelOrderNo != "" {
        err := s.channelClient.CancelPayment(ctx, payment.ChannelOrderNo)
        if err != nil {
            logger.Error("failed to cancel payment in channel", zap.Error(err))
        }
    }

    // 更新支付状态为失败
    payment.Status = model.PaymentStatusFailed
    payment.ErrorMsg = "Saga 补偿: 分布式事务回滚"
    return s.paymentRepo.Update(ctx, payment)
}
```

### 3. 数据库表结构

#### saga_instances 表

```sql
CREATE TABLE saga_instances (
    id UUID PRIMARY KEY,
    business_id VARCHAR(255) NOT NULL,
    business_type VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    current_step INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    metadata TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    compensated_at TIMESTAMP
);

CREATE INDEX idx_saga_instances_business_id ON saga_instances(business_id);
CREATE INDEX idx_saga_instances_status ON saga_instances(status);
```

#### saga_steps 表

```sql
CREATE TABLE saga_steps (
    id UUID PRIMARY KEY,
    saga_id UUID NOT NULL REFERENCES saga_instances(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    execute_data TEXT,
    compensate_data TEXT,
    result TEXT,
    error_message TEXT,
    executed_at TIMESTAMP,
    compensated_at TIMESTAMP,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retry_count INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saga_steps_saga_id ON saga_steps(saga_id);
CREATE INDEX idx_saga_steps_status ON saga_steps(status);
CREATE INDEX idx_saga_steps_next_retry_at ON saga_steps(next_retry_at);
```

## 执行流程

### 正常流程（成功）

```
1. Payment Gateway 创建 Saga
   ↓
2. 执行步骤1: CreateOrder
   ↓ (成功)
3. 执行步骤2: CallPaymentChannel
   ↓ (成功)
4. Saga 标记为 completed ✅
```

**数据库状态**:

```
saga_instances:
- status: completed
- current_step: 2

saga_steps:
- Step 1 (CreateOrder): status = completed
- Step 2 (CallPaymentChannel): status = completed
```

### 异常流程（补偿）

```
1. Payment Gateway 创建 Saga
   ↓
2. 执行步骤1: CreateOrder
   ↓ (成功)
3. 执行步骤2: CallPaymentChannel
   ↓ (失败 ❌)
4. 开始补偿流程
   ↓
5. 补偿步骤2: compensateCallPaymentChannel
   ↓ (成功)
6. 补偿步骤1: compensateCreateOrder (取消订单)
   ↓ (成功)
7. Saga 标记为 compensated ⚠️
```

**数据库状态**:

```
saga_instances:
- status: compensated
- current_step: 2
- error_message: "步骤 CallPaymentChannel 失败: ..."

saga_steps:
- Step 1 (CreateOrder): status = compensated
- Step 2 (CallPaymentChannel): status = failed, retry_count = 3
```

### 重试机制

```
1. 步骤失败
   ↓
2. retry_count < max_retry_count?
   ↓ YES
3. 计算下次重试时间（指数退避）
   next_retry_at = now + 2^retry_count 秒
   ↓
4. 等待到 next_retry_at
   ↓
5. 重试执行步骤
   ↓
6. 成功? → 继续下一步
   ↓ NO
7. retry_count++, 回到步骤2
   ↓
8. retry_count >= max_retry_count? → 开始补偿
```

## 集成方式

### 在 Payment Gateway 中使用

#### 1. 初始化 Saga Orchestrator

```go
// main.go
import "github.com/payment-platform/pkg/saga"

orchestrator := saga.NewSagaOrchestrator(database, redisClient)
sagaPaymentService := service.NewSagaPaymentService(
    orchestrator,
    paymentRepo,
    orderClient,
    channelClient,
)
```

#### 2. 在 CreatePayment 中调用 Saga

```go
// payment_service.go

// 创建支付记录后，使用 Saga 协调后续步骤
err := s.sagaPaymentService.ExecutePaymentSaga(ctx, payment)
if err != nil {
    logger.Error("saga execution failed", zap.Error(err))
    // Saga 会自动补偿，这里只需记录日志
    return nil, fmt.Errorf("支付流程失败: %w", err)
}
```

#### 3. 添加手动补偿 API

```go
// handler/saga_handler.go

// POST /api/v1/saga/payments/:payment_no/compensate
func (h *SagaHandler) CompensatePayment(c *gin.Context) {
    paymentNo := c.Param("payment_no")

    err := h.sagaPaymentService.CompensatePayment(c.Request.Context(), paymentNo)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "补偿成功"})
}
```

## 客户端补偿接口

### Order Service - CancelOrder

```go
// payment-gateway/internal/client/order_client.go

func (c *OrderClient) CancelOrder(ctx context.Context, orderNo string, reason string) error {
    path := fmt.Sprintf("/api/v1/orders/%s/cancel", orderNo)
    req := map[string]string{"reason": reason}

    resp, err := c.http.Post(ctx, path, req, nil)
    // ...
}
```

### Channel Adapter - CancelPayment

```go
// payment-gateway/internal/client/channel_client.go

func (c *ChannelClient) CancelPayment(ctx context.Context, channelTradeNo string) error {
    path := fmt.Sprintf("/api/v1/channel/payment/%s/cancel", channelTradeNo)

    resp, err := c.http.Post(ctx, path, nil, nil)
    // ...
}
```

## 监控和可观测性

### Saga 执行日志

```
INFO  saga execution started | saga_id=xxx | business_id=PAY-123 | business_type=payment
INFO  executing saga step | saga_id=xxx | step=0 | step_name=CreateOrder
INFO  saga step completed | saga_id=xxx | step_name=CreateOrder
INFO  executing saga step | saga_id=xxx | step=1 | step_name=CallPaymentChannel
ERROR saga step failed after max retries | saga_id=xxx | step_name=CallPaymentChannel | error=...
INFO  saga compensation started | saga_id=xxx | steps_to_compensate=1
INFO  compensating saga step | saga_id=xxx | step_name=CreateOrder
INFO  saga step compensated | saga_id=xxx | step_name=CreateOrder
INFO  saga compensation completed | saga_id=xxx
```

### 数据库查询

```sql
-- 查询失败的 Saga
SELECT * FROM saga_instances WHERE status = 'compensated' ORDER BY created_at DESC LIMIT 10;

-- 查询需要重试的步骤
SELECT * FROM saga_steps
WHERE status = 'failed' AND next_retry_at IS NOT NULL AND next_retry_at <= NOW()
LIMIT 100;

-- 查询某个支付的 Saga 详情
SELECT s.*, st.*
FROM saga_instances s
LEFT JOIN saga_steps st ON s.id = st.saga_id
WHERE s.business_id = 'PAY-20250124-123456'
ORDER BY st.step_order;
```

## 优势和限制

### 优势

✅ **最终一致性**: 通过补偿机制确保分布式事务最终一致
✅ **可观测性**: 所有 Saga 执行历史持久化到数据库，可追踪审计
✅ **可靠性**: 支持重试机制，自动处理瞬时故障
✅ **可扩展性**: 易于添加新步骤，支持复杂业务流程
✅ **解耦**: 服务间通过 API 调用，无需共享数据库

### 限制

⚠️ **非原子性**: 中间状态可见（如订单已创建但支付失败）
⚠️ **补偿复杂性**: 某些操作难以补偿（如已发送的邮件）
⚠️ **性能开销**: 每个步骤都需要持久化到数据库
⚠️ **补偿失败**: 补偿操作本身也可能失败，需要人工介入

## 最佳实践

### 1. 幂等性

确保所有步骤的执行函数和补偿函数都是幂等的：

```go
// ✅ 幂等的订单取消
func CancelOrder(orderNo string) error {
    // 检查订单状态，如果已取消则直接返回成功
    order := getOrder(orderNo)
    if order.Status == "cancelled" {
        return nil
    }
    // 执行取消逻辑
    return doCancel(order)
}
```

### 2. 补偿顺序

补偿必须按**逆序**执行：

```
Execute:    Step1 → Step2 → Step3
Compensate: Step3 → Step2 → Step1
```

### 3. 补偿可失败

补偿操作应该设计为"尽力而为"，即使补偿失败也不阻止其他步骤的补偿：

```go
func (s *SagaPaymentService) compensateCallPaymentChannel(...) error {
    // 尝试取消渠道支付
    if err := s.channelClient.CancelPayment(...); err != nil {
        logger.Error("failed to cancel in channel", zap.Error(err))
        // 记录错误但不返回，继续更新本地状态
    }

    // 更新本地支付状态（必须成功）
    return s.paymentRepo.Update(ctx, payment)
}
```

### 4. 监控和告警

设置告警规则：

```sql
-- 查询长时间未完成的 Saga
SELECT * FROM saga_instances
WHERE status = 'in_progress' AND created_at < NOW() - INTERVAL '1 hour';

-- 查询补偿失败的 Saga（需要人工介入）
SELECT * FROM saga_instances WHERE status = 'failed';
```

## 未来改进

### 短期 (1-2周)

- [ ] 添加 Saga Dashboard（查看所有 Saga 状态）
- [ ] 实现自动重试后台任务（扫描 next_retry_at）
- [ ] 添加 Prometheus 指标（saga_executions_total, saga_compensations_total）
- [ ] 编写单元测试和集成测试

### 中期 (1-2月)

- [ ] 支持异步步骤（通过 Kafka 消息队列）
- [ ] 实现 Saga 可视化流程图
- [ ] 添加手动重试/补偿功能（管理后台）
- [ ] 支持超时控制（步骤级别超时）

### 长期 (3-6月)

- [ ] 实现并行步骤（多个步骤同时执行）
- [ ] 支持嵌套 Saga（Saga 中包含子 Saga）
- [ ] 集成分布式追踪（Jaeger）
- [ ] 实现 Saga 编排 DSL（配置文件定义流程）

## 总结

### 已完成

✅ 实现了完整的 Saga 编排器框架
✅ 支持步骤定义、执行、补偿、重试
✅ 持久化 Saga 状态到 PostgreSQL
✅ 实现了支付流程的 Saga 集成
✅ 添加了 Order 和 Channel 客户端的补偿接口
✅ 创建了数据库迁移文件

### 技术亮点

- **编排器模式**: 中央控制，易于理解和维护
- **持久化状态**: 数据库记录，可追踪审计
- **自动重试**: 指数退避策略，处理瞬时故障
- **灵活补偿**: 支持自定义补偿逻辑

### 下一步

1. 在 Order Service 和 Channel Adapter 中实现 `/cancel` 接口
2. 在 Payment Gateway main.go 中集成 Saga Orchestrator
3. 运行数据库迁移创建 saga 表
4. 编写测试验证补偿流程

---

**文档版本**: 1.0
**创建时间**: 2025-01-24
**最后更新**: 2025-01-24
**维护者**: Payment Platform Team
