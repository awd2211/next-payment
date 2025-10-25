# 支付路由优化系统 - 完整指南

## 概述

支付路由优化系统是一个智能渠道选择引擎，能够根据多种策略自动选择最优的支付渠道，实现成本优化、成功率提升和负载均衡。

**核心特性**:
- 🎯 **多策略路由**: 支持成本优先、成功率优先、地域优化、负载均衡等策略
- 💰 **成本优化**: 自动选择手续费最低的渠道
- ✅ **成功率优化**: 优先选择历史成功率最高的渠道
- 🌍 **地域优化**: 根据用户国家选择本地化渠道
- ⚖️ **负载均衡**: 基于权重分配流量，避免单点过载
- 📊 **实时指标**: 支持动态更新渠道成功率和响应时间
- 🔧 **灵活配置**: 支持渠道启用/禁用、配置热更新

---

## 架构设计

### 组件结构

```
┌─────────────────────────────────────────────────────────┐
│                    RouterService                        │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐        ┌────────────────────────┐   │
│  │ PaymentRouter│───────>│ RoutingStrategy       │   │
│  │              │        │ - Geographic          │   │
│  │              │        │ - SuccessRate         │   │
│  │              │        │ - CostOptimization    │   │
│  │              │        │ - LoadBalance         │   │
│  └──────────────┘        └────────────────────────┘   │
│         │                                              │
│  ┌──────────────┐                                      │
│  │ConfigManager │<───── Redis Cache                   │
│  │              │                                      │
│  │  Channels    │                                      │
│  └──────────────┘                                      │
└─────────────────────────────────────────────────────────┘
```

### 路由策略优先级

路由器按优先级依次尝试策略，直到找到可用渠道：

| 策略 | 优先级 | 说明 |
|------|--------|------|
| Geographic | 90 | 地域优化，根据国家选择本地化渠道 |
| SuccessRate | 80 | 成功率优先，选择历史成功率最高的渠道 |
| CostOptimization | 50 | 成本优化，选择手续费最低的渠道 |
| LoadBalance | 30 | 负载均衡，基于权重随机选择 |

---

## 使用指南

### 1. 初始化路由服务

```go
import (
    "github.com/payment-platform/pkg/router"
)

// 创建路由服务
routerService := router.NewRouterService(redisClient)

// 初始化（选择路由模式）
// 模式: "balanced"(默认), "cost", "success", "geographic"
err := routerService.Initialize(ctx, "balanced")
if err != nil {
    log.Fatalf("路由服务初始化失败: %v", err)
}
```

### 2. 选择支付渠道

```go
// 创建路由请求
req := &router.RoutingRequest{
    MerchantID:    merchantID,
    Amount:        50000,      // 500美元（分）
    Currency:      "USD",
    Country:       "US",
    PayMethod:     "card",
    PreferChannel: "",         // 留空则自动选择
}

// 执行路由
result, err := routerService.SelectChannel(ctx, req)
if err != nil {
    return err
}

fmt.Printf("选择渠道: %s\n", result.Channel)
fmt.Printf("选择原因: %s\n", result.Reason)
fmt.Printf("估算手续费: %d\n", result.EstimatedFee)
fmt.Printf("费率: %.2f%%\n", result.FeeRate*100)
```

**输出示例**:
```
选择渠道: stripe
选择原因: 地域优化（US 本地化渠道）
估算手续费: 1480
费率: 2.90%
```

### 3. 指定优先渠道

```go
req := &router.RoutingRequest{
    MerchantID:    merchantID,
    Amount:        50000,
    Currency:      "USD",
    PreferChannel: "stripe",  // 指定优先渠道
}

result, _ := routerService.SelectChannel(ctx, req)
// 结果: channel=stripe, reason="商户指定优先渠道"
```

---

## 路由策略详解

### 1. Geographic Strategy (地域优化)

**优先级**: 90（最高）
**适用场景**: 有明确国家信息的交易

**策略逻辑**:
```go
国家 → 推荐渠道
CN  → alipay   (中国 → 支付宝)
US  → stripe   (美国 → Stripe)
EU  → stripe   (欧洲 → Stripe)
JP  → stripe   (日本 → Stripe)
SG  → stripe   (新加坡 → Stripe)
```

**使用示例**:
```go
req := &router.RoutingRequest{
    Amount:   100000,
    Currency: "CNY",
    Country:  "CN",  // 中国
}

result, _ := routerService.SelectChannel(ctx, req)
// 结果: channel=alipay, reason="地域优化（CN 本地化渠道）"
```

