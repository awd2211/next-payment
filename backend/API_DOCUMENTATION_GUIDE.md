# API Documentation Guide

## Overview

This document provides a comprehensive guide to the Swagger/OpenAPI documentation implemented across all microservices in the payment platform.

## Table of Contents

- [Quick Start](#quick-start)
- [Accessing Swagger UI](#accessing-swagger-ui)
- [Documentation Coverage](#documentation-coverage)
- [Adding API Documentation](#adding-api-documentation)
- [Regenerating Documentation](#regenerating-documentation)
- [Best Practices](#best-practices)

---

## Quick Start

### Prerequisites

Install the Swagger CLI tool:

```bash
cd backend
make install-swagger
```

This installs `swag` CLI at `~/go/bin/swag`.

### Generate Documentation

Generate Swagger documentation for all services:

```bash
cd backend
make swagger-docs
```

This command:
1. Scans all service directories
2. Parses Swagger annotations in `main.go` and handler files
3. Generates `docs.go`, `swagger.json`, and `swagger.yaml` in each service's `api-docs/` directory
4. Shows access URLs for each service's Swagger UI

---

## Accessing Swagger UI

Once services are running, access the interactive Swagger UI at:

| Service | Swagger UI URL | Port |
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

### Alternative Formats

**JSON Specification:**
```
http://localhost:{PORT}/swagger/swagger.json
```

**YAML Specification:**
```
http://localhost:{PORT}/swagger/swagger.yaml
```

---

## Documentation Coverage

### ✅ Fully Documented Services

| Service | Endpoints | Coverage | Notes |
|---------|-----------|----------|-------|
| **Admin Service** | 50+ | 95% | User management, RBAC, system config, audit logs |
| **Merchant Service** | 40+ | 95% | Merchant CRUD, KYC, settlement, dashboard |
| **Payment Gateway** | 10 | 100% | ✨ **New**: Full payment & refund flow, webhooks |
| **Order Service** | 15 | 80% | ✨ **New**: Order CRUD, status updates, statistics |
| **Channel Adapter** | 12 | 75% | Payment channel operations, exchange rates |
| **Notification Service** | 20 | 70% | Email, SMS, webhook notifications |
| **Merchant Auth Service** | 15 | 90% | API keys, merchant authentication |
| **KYC Service** | 12 | 85% | Document submission, verification, compliance |
| **Withdrawal Service** | 10 | 80% | Withdrawal requests, bank accounts, approvals |

### ⚠️ Minimal Documentation (Template Only)

| Service | Status | Priority |
|---------|--------|----------|
| Risk Service | ⚠️ Empty | Medium |
| Accounting Service | ⚠️ Empty | Medium |
| Analytics Service | ⚠️ Empty | Low |
| Config Service | ⚠️ Empty | Low |
| Cashier Service | ⚠️ Empty | Low |
| Settlement Service | ⚠️ Empty | Medium |

---

## Adding API Documentation

### Step 1: Add Service-Level Metadata (main.go)

Add these comments **above** the `main()` function in `cmd/main.go`:

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

### Step 2: Import Swagger Packages

Add to your `main.go`:

```go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "payment-platform/your-service/api-docs" // Generated docs
)
```

### Step 3: Register Swagger Route

Add after router initialization:

```go
application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

### Step 4: Document Handler Functions

Add Swagger annotations above each handler function:

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

### Annotation Reference

| Annotation | Description | Example |
|------------|-------------|---------|
| `@Summary` | Short description (1 line) | `Create Payment` |
| `@Description` | Detailed description | `Create a new payment transaction with validation` |
| `@Tags` | Group endpoints | `Payments`, `Orders`, `Refunds` |
| `@Accept` | Request content type | `json`, `xml` |
| `@Produce` | Response content type | `json` |
| `@Security` | Security scheme | `BearerAuth` (JWT), `ApiKeyAuth` |
| `@Param` | Parameter definition | `paymentNo path string true "Payment number"` |
| `@Success` | Success response | `200 {object} Response` |
| `@Failure` | Error response | `400 {object} Response` |
| `@Router` | Route path and method | `/payments [post]` |

### Parameter Types

```go
// Path parameter
//	@Param	id	path	string	true	"Record ID"

// Query parameter
//	@Param	page		query	int		false	"Page number"	default(1)
//	@Param	page_size	query	int		false	"Page size"		default(20)
//	@Param	status		query	string	false	"Filter by status"

// Body parameter
//	@Param	request	body	service.CreateInput	true	"Request body"

// Header parameter
//	@Param	X-Request-ID	header	string	false	"Request ID"
```

---

## Regenerating Documentation

### Regenerate All Services

```bash
cd backend
make swagger-docs
```

### Regenerate Single Service

```bash
cd backend/services/payment-gateway
~/go/bin/swag init -g cmd/main.go -o ./api-docs --parseDependency --parseInternal
```

### When to Regenerate

Regenerate Swagger docs when:
- ✅ Adding new API endpoints
- ✅ Changing request/response structures
- ✅ Updating parameter descriptions
- ✅ Modifying authentication schemes
- ✅ Changing route paths or HTTP methods

**Important:** Always run `make swagger-docs` before committing API changes.

---

## Best Practices

### 1. Consistent Naming

Use consistent naming across all services:

**Tags:**
- Use plural nouns: `Payments`, `Orders`, `Refunds`
- Group related endpoints under the same tag

**Summaries:**
- Use imperative mood: `Create Payment`, `Get Order`, `Update Status`
- Keep under 50 characters

**Descriptions:**
- Provide context and business logic
- Mention validations and constraints
- Include example use cases

### 2. Complete Parameter Documentation

Always document:
- ✅ Parameter name and type
- ✅ Required vs optional
- ✅ Default values (for optional parameters)
- ✅ Validation rules (min/max, format)
- ✅ Example values

**Example:**
```go
//	@Param	amount		query	int		true	"Payment amount in cents"		minimum(1)
//	@Param	currency	query	string	true	"Currency code (USD/EUR/CNY)"
//	@Param	page		query	int		false	"Page number"					default(1)		minimum(1)
//	@Param	page_size	query	int		false	"Items per page"				default(20)		minimum(1)	maximum(100)
```

### 3. Document All Status Codes

Document all possible HTTP status codes:

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

### 4. Use Response Models

Define clear response structures:

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

### 5. Document Authentication

Always specify security requirements:

```go
//	@Security	BearerAuth
```

For endpoints that don't require authentication:

```go
//	@Security	none
```

### 6. Add Examples

Use `example` tags in struct definitions:

```go
type CreatePaymentInput struct {
    Amount        int64  `json:"amount" binding:"required,gt=0" example:"10000"`
    Currency      string `json:"currency" binding:"required" example:"USD"`
    CustomerEmail string `json:"customer_email" binding:"email" example:"customer@example.com"`
}
```

### 7. Group by Tags

Organize endpoints logically:

**Payment Gateway:**
- `Payments` - Payment operations
- `Refunds` - Refund operations
- `Webhooks` - Callback handling

**Order Service:**
- `Orders` - Order management
- `Statistics` - Order analytics

**Merchant Service:**
- `Merchants` - Merchant CRUD
- `KYC` - KYC documents
- `Settlement` - Settlement operations

---

## Common Issues

### Issue 1: Cannot find type definition

**Error:**
```
ParseComment error: cannot find type definition: errors.SuccessResponse
```

**Solution:**
Use local struct types instead of external package types:
```go
//	@Success	200	{object}	Response  // ✅ Correct
//	@Success	200	{object}	errors.SuccessResponse  // ❌ Wrong
```

### Issue 2: Docs not updating

**Cause:** Generated files cached

**Solution:**
```bash
# Delete api-docs directory and regenerate
rm -rf api-docs/
make swagger-docs
```

### Issue 3: Import cycle error

**Cause:** Circular dependencies in package imports

**Solution:**
Use `--parseDependency` flag (already in Makefile):
```bash
swag init -g cmd/main.go -o ./api-docs --parseDependency --parseInternal
```

---

## Example: Full Payment Endpoint Documentation

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

## API Testing with Swagger UI

### 1. Authenticate

For endpoints requiring JWT authentication:

1. Click the **Authorize** button (lock icon) in Swagger UI
2. Enter: `Bearer YOUR_JWT_TOKEN`
3. Click **Authorize**
4. Click **Close**

### 2. Try Endpoints

1. Expand an endpoint
2. Click **Try it out**
3. Fill in required parameters
4. Click **Execute**
5. View response

### 3. View Models

Click **Schemas** at the bottom to view all request/response models.

---

## File Structure

After generating docs, each service has:

```
service-name/
├── cmd/
│   └── main.go              # Swagger metadata annotations
├── internal/
│   └── handler/
│       └── *_handler.go     # Endpoint annotations
└── api-docs/                # Generated by swag
    ├── docs.go              # Go code (imported by main.go)
    ├── swagger.json         # OpenAPI 2.0 JSON spec
    └── swagger.yaml         # OpenAPI 2.0 YAML spec
```

**Note:** The `api-docs/` directory is auto-generated. Do not edit these files directly.

---

## Integration with CI/CD

Add to your CI/CD pipeline:

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

## Next Steps

### Priority 1: Complete Documentation (New Endpoints)

- ✅ **Payment Gateway** - All payment and refund endpoints documented
- ✅ **Order Service** - Core order management endpoints documented
- ⏳ **Risk Service** - Risk assessment APIs (needs documentation)
- ⏳ **Accounting Service** - Ledger and accounting APIs (needs documentation)

### Priority 2: Enhanced Documentation

- Add more detailed descriptions
- Include error code references
- Add request/response examples
- Document rate limiting behavior
- Add webhook signature verification details

### Priority 3: API Versioning

- Plan API versioning strategy (v1, v2)
- Document breaking changes
- Maintain backward compatibility

---

## Resources

- **Swaggo Documentation**: https://github.com/swaggo/swag
- **OpenAPI 2.0 Spec**: https://swagger.io/specification/v2/
- **Swagger UI**: https://swagger.io/tools/swagger-ui/
- **JSON Schema**: https://json-schema.org/

---

## Support

For questions or issues with API documentation:
- Create an issue in the project repository
- Contact: support@payment-platform.com
- Slack: #api-documentation

---

**Last Updated:** 2025-10-24
**Version:** 1.0
**Status:** ✅ Production Ready
