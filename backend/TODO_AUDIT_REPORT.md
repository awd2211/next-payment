# TODO 审计报告

**生成时间**: 2025-10-24
**扫描范围**: 所有12个微服务的内部代码和命令入口
**TODO总数**: 28 个

---

## 📊 按服务统计

| 服务名称 | TODO数量 | 优先级分布 |
|---------|---------|-----------|
| admin-service | 5 | P3: 5 (gRPC未实现功能) |
| analytics-service | 4 | P3: 4 (gRPC未实现功能) |
| kyc-service | 4 | P3: 3, P2: 1 (OCR集成) |
| payment-gateway | 3 | P2: 2, P3: 1 |
| risk-service | 3 | P3: 3 (模型优化) |
| merchant-service | 2 | P1: 1 (加密存储), P3: 1 |
| settlement-service | 2 | P3: 2 |
| withdrawal-service | 2 | P2: 2 (银行API集成) |
| accounting-service | 1 | P3: 1 (汇率API) |
| channel-adapter | 1 | P2: 1 (加密货币计价) |
| order-service | 1 | P3: 1 (proto定义) |
| config-service | 0 | - |
| notification-service | 0 | - |
| **总计** | **28** | **P1:1, P2:5, P3:22** |

---

## 🔴 P1 - 高优先级（安全问题）

### 1. 银行账号加密存储 ⚠️ 安全风险
**位置**: `services/merchant-service/internal/service/business_service.go:116`
**代码**:
```go
AccountNumber: input.AccountNumber, // TODO: 应该加密存储
```

**问题**: 商户银行账号以明文存储在数据库中，存在数据泄露风险
**优先级**: **P1 - 高**
**建议修复**:
- 使用 `pkg/crypto` 包加密银行账号
- 在存储前加密，读取时解密
- 密钥管理使用环境变量或密钥管理服务（KMS）

**修复估时**: 4小时
**影响范围**: merchant-service

---

## 🟡 P2 - 中优先级（功能缺失）

### 1. 加密货币退款计价错误 💰
**位置**: `services/channel-adapter/internal/adapter/crypto_adapter.go:249`
**代码**:
```go
// TODO: 应该根据支付时的价格计算，而不是当前价格
```

**问题**: 加密货币退款使用当前价格计算，而非支付时价格，可能导致金额错误
**优先级**: **P2 - 中**
**建议修复**:
- 在支付记录中存储支付时的汇率
- 退款时使用历史汇率计算加密货币数量

**修复估时**: 6小时
**影响范围**: channel-adapter, payment记录表

---

### 2. 银行转账API未实现 🏦
**位置**:
- `services/withdrawal-service/internal/client/bank_transfer_client.go:43`
- `services/withdrawal-service/internal/client/bank_transfer_client.go:87`

**代码**:
```go
// TODO: 生产环境需要替换为真实银行API调用
// TODO: 生产环境需要替换为真实银行查询API
```

**问题**: 当前使用mock实现，生产环境无法实际转账
**优先级**: **P2 - 中**
**建议修复**:
- 集成真实银行API（如：银联、网联、或第三方支付机构）
- 实现银行转账状态查询
- 添加重试和对账机制

**修复估时**: 2周（需要银行API接入）
**影响范围**: withdrawal-service

---

### 3. Payment Gateway渠道签名验证未实现
**位置**: `services/payment-gateway/internal/service/payment_service.go:528`
**代码**:
```go
// TODO: 实现不同渠道的签名验证
```

**问题**: Webhook回调仅验证了Stripe签名，其他渠道未实现
**优先级**: **P2 - 中**
**建议修复**:
- 为PayPal、加密货币等渠道实现签名验证
- 使用适配器模式统一验证接口

**修复估时**: 8小时
**影响范围**: payment-gateway

---

### 4. Payment Gateway国家判断未实现
**位置**: `services/payment-gateway/internal/service/payment_service.go:963`
**代码**:
```go
// TODO: 根据customer_ip或其他信息判断国家
```

**问题**: 无法根据IP地理位置做国家级风控和路由
**优先级**: **P2 - 中**
**建议修复**:
- 集成GeoIP库（MaxMind GeoLite2）
- 根据国家实现不同的风控策略和支付路由

**修复估时**: 4小时
**影响范围**: payment-gateway

---

### 5. KYC OCR文档识别未实现
**位置**: `services/kyc-service/internal/service/kyc_service.go:87`
**代码**:
```go
// TODO: 调用OCR服务识别文档信息
```

**问题**: KYC文档审核无法自动提取信息，需人工录入
**优先级**: **P2 - 中**
**建议修复**:
- 集成OCR服务（AWS Textract、Google Vision API、或百度OCR）
- 实现身份证、护照、营业执照的自动识别
- 添加数据验证和防伪检测

