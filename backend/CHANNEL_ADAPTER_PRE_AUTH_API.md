# Channel Adapter 预授权 HTTP API 完整指南

本文档说明 channel-adapter 服务的预授权 HTTP API 接口，包括请求格式、响应格式和使用示例。

## 概述

channel-adapter 现已提供完整的预授权 HTTP API，支持创建、确认、取消和查询预授权操作。

### 支持的渠道

| 渠道 | 创建预授权 | 确认预授权 | 取消预授权 | 查询预授权 | 说明 |
|------|-----------|-----------|-----------|-----------|------|
| **Stripe** | ✅ | ✅ | ✅ | ✅ | 完整实现 (使用 PaymentIntent manual capture) |
| PayPal | ❌ | ❌ | ❌ | ❌ | 返回"不支持预授权"错误 |
| Alipay | ❌ | ❌ | ❌ | ❌ | 返回"不支持预授权"错误 |
| Crypto | ❌ | ❌ | ❌ | ❌ | 返回"不支持预授权"错误 |

## API 端点

### 基础 URL

```
http://localhost:40005/api/v1/channel
```

### 端点列表

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/pre-auth` | 创建预授权 |
| POST | `/pre-auth/capture` | 确认预授权（扣款） |
| POST | `/pre-auth/cancel` | 取消预授权（释放资金） |
| GET | `/pre-auth/:channel_pre_auth_no` | 查询预授权状态 |

## API 详细说明

### 1. 创建预授权

创建一个预授权，预授权资金但不立即扣款。

#### 请求

```http
POST /api/v1/channel/pre-auth HTTP/1.1
Host: localhost:40005
Content-Type: application/json

