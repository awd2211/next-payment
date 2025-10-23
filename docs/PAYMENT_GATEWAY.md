# 支付网关设计文档

## 概述

Payment Gateway是整个支付平台的核心服务，负责统一支付接口、路由策略、支付状态管理、回调处理等功能。

## 核心功能

### 1. 统一支付接口

提供统一的支付创建接口，商户无需关心底层支付渠道的差异。

#### 创建支付

**API：** `POST /api/v1/payments`

**请求示例：**
```json
{
  "order_no": "ORDER20240115001",
  "amount": 10000,
  "currency": "USD",
  "customer_email": "customer@example.com",
  "customer_name": "John Doe",
  "customer_phone": "+1234567890",
  "customer_ip": "192.168.1.100",
  "description": "购买商品A x2",
  "notify_url": "https://merchant.com/callback",
  "return_url": "https://merchant.com/success",
  "expire_minutes": 30,
  "language": "en",
  "channel": "stripe",
  "pay_method": "card",
  "extra": {
    "product_id": "123",
    "user_id": "456"
  }
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "payment_no": "PY202401151234567890ABC",
    "order_no": "ORDER20240115001",
    "amount": 10000,
    "currency": "USD",
    "status": "pending",
    "channel": "stripe",
    "pay_url": "https://checkout.stripe.com/pay/cs_...",
    "expired_at": "2024-01-15T12:30:00Z",
    "created_at": "2024-01-15T12:00:00Z"
  }
}
```

#### 查询支付

**API：** `GET /api/v1/payments/{payment_no}`

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "payment_no": "PY202401151234567890ABC",
    "order_no": "ORDER20240115001",
    "merchant_id": "uuid",
    "amount": 10000,
    "currency": "USD",
    "status": "success",
    "channel": "stripe",
    "channel_order_no": "pi_xxx",
    "customer_email": "customer@example.com",
    "customer_name": "John Doe",
    "paid_at": "2024-01-15T12:05:00Z",
    "created_at": "2024-01-15T12:00:00Z"
  }
}
```

---

### 2. 多货币支持与实时汇率

#### 支持的货币（32种全球主流货币）

**主要货币：**
- USD - 美元
- EUR - 欧元
- GBP - 英镑
- CNY - 人民币
- JPY - 日元
- KRW - 韩元
- HKD - 港币
- SGD - 新加坡元

**其他支持货币：**
- AUD, CAD, INR, BRL, MXN, RUB, TRY, ZAR
- CHF, SEK, NOK, DKK, PLN, CZK, HUF, THB
- IDR, MYR, PHP, VND, AED, SAR, ILS, EGP

#### 实时汇率服务

**汇率提供商：**
1. **ExchangeRate-API.com**（默认，免费）
2. **Fixer.io**（备选，需付费）
3. **Open Exchange Rates**（可扩展）

**缓存策略：**
- Redis缓存：10分钟TTL
- 内存缓存：10分钟TTL
- 自动失效与刷新

**API示例：**

获取汇率：
```go
rate, err := exchangeRateService.GetRate(ctx, "USD", "CNY")
// 返回：7.23456
```

货币转换：
```go
convertedAmount, err := exchangeRateService.Convert(ctx, 100.0, "USD", "EUR")
// 100 USD = 92.35 EUR
```

转换分为单位的金额：
```go
convertedCents, err := exchangeRateService.ConvertAmount(ctx, 10000, "USD", "CNY")
// 10000分 (100 USD) = 72346分 (723.46 CNY)
```

清除缓存：
```go
exchangeRateService.ClearCache(ctx)
```

#### 配置示例

**.env配置：**
```bash
# 汇率API配置
EXCHANGE_RATE_PROVIDER=exchangerate-api  # or fixer
EXCHANGE_RATE_API_KEY=your_api_key_here
EXCHANGE_RATE_CACHE_TTL=600              # 缓存时间（秒）

# 默认货币
DEFAULT_CURRENCY=USD

# 是否启用实时汇率
ENABLE_REAL_TIME_EXCHANGE=true
```

---

### 3. 支付路由策略

根据金额、地区、货币、渠道状态等条件，自动选择最优支付渠道。

#### 路由规则配置

**创建路由规则：** `POST /api/v1/payment-routes`

```json
{
  "name": "Stripe - 美元小额支付",
  "priority": 10,
  "channel": "stripe",
  "is_enabled": true,
  "conditions": {
    "currencies": ["USD", "EUR", "GBP"],
    "min_amount": 100,
    "max_amount": 100000,
    "countries": ["US", "CA", "GB", "AU"],
    "pay_methods": ["card"]
  },
  "description": "使用Stripe处理美元、欧元、英镑的小额卡支付"
}
```

#### 路由匹配规则

1. **优先级匹配**：按priority从高到低匹配
2. **金额范围**：支付金额在min_amount和max_amount之间
3. **货币类型**：支持的货币列表
4. **国家地区**：根据客户IP或地址判断
5. **支付方式**：卡支付、银行转账、电子钱包等

#### 路由示例

```
优先级: 10 → Stripe (USD小额)
优先级: 9  → PayPal (欧洲地区)
优先级: 8  → Crypto (大额支付)
优先级: 5  → Adyen (亚洲地区)
优先级: 1  → Stripe (默认)
```

---

### 4. 支付状态管理

#### 状态流转图

```
pending → processing → success
                   ↓
                 failed
                   ↓
               cancelled
                   ↓
                expired
