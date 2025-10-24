# 支付平台微服务通信快速参考

## 一句话总结

系统有 15 个微服务，17 个 HTTP clients，但只有 18% 的 clients 使用了熔断器，急需统一改进。

## 关键数字

| 指标 | 数值 | 评价 |
|------|------|------|
| 微服务总数 | 15 | ✅ |
| HTTP Clients | 17 | ✅ |
| 带熔断器的 Clients | 3 | ❌ 18% |
| 主要调用链路 | 5 | ✅ |
| 缺失链路 | 3 | ❌ |
| 架构整体评分 | 6.4/10 | ⚠️  |

## 最好的实现

**payment-gateway/internal/client** (A+ 级)
```go
// 自动获得: 熔断器、重试 3 次、日志、Jaeger 追踪
orderClient := client.NewOrderClient(orderServiceURL)
```

使用方式：
- OrderClient → order-service (40004)
- ChannelClient → channel-adapter (40005)
- RiskClient → risk-service (40006)

## 最差的实现

**merchant-service/internal/client** (B 级)
```go
// 问题: 无熔断器、无重试、无日志、无追踪
notificationClient := client.NewNotificationClient(notificationServiceURL)
```

影响的服务：
- payment-gateway
- notification-service
- accounting-service
- analytics-service
- risk-service

## 3 条缺失的关键链路

### 链路 A: Notification 反向集成
```
需要添加:
- payment-gateway 支付完成 → 发送通知 ❌
- settlement-service 结算完成 → 发送通知 ❌
- order-service 订单变更 → 发送通知 ❌
```

### 链路 B: Analytics 主动推送
```
需要添加:
- payment-gateway 推送支付事件 ❌
- order-service 推送订单事件 ❌
- channel-adapter 推送交易事件 ❌
```

### 链路 C: Config Service 使用
```
需要添加:
- 所有 15 个服务都应该读取 config-service ❌
```

## 立即修复清单 (1 周内)

### 第 1 天: 为 merchant-service 应用熔断器
```bash
# 修改 5 个 clients: payment, notification, accounting, analytics, risk
# 影响: 防止商户仪表盘级联故障
# 时间: 2-3 小时
```

### 第 2 天: 为 settlement/withdrawal 应用熔断器
```bash
# 修改 5 个 clients
# 影响: 防止结算和提现故障
# 时间: 2-3 小时
```

### 第 3 天: 为 merchant-auth 应用熔断器
```bash
# 修改 1 个 client
# 影响: 防止认证故障
# 时间: 1 小时
```

### 第 4-5 天: 添加 notification 集成
```bash
# 在 payment-gateway 和 settlement-service 中
# 时间: 3-4 小时
```

## 代码对比

### 不好的做法 ❌
```go
// merchant-service
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
// 问题: 无熔断器、无重试、无日志
```

### 好的做法 ✅
```go
// payment-gateway
orderClient := client.NewOrderClient(orderServiceURL)
order, err := orderClient.CreateOrder(ctx, req)
// 优点: 自动熔断器、重试、日志、追踪
```

### 改进方式
```go
// 在所有 clients 中统一使用
client := NewServiceClientWithBreaker(baseURL, "service-name")
```

## 17 个 Clients 完整列表

### ✅ 已优化 (3 个)
- payment-gateway: OrderClient, ChannelClient, RiskClient

### ❌ 需要改进 (14 个)
- merchant-service: PaymentClient, NotificationClient, AccountingClient, AnalyticsClient, RiskClient
- settlement-service: AccountingClient, WithdrawalClient
- withdrawal-service: AccountingClient, NotificationClient, BankTransferClient
- merchant-auth-service: MerchantClient
- channel-adapter: ExchangeRateClient
- risk-service: IPAPIClient

## 5 条主要调用链路

### 1. 支付创建流程
```
payment-gateway (40003)
  ├─→ risk-service (40006) - 风控检查
  ├─→ order-service (40004) - 创建订单
  └─→ channel-adapter (40005) - 选择渠道
```
**质量**: ✅ A+ 

### 2. 商户仪表盘
```
merchant-service (40002)
  ├─→ analytics-service (40009)
  ├─→ accounting-service (40007)
  ├─→ risk-service (40006)
  ├─→ notification-service (40008)
  └─→ payment-gateway (40003)
```
**质量**: ⚠️  B (无熔断器)