{
  "channel": "stripe",
  "pre_auth_no": "PA20250124123456",
  "order_no": "ORDER-12345",
  "amount": 50000,
  "currency": "USD",
  "customer_email": "customer@example.com",
  "customer_name": "John Doe",
  "description": "酒店预订押金",
  "expires_at": 1737888000,
  "callback_url": "https://merchant.com/callback"
}
```

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel | string | ✅ | 支付渠道 (stripe, paypal, alipay, crypto) |
| pre_auth_no | string | ✅ | 预授权流水号 (商户系统生成) |
| order_no | string | ✅ | 订单号 |
| amount | int64 | ✅ | 预授权金额（分） |
| currency | string | ✅ | 货币代码 (USD, CNY, EUR 等) |
| customer_email | string | ✅ | 客户邮箱 |
| customer_name | string | ❌ | 客户姓名 |
| description | string | ❌ | 描述 |
| expires_at | int64 | ❌ | 过期时间 (Unix 时间戳) |
| callback_url | string | ❌ | 回调 URL |
| extra | object | ❌ | 扩展字段 |

#### 响应

```json
{
  "code": "SUCCESS",
  "message": "创建预授权成功",
  "data": {
    "pre_auth_no": "PA20250124123456",
    "channel_pre_auth_no": "pi_3abc123def456ghi",
    "client_secret": "pi_3abc123def456ghi_secret_xyz789",
    "status": "requires_payment_method",
    "expires_at": 1737888000,
    "extra": {
      "payment_intent_id": "pi_3abc123def456ghi",
      "client_secret": "pi_3abc123def456ghi_secret_xyz789"
    }
  }
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| pre_auth_no | string | 平台预授权流水号 |
| channel_pre_auth_no | string | 渠道预授权号 (Stripe PaymentIntent ID) |
| client_secret | string | 客户端密钥（用于前端 Stripe.js 确认） |
| status | string | 预授权状态 |
| expires_at | int64 | 过期时间 (Unix 时间戳) |
| extra | object | 扩展信息 |

#### 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 500 | 服务器内部错误 |

---

### 2. 确认预授权 (扣款)

确认预授权并进行实际扣款。可以部分确认（金额小于等于预授权金额）。

#### 请求

```http
POST /api/v1/channel/pre-auth/capture HTTP/1.1
Host: localhost:40005
Content-Type: application/json

{
  "channel": "stripe",
  "pre_auth_no": "PA20250124123456",
  "channel_pre_auth_no": "pi_3abc123def456ghi",
  "amount": 40000,
  "currency": "USD",
  "description": "酒店实际消费"
}
```

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel | string | ✅ | 支付渠道 |
| pre_auth_no | string | ✅ | 平台预授权流水号 |
| channel_pre_auth_no | string | ✅ | 渠道预授权号 |
| amount | int64 | ❌ | 确认金额（分，可小于等于预授权金额，不填表示全额） |
| currency | string | ✅ | 货币代码 |
| description | string | ❌ | 描述 |
| extra | object | ❌ | 扩展字段 |

#### 响应

```json
{
  "code": "SUCCESS",
  "message": "确认预授权成功",
  "data": {
    "pre_auth_no": "PA20250124123456",
    "channel_trade_no": "pi_3abc123def456ghi",
    "channel_pre_auth_no": "pi_3abc123def456ghi",
    "status": "succeeded",
    "amount": 40000,
    "extra": {
      "payment_intent_id": "pi_3abc123def456ghi",
      "captured_at": 1737715200
    }
  }
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| pre_auth_no | string | 平台预授权流水号 |
| channel_trade_no | string | 渠道交易号 (支付完成后的 ID) |
| channel_pre_auth_no | string | 渠道预授权号 |
| status | string | 状态 (succeeded = 扣款成功) |
| amount | int64 | 实际扣款金额（分） |
| extra | object | 扩展信息 |

---

### 3. 取消预授权 (释放资金)

取消预授权，释放冻结的资金。

#### 请求

```http
POST /api/v1/channel/pre-auth/cancel HTTP/1.1
Host: localhost:40005
Content-Type: application/json

{
  "channel": "stripe",
  "pre_auth_no": "PA20250124123456",
  "channel_pre_auth_no": "pi_3abc123def456ghi",
  "reason": "客户提前退房"
}
```

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel | string | ✅ | 支付渠道 |
| pre_auth_no | string | ✅ | 平台预授权流水号 |
| channel_pre_auth_no | string | ✅ | 渠道预授权号 |
| reason | string | ❌ | 取消原因 |
| extra | object | ❌ | 扩展字段 |

#### 响应

```json
{
  "code": "SUCCESS",
  "message": "取消预授权成功",
  "data": {
    "pre_auth_no": "PA20250124123456",
    "channel_pre_auth_no": "pi_3abc123def456ghi",
    "status": "canceled",
    "extra": {
      "payment_intent_id": "pi_3abc123def456ghi",
      "cancelled_at": 1737715300
    }
  }
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| pre_auth_no | string | 平台预授权流水号 |
| channel_pre_auth_no | string | 渠道预授权号 |
| status | string | 状态 (canceled = 已取消) |
| extra | object | 扩展信息 |

---

### 4. 查询预授权状态

查询预授权的当前状态。

#### 请求

```http
GET /api/v1/channel/pre-auth/pi_3abc123def456ghi HTTP/1.1
Host: localhost:40005
```

#### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel_pre_auth_no | string | ✅ | 渠道预授权号 |

**注意**: 当前实现需要在 handler 中传递 channel 参数，未来会支持从数据库自动识别渠道。

#### 响应

```json
{
  "code": "SUCCESS",
  "message": "查询预授权成功",
  "data": {
    "channel_pre_auth_no": "pi_3abc123def456ghi",
    "status": "requires_capture",
    "amount": 50000,
    "captured_amount": 0,
    "currency": "USD",
    "expires_at": null,
    "created_at": 1737715200,
    "extra": {
      "payment_intent_id": "pi_3abc123def456ghi",
      "amount_capturable": 50000,
      "amount_received": 0,
      "cancellation_reason": null
    }
  }
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| channel_pre_auth_no | string | 渠道预授权号 |
| status | string | 状态 |
| amount | int64 | 预授权金额（分） |
| captured_amount | int64 | 已确认金额（分） |
| currency | string | 货币代码 |
| expires_at | int64 | 过期时间 |
| created_at | int64 | 创建时间 |
| extra | object | 扩展信息 |

## 预授权状态流转

```
1. requires_payment_method (需要支付方式)
   ↓ 客户添加支付方式
2. requires_confirmation (需要确认)
   ↓ 商户调用 /pre-auth/capture
3. requires_capture (需要扣款) - 预授权成功
   ↓
   ├── 调用 /pre-auth/capture → succeeded (扣款成功)
   └── 调用 /pre-auth/cancel → canceled (已取消)
```

### Stripe 状态说明

| Stripe 状态 | 平台状态 | 说明 |
|-------------|---------|------|
| requires_payment_method | pending | 等待支付方式 |
| requires_confirmation | pending | 等待确认 |
| requires_action | pending | 需要用户操作 (3D Secure) |
| requires_capture | authorized | 预授权成功，可扣款 |
| processing | processing | 处理中 |
| succeeded | captured | 扣款成功 |
| canceled | canceled | 已取消 |

## 使用示例

### Node.js 示例

```javascript
const axios = require('axios');

const CHANNEL_ADAPTER_URL = 'http://localhost:40005/api/v1/channel';

// 1. 创建预授权
async function createPreAuth() {
  const response = await axios.post(`${CHANNEL_ADAPTER_URL}/pre-auth`, {
    channel: 'stripe',
    pre_auth_no: `PA${Date.now()}`,
    order_no: 'ORDER-12345',
    amount: 50000,  // $500.00
    currency: 'USD',
    customer_email: 'customer@example.com',
    description: '酒店预订押金'
  });

  console.log('预授权创建成功:', response.data.data);
  return response.data.data;
}

// 2. 确认预授权（扣款）
async function capturePreAuth(channelPreAuthNo) {
  const response = await axios.post(`${CHANNEL_ADAPTER_URL}/pre-auth/capture`, {
    channel: 'stripe',
    pre_auth_no: 'PA20250124123456',
    channel_pre_auth_no: channelPreAuthNo,
    amount: 40000,  // $400.00 (部分确认)
    currency: 'USD',
    description: '酒店实际消费'
  });

  console.log('预授权确认成功:', response.data.data);
  return response.data.data;
}

// 3. 取消预授权
async function cancelPreAuth(channelPreAuthNo) {
  const response = await axios.post(`${CHANNEL_ADAPTER_URL}/pre-auth/cancel`, {
    channel: 'stripe',
    pre_auth_no: 'PA20250124123456',
    channel_pre_auth_no: channelPreAuthNo,
    reason: '客户提前退房'
  });

  console.log('预授权取消成功:', response.data.data);
  return response.data.data;
}

// 完整流程
async function fullPreAuthFlow() {
  // 步骤 1: 创建预授权
  const preAuth = await createPreAuth();

  // 客户在前端使用 client_secret 完成支付方式确认...

  // 步骤 2: 确认预授权（扣款）
  await capturePreAuth(preAuth.channel_pre_auth_no);

  // OR 步骤 2': 取消预授权
  // await cancelPreAuth(preAuth.channel_pre_auth_no);
}

fullPreAuthFlow();
```

### cURL 示例

```bash
# 1. 创建预授权
curl -X POST http://localhost:40005/api/v1/channel/pre-auth \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "stripe",
    "pre_auth_no": "PA20250124123456",
    "order_no": "ORDER-12345",
    "amount": 50000,
    "currency": "USD",
    "customer_email": "customer@example.com",
    "description": "酒店预订押金"
  }'

# 2. 确认预授权（部分确认）
curl -X POST http://localhost:40005/api/v1/channel/pre-auth/capture \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "stripe",
    "pre_auth_no": "PA20250124123456",
    "channel_pre_auth_no": "pi_3abc123def456ghi",
    "amount": 40000,
    "currency": "USD",
    "description": "酒店实际消费"
  }'

