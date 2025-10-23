# Notification Service 完整指南

## 📋 概述

Notification Service（通知服务）是支付平台的核心服务之一，负责处理所有邮件、短信和Webhook通知。本服务已完成100%开发，包含完整的API和后台任务处理。

## ✅ 已完成功能

### 核心功能
- ✅ **邮件发送**：支持SMTP和Mailgun提供商
- ✅ **短信发送**：支持Twilio和Mock SMS（测试用）
- ✅ **Webhook投递**：支持异步Webhook事件推送和重试机制
- ✅ **模板管理**：支持系统模板和商户自定义模板
- ✅ **通知历史**：完整的通知记录和查询功能
- ✅ **后台任务**：自动处理待发送通知和失败重试

### 8个系统默认模板
1. **merchant_welcome** - 商户注册欢迎邮件
2. **kyc_approved** - KYC审核通过通知
3. **kyc_rejected** - KYC审核拒绝通知
4. **merchant_frozen** - 商户账号冻结通知
5. **password_reset** - 密码重置邮件
6. **payment_success** - 支付成功通知
7. **payment_failed** - 支付失败通知
8. **refund_completed** - 退款完成通知

## 🚀 快速开始

### 1. 数据库初始化

```bash
# 创建数据库
docker exec payment-postgres psql -U postgres -c "CREATE DATABASE payment_notify;"

# 运行服务（会自动执行AutoMigrate）
GOWORK=/home/eric/payment/backend/go.work go run ./cmd/main.go

# 导入系统默认模板
docker exec -i payment-postgres psql -U postgres -d payment_notify < migrations/001_seed_templates.sql
```

### 2. 环境变量配置

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_notify
DB_SSL_MODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# 服务端口
PORT=8007

# JWT认证
JWT_SECRET=your-secret-key-change-in-production

# SMTP邮件配置（可选）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@payment-platform.com

# Mailgun邮件配置（可选）
MAILGUN_DOMAIN=mg.yourdomain.com
MAILGUN_API_KEY=your-mailgun-api-key
MAILGUN_FROM=noreply@yourdomain.com

# Twilio短信配置（可选）
TWILIO_ACCOUNT_SID=your-account-sid
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_FROM=+1234567890
```

### 3. 启动服务

```bash
cd /home/eric/payment/backend/services/notification-service

# 使用Air热重载（开发环境）
DB_HOST=localhost DB_PORT=40432 DB_USER=postgres DB_PASSWORD=postgres \
DB_NAME=payment_notify REDIS_HOST=localhost REDIS_PORT=40379 \
PORT=8007 ~/go/bin/air

# 或直接运行（生产环境）
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres \
DB_NAME=payment_notify REDIS_HOST=localhost REDIS_PORT=6379 \
PORT=8007 /tmp/notification-service
```

## 📡 API接口文档

### 认证说明
除了健康检查和Swagger文档外，所有API都需要JWT认证。请在请求头中添加：
```
Authorization: Bearer <your-jwt-token>
```

### 1. 通知发送 API

#### 1.1 发送邮件
```http
POST /api/v1/notifications/email
Content-Type: application/json
Authorization: Bearer <token>

{
  "to": ["user@example.com"],
  "subject": "测试邮件",
  "html_body": "<p>这是一封测试邮件</p>",
  "text_body": "这是一封测试邮件",
  "provider": "smtp",
  "priority": 5
}
```

#### 1.2 使用模板发送邮件
```http
POST /api/v1/notifications/email/template
Content-Type: application/json
Authorization: Bearer <token>

{
  "to": ["merchant@example.com"],
  "template_code": "merchant_welcome",
  "template_data": {
    "merchant_name": "测试商户",
    "merchant_id": "123456",
    "email": "merchant@example.com"
  },
  "provider": "smtp",
  "priority": 5
}
```

#### 1.3 发送短信
```http
POST /api/v1/notifications/sms
Content-Type: application/json
Authorization: Bearer <token>

{
  "to": "+8613800138000",
  "content": "您的验证码是：123456",
  "provider": "twilio",
  "priority": 9
}
```

#### 1.4 发送Webhook
```http
POST /api/v1/notifications/webhook
Content-Type: application/json
Authorization: Bearer <token>

{
  "event_type": "payment.success",
  "event_id": "evt_123456",
  "data": {
    "order_no": "ORD20240101001",
    "amount": 10000,
    "currency": "USD"
  }
}
```

### 2. 通知查询 API

#### 2.1 获取通知详情
```http
GET /api/v1/notifications/{notification_id}
Authorization: Bearer <token>
```

#### 2.2 列出通知列表
```http
GET /api/v1/notifications?merchant_id={merchant_id}&type=payment&channel=email&status=sent&page=1&page_size=20
Authorization: Bearer <token>
```

### 3. 模板管理 API

#### 3.1 创建模板
```http
POST /api/v1/templates
Content-Type: application/json
Authorization: Bearer <token>