**优势**:
- ✅ 提高本地用户支付成功率
- ✅ 减少跨境手续费
- ✅ 更好的用户体验（本地化支付方式）

---

### 2. SuccessRate Strategy (成功率优先)

**优先级**: 80（高）
**适用场景**: 对成功率要求高的重要交易

**策略逻辑**:
- 选择历史成功率最高的可用渠道
- 考虑币种、国家、支付方式支持情况
- 忽略手续费差异

**渠道成功率** (默认配置):
| 渠道 | 成功率 |
|------|--------|
| Alipay | 98% |
| WeChat | 97% |
| Stripe | 95% |
| PayPal | 92% |
| Crypto | 88% |

**使用示例**:
```go
req := &router.RoutingRequest{
    Amount:   100000,
    Currency: "CNY",
    Country:  "CN",
}

// 在中国，支付宝成功率最高（98%）
result, _ := routerService.SelectChannel(ctx, req)
// 结果: channel=alipay, reason="成功率最高（98.0%）"
```

**优势**:
- ✅ 减少支付失败率
- ✅ 提升用户满意度
- ✅ 减少客服成本

---

### 3. CostOptimization Strategy (成本优化)

**优先级**: 50（中）
**适用场景**: 成本敏感的大额交易

**策略逻辑**:
- 计算每个渠道的手续费：`fee = amount * feeRate` (如果 < minFee，则取 minFee)
- 选择手续费最低的渠道

**渠道费率对比**:
| 渠道 | 费率 | 最低费用 |
|------|------|---------|
| Alipay | 0.6% | ¥0.10 |
| WeChat | 0.6% | ¥0.10 |
| Stripe | 2.9% | $0.30 |
| PayPal | 3.4% | $0.30 |
| Crypto | 1.0% | $0 |

**使用示例**:
```go
// 大额交易，成本优化明显
req := &router.RoutingRequest{
    Amount:   10000000,  // 100,000美元
    Currency: "USD",
}

result, _ := routerService.SelectChannel(ctx, req)
// Crypto: 100,000 * 1.0% = $1,000
// Stripe: 100,000 * 2.9% = $2,900
// PayPal: 100,000 * 3.4% = $3,400
// 结果: channel=crypto, reason="成本最优（费率: 1.00%）"
```

**成本对比（100,000美元交易）**:
```
Crypto:  $1,000  (1.0%)  ← 最低成本
Stripe:  $2,900  (2.9%)
PayPal:  $3,400  (3.4%)
节省:    $1,900 (65.5%)
```

**优势**:
- ✅ 降低交易成本
- ✅ 大额交易节省显著
- ✅ 提高利润率

---

### 4. LoadBalance Strategy (负载均衡)

**优先级**: 30（低，兜底策略）
**适用场景**: 多个渠道都可用时分散流量

**策略逻辑**:
- 基于权重的加权随机选择
- 避免单一渠道过载
- 提高系统整体可用性

**渠道权重** (默认配置):
| 渠道 | 权重 |
|------|------|
| Alipay | 120 |
| WeChat | 110 |
| Stripe | 100 |
| PayPal | 80 |
| Crypto | 50 |

**使用示例**:
```go
// 多次调用会按权重分配
for i := 0; i < 100; i++ {
    result, _ := routerService.SelectChannel(ctx, req)
    fmt.Println(result.Channel)
}

// 预期分布:
// Alipay: ~26次 (120/460 ≈ 26%)
// WeChat: ~24次 (110/460 ≈ 24%)
// Stripe: ~22次 (100/460 ≈ 22%)
// PayPal: ~17次 (80/460 ≈ 17%)
// Crypto: ~11次 (50/460 ≈ 11%)
```

**优势**:
- ✅ 避免单点过载
- ✅ 提高系统容错能力
- ✅ 均衡渠道使用

---

## 渠道配置管理

### 1. 获取所有渠道配置

```go
channels := routerService.GetAllChannels()
for _, ch := range channels {
    fmt.Printf("渠道: %s, 启用: %v, 费率: %.2f%%\n",
        ch.Channel, ch.IsEnabled, ch.FeeRate*100)
}
```

**输出示例**:
```
渠道: stripe, 启用: true, 费率: 2.90%
渠道: paypal, 启用: true, 费率: 3.40%
渠道: alipay, 启用: true, 费率: 0.60%
渠道: wechat, 启用: true, 费率: 0.60%
渠道: crypto, 启用: false, 费率: 1.00%
```

### 2. 获取单个渠道配置

