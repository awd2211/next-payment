# 邮件集成文档

## 功能概述

Payment Platform 集成了完善的邮件发送功能，支持：

✅ **SMTP** - 使用自己的邮箱服务器（如 Gmail、Outlook、企业邮箱）
✅ **Mailgun** - 专业的邮件发送服务（推荐用于生产环境）
✅ **邮件模板管理** - 管理员可在后台可视化配置邮件模板
✅ **变量替换** - 支持动态变量（如 `{{.Name}}`, `{{.Email}}`）
✅ **HTML + 纯文本** - 同时支持HTML和纯文本格式
✅ **邮件日志** - 记录所有邮件发送状态，便于追踪
✅ **批量发送** - 支持批量发送邮件
✅ **定时发送** - Mailgun支持定时发送
✅ **附件支持** - 支持添加附件

---

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│               Admin Portal（管理员后台）                  │
│   - 邮件模板管理（CRUD）                                  │
│   - 可视化编辑器（HTML/CSS）                              │
│   - 测试邮件发送                                          │
│   - 查看邮件发送日志                                      │
└─────────────────────────────────────────────────────────┘
                         ↓ REST API
┌─────────────────────────────────────────────────────────┐
│              Admin Service（运营管理服务）                │
│   - EmailTemplateService（邮件模板服务）                 │
│   - EmailTemplateRepository（数据访问层）                │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                pkg/email（共享邮件库）                    │
│   - EmailClient（统一客户端）                            │
│   - SMTPProvider（SMTP提供商）                           │
│   - MailgunProvider（Mailgun提供商）                     │
│   - Template Engine（模板引擎）                          │
└─────────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────┬──────────────────────────────────┐
│   SMTP Server        │   Mailgun API                    │
│   (Gmail/Outlook)    │   (mail.mailgun.com)             │
└──────────────────────┴──────────────────────────────────┘
```

---

## 快速开始

### 1. 配置环境变量

编辑 `.env` 文件：

```bash
# 选择邮件提供商：smtp 或 mailgun
EMAIL_PROVIDER=smtp

# SMTP配置（使用Gmail示例）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password  # 需要生成应用专用密码
SMTP_FROM=noreply@payment-platform.com
SMTP_FROM_NAME=Payment Platform

# Mailgun配置（推荐生产环境）
MAILGUN_DOMAIN=mail.yourdomain.com
MAILGUN_API_KEY=your-mailgun-api-key
MAILGUN_FROM=noreply@yourdomain.com
MAILGUN_FROM_NAME=Payment Platform
MAILGUN_EU_REGION=false

# 邮件模板路径
EMAIL_TEMPLATE_PATH=./templates/email
```

### 2. 初始化邮件模板

启动服务后，调用API初始化默认模板：

```bash
curl -X POST http://localhost:8001/api/v1/email-templates/init-defaults \
  -H "Authorization: Bearer <admin-token>"
```

### 3. 发送测试邮件

```bash
curl -X POST http://localhost:8001/api/v1/email-templates/send-template \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["user@example.com"],
    "template_code": "welcome",
    "data": {
      "Name": "张三",
      "Email": "user@example.com",
      "DashboardURL": "https://dashboard.payment-platform.com"
    }
  }'
