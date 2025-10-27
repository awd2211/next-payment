# 19个微服务 TODO 检查报告

## 总览统计

| 服务名称 | TODO数量 | 状态 |
|---------|---------|------|
| payment-gateway | 2 | ⚠️ 需处理 |
| channel-adapter | 1 | ⚠️ 需处理 |
| accounting-service | 1 | ⚠️ 需处理 |
| risk-service | 2 | ⚠️ 需处理 |
| analytics-service | 4 | ⚠️ 需处理 |
| config-service | 1 | ⚠️ 需处理 |
| merchant-auth-service | 1 | ⚠️ 需处理 |
| settlement-service | 4 (3已修复) | ✅ 大部分完成 |
| kyc-service | 2 | ⚠️ 需处理 |
| admin-bff-service | 5 | ⚠️ 需处理 |
| merchant-bff-service | 1 | ⚠️ 需处理 |
| merchant-policy-service | 1 | ⚠️ 需处理 |
| order-service | 0 | ✅ 无TODO |
| notification-service | 0 | ✅ 无TODO |
| cashier-service | 0 | ✅ 无TODO |
| dispute-service | 0 | ✅ 无TODO |
| reconciliation-service | 0 | ✅ 无TODO |
| withdrawal-service | 0 | ✅ 无TODO |
| merchant-quota-service | 0 | ✅ 无TODO |

**总计**: 25 个 TODO (其中 3 个已标记为 FIXED)

---

## 详细 TODO 清单

### 1. payment-gateway (2个)

#### 1.1 国家判断逻辑
**文件**: `internal/service/payment_service.go:1383`
```go
// TODO: 根据customer_ip或其他信息判断国家
```
**优先级**: 中
**说明**: 需要实现基于IP地址的国家判断功能

#### 1.2 Webhook商户密钥获取
**文件**: `internal/service/webhook_notification_service.go:177`
```go
// TODO: 从数据库或缓存获取 merchant secret
```
**优先级**: 高
**说明**: Webhook验证需要动态获取商户密钥

---

### 2. channel-adapter (1个)

#### 2.1 预授权渠道选择
**文件**: `internal/service/channel_service.go:728`
```go
// TODO: 从数据库获取预授权记录以确定使用哪个渠道
```
**优先级**: 中
**说明**: 预授权完成需要知道原始使用的支付渠道

---

### 3. accounting-service (1个)

#### 3.1 实时汇率API调用
**文件**: `internal/service/account_service.go:1655`
```go
// TODO: 调用 channel-adapter 的汇率API
```
**优先级**: 中
**说明**: 跨币种核算需要实时汇率

---

### 4. risk-service (2个)

#### 4.1 匹配规则列表
**文件**: `internal/grpc/risk_server.go:58`
```go
// TODO: 添加匹配的规则列表
```
**优先级**: 低
**说明**: 返回触发的风控规则详情

#### 4.2 支付结果反馈机制
**文件**: `internal/grpc/risk_server.go:117`
```go
// TODO: 实现支付结果反馈机制，用于优化风控模型
```
**优先级**: 低
**说明**: 风控模型自学习优化

---

### 5. analytics-service (4个)

#### 5.1 跨商户统计
**文件**: `internal/grpc/analytics_server.go:169`
```go
// TODO: 需要在repository层实现跨商户的统计查询
```
**优先级**: 低
**说明**: 平台级别的统计分析

#### 5.2 系统健康检查
**文件**: `internal/grpc/analytics_server.go:226`
```go
// TODO: 需要实现系统健康检查逻辑
```
**优先级**: 中
**说明**: Analytics服务的健康监控

#### 5.3 报表生成
**文件**: `internal/grpc/analytics_server.go:247`
```go
// TODO: 需要实现报表生成功能
```
**优先级**: 低
**说明**: 定时报表生成功能

#### 5.4 报表存储查询
**文件**: `internal/grpc/analytics_server.go:269`
```go
// TODO: 需要实现报表存储和查询功能
```
**优先级**: 低
**说明**: 历史报表管理

---

### 6. config-service (1个)

#### 6.1 状态更新逻辑
**文件**: `internal/service/health_checker.go:144`
```go
// TODO: 实现状态更新逻辑
```
**优先级**: 中
**说明**: 健康检查状态更新

---

### 7. merchant-auth-service (1个)

#### 7.1 API Key归属验证
**文件**: `internal/service/api_key_service.go:102`
```go
// TODO: 验证key属于该merchant
```
**优先级**: 高 🔴
**说明**: **安全问题** - 需要验证API Key的所有权

---

### 8. settlement-service (4个,3个已修复 ✅)

#### 8.1 商户列表查询 ✅
**文件**: `internal/service/auto_settlement_task.go:106`
```go
// FIXED TODO #1: 从merchant-config-service查询启用自动结算的商户列表
```
**状态**: ✅ 已实现

#### 8.2 退款数据获取 ✅
**文件**: `internal/service/auto_settlement_task.go:191`
```go
// FIXED TODO #2: 从accounting service获取退款数据
```
**状态**: ✅ 已实现