{
  "code": "custom_template",
  "name": "自定义模板",
  "type": "marketing",
  "channel": "email",
  "subject": "促销活动 - {{campaign_name}}",
  "content": "<html><body><h1>{{campaign_name}}</h1><p>{{content}}</p></body></html>",
  "description": "营销活动邮件模板",
  "variables": "[\"campaign_name\", \"content\"]",
  "is_enabled": true
}
```

#### 3.2 获取模板
```http
GET /api/v1/templates/{template_code}
Authorization: Bearer <token>
```

#### 3.3 列出模板
```http
GET /api/v1/templates
Authorization: Bearer <token>
```

#### 3.4 更新模板
```http
PUT /api/v1/templates/{template_id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "更新后的模板名称",
  "subject": "新的主题",
  "content": "新的内容",
  "is_enabled": true
}
```

#### 3.5 删除模板
```http
DELETE /api/v1/templates/{template_id}
Authorization: Bearer <token>
```

### 4. Webhook端点管理 API

#### 4.1 创建Webhook端点
```http
POST /api/v1/webhooks/endpoints
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "支付回调端点",
  "url": "https://merchant.example.com/webhooks/payment",
  "secret": "your-webhook-secret",
  "events": "[\"payment.success\", \"payment.failed\", \"refund.completed\"]",
  "is_enabled": true,
  "timeout": 30,
  "max_retry": 3,
  "description": "接收支付相关事件"
}
```

#### 4.2 列出Webhook端点
```http
GET /api/v1/webhooks/endpoints
Authorization: Bearer <token>
```

#### 4.3 更新Webhook端点
```http
PUT /api/v1/webhooks/endpoints/{endpoint_id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "更新后的端点",
  "url": "https://new-url.example.com/webhooks",
  "is_enabled": false
}
```

#### 4.4 删除Webhook端点
```http
DELETE /api/v1/webhooks/endpoints/{endpoint_id}
Authorization: Bearer <token>
```

#### 4.5 查询Webhook投递记录
```http
GET /api/v1/webhooks/deliveries?endpoint_id={endpoint_id}&status=delivered&page=1&page_size=20
Authorization: Bearer <token>
```

### 5. 系统接口

#### 5.1 健康检查（无需认证）
```http
GET /health
```

#### 5.2 API文档（无需认证）
```http
GET /swagger/index.html
```

## 🔧 与其他服务集成

### 在merchant-service中使用

```go
// 发送商户注册欢迎邮件
func (s *merchantService) sendWelcomeEmail(merchant *model.Merchant) error {
    notifyURL := config.GetEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8007")

    reqBody := map[string]interface{}{
        "to": []string{merchant.Email},
        "template_code": "merchant_welcome",
        "template_data": map[string]interface{}{
            "merchant_name": merchant.LegalName,
            "merchant_id": merchant.ID.String(),
            "email": merchant.Email,
        },
        "provider": "smtp",
        "priority": 5,
    }

    // 发送HTTP请求到notification-service
    // ... (使用httpclient发送POST请求)
}

// 发送KYC审核通过通知
func (s *merchantService) sendKYCApprovedEmail(merchant *model.Merchant) error {
    // 类似上面的逻辑，使用 kyc_approved 模板
}
```

## 📊 数据库表结构

### notifications
- 通知记录表，存储所有发送的通知
- 字段：id, merchant_id, type, channel, recipient, subject, content, status, priority, retry_count等

### notification_templates
- 通知模板表，存储系统模板和商户自定义模板
- 字段：id, merchant_id, code, name, type, channel, subject, content, variables, is_system等

### webhook_endpoints
- Webhook端点配置表
- 字段：id, merchant_id, name, url, secret, events, is_enabled, timeout, max_retry等

### webhook_deliveries
- Webhook投递记录表
- 字段：id, endpoint_id, merchant_id, event_type, payload, status, http_status, response_body等

## 🔄 后台任务

服务启动后会自动运行两个后台任务（每分钟执行一次）：

1. **ProcessPendingNotifications**：处理待发送的通知（每次最多处理100条）
2. **ProcessPendingWebhookDeliveries**：处理待投递的Webhook（包括失败重试）

## 🎯 优先级说明

通知优先级范围：0-9（9为最高优先级）
- **9**: 紧急通知（如安全告警、账号冻结）
- **7-8**: 重要通知（如支付成功、KYC审核结果）
- **5-6**: 一般通知（如订单更新）
- **3-4**: 低优先级通知（如营销邮件）
- **0-2**: 最低优先级

## 🔐 安全特性

1. **JWT认证**：所有业务API都需要JWT token
2. **Webhook签名**：使用HMAC-SHA256签名验证Webhook请求
3. **限流保护**：使用Redis实现分布式限流（100请求/分钟）
4. **请求ID追踪**：每个请求都有唯一的Request ID用于日志追踪

## 📈 监控和日志

- 使用Zap结构化日志
- 所有请求都有Request ID
- 后台任务错误会记录到日志
- 可通过健康检查端点监控服务状态

## 🐛 故障排查

### 编译失败
```bash
# 清除缓存后重新编译
cd /home/eric/payment/backend/services/notification-service
go clean -cache
GOWORK=/home/eric/payment/backend/go.work go mod tidy
GOWORK=/home/eric/payment/backend/go.work go build ./cmd/main.go
```

### 邮件发送失败
1. 检查SMTP配置是否正确
2. 确认邮箱密码（Gmail需要使用应用专用密码）
3. 查看notification记录表的error_message字段

### Webhook投递失败
1. 检查目标URL是否可访问
2. 确认Webhook端点的secret配置正确
3. 查看webhook_deliveries表的error_message和http_status字段

## 🎉 总结

Notification Service已经100%完成开发，包含：
- ✅ 3种通知渠道（Email、SMS、Webhook）
- ✅ 完整的CRUD API（18个端点）
- ✅ 8个系统默认模板
- ✅ JWT认证和安全防护
- ✅ 后台任务和失败重试机制
- ✅ 完整的通知历史记录

服务已经可以立即投入使用，支持merchant-service、payment-gateway等其他服务的通知需求！