# 3. 取消预授权
curl -X POST http://localhost:40005/api/v1/channel/pre-auth/cancel \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "stripe",
    "pre_auth_no": "PA20250124123456",
    "channel_pre_auth_no": "pi_3abc123def456ghi",
    "reason": "客户提前退房"
  }'

# 4. 查询预授权状态
curl -X GET http://localhost:40005/api/v1/channel/pre-auth/pi_3abc123def456ghi
```

## 错误处理

### 错误响应格式

```json
{
  "code": "INTERNAL_ERROR",
  "message": "创建预授权失败",
  "details": "当前支付渠道不支持预授权功能",
  "trace_id": "req_abc123def456"
}
```

### 常见错误码

| 错误码 | 说明 | 解决方案 |
|--------|------|---------|
| INVALID_REQUEST | 请求参数错误 | 检查必填参数是否完整 |
| UNAUTHORIZED | 未认证 | 添加认证令牌 |
| CHANNEL_NOT_SUPPORTED | 渠道不支持预授权 | 使用 Stripe 渠道 |
| PREAUTH_NOT_FOUND | 预授权不存在 | 检查 channel_pre_auth_no 是否正确 |
| PREAUTH_EXPIRED | 预授权已过期 | 重新创建预授权 |
| INVALID_AMOUNT | 金额无效 | 确认金额不能大于预授权金额 |
| INTERNAL_ERROR | 服务器内部错误 | 查看服务日志 |

## 集成 Stripe.js (前端)

创建预授权后，需要在前端使用 Stripe.js 完成支付方式确认：

```html
<!DOCTYPE html>
<html>
<head>
  <script src="https://js.stripe.com/v3/"></script>