```

---

## API接口文档

### 邮件模板管理

#### 1. 创建邮件模板

```http
POST /api/v1/email-templates
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "custom_template",
  "name": "自定义模板",
  "subject": "欢迎使用{{.ProductName}}",
  "html_content": "<html><body><h1>您好，{{.Name}}！</h1></body></html>",
  "text_content": "您好，{{.Name}}！",
  "category": "notification",
  "description": "自定义通知模板",
  "variables": [
    {
      "name": "Name",
      "placeholder": "{{.Name}}",
      "description": "用户名",
      "required": true
    },
    {
      "name": "ProductName",
      "placeholder": "{{.ProductName}}",
      "description": "产品名称",
      "required": true
    }
  ]
}
```

**响应：**
```json
{
  "data": {
    "id": "uuid",
    "code": "custom_template",
    "name": "自定义模板",
    "subject": "欢迎使用{{.ProductName}}",
    "html_content": "...",
    "category": "notification",
    "is_active": true,
    "created_at": "2025-01-15T10:00:00Z"
  }
}
```

#### 2. 获取模板列表

```http
GET /api/v1/email-templates?page=1&page_size=20&category=account&is_active=true
Authorization: Bearer <token>
```

**响应：**
```json
{
  "data": [
    {
      "id": "uuid",
      "code": "welcome",
      "name": "欢迎邮件",
      "subject": "欢迎加入 Payment Platform",
      "category": "account",
      "is_active": true,
      "is_system": true
    }
  ],
  "total": 10,
  "page": 1,
  "page_size": 20
}
```

#### 3. 更新模板

```http
PUT /api/v1/email-templates/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "欢迎邮件（新版）",
  "subject": "欢迎加入 {{.PlatformName}}",
  "html_content": "<html>...</html>",
  "is_active": true
}
```

#### 4. 测试模板渲染

```http
POST /api/v1/email-templates/:id/test
Authorization: Bearer <token>
Content-Type: application/json

{
  "data": {
    "Name": "测试用户",
    "Email": "test@example.com",
    "PlatformName": "Payment Platform"
  }
}
```

**响应：**
```json
{
  "html": "<html><body><h1>欢迎，测试用户！</h1>...</body></html>"
}
```

#### 5. 删除模板

```http
DELETE /api/v1/email-templates/:id
Authorization: Bearer <token>
```

---

### 邮件发送

#### 1. 使用模板发送邮件

```http
POST /api/v1/email-templates/send-template
Authorization: Bearer <token>
Content-Type: application/json

{
  "to": ["user1@example.com", "user2@example.com"],
  "template_code": "welcome",
  "data": {
    "Name": "张三",
    "Email": "user1@example.com",
    "Username": "zhangsan",
    "CreatedAt": "2025-01-15 10:00:00",
    "DashboardURL": "https://dashboard.payment-platform.com"
  }
}
```

#### 2. 直接发送邮件（不使用模板）

```http
POST /api/v1/email-templates/send
Authorization: Bearer <token>
Content-Type: application/json

{
  "to": ["user@example.com"],
  "subject": "重要通知",
  "html_content": "<html><body><h1>这是一封重要通知</h1></body></html>",
  "text_content": "这是一封重要通知"
}
```

---

### 邮件日志

#### 查看邮件发送日志

```http
GET /api/v1/email-templates/logs?page=1&page_size=20&status=sent&to=user@example.com
Authorization: Bearer <token>
```

**响应：**
```json
{
  "data": [
    {
      "id": "uuid",
      "template_id": "uuid",
      "to": "user@example.com",
      "subject": "欢迎加入 Payment Platform",
      "status": "sent",
      "provider": "smtp",
      "sent_at": "2025-01-15T10:05:00Z",
      "created_at": "2025-01-15T10:00:00Z"
    },
    {
      "id": "uuid",
      "to": "error@example.com",
      "subject": "支付成功通知",
      "status": "failed",
      "provider": "mailgun",
      "error_msg": "Invalid email address",
      "created_at": "2025-01-15T09:00:00Z"
    }
  ],
  "total": 1250,
  "page": 1,
  "page_size": 20
}
```

---

## 预定义模板

系统内置了以下邮件模板：

| 模板代码 | 名称 | 用途 | 可用变量 |
|---------|------|------|---------|
| `welcome` | 欢迎邮件 | 用户注册成功 | Name, Email, Username, DashboardURL |
| `verify_email` | 邮箱验证 | 验证邮箱地址 | Name, VerificationCode, VerificationURL, ExpiresIn |
| `reset_password` | 重置密码 | 找回密码 | Name, ResetURL, ExpiresIn |
| `payment_success` | 支付成功 | 支付成功通知 | CustomerName, Amount, Currency, OrderNo, PaymentID, PaidAt |
| `payment_failed` | 支付失败 | 支付失败通知 | CustomerName, Amount, Currency, OrderNo, FailureReason |
| `refund_completed` | 退款完成 | 退款成功通知 | CustomerName, RefundAmount, Currency, OrderNo, RefundID |
| `merchant_approved` | 商户审核通过 | 商户KYC通过 | Name, MerchantName, DashboardURL |
| `merchant_rejected` | 商户审核拒绝 | 商户KYC拒绝 | Name, MerchantName, Reason |
| `invoice` | 账单 | 月度账单 | MerchantName, Month, TotalAmount, Currency, InvoiceURL |

---

## 模板变量说明

### 变量语法

使用 Go template 语法：`{{.VariableName}}`

### 条件判断

```html
{{if .IsVIP}}
  <p>尊贵的VIP用户，您好！</p>
{{else}}
  <p>尊敬的用户，您好！</p>
{{end}}
```

### 循环遍历

```html
<ul>
{{range .Items}}
  <li>{{.Name}} - {{.Price}}</li>
{{end}}
</ul>
```

### 常用函数

```html
<!-- 格式化日期 -->
{{.CreatedAt.Format "2006-01-02 15:04:05"}}