```

#### 状态说明

| 状态 | 说明 | 可执行操作 |
|-----|------|-----------|
| pending | 待支付 | 取消、支付 |
| processing | 处理中 | 查询 |
| success | 支付成功 | 退款 |
| failed | 支付失败 | 重新支付 |
| cancelled | 已取消 | 无 |
| expired | 已过期 | 重新创建 |

#### 超时处理

- 默认超时时间：30分钟
- 可配置：1-1440分钟
- 超时后自动设置为expired状态
- 定时任务每分钟检查一次

---

### 5. 回调处理

#### 异步通知（Webhook）

**流程：**
1. 渠道回调 → Payment Gateway
2. 验证签名
3. 更新支付状态
4. 记录回调日志
5. 通知商户（异步）

**商户回调格式：**
```json
{
  "event": "payment.success",
  "payment_no": "PY202401151234567890ABC",
  "order_no": "ORDER20240115001",
  "amount": 10000,
  "currency": "USD",
  "status": "success",
  "channel": "stripe",
  "channel_order_no": "pi_xxx",
  "paid_at": "2024-01-15T12:05:00Z",
  "signature": "sha256_signature"
}
```

**签名验证：**
```
signature = HMAC-SHA256(payload, merchant_secret)
```

#### 回调重试策略

| 次数 | 延迟时间 |
|-----|---------|
| 1 | 立即 |
| 2 | 1分钟后 |
| 3 | 5分钟后 |
| 4 | 15分钟后 |
| 5 | 30分钟后 |
| 6 | 1小时后 |
| 7 | 2小时后 |
| 8 | 6小时后 |

最多重试8次，超过则标记为失败。

---

### 6. 退款处理

#### 创建退款

**API：** `POST /api/v1/refunds`

```json
{
  "payment_no": "PY202401151234567890ABC",
  "amount": 5000,
  "reason": "商品质量问题",
  "description": "客户要求部分退款",
  "operator_type": "merchant"
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "refund_no": "RF202401151234567890XYZ",
    "payment_no": "PY202401151234567890ABC",
    "amount": 5000,
    "currency": "USD",
    "status": "pending",
    "reason": "商品质量问题",
    "created_at": "2024-01-15T14:00:00Z"
  }
}
```

#### 退款限制

- 只能退款成功的支付
- 累计退款金额不能超过支付金额
- 支持部分退款
- 支持多次退款

#### 退款状态

- `pending` - 待处理
- `processing` - 处理中
- `success` - 退款成功
- `failed` - 退款失败

---

### 7. 数据模型

#### payments（支付记录表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| merchant_id | UUID | 商户ID |
| order_no | VARCHAR(64) | 商户订单号 |
| payment_no | VARCHAR(64) | 平台支付流水号 |
| channel | VARCHAR(50) | 支付渠道 |
| channel_order_no | VARCHAR(128) | 渠道订单号 |
| amount | BIGINT | 金额（分） |
| currency | VARCHAR(10) | 货币类型 |
| status | VARCHAR(20) | 状态 |
| pay_method | VARCHAR(50) | 支付方式 |
| customer_email | VARCHAR(255) | 客户邮箱 |
| customer_name | VARCHAR(100) | 客户姓名 |
| customer_phone | VARCHAR(20) | 客户手机 |
| customer_ip | VARCHAR(50) | 客户IP |
| description | TEXT | 商品描述 |
| notify_url | VARCHAR(500) | 异步通知URL |
| return_url | VARCHAR(500) | 同步跳转URL |
| extra | JSONB | 扩展信息 |
| error_code | VARCHAR(50) | 错误码 |
| error_msg | TEXT | 错误信息 |
| notify_status | VARCHAR(20) | 通知状态 |
| notify_times | INTEGER | 通知次数 |
| last_notify_at | TIMESTAMPTZ | 最后通知时间 |
| paid_at | TIMESTAMPTZ | 支付完成时间 |
| expired_at | TIMESTAMPTZ | 过期时间 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

**索引：**
- order_no（唯一）
- payment_no（唯一）
- merchant_id
- status
- channel
- created_at

#### refunds（退款记录表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| payment_id | UUID | 关联支付ID |
| merchant_id | UUID | 商户ID |
| refund_no | VARCHAR(64) | 退款单号 |
| channel_refund_no | VARCHAR(128) | 渠道退款单号 |
| amount | BIGINT | 退款金额 |
| currency | VARCHAR(10) | 货币类型 |
| status | VARCHAR(20) | 状态 |
| reason | VARCHAR(200) | 退款原因 |
| description | TEXT | 退款说明 |
| operator_id | UUID | 操作人ID |
| operator_type | VARCHAR(20) | 操作人类型 |
| error_code | VARCHAR(50) | 错误码 |
| error_msg | TEXT | 错误信息 |
| refunded_at | TIMESTAMPTZ | 退款完成时间 |
| created_at | TIMESTAMPTZ | 创建时间 |

#### payment_callbacks（回调记录表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| payment_id | UUID | 支付ID |
| channel | VARCHAR(50) | 支付渠道 |
| event | VARCHAR(50) | 事件类型 |
| raw_data | TEXT | 原始回调数据 |
| signature | VARCHAR(500) | 签名 |
| is_verified | BOOLEAN | 是否验证通过 |
| is_processed | BOOLEAN | 是否已处理 |
| error_msg | TEXT | 错误信息 |
| created_at | TIMESTAMPTZ | 创建时间 |

#### payment_routes（路由规则表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 规则名称 |
| priority | INTEGER | 优先级 |
| channel | VARCHAR(50) | 目标渠道 |
| conditions | JSONB | 路由条件 |
| is_enabled | BOOLEAN | 是否启用 |
| description | TEXT | 规则描述 |
| created_at | TIMESTAMPTZ | 创建时间 |

---

## API完整列表

### 支付管理
- `POST /api/v1/payments` - 创建支付
- `GET /api/v1/payments/{payment_no}` - 查询支付
- `GET /api/v1/payments` - 支付列表
- `POST /api/v1/payments/{payment_no}/cancel` - 取消支付

### 退款管理
- `POST /api/v1/refunds` - 创建退款
- `GET /api/v1/refunds/{refund_no}` - 查询退款
- `GET /api/v1/refunds` - 退款列表

### 回调处理
- `POST /api/v1/callbacks/stripe` - Stripe回调
- `POST /api/v1/callbacks/paypal` - PayPal回调
- `POST /api/v1/callbacks/crypto` - 加密货币回调

### 路由管理
- `POST /api/v1/payment-routes` - 创建路由规则
- `GET /api/v1/payment-routes` - 路由规则列表
- `PUT /api/v1/payment-routes/{id}` - 更新路由规则
- `DELETE /api/v1/payment-routes/{id}` - 删除路由规则

### 汇率管理
- `GET /api/v1/exchange-rates/{from}/{to}` - 获取汇率
- `GET /api/v1/exchange-rates/{base}` - 获取所有汇率
- `POST /api/v1/exchange-rates/convert` - 货币转换
- `POST /api/v1/exchange-rates/clear-cache` - 清除汇率缓存

---

## 安全机制

### 1. API签名验证

**请求签名：**
```
signature = MD5(api_key + timestamp + nonce + request_body + api_secret)
```

**Header：**
```
X-API-Key: pk_test_xxxxx
X-Timestamp: 1642234567
X-Nonce: random_string
X-Signature: md5_signature
```

### 2. IP白名单

- 支持商户配置IP白名单
- 只允许白名单IP调用API
- 默认禁用，需商户手动开启

### 3. 频率限制

| 级别 | 限制 |
|-----|------|
| 创建支付 | 100次/分钟/商户 |
| 查询支付 | 300次/分钟/商户 |
| 创建退款 | 30次/分钟/商户 |
| 查询汇率 | 100次/分钟/商户 |

### 4. 金额限制

| 类型 | 限制 |
|-----|------|
| 单笔最小金额 | 1分 |
| 单笔最大金额 | 1000000分（10000元）|
| 每日累计限额 | 可配置 |
| 单次退款限额 | 不超过原支付金额 |

---

## 技术栈

- **语言**：Go 1.21+
- **框架**：Gin（HTTP）
- **数据库**：PostgreSQL
- **缓存**：Redis
- **消息队列**：Kafka
- **汇率API**：ExchangeRate-API.com / Fixer.io

---

## 未来扩展

### 短期
- [ ] 更多支付渠道集成（Adyen, Square）
- [ ] 分账功能
- [ ] 预授权支付
- [ ] 定期扣款

### 中期
- [ ] 智能路由优化（机器学习）
- [ ] 多维度风控规则
- [ ] 实时数据分析
- [ ] A/B测试

### 长期
- [ ] 跨境支付优化
- [ ] 区块链支付
- [ ] 数字货币钱包
- [ ] 全球化部署
