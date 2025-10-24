# Swagger Quick Reference Card

## TL;DR

```bash
# Install swag CLI
make install-swagger

# Generate all Swagger docs
make swagger-docs

# Access Swagger UI
http://localhost:{SERVICE_PORT}/swagger/index.html
```

---

## Common Annotations

### Service-Level (cmd/main.go)

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

### Endpoint-Level (handlers)

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

## Parameter Types

| Type | Syntax | Example |
|------|--------|---------|
| **Path** | `@Param name path type required "desc"` | `@Param id path string true "User ID"` |
| **Query** | `@Param name query type required "desc"` | `@Param page query int false "Page number" default(1)` |
| **Body** | `@Param name body type required "desc"` | `@Param request body CreateInput true "Request"` |
| **Header** | `@Param name header type required "desc"` | `@Param X-API-Key header string true "API Key"` |

---

## Data Types

| Swagger Type | Go Type |
|--------------|---------|
| `integer` | `int`, `int32`, `int64` |
| `number` | `float32`, `float64` |
| `string` | `string` |
| `boolean` | `bool` |
| `array` | `[]Type` |
| `object` | `struct`, `map` |

---

## Response Examples

### Simple Response
```go
//	@Success	200	{object}	Response
```

### Response with Nested Data
```go
//	@Success	200	{object}	Response{data=model.User}
```

### Array Response
```go
//	@Success	200	{array}		model.User
```

### Paginated Response
```go
//	@Success	200	{object}	PageResponse{list=[]model.User}
```

---

## Common Tags

| Tag | Use Case |
|-----|----------|
| `Payments` | Payment operations |
| `Orders` | Order management |
| `Refunds` | Refund processing |
| `Merchants` | Merchant management |
| `Users` | User administration |
| `Statistics` | Analytics & reports |
| `Webhooks` | Callback handlers |

---

## Validation Constraints

```go
//	@Param	amount	query	int		true	"Amount"	minimum(1)	maximum(999999)
//	@Param	email	query	string	true	"Email"		format(email)
//	@Param	status	query	string	false	"Status"	Enums(pending, success, failed)
//	@Param	page	query	int		false	"Page"		default(1)	minimum(1)
```

---

## Security

### JWT Authentication
```go
//	@Security	BearerAuth
```

### No Authentication
```go
//	@Security	none
```

### API Key
```go
//	@Security	ApiKeyAuth
```

---

## Struct Examples

### Example Tags in Structs
```go
type CreateInput struct {
    Amount   int64  `json:"amount" binding:"required" example:"10000"`
    Currency string `json:"currency" binding:"required" example:"USD"`
    Email    string `json:"email" binding:"email" example:"user@example.com"`
}
```

---

## Regeneration

### All Services
```bash
cd backend && make swagger-docs
```

### Single Service
```bash
cd backend/services/payment-gateway
~/go/bin/swag init -g cmd/main.go -o ./api-docs --parseDependency --parseInternal
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "Cannot find type" error | Use local types, not external package types |
| Docs not updating | Delete `api-docs/` and regenerate |
| Import cycle | Use `--parseDependency` flag (default in Makefile) |
| Missing models | Ensure struct is exported (starts with capital letter) |

---

## Service URLs

| Service | URL |
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

## Complete Example

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

**See also:** [API_DOCUMENTATION_GUIDE.md](API_DOCUMENTATION_GUIDE.md) for detailed guide