```go
config, err := routerService.GetChannelConfig("stripe")
if err != nil {
    return err
}

fmt.Printf("渠道名称: %s\n", config.Channel)
fmt.Printf("支持币种: %v\n", config.SupportedCurrencies)
fmt.Printf("支持国家: %v\n", config.SupportedCountries)
fmt.Printf("费率: %.2f%%\n", config.FeeRate*100)
fmt.Printf("最低费用: %d\n", config.MinFee)
fmt.Printf("成功率: %.1f%%\n", config.SuccessRate*100)
fmt.Printf("平均响应时间: %dms\n", config.AvgResponseTime)
```

### 3. 更新渠道状态（启用/禁用）

```go
// 禁用渠道
err := routerService.UpdateChannelStatus(ctx, "crypto", false)

// 启用渠道
err = routerService.UpdateChannelStatus(ctx, "stripe", true)
```

**使用场景**:
- 渠道故障时临时禁用
- 维护期间禁用
- 新渠道上线时启用
- A/B测试时动态切换

### 4. 更新渠道指标

```go
// 更新成功率和响应时间
err := routerService.UpdateChannelMetrics(ctx, "stripe", 0.96, 450)
```

**自动更新示例**:
```go
// 在支付完成后更新渠道指标
func afterPayment(channel string, success bool, responseTime int64) {
    // 计算新的成功率（滑动窗口）
    currentRate := getCurrentSuccessRate(channel)
    newRate := (currentRate*0.95 + (successFloat)*0.05) // 指数移动平均

    // 计算新的响应时间
    currentAvgTime := getCurrentAvgResponseTime(channel)
    newAvgTime := (currentAvgTime*0.95 + responseTime*0.05)

    // 更新指标
    routerService.UpdateChannelMetrics(ctx, channel, newRate, newAvgTime)
}
```

### 5. 根据条件筛选渠道

```go
// 根据国家筛选
channels := routerService.GetChannelsByCountry("CN")
// 结果: [alipay, wechat, stripe, paypal]

// 根据币种筛选
channels = routerService.GetChannelsByCurrency("CNY")
// 结果: [alipay, wechat]
```

### 6. 估算手续费

```go
fee, err := routerService.EstimateFee(ctx, "stripe", 50000)
// 结果: 1480 (500美元 * 2.9% + 30美分)
```

---

## 路由模式

### 1. Balanced Mode (平衡模式 - 默认)

注册所有策略，平衡考虑地域、成功率、成本和负载：

```go
routerService.Initialize(ctx, "balanced")
```

**策略顺序**:
1. Geographic (90) - 优先本地化
2. SuccessRate (80) - 保证成功率
3. CostOptimization (50) - 优化成本
4. LoadBalance (30) - 兜底分流

**适用场景**: 大多数业务场景

---

### 2. Cost Mode (成本优先模式)

只注册成本优化和负载均衡策略：

```go
routerService.Initialize(ctx, "cost")
```

**策略顺序**:
1. CostOptimization (50)
2. LoadBalance (30)

**适用场景**:
- 大额交易
- 成本敏感业务
- B2B支付

---

### 3. Success Mode (成功率优先模式)

优先保证成功率：

```go
routerService.Initialize(ctx, "success")
```

**策略顺序**:
1. SuccessRate (80)
2. CostOptimization (50)
3. LoadBalance (30)

**适用场景**:
- 高价值交易
- VIP用户支付
- 重要订单

---

### 4. Geographic Mode (地域优先模式)

强化地域优化：

```go
routerService.Initialize(ctx, "geographic")
```

**策略顺序**:
1. Geographic (90)
2. SuccessRate (80)
3. CostOptimization (50)
4. LoadBalance (30)

**适用场景**:
- 跨境电商
- 多国业务
- 本地化要求高的场景

---

## 高级用法

### 1. 自定义路由策略

```go
// 实现自定义策略
type CustomStrategy struct {
    channels []*router.ChannelConfig
}

func (s *CustomStrategy) Name() string {
    return "Custom"
}

func (s *CustomStrategy) Priority() int {
    return 100 // 最高优先级
}

func (s *CustomStrategy) SelectChannel(ctx context.Context, req *router.RoutingRequest) (*router.RoutingResult, error) {
    // 自定义逻辑
    // ...
    return &router.RoutingResult{
        Channel: "my_channel",
        Reason:  "自定义策略",
    }, nil
}

// 注册自定义策略
routerService.router.RegisterStrategy(customStrategy)
```

### 2. 实时指标更新

