# Merchant Service Refactoring Plan

## 问题诊断

当前 `merchant-service` 承担了 **11 个职责**，严重违反单一职责原则：

| 模型 | 当前位置 | 应归属服务 | 优先级 | 理由 |
|------|---------|-----------|-------|-----|
| Merchant | merchant-service | ✅ 保留 | - | 核心商户信息 |
| APIKey | merchant-service | merchant-auth-service | P0 | 认证域职责 |
| KYCDocument | merchant-service | kyc-service | P1 | 独立业务域 |
| BusinessQualification | merchant-service | kyc-service | P1 | 独立业务域 |
| SettlementAccount | merchant-service | settlement-service | P1 | 财务域职责 |
| MerchantFeeConfig | merchant-service | merchant-config-service | P2 | 配置域职责 |
| MerchantTransactionLimit | merchant-service | merchant-config-service | P2 | 配置域职责 |
| ChannelConfig | merchant-service | merchant-config-service | P2 | 配置域职责 |
| MerchantUser | merchant-service | merchant-team-service | P3 | 团队管理域 |
| MerchantContract | merchant-service | contract-service | P3 | 合同管理域 |

## Phase 1: 迁移 APIKey 到 merchant-auth-service (P0)

### 目标
- 将 `APIKey` 模型和相关逻辑从 merchant-service 迁移到 merchant-auth-service
- 修改 payment-gateway 的签名验证中间件调用新服务

### 步骤

#### 1.1 在 merchant-auth-service 中添加 APIKey 模型

**文件**: `backend/services/merchant-auth-service/internal/model/api_key.go`

```go
package model

import (
    "time"
    "github.com/google/uuid"
)

type APIKey struct {
    ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    MerchantID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`
    APIKey      string     `gorm:"type:varchar(64);unique;not null;index" json:"api_key"`
    APISecret   string     `gorm:"type:varchar(128);not null" json:"api_secret,omitempty"`
    Name        string     `gorm:"type:varchar(100)" json:"name"`
    Environment string     `gorm:"type:varchar(20);not null;index" json:"environment"`
    IsActive    bool       `gorm:"default:true" json:"is_active"`
    LastUsedAt  *time.Time `gorm:"type:timestamptz" json:"last_used_at"`
    ExpiresAt   *time.Time `gorm:"type:timestamptz" json:"expires_at"`
    CreatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
    UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

func (APIKey) TableName() string {
    return "api_keys"
}
```

#### 1.2 创建 APIKey 仓储层

**文件**: `backend/services/merchant-auth-service/internal/repository/api_key_repository.go`

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "payment-platform/merchant-auth-service/internal/model"
)

type APIKeyRepository interface {
    GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error)
    UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error
}

type apiKeyRepository struct {
    db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
    return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error) {
    var key model.APIKey
    err := r.db.WithContext(ctx).
        Where("api_key = ? AND is_active = ?", apiKey, true).
        First(&key).Error
    return &key, err
}

func (r *apiKeyRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&model.APIKey{}).
        Where("id = ?", id).
        Update("last_used_at", time.Now()).Error
}
```

#### 1.3 创建 APIKey 服务层

**文件**: `backend/services/merchant-auth-service/internal/service/api_key_service.go`

```go
package service

import (
    "context"
    "errors"
    "time"
    "github.com/google/uuid"
    "payment-platform/merchant-auth-service/internal/model"
    "payment-platform/merchant-auth-service/internal/repository"
    "github.com/payment-platform/pkg/crypto"
)

type APIKeyService interface {
    ValidateAPIKey(ctx context.Context, apiKey, signature, payload string) (*model.APIKey, error)
}

type apiKeyService struct {
    repo repository.APIKeyRepository
}

func NewAPIKeyService(repo repository.APIKeyRepository) APIKeyService {
    return &apiKeyService{repo: repo}
}

func (s *apiKeyService) ValidateAPIKey(ctx context.Context, apiKey, signature, payload string) (*model.APIKey, error) {
    // 1. 查询 API Key
    key, err := s.repo.GetByAPIKey(ctx, apiKey)
    if err != nil {
        return nil, errors.New("invalid api key")
    }

    // 2. 检查是否过期
    if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
        return nil, errors.New("api key expired")
    }

    // 3. 验证签名
    expectedSignature := crypto.HmacSHA256(payload, key.APISecret)
    if signature != expectedSignature {
        return nil, errors.New("invalid signature")
    }

    // 4. 更新最后使用时间（异步）
    go s.repo.UpdateLastUsedAt(context.Background(), key.ID)

    return key, nil
}
```