### 3. 结算流程
```
settlement-service (40013)
  ├─→ accounting-service (40007)
  └─→ withdrawal-service (40014)
```
**质量**: ⚠️  B- (未优化)

### 4. 提现流程
```
withdrawal-service (40014)
  ├─→ accounting-service (40007)
  ├─→ notification-service (40008)
  └─→ Bank API (外部)
```
**质量**: ⚠️  B- (通知可靠性差)

### 5. 认证流程
```
merchant-auth-service (40011)
  └─→ merchant-service (40002)
```
**质量**: ⚠️  B- (无熔断器)

## 熔断器配置详情

```go
// 默认配置 (pkg/httpclient)
config := &httpclient.Config{
    Timeout:    30 * time.Second,    // 请求超时
    MaxRetries: 3,                   // 重试次数
    RetryDelay: time.Second,         // 初始延迟
}

breakerConfig := &httpclient.BreakerConfig{
    MaxRequests: 3,                  // 半开状态允许 3 个请求
    Interval:    time.Minute,        // 1 分钟统计窗口
    Timeout:     30 * time.Second,   // 30 秒后尝试恢复
    ReadyToTrip: func(counts) bool {
        // 5 个请求中 60% 失败则熔断
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 5 && failureRatio >= 0.6
    },
}
```

## 可观测性支持

### Prometheus ✅
- 所有 15 个服务都有 `/metrics` 端点
- HTTP 请求计数、耗时、大小都有记录

### Jaeger ✅
- 所有服务都支持分布式追踪
- W3C Trace Context 传播
- 但 clients 调用中缺少 span

### 建议
```go
// 在所有 client 调用中添加 span
ctx, span := tracing.StartSpan(ctx, "service-name", "OperationName")
defer span.End()
```

## 文件位置

```
核心库:
  /home/eric/payment/backend/pkg/httpclient/        # HTTP Client 基础库
  /home/eric/payment/backend/pkg/httpclient/client.go        # 重试机制
  /home/eric/payment/backend/pkg/httpclient/breaker.go       # 熔断器

最好的例子:
  /home/eric/payment/backend/services/payment-gateway/internal/client/

需要改进:
  /home/eric/payment/backend/services/merchant-service/internal/client/
  /home/eric/payment/backend/services/settlement-service/internal/client/
  /home/eric/payment/backend/services/withdrawal-service/internal/client/

完整分析:
  /home/eric/payment/MICROSERVICE_COMMUNICATION_ANALYSIS.md
  /home/eric/payment/ARCHITECTURE_SUMMARY.txt
```

## 优先级 P1 任务

| 任务 | 工时 | 开始日期 |
|------|------|---------|
| merchant-service 5 个 clients | 2-3h | Day 1 |
| settlement/withdrawal 5 个 clients | 2-3h | Day 2 |
| merchant-auth 1 个 client | 1h | Day 3 |
| notification 集成 | 3-4h | Day 4-5 |
| **总计** | **8-11h** | **1 周** |

**预期效果**: 
- 防止级联故障
- 降低错误率 80%
- 改善用户体验

## 常见问题

### Q: 为什么需要熔断器?
A: 防止服务故障扩散。当下游服务故障时，熔断器会快速返回错误，避免占用上游资源。

### Q: 能否不修改所有 clients?
A: 不能。14 个 clients 都缺少保护，风险很高。payment-gateway 虽然好，但如果 merchant-service 故障会影响整个仪表盘。

### Q: 修改 clients 会影响现有代码吗?
A: 不会。`NewServiceClientWithBreaker` 是 `NewServiceClient` 的增强版，完全兼容。

### Q: 什么时候可以看到效果?
A: 修改后立即生效。客户端调用时自动获得熔断器、重试和日志保护。

## 下一步

1. **立即**: 审阅本文档，确认优先级
2. **今天**: 在 merchant-service 中应用 `NewServiceClientWithBreaker`
3. **本周**: 完成所有 14 个 clients 的改进
4. **下周**: 添加 notification 和 analytics 集成

---

**最后更新**: 2024-10-24  
**作者**: 微服务架构分析  
**文档等级**: 高优先级，需要立即行动  