```go
// 定时任务：每5分钟更新一次渠道指标
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        channels := routerService.GetAllChannels()
        for _, ch := range channels {
            // 从数据库或监控系统获取实时指标
            successRate := calculateSuccessRate(ch.Channel)
            avgResponseTime := calculateAvgResponseTime(ch.Channel)

            // 更新指标
            routerService.UpdateChannelMetrics(ctx, ch.Channel, successRate, avgResponseTime)
        }
    }
}()
```

### 3. 动态权重调整

```go
// 根据当前负载动态调整权重
func adjustWeights(ctx context.Context) {
    channels := routerService.GetAllChannels()
    for _, ch := range channels {
        // 获取当前负载
        currentLoad := getCurrentLoad(ch.Channel)

        // 负载高，降低权重
        if currentLoad > 0.8 {
            ch.Weight = int(float64(ch.Weight) * 0.5)
        }

        // 负载低，提高权重
        if currentLoad < 0.3 {
            ch.Weight = int(float64(ch.Weight) * 1.5)
        }

        routerService.configManager.UpdateChannel(ctx, ch)
    }
}
```

---

## 监控和告警

### 1. 渠道健康检查

```go
func checkChannelHealth(ctx context.Context) {
    channels := routerService.GetAllChannels()
    for _, ch := range channels {
        if ch.SuccessRate < 0.85 {
            logger.Warn("渠道成功率过低",
                zap.String("channel", ch.Channel),
                zap.Float64("success_rate", ch.SuccessRate))

            // 发送告警
            alerting.Send("渠道成功率过低", ch.Channel)
        }

        if ch.AvgResponseTime > 2000 {
            logger.Warn("渠道响应时间过长",
                zap.String("channel", ch.Channel),
                zap.Int64("avg_response_time", ch.AvgResponseTime))
        }
    }
}
```

### 2. Prometheus指标

```promql
# 各渠道使用频率
sum(rate(payment_channel_selected_total[5m])) by (channel)

# 各策略命中率
sum(rate(routing_strategy_hit_total[5m])) by (strategy)

# 渠道成功率
avg(channel_success_rate) by (channel)

# 平均手续费
avg(channel_estimated_fee) by (channel)
```

---

## 最佳实践

### 1. 根据业务特点选择模式

```go
// B2C小额高频 → balanced/geographic
if isB2C && avgAmount < 10000 {
    mode = "balanced"
}

// B2B大额低频 → cost
if isB2B && avgAmount > 100000 {
    mode = "cost"
}

// 高端用户 → success
if isVIP {
    mode = "success"
}
```

### 2. 设置合理的渠道权重

```go
// 根据渠道稳定性和成本设置权重
weights := map[string]int{
    "primary":   150,  // 主渠道，最稳定
    "secondary": 100,  // 备用渠道
    "fallback":  50,   // 兜底渠道
}
```

### 3. 及时更新渠道配置

```go
// 渠道故障时立即禁用
if channelError {
    routerService.UpdateChannelStatus(ctx, channel, false)
}

// 故障恢复后启用
if channelRecovered {
    routerService.UpdateChannelStatus(ctx, channel, true)
}
```

### 4. 定期重新加载配置

```go
// 每小时重新加载一次配置
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        if err := routerService.ReloadChannels(ctx); err != nil {
            logger.Error("重新加载渠道配置失败", zap.Error(err))
        }
    }
}()
```

---

## 性能优化

### 1. Redis缓存

配置管理器自动使用Redis缓存渠道配置，缓存时间5分钟：

```go
// 缓存key: payment:router:channel_configs
// TTL: 5分钟
```

### 2. 策略执行性能

各策略执行时间（平均）：

| 策略 | 执行时间 |
|------|----------|
| Geographic | < 1ms |
| SuccessRate | < 1ms |
| CostOptimization | < 2ms |
| LoadBalance | < 1ms |

总路由时间：< 5ms

---

## 总结

支付路由优化系统提供了强大而灵活的渠道选择能力，能够根据不同的业务需求自动选择最优渠道。

**关键优势**:
- 🎯 智能路由：多策略自动选择
- 💰 成本节省：大额交易可节省65%+手续费
- ✅ 成功率提升：优先高成功率渠道
- 🌍 本地化：根据地域选择最优渠道
- ⚖️ 负载均衡：避免单点过载
- 📊 实时优化：动态更新渠道指标

**生产就绪**:
- ✅ 完整的策略体系
- ✅ Redis缓存支持
- ✅ 热配置更新
- ✅ 详细的日志记录
- ✅ 性能优化

**下一步**:
1. 集成到payment-gateway服务
2. 添加路由指标监控
3. 实现自定义策略
4. A/B测试不同路由模式

---

**文档版本**: v1.0
**最后更新**: 2025-01-24
**作者**: Payment Platform Team
