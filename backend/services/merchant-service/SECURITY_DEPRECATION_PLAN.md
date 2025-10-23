# Merchant Service 安全功能下线计划

**日期**: 2025-10-23
**状态**: 已完成分析，准备执行
**原因**: 安全功能已迁移至独立的 `merchant-auth-service` (端口 8011)

---

## 📊 影响分析

### ✅ 可以安全移除的文件

以下文件可以完全删除，不会影响其他功能：

1. **Handler 层**
   - `internal/handler/security_handler.go` - Security HTTP 端点

2. **Service 层**
   - `internal/service/security_service.go` - 安全业务逻辑
   - `internal/service/auth_service.go` - 认证服务（**已注释，未使用**）
   - `internal/service/notification_service.go` - 通知服务（**未初始化，未使用**）

3. **Repository 层**
   - `internal/repository/security_repository.go` - 安全数据访问

4. **Model 层**
   - `internal/model/security.go` - 5个安全相关模型：
     - `TwoFactorAuth`
     - `LoginActivity`
     - `SecuritySettings`
     - `PasswordHistory`
     - `Session`

### 📌 需要修改的文件

#### 1. `cmd/main.go`

**需要移除的代码**：

```go
// Line 72: AutoMigrate中移除
&model.SecuritySettings{},
&model.TwoFactorAuth{},
&model.LoginActivity{},
&model.PasswordHistory{},
&model.Session{},

// Line 112-113: 移除SecurityRepository初始化
securityRepo := repository.NewSecurityRepository(database)

// Line 123: MerchantService构造函数中移除securityRepo参数
merchantService := service.NewMerchantService(merchantRepo, apiKeyRepo, securityRepo, jwtManager)
// 改为：
merchantService := service.NewMerchantService(merchantRepo, apiKeyRepo, jwtManager)

// Line 126-127: 移除SecurityService初始化
securityService := service.NewSecurityService(securityRepo, merchantRepo)

// Line 141: 移除已注释的AuthService（直接删除整行）
// authService := service.NewAuthService(merchantRepo, securityRepo, jwtManager)

// Line 147-148: 移除SecurityHandler初始化
securityHandler := handler.NewSecurityHandler(securityService)

// Line 194: 移除路由注册
securityHandler.RegisterRoutes(api, authMiddleware)
```

#### 2. `internal/service/merchant_service.go`

**需要修改的代码**：

```go
// Line 39: 移除字段
securityRepo repository.SecurityRepository

// Line 45-48: 修改构造函数
func NewMerchantService(
	merchantRepo repository.MerchantRepository,
	apiKeyRepo repository.APIKeyRepository,
	securityRepo repository.SecurityRepository,  // 删除这行
	jwtManager *auth.JWTManager,
) MerchantService {
	return &merchantService{
		merchantRepo: merchantRepo,
		apiKeyRepo:   apiKeyRepo,
		securityRepo: securityRepo,  // 删除这行
		jwtManager:   jwtManager,
	}
}
```

### ⚠️ 依赖关系分析

**✅ 无依赖风险**:
- `MerchantService` 虽然接受 `securityRepo` 参数，但**从未使用**
- `AuthService` 虽然依赖 `securityRepo`，但**已被注释，未初始化**
- `NotificationService` 虽然依赖 `securityRepo`，但**未初始化，未使用**

**✅ 路由独立**:
- 安全相关路由 (`/api/v1/security/*`) 完全独立
- 移除后不影响其他API端点

---

## 🚀 执行步骤

### Step 1: 备份当前代码
```bash
cd /home/eric/payment/backend/services/merchant-service
git diff > /tmp/merchant-service-before-cleanup.patch
```

### Step 2: 删除安全相关文件
```bash
rm internal/handler/security_handler.go
rm internal/service/security_service.go
rm internal/service/auth_service.go
rm internal/service/notification_service.go
rm internal/repository/security_repository.go
rm internal/model/security.go
```

### Step 3: 修改 `cmd/main.go`

移除以下内容：
1. AutoMigrate 中的5个安全模型
2. SecurityRepository 初始化
3. SecurityService 初始化
4. SecurityHandler 初始化
5. SecurityHandler 路由注册
6. AuthService 注释行（直接删除）

### Step 4: 修改 `internal/service/merchant_service.go`

移除：
1. `securityRepo` 字段
2. 构造函数中的 `securityRepo` 参数

### Step 5: 重新编译测试
```bash
cd /home/eric/payment/backend/services/merchant-service
go mod tidy
go build -o /tmp/merchant-service-clean ./cmd/main.go

# 测试服务启动
/tmp/merchant-service-clean
```

### Step 6: 验证功能
```bash
# 1. 健康检查
curl http://localhost:8002/health

# 2. 测试merchant相关功能（确保未受影响）
curl http://localhost:8002/api/v1/merchant

# 3. 确认安全端点已移除（应该返回404）
curl http://localhost:8002/api/v1/security/settings

# 4. 确认merchant-auth-service正常工作
curl http://localhost:8011/api/v1/security/settings -H "Authorization: Bearer <token>"
```

---

## 📋 验证清单

执行完成后，确认以下事项：

- [ ] merchant-service 成功编译
- [ ] merchant-service 启动无错误
- [ ] `/health` 端点正常
- [ ] `/api/v1/merchant/*` 端点正常工作
- [ ] `/api/v1/security/*` 端点返回 404（已移除）
- [ ] `merchant-auth-service` 的 `/api/v1/security/*` 正常工作
- [ ] 无编译警告或错误
- [ ] 代码可以通过 `go test ./...`

---

## 🔄 回滚方案

如果出现问题，可以快速回滚：

```bash
cd /home/eric/payment/backend/services/merchant-service
git apply /tmp/merchant-service-before-cleanup.patch
go build -o /tmp/merchant-service ./cmd/main.go
```

---

## 📝 预期结果

- **代码减少**: 约 500+ 行代码
- **文件减少**: 6 个文件
- **依赖简化**: 移除 SecurityRepository 依赖
- **职责分离**: merchant-service 专注于商户管理，安全功能由 merchant-auth-service 独立负责
- **维护性**: 更清晰的代码结构，更易于维护

---

## ✅ 完成标准

当以下条件全部满足时，视为完成：

1. ✅ 所有安全相关文件已删除
2. ✅ merchant-service 编译无错误
3. ✅ merchant-service 运行无错误
4. ✅ 核心功能（商户管理）正常工作
5. ✅ 安全端点已从 merchant-service 移除
6. ✅ merchant-auth-service 提供所有安全功能
7. ✅ 测试通过

---

**文档版本**: v1.0
**最后更新**: 2025-10-23
**执行人**: 待定
