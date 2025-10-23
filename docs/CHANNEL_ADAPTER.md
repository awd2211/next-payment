# 渠道适配服务文档 (Channel Adapter Service)

## 📋 目录

- [服务概述](#服务概述)
- [核心功能](#核心功能)
- [技术架构](#技术架构)
- [Stripe 适配器](#stripe-适配器)
- [数据模型](#数据模型)
- [API 接口](#api-接口)
- [Webhook 处理](#webhook-处理)
- [部署配置](#部署配置)

---

## 服务概述

**Channel Adapter Service** 是支付渠道适配服务，负责对接各种第三方支付渠道（Stripe、PayPal、加密货币等），提供统一的支付接口。

### 主要职责

- 🔌 **渠道对接** - 适配不同支付渠道的 API
- 💳 **支付处理** - 创建支付、查询状态、取消支付
- 💰 **退款管理** - 创建退款、查询退款状态
- 🔔 **Webhook 处理** - 接收并处理支付渠道的回调通知
- ⚙️ **配置管理** - 管理各渠道的配置信息（API密钥、费率等）
- 📊 **交易记录** - 记录所有渠道交易详情

### 技术栈

- **语言**: Go 1.21
- **HTTP 框架**: Gin
- **数据库**: PostgreSQL + GORM
- **支付SDK**: stripe-go v76
- **端口**: 8003

---

## 核心功能

### 1. 支付操作

#### 创建支付
```go
// 请求
type CreatePaymentRequest struct {
    MerchantID    uuid.UUID              // 商户ID
    Channel       string                 // 渠道：stripe, paypal, crypto
    PaymentNo     string                 // 支付流水号
    OrderNo       string                 // 订单号
    Amount        int64                  // 金额（分）
    Currency      string                 // 货币
    CustomerEmail string                 // 客户邮箱
    CustomerName  string                 // 客户姓名
    Description   string                 // 描述
    SuccessURL    string                 // 成功跳转URL
    CancelURL     string                 // 取消跳转URL
    CallbackURL   string                 // 回调URL
}

// 响应
type CreatePaymentResponse struct {
    PaymentNo      string  // 支付流水号
    ChannelTradeNo string  // 渠道交易号
    ClientSecret   string  // 客户端密钥（给前端使用）
    PaymentURL     string  // 支付URL（重定向方式）
    QRCodeURL      string  // 二维码URL
    Status         string  // 状态
}
```

#### 查询支付
```go
type QueryPaymentResponse struct {
    PaymentNo            string                 // 支付流水号
    ChannelTradeNo       string                 // 渠道交易号
    Status               string                 // 状态
    Amount               int64                  // 金额（分）
    Currency             string                 // 货币
    PaymentMethod        string                 // 支付方式
    PaymentMethodDetails map[string]interface{} // 支付方式详情
    PaidAt               *time.Time             // 支付时间
}
```

### 2. 退款操作

#### 创建退款
```go
type CreateRefundRequest struct {
    MerchantID uuid.UUID // 商户ID
    RefundNo   string    // 退款流水号
    PaymentNo  string    // 原支付流水号
    Amount     int64     // 退款金额（分）
    Currency   string    // 货币
    Reason     string    // 退款原因
}

type CreateRefundResponse struct {
    RefundNo        string // 退款流水号
    ChannelRefundNo string // 渠道退款号
    Status          string // 状态
}
```

### 3. Webhook 处理

支持处理各渠道的 Webhook 回调，包括：

- **支付成功** - payment.success
- **支付失败** - payment.failed
- **支付取消** - payment.cancelled
- **退款成功** - refund.success
- **退款失败** - refund.failed

---

## 技术架构

### 分层架构

```
┌─────────────────────────────────────────┐
│            HTTP Handler Layer            │
│        (Gin Routes & Controllers)        │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│            Service Layer                 │
│      (Business Logic & Orchestration)    │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│           Adapter Layer                  │
│  (Stripe, PayPal, Crypto Adapters)      │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Repository Layer                 │
│       (Database Operations - GORM)       │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│            PostgreSQL                    │
└─────────────────────────────────────────┘
```

### 适配器模式

使用适配器模式统一不同支付渠道的接口：

```go
// PaymentAdapter 支付适配器接口
type PaymentAdapter interface {
    GetChannel() string
    CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)
    QueryPayment(ctx context.Context, channelTradeNo string) (*QueryPaymentResponse, error)
    CancelPayment(ctx context.Context, channelTradeNo string) error
    CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error)
    QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error)
    VerifyWebhook(ctx context.Context, signature string, body []byte) (bool, error)
    ParseWebhook(ctx context.Context, body []byte) (*WebhookEvent, error)
}

// AdapterFactory 适配器工厂
type AdapterFactory struct {
    adapters map[string]PaymentAdapter
}
```

---

## Stripe 适配器

### Stripe 配置

```go
type StripeConfig struct {
    APIKey              string // Stripe API 密钥
    WebhookSecret       string // Webhook 签名密钥
    PublishableKey      string // 可发布密钥（给前端使用）
    StatementDescriptor string // 账单描述符
    SuccessURL          string // 支付成功跳转URL
    CancelURL           string // 支付取消跳转URL
    CaptureMethod       string // 捕获方式：automatic, manual
}
```

### Stripe 创建支付流程

1. **创建 PaymentIntent**
   ```go
   params := &stripe.PaymentIntentParams{
       Amount:      stripe.Int64(req.Amount),
       Currency:    stripe.String(req.Currency),
       Description: stripe.String(req.Description),
       Metadata: map[string]string{
           "payment_no": req.PaymentNo,
           "order_no":   req.OrderNo,
       },
   }
   params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
       Enabled: stripe.Bool(true),
   }
   pi, err := paymentintent.New(params)
   ```

2. **返回 ClientSecret**
   - 前端使用 `client_secret` 调用 Stripe.js 完成支付
   - 支持多种支付方式（信用卡、借记卡、数字钱包等）

3. **接收 Webhook 回调**
   - Stripe 发送 `payment_intent.succeeded` 事件
   - 验证签名并更新订单状态

### Stripe 金额转换

Stripe 对不同货币的金额单位不同：

```go
// 零小数位货币（如 JPY、KRW）- 金额单位是主单位
// 两位小数货币（如 USD、EUR）- 金额单位是分（cents）
// 三位小数货币（如 BHD、KWD）- 金额单位是 1/1000

func ConvertAmountToStripe(amount int64, currency string) int64 {
    zeroDecimalCurrencies := map[string]bool{
        "JPY": true, "KRW": true, "VND": true, ...
    }

    if zeroDecimalCurrencies[currency] {
        return amount / 100 // 从分转换为主单位
    }

    return amount // 其他货币直接使用分
}
```

### Stripe Webhook 事件

支持的事件类型：

| 事件类型 | 说明 | 处理动作 |
|---------|------|---------|
| `payment_intent.succeeded` | 支付成功 | 更新订单状态为已支付 |
| `payment_intent.payment_failed` | 支付失败 | 更新订单状态为失败 |
| `payment_intent.canceled` | 支付取消 | 更新订单状态为已取消 |
| `charge.refunded` | 退款成功 | 更新订单状态为已退款 |

### Stripe 状态映射

```go
func convertStripeStatus(status stripe.PaymentIntentStatus) string {
    switch status {
    case stripe.PaymentIntentStatusRequiresPaymentMethod,
         stripe.PaymentIntentStatusRequiresConfirmation,
         stripe.PaymentIntentStatusRequiresAction:
        return PaymentStatusPending
    case stripe.PaymentIntentStatusProcessing:
        return PaymentStatusProcessing
    case stripe.PaymentIntentStatusSucceeded:
        return PaymentStatusSuccess
    case stripe.PaymentIntentStatusCanceled:
        return PaymentStatusCancelled
    default:
        return PaymentStatusFailed
    }
}
```

---

## 数据模型

### 1. ChannelConfig - 渠道配置

```sql
CREATE TABLE channel_configs (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,          -- 商户ID
    channel VARCHAR(50) NOT NULL,       -- 渠道：stripe, paypal, crypto
    is_enabled BOOLEAN DEFAULT true,    -- 是否启用
    mode VARCHAR(20) NOT NULL,          -- 模式：test, live
    config JSONB NOT NULL,              -- 配置信息（JSON加密存储）
    fee_rate DECIMAL(10,4) DEFAULT 0,   -- 费率（百分比）
    fixed_fee BIGINT DEFAULT 0,         -- 固定手续费（分）
    min_amount BIGINT DEFAULT 0,        -- 最小金额（分）
    max_amount BIGINT,                  -- 最大金额（分）
    currencies JSONB,                   -- 支持的货币列表
    countries JSONB,                    -- 支持的国家列表
    priority INTEGER DEFAULT 0,         -- 优先级
    extra JSONB,                        -- 扩展信息
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

### 2. Transaction - 交易记录

```sql
CREATE TABLE channel_transactions (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    order_no VARCHAR(64),
    payment_no VARCHAR(64) NOT NULL,           -- 支付流水号
    channel VARCHAR(50) NOT NULL,              -- 渠道
    channel_trade_no VARCHAR(200),             -- 渠道交易号
    transaction_type VARCHAR(20) NOT NULL,     -- 交易类型：payment, refund
    amount BIGINT NOT NULL,                    -- 金额（分）
    currency VARCHAR(10) NOT NULL,             -- 货币
    status VARCHAR(20) NOT NULL,               -- 状态
    customer_email VARCHAR(255),
    customer_name VARCHAR(100),
    payment_method VARCHAR(50),                -- 支付方式
    payment_method_details JSONB,              -- 支付方式详情
    fee_amount BIGINT DEFAULT 0,               -- 手续费（分）
    net_amount BIGINT,                         -- 净额（分）
    error_code VARCHAR(50),                    -- 错误码
    error_message TEXT,                        -- 错误信息
    request_data JSONB,                        -- 请求数据
    response_data JSONB,                       -- 响应数据
    webhook_data JSONB,                        -- Webhook数据
    extra JSONB,
    processed_at TIMESTAMPTZ,                  -- 处理时间
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

### 3. WebhookLog - Webhook 日志

```sql
CREATE TABLE webhook_logs (
    id UUID PRIMARY KEY,
    merchant_id UUID,
    channel VARCHAR(50) NOT NULL,
    event_id VARCHAR(200),                  -- 事件ID
    event_type VARCHAR(100),                -- 事件类型
    payment_no VARCHAR(64),                 -- 支付流水号
    signature TEXT,                         -- 签名
    is_verified BOOLEAN DEFAULT false,      -- 是否验证通过
    is_processed BOOLEAN DEFAULT false,     -- 是否已处理
    request_body JSONB,                     -- 请求体
    request_headers JSONB,                  -- 请求头
    process_result TEXT,                    -- 处理结果
    retry_count INTEGER DEFAULT 0,          -- 重试次数
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);
```

---

## API 接口

### 1. 创建支付

**请求**
```http
POST /api/v1/channel/payments
Content-Type: application/json
Authorization: Bearer {token}

{
    "channel": "stripe",
    "payment_no": "PY20240115123456ABCD",
    "order_no": "OD20240115123456XYZ",
    "amount": 10000,
    "currency": "USD",
    "customer_email": "customer@example.com",
    "customer_name": "John Doe",
    "description": "Order #12345",
    "success_url": "https://example.com/success",
    "cancel_url": "https://example.com/cancel",
    "callback_url": "https://example.com/callback"
}
```

**响应**
```json
{
    "payment_no": "PY20240115123456ABCD",
    "channel_trade_no": "pi_3abc123def456",
    "client_secret": "pi_3abc123def456_secret_xyz",
    "status": "pending",
    "extra": {
        "payment_intent_id": "pi_3abc123def456"
    }
}
```

### 2. 查询支付

**请求**
```http
GET /api/v1/channel/payments/{payment_no}
Authorization: Bearer {token}
```

**响应**
```json
{
    "payment_no": "PY20240115123456ABCD",
    "channel_trade_no": "pi_3abc123def456",
    "status": "success",
    "amount": 10000,
    "currency": "USD",
    "payment_method": "card",
    "payment_method_details": {
        "brand": "visa",
        "last4": "4242",
        "exp_month": 12,
        "exp_year": 2025,
        "country": "US"
    },
    "paid_at": "2024-01-15T12:34:56Z"
}
```

### 3. 取消支付

**请求**
```http
POST /api/v1/channel/payments/{payment_no}/cancel
Authorization: Bearer {token}
```

**响应**
```json
{
    "message": "取消成功"
}
```

### 4. 创建退款

**请求**
```http
POST /api/v1/channel/refunds
Content-Type: application/json
Authorization: Bearer {token}

{
    "refund_no": "RF20240115123456ABCD",
    "payment_no": "PY20240115123456ABCD",
    "amount": 10000,
    "currency": "USD",
    "reason": "Customer requested refund"
}
```

**响应**
```json
{
    "refund_no": "RF20240115123456ABCD",
    "channel_refund_no": "re_3abc123def456",
    "status": "processing"
}
```

### 5. 查询退款

**请求**
```http
GET /api/v1/channel/refunds/{refund_no}
Authorization: Bearer {token}
```

**响应**
```json
{
    "refund_no": "RF20240115123456ABCD",
    "channel_refund_no": "re_3abc123def456",
    "status": "refunded",
    "amount": 10000,
    "currency": "USD",
    "refunded_at": "2024-01-15T13:00:00Z"
}
```

---

## Webhook 处理

### Stripe Webhook 端点

```http
POST /api/v1/webhooks/stripe
Content-Type: application/json
Stripe-Signature: t=1234567890,v1=abc123def456...

{
    "id": "evt_abc123",
    "type": "payment_intent.succeeded",
    "data": {
        "object": {
            "id": "pi_abc123",
            "amount": 10000,
            "currency": "usd",
            "status": "succeeded",
            "metadata": {
                "payment_no": "PY20240115123456ABCD",
                "order_no": "OD20240115123456XYZ"
            }
        }
    }
}
```

### Webhook 处理流程

1. **接收回调** - 接收 Stripe 发送的 Webhook 请求
2. **验证签名** - 使用 Webhook Secret 验证签名
3. **保存日志** - 记录完整的 Webhook 数据
4. **幂等性检查** - 根据 event_id 检查是否已处理
5. **解析事件** - 解析 Webhook 数据
6. **更新状态** - 更新交易状态
7. **返回响应** - 返回 200 OK

```go
func (s *channelService) HandleWebhook(ctx context.Context, channel string, signature string, body []byte, headers map[string]string) error {
    // 1. 获取适配器
    adpt, ok := s.adapterFactory.GetAdapter(channel)

    // 2. 验证签名
    verified, err := adpt.VerifyWebhook(ctx, signature, body)

    // 3. 解析 Webhook 数据
    event, err := adpt.ParseWebhook(ctx, body)

    // 4. 幂等性检查
    existingLog, _ := s.repo.GetWebhookLog(ctx, event.EventID)
    if existingLog != nil && existingLog.IsProcessed {
        return nil // 已处理，直接返回
    }

    // 5. 保存日志
    log := &model.WebhookLog{...}
    s.repo.CreateWebhookLog(ctx, log)

    // 6. 处理事件
    s.processWebhookEvent(ctx, event)

    // 7. 标记为已处理
    log.IsProcessed = true
    s.repo.UpdateWebhookLog(ctx, log)

    return nil
}
```

### Webhook 重试机制

如果 Webhook 处理失败，系统会自动重试：

- **最大重试次数**: 3次
- **重试间隔**: 指数退避（1分钟、5分钟、30分钟）
- **超时处理**: 超过3次重试后标记为失败，需要人工介入

```go
func (s *channelService) ProcessPendingWebhooks(ctx context.Context) error {
    // 获取未处理的 Webhook 列表（retry_count < 3）
    logs, err := s.repo.ListUnprocessedWebhooks(ctx, 100)

    for _, log := range logs {
        // 处理事件
        if err := s.processWebhookEvent(ctx, &event); err != nil {
            log.RetryCount++
        } else {
            log.IsProcessed = true
        }
        s.repo.UpdateWebhookLog(ctx, log)
    }

    return nil
}
```

---

## 部署配置

### 环境变量

```bash
# 数据库配置
DATABASE_URL=postgres://user:pass@localhost:5432/payment_platform?sslmode=disable

# 服务配置
PORT=8003

# Stripe 配置
STRIPE_API_KEY=sk_test_xxx                     # Stripe API 密钥
STRIPE_WEBHOOK_SECRET=whsec_xxx                # Webhook 签名密钥
STRIPE_PUBLISHABLE_KEY=pk_test_xxx             # 可发布密钥

# PayPal 配置（未来支持）
PAYPAL_CLIENT_ID=xxx
PAYPAL_CLIENT_SECRET=xxx
PAYPAL_MODE=sandbox                            # sandbox 或 live
```

### Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o channel-adapter ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/channel-adapter .
EXPOSE 8003
CMD ["./channel-adapter"]
```

### Docker Compose

```yaml
channel-adapter:
  build: ./backend/services/channel-adapter
  ports:
    - "8003:8003"
  environment:
    - DATABASE_URL=postgres://postgres:postgres@postgres:5432/payment_platform?sslmode=disable
    - STRIPE_API_KEY=${STRIPE_API_KEY}
    - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
    - STRIPE_PUBLISHABLE_KEY=${STRIPE_PUBLISHABLE_KEY}
  depends_on:
    - postgres
  restart: unless-stopped
```

### 配置 Stripe Webhook

1. **登录 Stripe Dashboard**
   - 访问 https://dashboard.stripe.com/webhooks

2. **添加端点**
   - 点击 "Add endpoint"
   - URL: `https://your-domain.com/api/v1/webhooks/stripe`
   - 选择事件：
     - `payment_intent.succeeded`
     - `payment_intent.payment_failed`
     - `payment_intent.canceled`
     - `charge.refunded`

3. **获取 Webhook Secret**
   - 创建后会显示 Webhook 签名密钥（whsec_xxx）
   - 将密钥配置到环境变量 `STRIPE_WEBHOOK_SECRET`

4. **测试 Webhook**
   - 使用 Stripe CLI 测试：
   ```bash
   stripe listen --forward-to localhost:8003/api/v1/webhooks/stripe
   stripe trigger payment_intent.succeeded
   ```

---

## 监控与日志

### 关键指标

- **支付成功率** - 成功支付数 / 总支付数
- **平均响应时间** - API 响应时间
- **Webhook 处理延迟** - 接收到处理完成的时间
- **错误率** - 失败请求数 / 总请求数
- **渠道可用性** - 各渠道的健康状态

### 日志记录

```go
// 记录所有交易
log.Info("创建支付",
    "payment_no", paymentNo,
    "channel", channel,
    "amount", amount,
    "currency", currency,
)

// 记录 Webhook 事件
log.Info("处理 Webhook",
    "event_id", eventID,
    "event_type", eventType,
    "channel", channel,
)

// 记录错误
log.Error("支付失败",
    "payment_no", paymentNo,
    "error", err,
)
```

---

## 安全最佳实践

### 1. API 密钥管理
- ✅ 使用环境变量存储密钥，不要硬编码
- ✅ 使用 HashiCorp Vault 或 AWS Secrets Manager 管理敏感信息
- ✅ 定期轮换 API 密钥
- ✅ 数据库中的配置字段使用 AES-256 加密

### 2. Webhook 安全
- ✅ 始终验证 Webhook 签名
- ✅ 使用 HTTPS
- ✅ 实现幂等性检查
- ✅ 设置请求超时

### 3. 数据保护
- ✅ 不存储完整的信用卡号
- ✅ 记录日志时脱敏敏感信息
- ✅ 使用 TLS 1.3 加密传输
- ✅ 定期备份交易数据

### 4. 错误处理
- ✅ 不在错误信息中暴露敏感数据
- ✅ 记录详细的错误日志供调试
- ✅ 向用户返回友好的错误提示

---

## 常见问题

### Q1: 如何添加新的支付渠道？

1. 实现 `PaymentAdapter` 接口
2. 在 AdapterFactory 中注册适配器
3. 创建渠道配置
4. 配置 Webhook 端点

### Q2: Webhook 丢失怎么办？

- 系统会定期处理未处理的 Webhook（retry_count < 3）
- 可以手动调用查询接口同步状态
- Stripe 会重试发送 Webhook（最多3天）

### Q3: 如何处理并发问题？

- 使用数据库的唯一索引防止重复创建
- Webhook 事件使用 event_id 去重
- 交易状态更新使用乐观锁

### Q4: 支持哪些货币？

目前支持 32 种主流货币：
- 零小数位: JPY, KRW, VND等
- 两位小数: USD, EUR, GBP, CNY等
- 三位小数: BHD, KWD等

---

## 总结

Channel Adapter Service 提供了统一的支付渠道适配层，目前已完整实现 Stripe 支付渠道的对接，支持创建支付、查询、取消、退款等完整功能，并提供可靠的 Webhook 处理机制。

**已实现功能**:
- ✅ Stripe 完整适配
- ✅ 支付创建与查询
- ✅ 退款管理
- ✅ Webhook 处理与重试
- ✅ 交易记录与日志
- ✅ 多货币支持

**未来扩展**:
- ⏳ PayPal 适配器
- ⏳ 加密货币支付
- ⏳ 支付宝/微信支付（国内）
- ⏳ 更多支付方式