#### 1.4 添加 HTTP API 端点

**文件**: `backend/services/merchant-auth-service/internal/handler/api_key_handler.go`

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "payment-platform/merchant-auth-service/internal/service"
)

type APIKeyHandler struct {
    service service.APIKeyService
}

func NewAPIKeyHandler(service service.APIKeyService) *APIKeyHandler {
    return &APIKeyHandler{service: service}
}

// ValidateSignature 验证 API 签名（供 payment-gateway 调用）
func (h *APIKeyHandler) ValidateSignature(c *gin.Context) {
    var req struct {
        APIKey    string `json:"api_key" binding:"required"`
        Signature string `json:"signature" binding:"required"`
        Payload   string `json:"payload" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    key, err := h.service.ValidateAPIKey(c.Request.Context(), req.APIKey, req.Signature, req.Payload)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "valid":       true,
        "merchant_id": key.MerchantID,
        "environment": key.Environment,
    })
}

// RegisterRoutes 注册路由
func RegisterAPIKeyRoutes(r *gin.RouterGroup, handler *APIKeyHandler) {
    r.POST("/validate-signature", handler.ValidateSignature)
}
```

#### 1.5 修改 payment-gateway 签名中间件

**文件**: `backend/services/payment-gateway/internal/middleware/signature.go`

修改为调用 merchant-auth-service 的 HTTP API：

```go
package middleware

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/payment-platform/pkg/logger"
)

type SignatureMiddleware struct {
    authServiceURL string
}

func NewSignatureMiddleware(authServiceURL string) *SignatureMiddleware {
    return &SignatureMiddleware{authServiceURL: authServiceURL}
}

func (m *SignatureMiddleware) Verify() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-Api-Key")
        signature := c.GetHeader("X-Signature")

        if apiKey == "" || signature == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing api key or signature"})
            c.Abort()
            return
        }

        // 读取请求体
        bodyBytes, _ := io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

        // 调用 merchant-auth-service 验证
        reqBody := map[string]string{
            "api_key":   apiKey,
            "signature": signature,
            "payload":   string(bodyBytes),
        }
        reqJSON, _ := json.Marshal(reqBody)

        resp, err := http.Post(
            m.authServiceURL+"/api/v1/auth/validate-signature",
            "application/json",
            bytes.NewBuffer(reqJSON),
        )
        if err != nil || resp.StatusCode != http.StatusOK {
            logger.Error("signature validation failed", zap.Error(err))
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
            c.Abort()
            return
        }

        var result struct {
            Valid       bool   `json:"valid"`
            MerchantID  string `json:"merchant_id"`
            Environment string `json:"environment"`
        }
        json.NewDecoder(resp.Body).Decode(&result)

        c.Set("merchant_id", result.MerchantID)
        c.Set("environment", result.Environment)
        c.Next()
    }
}
```

#### 1.6 数据迁移脚本

**文件**: `backend/scripts/migrate_api_keys.sql`

```sql
-- 从 payment_merchant 复制 api_keys 表到 payment_merchant_auth

-- 1. 备份原表
CREATE TABLE payment_merchant.api_keys_backup AS
SELECT * FROM payment_merchant.api_keys;

-- 2. 在新数据库创建表
\c payment_merchant_auth

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    api_secret VARCHAR(128) NOT NULL,
    name VARCHAR(100),
    environment VARCHAR(20) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_keys_merchant_id ON api_keys(merchant_id);
CREATE INDEX idx_api_keys_api_key ON api_keys(api_key);
CREATE INDEX idx_api_keys_environment ON api_keys(environment);

-- 3. 复制数据
\c payment_merchant
COPY (SELECT * FROM api_keys) TO '/tmp/api_keys.csv' CSV HEADER;

\c payment_merchant_auth
COPY api_keys FROM '/tmp/api_keys.csv' CSV HEADER;

-- 4. 验证数据
SELECT COUNT(*) FROM payment_merchant.api_keys;
SELECT COUNT(*) FROM payment_merchant_auth.api_keys;

-- 5. 删除原表（谨慎！先验证新服务工作正常）
-- DROP TABLE payment_merchant.api_keys;
```

### 测试计划

#### 单元测试
```bash
cd backend/services/merchant-auth-service
go test ./internal/service -run TestValidateAPIKey
```

#### 集成测试
```bash
# 1. 启动 merchant-auth-service
PORT=40011 go run cmd/main.go

# 2. 测试签名验证 API
curl -X POST http://localhost:40011/api/v1/auth/validate-signature \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "test_key",
    "signature": "calculated_signature",
    "payload": "{\"amount\":100}"
  }'

# 3. 测试 payment-gateway 调用
curl -X POST http://localhost:40003/api/v1/payments \
  -H "X-Api-Key: test_key" \
  -H "X-Signature: calculated_signature" \
  -d '{"amount":100,"currency":"USD"}'
```

### 回滚方案

如果迁移失败：
1. 恢复 merchant-service 的 APIKey 代码
2. 恢复 payment-gateway 的原签名中间件
3. 从备份恢复 api_keys 表

```sql
\c payment_merchant
DROP TABLE IF EXISTS api_keys;
CREATE TABLE api_keys AS SELECT * FROM api_keys_backup;
```

---

## Phase 2: 创建 kyc-service (P1)

### 目标
- 创建独立的 KYC 服务
- 迁移 `KYCDocument` 和 `BusinessQualification`

### 步骤

#### 2.1 创建服务骨架
```bash
cd backend/services
mkdir -p kyc-service/{cmd,internal/{model,repository,service,handler}}
cd kyc-service
go mod init payment-platform/kyc-service
```

#### 2.2 复制模型文件
从 merchant-service 复制：
- `KYCDocument`
- `BusinessQualification`

#### 2.3 实现服务层
- KYC 文档上传/审核
- OCR 识别（集成第三方服务）
- 业务资质验证

#### 2.4 数据迁移
```sql
-- 迁移到 payment_kyc 数据库
```

---

## Phase 3: 迁移 SettlementAccount (P1)

### 目标
- 将 `SettlementAccount` 迁移到 settlement-service

### 步骤
（类似 Phase 1 的流程）

---

## Phase 4: 创建 merchant-config-service (P2)

### 目标
- 创建配置管理服务
- 迁移 `MerchantFeeConfig`, `MerchantTransactionLimit`, `ChannelConfig`

---

## Phase 5: 评估是否需要 merchant-team-service (P3)

### 评估标准
- 如果商户子账户功能复杂（SSO、细粒度权限），则独立
- 如果简单，可以保留在 merchant-service

---

## 迁移检查清单

每完成一个 Phase 后，检查：

- [ ] 新服务编译通过
- [ ] 数据迁移成功（行数一致）
- [ ] API 测试通过（Postman/curl）
- [ ] 旧服务删除相关代码
- [ ] 更新 CLAUDE.md 文档
- [ ] 更新架构图
- [ ] 监控指标正常（Prometheus）
- [ ] 日志输出正常
- [ ] 回滚脚本已测试

---

## 预期结果

### 迁移前
- merchant-service: 11 个模型, 2030 行代码

### 迁移后
- merchant-service: 1 个模型 (Merchant), ~300 行代码
- merchant-auth-service: 1 个模型 (APIKey)
- kyc-service: 2 个模型 (KYCDocument, BusinessQualification)
- settlement-service: 1 个模型 (SettlementAccount)
- merchant-config-service: 3 个模型 (FeeConfig, TransactionLimit, ChannelConfig)

**代码减少**: ~70% 🎉
**职责清晰**: ✅
**可维护性**: ⬆️⬆️⬆️