#### 8.3 通知发送 ✅
**文件**: `internal/service/auto_settlement_task.go:394`
```go
// FIXED TODO #3: 实际调用notification client发送通知
```
**状态**: ✅ 已实现

#### 8.4 待结算金额计算
**文件**: `internal/grpc/settlement_server.go:329`
```go
// TODO: Calculate pending amount
```
**优先级**: 中
**说明**: 计算待结算金额

---

### 9. kyc-service (2个)

#### 9.1 邮箱验证集成
**文件**: `internal/grpc/kyc_server.go:180`
```go
// TODO: integrate with merchant service
```
**优先级**: 中
**说明**: 邮箱验证状态同步

#### 9.2 手机验证集成
**文件**: `internal/grpc/kyc_server.go:181`
```go
// TODO: integrate with merchant service
```
**优先级**: 中
**说明**: 手机验证状态同步

---

### 10. admin-bff-service (5个)

#### 10.1 商户审核逻辑
**文件**: `internal/grpc/admin_server.go:353`
```go
// TODO: 实现商户审核逻辑
```
**优先级**: 高
**说明**: 商户审核功能

#### 10.2 商户审核列表
**文件**: `internal/grpc/admin_server.go:360`
```go
// TODO: 实现商户审核列表查询
```
**优先级**: 高
**说明**: 审核队列管理

#### 10.3 审批流程创建
**文件**: `internal/grpc/admin_server.go:430`
```go
// TODO: 实现审批流程创建逻辑
```
**优先级**: 中
**说明**: 工作流引擎

#### 10.4 审批处理
**文件**: `internal/grpc/admin_server.go:436`
```go
// TODO: 实现审批处理逻辑
```
**优先级**: 中
**说明**: 审批动作执行

#### 10.5 审批列表查询
**文件**: `internal/grpc/admin_server.go:442`
```go
// TODO: 实现审批列表查询
```
**优先级**: 中
**说明**: 审批记录查询

#### 10.6 Loki日志发送
**文件**: `internal/logging/structured_logger.go:262`
```go
// TODO: 实际发送HTTP请求到Loki
```
**优先级**: 低
**说明**: 日志聚合集成

---

### 11. merchant-bff-service (1个)

#### 11.1 Loki日志发送
**文件**: `internal/logging/structured_logger.go:262`
```go
// TODO: 实际发送HTTP请求到Loki
```
**优先级**: 低
**说明**: 日志聚合集成

---

### 12. merchant-policy-service (1个)

#### 12.1 渠道策略Repository
**文件**: `cmd/main.go:110`
```go
// TODO: 下阶段实现
```
**优先级**: 低
**说明**: 渠道级别的策略配置

---

## 优先级分类

### 🔴 高优先级 (需立即处理)

1. **merchant-auth-service**: API Key归属验证 (安全问题)
2. **payment-gateway**: Webhook商户密钥获取
3. **admin-bff-service**: 商户审核逻辑和列表查询

### 🟡 中优先级 (建议处理)

1. **payment-gateway**: 国家判断逻辑
2. **channel-adapter**: 预授权渠道选择
3. **accounting-service**: 实时汇率API调用
4. **config-service**: 状态更新逻辑
5. **settlement-service**: 待结算金额计算
6. **kyc-service**: 邮箱和手机验证集成 (2个)
7. **admin-bff-service**: 审批流程相关 (3个)

### 🟢 低优先级 (可延后处理)

1. **risk-service**: 规则列表和反馈机制 (2个)
2. **analytics-service**: 跨商户统计、报表功能 (4个)
3. **admin-bff-service**: Loki日志发送
4. **merchant-bff-service**: Loki日志发送
5. **merchant-policy-service**: 渠道策略Repository

---

## 建议行动计划

### Phase 1: 安全修复 (1-2天)
- [ ] merchant-auth-service: 实现API Key归属验证
- [ ] payment-gateway: 实现Webhook商户密钥动态获取

### Phase 2: 核心功能完善 (3-5天)
- [ ] admin-bff-service: 实现商户审核流程
- [ ] payment-gateway: 实现IP地址国家判断
- [ ] channel-adapter: 实现预授权渠道记录
- [ ] accounting-service: 集成汇率API

### Phase 3: 增强功能 (5-7天)
- [ ] settlement-service: 实现待结算金额计算
- [ ] kyc-service: 集成商户验证状态
- [ ] config-service: 完善健康检查
- [ ] admin-bff-service: 实现审批工作流

### Phase 4: 可观测性优化 (按需)
- [ ] admin-bff-service: 集成Loki日志
- [ ] merchant-bff-service: 集成Loki日志
- [ ] risk-service: 实现规则匹配详情
- [ ] analytics-service: 实现报表功能

---

## 总结

✅ **完成度**: 12/19 服务 (63%) 无TODO  
⚠️ **待处理**: 7 服务包含 25 个TODO  
🔴 **安全问题**: 1 个 (merchant-auth-service)  
📊 **预计工时**: 15-20 天 (按优先级分阶段实施)

**建议**: 优先处理高优先级TODO,特别是安全相关的API Key验证问题。