<!-- 金额格式化 -->
{{printf "%.2f" .Amount}}
```

---

## 配置指南

### 使用Gmail SMTP

1. **启用两步验证**：https://myaccount.google.com/security
2. **生成应用专用密码**：https://myaccount.google.com/apppasswords
3. **配置环境变量**：

```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=<生成的16位密码>
SMTP_FROM=noreply@payment-platform.com
SMTP_FROM_NAME=Payment Platform
```

### 使用Mailgun

1. **注册Mailgun账号**：https://signup.mailgun.com
2. **验证域名**：添加DNS记录
3. **获取API Key**：Settings → API Keys
4. **配置环境变量**：

```bash
EMAIL_PROVIDER=mailgun
MAILGUN_DOMAIN=mail.yourdomain.com
MAILGUN_API_KEY=<your-api-key>
MAILGUN_FROM=noreply@yourdomain.com
MAILGUN_FROM_NAME=Payment Platform
MAILGUN_EU_REGION=false
```

### 使用AWS SES

可扩展支持AWS SES，只需实现 `EmailProvider` 接口：

```go
type SESProvider struct {}

func (p *SESProvider) Send(to []string, subject, htmlBody, textBody string, attachments []Attachment) error {
    // 调用AWS SES API
}
```

---

## 最佳实践

### 1. 邮件模板设计

✅ **使用响应式设计**
```html
<style>
  @media only screen and (max-width: 600px) {
    .container { width: 100% !important; }
  }
</style>
```

✅ **提供纯文本备份**
```
主题：支付成功
正文：
您好，{{.Name}}！
您的支付已成功完成。
金额：{{.Currency}} {{.Amount}}
订单号：{{.OrderNo}}
```

✅ **测试多个邮箱客户端**
- Gmail
- Outlook
- Apple Mail
- 手机邮箱

### 2. 发送频率控制

```go
// 使用限流防止滥用
rateLimiter := middleware.NewRateLimiter(redis, 10, time.Minute) // 10封/分钟
```

### 3. 邮件日志监控

```sql
-- 查询过去24小时发送失败的邮件
SELECT * FROM email_logs
WHERE status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours';