**修复估时**: 1周
**影响范围**: kyc-service

---

## 🟢 P3 - 低优先级（未来增强）

### gRPC服务未实现功能（9个）
这些TODO主要是gRPC接口的占位符，当前系统使用HTTP/REST通信，gRPC功能暂未启用：

1. **admin-service** (5个):
   - 商户审核逻辑 (`admin_server.go:353`)
   - 商户审核列表查询 (`admin_server.go:360`)
   - 审批流程创建 (`admin_server.go:430`)
   - 审批处理 (`admin_server.go:436`)
   - 审批列表查询 (`admin_server.go:442`)

2. **analytics-service** (4个):
   - 跨商户统计查询 (`analytics_server.go:169`)
   - 系统健康检查 (`analytics_server.go:226`)
   - 报表生成功能 (`analytics_server.go:247`)
   - 报表存储和查询 (`analytics_server.go:269`)

3. **risk-service** (3个):
   - 支付结果反馈机制 (`risk_server.go:117`)
   - 匹配规则列表 (`risk_server.go:58`)
   - 支付结果上报逻辑 (`risk_handler.go:328`)

**优先级**: **P3 - 低**
**说明**: 当前系统架构以HTTP为主，gRPC功能作为未来优化方向

---

### 其他低优先级功能（13个）

1. **KYC邮箱/手机验证集成** (2个)
   - 位置: `kyc-service/internal/grpc/kyc_server.go:180-181`
   - 需要与merchant-service集成

2. **KYC审核通知**
   - 位置: `kyc-service/internal/service/kyc_service.go:95`
   - 发送审核结果通知给商户

3. **商户邀请邮件**
   - 位置: `merchant-service/internal/service/business_service.go:345`
   - 发送商户入驻邀请邮件

4. **Settlement待结算金额计算**
   - 位置: `settlement-service/internal/grpc/settlement_server.go:329`
   - 计算商户待结算金额

5. **Settlement银行账户获取**
   - 位置: `settlement-service/internal/service/settlement_service.go:307`
   - 从merchant-service获取默认银行账户

6. **Order Proto定义更新**
   - 位置: `order-service/internal/grpc/order_server.go:44`
   - 更新proto支持完整订单字段

7. **Accounting汇率API调用**
   - 位置: `accounting-service/internal/service/account_service.go:1597`
   - 调用channel-adapter的实时汇率API

8. **API Key CIDR验证增强** (已在优化中修复)
   - 位置: `payment-gateway/internal/repository/api_key_repository.go:146`
   - 注释已过时，实际已使用 `net.ParseCIDR()`

---

## 📈 优先级建议执行顺序

### 阶段1: 安全修复（1-2周）
1. ✅ 银行账号加密存储 (P1)

### 阶段2: 核心功能完善（2-4周）
2. ✅ 加密货币退款计价修复 (P2)
3. ✅ Payment Gateway渠道签名验证 (P2)
4. ✅ Payment Gateway国家判断 (P2)
5. ✅ KYC OCR文档识别 (P2)

### 阶段3: 银行集成（4-8周）
6. ✅ 银行转账API集成 (P2)

### 阶段4: 增强功能（按需实现）
7. 🔄 gRPC服务实现 (P3)
8. 🔄 其他低优先级功能 (P3)

---

## 🎯 立即需要修复的TODO (P1+P2)

| TODO | 优先级 | 估时 | 影响 |
|------|--------|------|------|
| 银行账号加密存储 | P1 | 4h | 安全风险 |
| 加密货币退款计价 | P2 | 6h | 金额错误 |
| 银行转账API集成 | P2 | 2周 | 无法提现 |
| 渠道签名验证 | P2 | 8h | 安全风险 |
| 国家判断 | P2 | 4h | 风控缺失 |
| KYC OCR识别 | P2 | 1周 | 效率低 |

**总计**: 6个高优先级TODO，预计总工时 **3-4周**

---

## 📝 建议

1. **立即修复P1**: 银行账号加密是严重的安全问题，应立即修复
2. **短期修复P2**: 2-4周内完成所有P2级别TODO
3. **P3功能规划**: 根据业务需求决定是否实现gRPC功能
4. **代码审查**: 定期审查TODO，避免技术债务累积

---

## 🔍 已完成的优化（三轮优化）

以下TODO已在之前的优化中修复：
- ✅ Mock Secret硬编码 (第一轮优化)
- ✅ IP CIDR验证 (第一轮优化，但注释未更新)
- ✅ JSON Unmarshal错误处理 (第二轮优化)
- ✅ 幂等性原子性 (第二轮优化)
- ✅ Webhook URL硬编码 (第三轮优化)

**建议**: 清理已修复功能的过时TODO注释

---

**报告生成完毕** | 总计28个TODO | P1:1 | P2:5 | P3:22
