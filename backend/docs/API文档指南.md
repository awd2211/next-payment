# API 文档指南

## 概述

本文档提供了支付平台所有微服务中实现的 Swagger/OpenAPI 文档的全面指南。

## 目录

- [快速开始](#快速开始)
- [访问 Swagger UI](#访问-swagger-ui)
- [文档覆盖率](#文档覆盖率)
- [添加 API 文档](#添加-api-文档)
- [重新生成文档](#重新生成文档)
- [最佳实践](#最佳实践)

---

## 快速开始

### 前置要求

安装 Swagger CLI 工具：

```bash
cd backend
make install-swagger
```

这会在 `~/go/bin/swag` 安装 `swag` CLI。

### 生成文档

为所有服务生成 Swagger 文档：

```bash
cd backend
make swagger-docs
```

此命令会：
1. 扫描所有服务目录
2. 解析 `main.go` 和 handler 文件中的 Swagger 注解
3. 在每个服务的 `api-docs/` 目录中生成 `docs.go`、`swagger.json` 和 `swagger.yaml`
4. 显示每个服务的 Swagger UI 访问 URL

---

## 访问 Swagger UI

服务运行后，可以通过以下地址访问交互式 Swagger UI：

| 服务 | Swagger UI URL | 端口 |
|---------|---------------|------|
| **Admin Service** | http://localhost:40001/swagger/index.html | 40001 |
| **Merchant Service** | http://localhost:40002/swagger/index.html | 40002 |
| **Payment Gateway** | http://localhost:40003/swagger/index.html | 40003 |
| **Order Service** | http://localhost:40004/swagger/index.html | 40004 |
| **Channel Adapter** | http://localhost:40005/swagger/index.html | 40005 |
| **Risk Service** | http://localhost:40006/swagger/index.html | 40006 |
| **Accounting Service** | http://localhost:40007/swagger/index.html | 40007 |
| **Notification Service** | http://localhost:40008/swagger/index.html | 40008 |
| **Analytics Service** | http://localhost:40009/swagger/index.html | 40009 |
| **Config Service** | http://localhost:40010/swagger/index.html | 40010 |
| **Merchant Auth Service** | http://localhost:40011/swagger/index.html | 40011 |
| **Settlement Service** | http://localhost:40013/swagger/index.html | 40013 |
| **Withdrawal Service** | http://localhost:40014/swagger/index.html | 40014 |
| **KYC Service** | http://localhost:40015/swagger/index.html | 40015 |

### 其他格式

**JSON 规范：**
```
http://localhost:{PORT}/swagger/swagger.json
```

**YAML 规范：**
```
http://localhost:{PORT}/swagger/swagger.yaml
```

---

## 文档覆盖率

### ✅ 完整文档的服务

| 服务 | 端点数 | 覆盖率 | 备注 |
|---------|-----------|----------|-------|
| **Admin Service** | 50+ | 95% | 用户管理、RBAC、系统配置、审计日志 |
| **Merchant Service** | 40+ | 95% | 商户 CRUD、KYC、结算、仪表板 |
| **Payment Gateway** | 10 | 100% | ✨ **新增**：完整支付和退款流程、webhooks |
| **Order Service** | 15 | 80% | ✨ **新增**：订单 CRUD、状态更新、统计 |
| **Channel Adapter** | 12 | 75% | 支付渠道操作、汇率 |
| **Notification Service** | 20 | 70% | Email、SMS、webhook 通知 |
| **Merchant Auth Service** | 15 | 90% | API 密钥、商户认证 |
| **KYC Service** | 12 | 85% | 文档提交、验证、合规 |
| **Withdrawal Service** | 10 | 80% | 提现请求、银行账户、审批 |

### ⚠️ 最小文档（仅模板）

| 服务 | 状态 | 优先级 |
|---------|--------|----------|
| Risk Service | ⚠️ 空 | 中 |
| Accounting Service | ⚠️ 空 | 中 |
| Analytics Service | ⚠️ 空 | 低 |
| Config Service | ⚠️ 空 | 低 |
| Cashier Service | ⚠️ 空 | 低 |
| Settlement Service | ⚠️ 空 | 中 |

---

## 添加 API 文档

### 步骤 1：添加服务级元数据（main.go）

在 `cmd/main.go` 中的 `main()` 函数**上方**添加这些注释：

```go
//	@title						Your Service API
//	@version					1.0
//	@description				Service description here
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40XXX
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
    // ...
}
```

### 步骤 2：导入 Swagger 包

在 `main.go` 中添加：

```go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "payment-platform/your-service/api-docs" // Generated docs
)
```

### 步骤 3：注册 Swagger 路由

在路由器初始化后添加：

```go
application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

### 步骤 4：文档化处理函数

在每个处理函数上方添加 Swagger 注解：

```go
// CreatePayment creates a new payment
//
//	@Summary		Create Payment
//	@Description	Create a new payment transaction
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		service.CreatePaymentInput	true	"Payment creation request"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
    // Implementation...
}
```

### 注解参考

| 注解 | 描述 | 示例 |
|------------|-------------|---------|
| `@Summary` | 简短描述（1 行） | `Create Payment` |
| `@Description` | 详细描述 | `Create a new payment transaction with validation` |
| `@Tags` | 分组端点 | `Payments`、`Orders`、`Refunds` |
| `@Accept` | 请求内容类型 | `json`、`xml` |
| `@Produce` | 响应内容类型 | `json` |
| `@Security` | 安全方案 | `BearerAuth` (JWT)、`ApiKeyAuth` |
| `@Param` | 参数定义 | `paymentNo path string true "Payment number"` |
| `@Success` | 成功响应 | `200 {object} Response` |
| `@Failure` | 错误响应 | `400 {object} Response` |
| `@Router` | 路由路径和方法 | `/payments [post]` |

### 参数类型

```go
// 路径参数
//	@Param	id	path	string	true	"Record ID"

// 查询参数
//	@Param	page		query	int		false	"Page number"	default(1)
//	@Param	page_size	query	int		false	"Page size"		default(20)
//	@Param	status		query	string	false	"Filter by status"

// Body 参数
//	@Param	request	body	service.CreateInput	true	"Request body"

// Header 参数
//	@Param	X-Request-ID	header	string	false	"Request ID"
```

---

## 重新生成文档

### 重新生成所有服务

```bash
cd backend
make swagger-docs
```

### 重新生成单个服务

```bash
cd backend/services/payment-gateway
~/go/bin/swag init -g cmd/main.go -o ./api-docs --parseDependency --parseInternal
```

### 何时重新生成

在以下情况下重新生成 Swagger 文档：
- ✅ 添加新的 API 端点
- ✅ 更改请求/响应结构
- ✅ 更新参数描述
- ✅ 修改认证方案
- ✅ 更改路由路径或 HTTP 方法

**重要：** 在提交 API 更改之前始终运行 `make swagger-docs`。

---

## 最佳实践

### 1. 一致的命名

在所有服务中使用一致的命名：

**标签：**
- 使用复数名词：`Payments`、`Orders`、`Refunds`
- 将相关端点归类到同一标签下

**摘要：**
- 使用祈使语气：`Create Payment`、`Get Order`、`Update Status`
- 保持在 50 个字符以内

**描述：**
- 提供上下文和业务逻辑
- 提及验证和约束
- 包含示例用例

### 2. 完整的参数文档

始终记录：
- ✅ 参数名称和类型
- ✅ 必需 vs 可选
- ✅ 默认值（对于可选参数）
- ✅ 验证规则（min/max、format）
- ✅ 示例值

**示例：**
```go
//	@Param	amount		query	int		true	"Payment amount in cents"		minimum(1)
//	@Param	currency	query	string	true	"Currency code (USD/EUR/CNY)"
//	@Param	page		query	int		false	"Page number"					default(1)		minimum(1)
//	@Param	page_size	query	int		false	"Items per page"				default(20)		minimum(1)	maximum(100)
```

### 3. 记录所有状态码

记录所有可能的 HTTP 状态码：

```go
//	@Success	200		{object}	Response				"Success"
//	@Success	201		{object}	Response				"Created"
//	@Failure	400		{object}	Response				"Bad Request"
//	@Failure	401		{object}	Response				"Unauthorized"
//	@Failure	403		{object}	Response				"Forbidden"
//	@Failure	404		{object}	Response				"Not Found"
//	@Failure	409		{object}	Response				"Conflict"
//	@Failure	500		{object}	Response				"Internal Server Error"
```

### 4. 使用响应模型

定义清晰的响应结构：

```go
type Response struct {
    Code    int         `json:"code" example:"0"`
    Message string      `json:"message" example:"success"`
    Data    interface{} `json:"data,omitempty"`
    TraceID string      `json:"trace_id,omitempty" example:"abc123"`
}

type PageResponse struct {
    List     interface{} `json:"list"`
    Total    int64       `json:"total" example:"100"`
    Page     int         `json:"page" example:"1"`
    PageSize int         `json:"page_size" example:"20"`
}
```

### 5. 记录认证

始终指定安全要求：

```go
//	@Security	BearerAuth
```

对于不需要认证的端点：

```go
//	@Security	none
```

### 6. 添加示例

在结构定义中使用 `example` 标签：

```go
type CreatePaymentInput struct {
    Amount        int64  `json:"amount" binding:"required,gt=0" example:"10000"`
    Currency      string `json:"currency" binding:"required" example:"USD"`
    CustomerEmail string `json:"customer_email" binding:"email" example:"customer@example.com"`
}
```

### 7. 按标签分组

逻辑地组织端点：

**Payment Gateway：**
- `Payments` - 支付操作
- `Refunds` - 退款操作
- `Webhooks` - 回调处理

**Order Service：**
- `Orders` - 订单管理
- `Statistics` - 订单分析

**Merchant Service：**
- `Merchants` - 商户 CRUD
- `KYC` - KYC 文档
- `Settlement` - 结算操作

---

## 常见问题

### 问题 1：找不到类型定义

**错误：**
```
ParseComment error: cannot find type definition: errors.SuccessResponse
```

**解决方案：**
使用本地结构体类型而不是外部包类型：
```go
//	@Success	200	{object}	Response  // ✅ 正确
//	@Success	200	{object}	errors.SuccessResponse  // ❌ 错误
```

### 问题 2：文档未更新

**原因：** 生成的文件被缓存

**解决方案：**
```bash
# 删除 api-docs 目录并重新生成
rm -rf api-docs/
make swagger-docs
```

### 问题 3：导入循环错误

**原因：** 包导入中的循环依赖

**解决方案：**
使用 `--parseDependency` 标志（已在 Makefile 中）：
```bash
swag init -g cmd/main.go -o ./api-docs --parseDependency --parseInternal
```

---

## 示例：完整的支付端点文档

```go
// CreatePayment creates a new payment
//
//	@Summary		Create Payment
//	@Description	Create a new payment transaction supporting Stripe, PayPal, and other channels
//	@Description	- Validates merchant ID and payment amount
//	@Description	- Performs risk assessment
//	@Description	- Creates order in order-service
//	@Description	- Routes to appropriate payment channel
//	@Description	- Returns payment URL for customer redirect
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		service.CreatePaymentInput	true	"Payment creation request"
//	@Success		200		{object}	Response{data=model.Payment}	"Payment created successfully"
//	@Failure		400		{object}	Response						"Invalid request parameters"
//	@Failure		401		{object}	Response						"Unauthorized - invalid or missing token"
//	@Failure		403		{object}	Response						"Forbidden - merchant not active"
//	@Failure		422		{object}	Response						"Unprocessable - risk check failed"
//	@Failure		500		{object}	Response						"Internal server error"
//	@Router			/payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
    // Implementation...
}
```

---

## 使用 Swagger UI 测试 API

### 1. 认证

对于需要 JWT 认证的端点：

1. 点击 Swagger UI 中的 **Authorize** 按钮（锁图标）
2. 输入：`Bearer YOUR_JWT_TOKEN`
3. 点击 **Authorize**
4. 点击 **Close**

### 2. 尝试端点

1. 展开一个端点
2. 点击 **Try it out**
3. 填写必需参数
4. 点击 **Execute**
5. 查看响应

### 3. 查看模型

点击底部的 **Schemas** 查看所有请求/响应模型。

---

## 文件结构

生成文档后，每个服务都有：

```
service-name/
├── cmd/
│   └── main.go              # Swagger 元数据注解
├── internal/
│   └── handler/
│       └── *_handler.go     # 端点注解
└── api-docs/                # 由 swag 生成
    ├── docs.go              # Go 代码（由 main.go 导入）
    ├── swagger.json         # OpenAPI 2.0 JSON 规范
    └── swagger.yaml         # OpenAPI 2.0 YAML 规范
```

**注意：** `api-docs/` 目录是自动生成的。请勿直接编辑这些文件。

---

## 与 CI/CD 集成

添加到您的 CI/CD 管道：

```yaml
# .github/workflows/ci.yml
- name: Generate Swagger docs
  run: |
    cd backend
    make install-swagger
    make swagger-docs

- name: Validate Swagger specs
  run: |
    # Validate all swagger.json files
    for spec in services/*/api-docs/swagger.json; do
      swagger-cli validate "$spec"
    done
```

---

## 下一步

### 优先级 1：完成文档（新端点）

- ✅ **Payment Gateway** - 所有支付和退款端点已文档化
- ✅ **Order Service** - 核心订单管理端点已文档化
- ⏳ **Risk Service** - 风险评估 API（需要文档）
- ⏳ **Accounting Service** - 账本和会计 API（需要文档）

### 优先级 2：增强文档

- 添加更详细的描述
- 包含错误代码参考
- 添加请求/响应示例
- 记录速率限制行为
- 添加 webhook 签名验证详情

### 优先级 3：API 版本控制

- 规划 API 版本控制策略（v1、v2）
- 记录破坏性更改
- 维护向后兼容性

---

## 资源

- **Swaggo 文档**：https://github.com/swaggo/swag
- **OpenAPI 2.0 规范**：https://swagger.io/specification/v2/
- **Swagger UI**：https://swagger.io/tools/swagger-ui/
- **JSON Schema**：https://json-schema.org/

---

## 支持

关于 API 文档的问题或疑问：
- 在项目仓库中创建 issue
- 联系：support@payment-platform.com
- Slack：#api-documentation

---

**最后更新：** 2025-10-24
**版本：** 1.0
**状态：** ✅ 生产就绪