-- 统计各模板发送成功率
SELECT
  template_id,
  COUNT(*) as total,
  SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as success,
  ROUND(100.0 * SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM email_logs
GROUP BY template_id;
```

### 4. 退信处理

```go
// Mailgun提供退信列表API
bounces, err := mailgunProvider.GetBounces()
for _, bounce := range bounces {
    // 将退信地址加入黑名单
    blacklist.Add(bounce.Address)
}
```

---

## 故障排查

### 问题1：SMTP连接超时

**症状**：`dial tcp: i/o timeout`

**解决方案：**
1. 检查防火墙是否允许端口587（STARTTLS）或465（SSL）
2. 检查SMTP服务器地址是否正确
3. 尝试使用telnet测试连接：
```bash
telnet smtp.gmail.com 587
```

### 问题2：认证失败

**症状**：`535 Authentication failed`

**解决方案：**
1. Gmail需要使用应用专用密码，不是账号密码
2. 检查用户名和密码是否正确
3. 确认账号已启用SMTP访问

### 问题3：邮件进垃圾箱

**解决方案：**
1. **配置SPF记录**：
```
v=spf1 include:_spf.google.com ~all
```

2. **配置DKIM签名**（Mailgun自动配置）

3. **配置DMARC记录**：
```
v=DMARC1; p=none; rua=mailto:dmarc@yourdomain.com
```

4. **避免垃圾词汇**：
   - ❌ "免费"、"中奖"、"点击这里"
   - ✅ 使用正式语言

### 问题4：Mailgun API错误

**症状**：`401 Unauthorized`

**解决方案：**
1. 检查API Key是否正确
2. 确认域名已验证
3. 检查是否使用了正确的区域（US/EU）

---

## 性能优化

### 1. 批量发送

```go
// 使用批量发送API
messages := []*email.EmailMessage{
    {To: []string{"user1@example.com"}, Subject: "...", HTMLBody: "..."},
    {To: []string{"user2@example.com"}, Subject: "...", HTMLBody: "..."},
}
smtpProvider.BatchSend(messages)
```

### 2. 异步发送

```go
// 使用Kafka异步发送
go func() {
    kafkaProducer.Send("email-queue", emailData)
}()
```

### 3. 连接池

SMTP Provider 已内置连接池，复用TCP连接。

---

## 数据库表结构

### email_templates（邮件模板表）

```sql
CREATE TABLE email_templates (
    id UUID PRIMARY KEY,
    code VARCHAR(100) UNIQUE,        -- 模板代码
    name VARCHAR(100),                -- 模板名称
    subject VARCHAR(255),             -- 邮件主题
    html_content TEXT,                -- HTML内容
    text_content TEXT,                -- 纯文本内容
    category VARCHAR(50),             -- 分类
    variables JSONB,                  -- 可用变量
    is_active BOOLEAN,                -- 是否启用
    is_system BOOLEAN,                -- 系统内置
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

### email_logs（邮件日志表）

```sql
CREATE TABLE email_logs (
    id UUID PRIMARY KEY,
    template_id UUID,                 -- 模板ID
    to_address VARCHAR(255),          -- 收件人
    subject VARCHAR(255),             -- 主题
    status VARCHAR(20),               -- pending/sent/failed
    provider VARCHAR(50),             -- smtp/mailgun
    error_msg TEXT,                   -- 错误信息
    sent_at TIMESTAMPTZ,              -- 发送时间
    created_at TIMESTAMPTZ
);
```

---

## 扩展功能

### 1. 添加新的邮件提供商

实现 `EmailProvider` 接口：

```go
type SendGridProvider struct {
    apiKey string
}

func (p *SendGridProvider) Send(to []string, subject, htmlBody, textBody string, attachments []Attachment) error {
    // 调用SendGrid API
}
```

### 2. 富文本编辑器

集成前端编辑器（如Quill、TinyMCE）用于可视化编辑邮件模板。

### 3. A/B测试

```go
// 创建两个版本的模板进行测试
templateA := "welcome_v1"
templateB := "welcome_v2"

// 随机选择
if rand.Float32() < 0.5 {
    SendTemplateEmail(to, templateA, data)
} else {
    SendTemplateEmail(to, templateB, data)
}
```

---

## 参考资料

- [Mailgun文档](https://documentation.mailgun.com/en/latest/)
- [Gmail SMTP配置](https://support.google.com/mail/answer/7126229)
- [Go模板语法](https://pkg.go.dev/text/template)
- [邮件HTML最佳实践](https://www.campaignmonitor.com/css/)

---

## 联系支持

如有问题，请联系：
- 技术支持：support@payment-platform.com
- 文档：https://docs.payment-platform.com