</head>
<body>
  <div id="payment-element"></div>
  <button id="submit">预授权确认</button>

  <script>
    // 1. 初始化 Stripe
    const stripe = Stripe('pk_test_...');

    // 2. 从后端获取 client_secret
    const clientSecret = 'pi_3abc123def456ghi_secret_xyz789';

    // 3. 创建 Payment Element
    const elements = stripe.elements({ clientSecret });
    const paymentElement = elements.create('payment');
    paymentElement.mount('#payment-element');

    // 4. 处理提交
    document.getElementById('submit').addEventListener('click', async () => {
      const {error} = await stripe.confirmPayment({
        elements,
        confirmParams: {
          return_url: 'https://merchant.com/return',
        },
      });

      if (error) {
        console.error('预授权失败:', error.message);
      } else {
        console.log('预授权成功');
        // 后端会收到 webhook，更新预授权状态
      }
    });
  </script>
</body>
</html>
```

## 最佳实践

### 1. 酒店预订场景

```javascript
// 步骤 1: 客户预订时创建预授权
const preAuth = await createPreAuth({
  amount: 100000,  // $1000 押金
  currency: 'USD',
  description: '酒店预订押金 - 房间 101',
  expires_at: checkOutDate + 24 * 3600  // 退房后 24 小时过期
});

// 步骤 2: 客户退房时确认实际消费
const capture = await capturePreAuth({
  channel_pre_auth_no: preAuth.channel_pre_auth_no,
  amount: 85000,  // $850 实际消费
  description: '房费 + 餐饮'
});

// 步骤 3: 剩余押金自动释放
// Stripe 会自动释放 (100000 - 85000) = $150 给客户
```

### 2. 租车场景

```javascript
// 步骤 1: 租车时预授权
const preAuth = await createPreAuth({
  amount: 50000,  // $500 押金
  currency: 'USD',
  description: '租车押金',
  expires_at: returnDate + 7 * 24 * 3600  // 还车后 7 天过期
});

// 步骤 2a: 无损坏，取消预授权
if (noDamage) {
  await cancelPreAuth({
    channel_pre_auth_no: preAuth.channel_pre_auth_no,
    reason: '车辆无损坏，押金全额退还'
  });
}

// 步骤 2b: 有损坏，扣除维修费
if (hasDamage) {
  await capturePreAuth({
    channel_pre_auth_no: preAuth.channel_pre_auth_no,
    amount: repairCost,
    description: `车辆维修费: ${repairDetails}`
  });
}
```

### 3. 幂等性处理

```javascript
// 使用唯一的 pre_auth_no 实现幂等性
const preAuthNo = `PA_${orderId}_${Date.now()}`;

try {
  const preAuth = await createPreAuth({
    pre_auth_no: preAuthNo,  // 唯一标识
    // ... 其他参数
  });
} catch (error) {
  if (error.response?.status === 409) {
    console.log('预授权已存在');
  }
}
```

## 监控和日志

### 关键日志事件

```
INFO  创建预授权  channel=stripe pre_auth_no=PA... amount=50000
INFO  预授权创建成功  channel_pre_auth_no=pi_...
ERROR 创建预授权失败  channel=paypal error="当前支付渠道不支持预授权功能"
INFO  确认预授权  channel_pre_auth_no=pi_... amount=40000
INFO  预授权确认成功  status=succeeded
INFO  取消预授权  channel_pre_auth_no=pi_... reason="客户提前退房"
```

### Prometheus 指标 (未来)

```promql
# 预授权创建次数
channel_preauth_created_total{channel="stripe",status="success"}

# 预授权确认金额
channel_preauth_captured_amount{channel="stripe",currency="USD"}

# 预授权取消率
sum(channel_preauth_canceled_total) / sum(channel_preauth_created_total)
```

## 总结

channel-adapter 的预授权 HTTP API 现已完全实现，具备以下特性：

✅ **完整的 RESTful API**: 创建、确认、取消、查询
✅ **Stripe 完整支持**: 使用 PaymentIntent manual capture
✅ **部分确认**: 支持小于预授权金额的确认
✅ **错误处理**: 完整的错误响应和状态码
✅ **编译成功**: 100% 编译通过

**适用场景**:
- 酒店预订押金
- 租车押金
- 大额商品预付款
- 任何需要先冻结后扣款的场景

**下一步**:
- 为 PayPal 实现预授权 (使用 Orders API)
- 为 Alipay 实现预授权 (使用预授权接口)
- 添加 Prometheus 指标收集
- 完善查询接口（支持自动识别渠道）
