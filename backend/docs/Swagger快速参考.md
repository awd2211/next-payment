# Swagger 快速参考卡

## 速查

```bash
# 安装 swag CLI
make install-swagger

# 生成所有 Swagger 文档
make swagger-docs

# 访问 Swagger UI
http://localhost:{SERVICE_PORT}/swagger/index.html
```

---

## 常用注解

### 服务级别（cmd/main.go）

```go
//	@title			Service Name API
//	@version		1.0
//	@description	Service description
//	@host			localhost:40XXX
//	@BasePath		/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in				header
//	@name			Authorization

func main() { }
```

### 端点级别（handlers）

```go
//	@Summary		Short description
//	@Description	Detailed description
//	@Tags			GroupName
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name	in		type	required	"description"
//	@Success		200		{object}	ResponseType
//	@Failure		400		{object}	ErrorType
//	@Router			/path [method]
func Handler(c *gin.Context) { }
```

---

## 参数类型

| 类型 | 语法 | 示例 |
|------|--------|---------|
| **Path** | `@Param name path type required "desc"` | `@Param id path string true "User ID"` |
| **Query** | `@Param name query type required "desc"` | `@Param page query int false "Page number" default(1)` |
| **Body** | `@Param name body type required "desc"` | `@Param request body CreateInput true "Request"` |
| **Header** | `@Param name header type required "desc"` | `@Param X-API-Key header string true "API Key"` |

---

## 数据类型

| Swagger 类型 | Go 类型 |
|--------------|---------|
| `integer` | `int`、`int32`、`int64` |
| `number` | `float32`、`float64` |
| `string` | `string` |
| `boolean` | `bool` |
| `array` | `[]Type` |
| `object` | `struct`、`map` |

---

## 响应示例

### 简单响应
```go
//	@Success	200	{object}	Response
```

### 嵌套数据响应
```go
//	@Success	200	{object}	Response{data=model.User}
```

### 数组响应
```go
//	@Success	200	{array}		model.User
```

### 分页响应
```go
//	@Success	200	{object}	PageResponse{list=[]model.User}
```

---

## 常用标签

| 标签 | 用例 |
|-----|----------|
| `Payments` | 支付操作 |
| `Orders` | 订单管理 |
| `Refunds` | 退款处理 |
| `Merchants` | 商户管理 |
| `Users` | 用户管理 |
| `Statistics` | 分析和报告 |
| `Webhooks` | 回调处理 |

---

## 验证约束

```go
//	@Param	amount	query	int		true	"Amount"	minimum(1)	maximum(999999)
//	@Param	email	query	string	true	"Email"		format(email)
//	@Param	status	query	string	false	"Status"	Enums(pending, success, failed)
//	@Param	page	query	int		false	"Page"		default(1)	minimum(1)
```

---

## 安全性

### JWT 认证
```go
//	@Security	BearerAuth
```

### 无认证
```go
//	@Security	none
```

### API 密钥
```go
//	@Security	ApiKeyAuth
```

---

## 结构示例

### 结构体中的示例标签
```go
type CreateInput struct {
    Amount   int64  `json:"amount" binding:"required" example:"10000"`
    Currency string `json:"currency" binding:"required" example:"USD"`
    Email    string `json:"email" binding:"email" example:"user@example.com"`
}
```

---

## 重新生成

### 所有服务
```bash
cd backend && make swagger-docs
```

### 单个服务
```bash
cd backend/services/payment-gateway
~/go/bin/swag init -g cmd/main.go -o ./api-docs --parseDependency --parseInternal
```

---

## 故障排除

| 问题 | 解决方案 |
|-------|----------|
| "Cannot find type" 错误 | 使用本地类型，而非外部包类型 |
| 文档未更新 | 删除 `api-docs/` 并重新生成 |
| 导入循环 | 使用 `--parseDependency` 标志（Makefile 中默认） |
| 缺失模型 | 确保结构体已导出（以大写字母开头） |

---

## 服务 URL

| 服务 | URL |
|---------|-----|
| Admin | http://localhost:40001/swagger/index.html |
| Merchant | http://localhost:40002/swagger/index.html |
| Payment Gateway | http://localhost:40003/swagger/index.html |
| Order | http://localhost:40004/swagger/index.html |
| Channel Adapter | http://localhost:40005/swagger/index.html |
| Risk | http://localhost:40006/swagger/index.html |
| Accounting | http://localhost:40007/swagger/index.html |
| Notification | http://localhost:40008/swagger/index.html |

---

## 完整示例

```go
// CreatePayment creates a new payment
//
//	@Summary		Create Payment
//	@Description	Create a new payment transaction
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreatePaymentInput	true	"Payment request"
//	@Success		200		{object}	Response{data=model.Payment}
//	@Failure		400		{object}	Response	"Bad request"
//	@Failure		401		{object}	Response	"Unauthorized"
//	@Failure		500		{object}	Response	"Internal error"
//	@Router			/payments [post]
func (h *Handler) CreatePayment(c *gin.Context) {
    // ...
}
```

---

**另请参阅：** [API_DOCUMENTATION_GUIDE.md](API_DOCUMENTATION_GUIDE.md) 详细指南
