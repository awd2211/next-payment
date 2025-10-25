# API 对齐快速修复指南

**文档目的**: 提供快速的代码修复示例和步骤  
**预计修复时间**: 2-3 小时  
**优先级**: 按列出的顺序执行

---

## 问题 1: Accounting Service 路径错误 (🔴 高优先级)

### 问题描述
- 前端: `/accounting/entries`, `/accounting/balances` 等
- 后端: 可能注册为 `/api/v1/accounting/...`

### 验证步骤

```bash
cd /home/eric/payment/backend/services/accounting-service

# 检查路由注册
grep -n "api :=" cmd/main.go
grep -n "accountHandler" cmd/main.go
grep -A 10 "RegisterRoutes" internal/handler/account_handler.go
```

### 修复方案 A: 前端路径修复 (推荐)

**文件**: `frontend/admin-portal/src/services/accountingService.ts`

```typescript
// 修改前
const api = axios.create({
  baseURL: '/api/v1',  // 使用默认前缀
})

export const accountingService = {
  listEntries: (params) => {
    return request.get('/accounting/entries', { params })  // ❌ 缺少 /api/v1
  },
  // ...
}

// 修改后
export const accountingService = {
  listEntries: (params) => {
    return request.get('/api/v1/accounting/entries', { params })  // ✅ 完整路径
  },
  getEntry: (id) => {
    return request.get(`/api/v1/accounting/entries/${id}`)
  },
  createEntry: (data) => {
    return request.post('/api/v1/accounting/entries', data)
  },
  listBalances: (params) => {
    return request.get('/api/v1/accounting/balances', { params })
  },
  getLedger: (params) => {
    return request.get('/api/v1/accounting/ledger', { params })
  },
  getGeneralLedger: (params) => {
    return request.get('/api/v1/accounting/general-ledger', { params })
  },
  getSummary: (params) => {
    return request.get('/api/v1/accounting/summary', { params })
  },
  getBalanceSheet: (params) => {
    return request.get('/api/v1/accounting/balance-sheet', { params })
  },
  getIncomeStatement: (params) => {
    return request.get('/api/v1/accounting/income-statement', { params })
  },
  getCashFlow: (params) => {
    return request.get('/api/v1/accounting/cash-flow', { params })
  },
  closeMonth: (params) => {
    return request.post('/api/v1/accounting/close-month', params)
  },
  getChartOfAccounts: () => {
    return request.get('/api/v1/accounting/chart-of-accounts')
  },
}
```

### 修复方案 B: 后端路由修复 (如果路由确实有问题)

**文件**: `backend/services/accounting-service/internal/handler/account_handler.go`

检查 `RegisterRoutes` 方法:

```go
// 确保路由注册正确
func (h *AccountHandler) RegisterRoutes(r *gin.RouterGroup) {
  accounting := r.Group("/accounting")  // 相对路由，会拼接为 /api/v1/accounting
  {
    accounting.GET("/entries", h.ListEntries)
    accounting.GET("/entries/:id", h.GetEntry)
    accounting.POST("/entries", h.CreateEntry)
    accounting.GET("/balances", h.ListBalances)
    accounting.GET("/ledger", h.GetLedger)
    // ...
  }
}
```

**注意**: 如果 main.go 中的注册是这样:
```go
api := application.Router.Group("/api/v1")
accountingHandler.RegisterRoutes(api)
```

那么路由会自动拼接为 `/api/v1/accounting/...`, 前端需要调整。

---

## 问题 2: Channel 配置管理接口 (🔴 高优先级)

### 问题描述
前端需要创建/修改/删除渠道配置，但后端只有查询接口。

### 修复方案: 后端添加 CRUD 接口

**文件**: `backend/services/channel-adapter/internal/handler/channel_handler.go`

```go
// 在 RegisterRoutes 方法中添加

func (h *ChannelHandler) RegisterRoutes(router *gin.Engine) {
  api := router.Group("/api/v1")
  {
    // 现有的查询接口
    api.GET("/channel/config", h.ListChannelConfigs)
    api.GET("/channel/config/:channel", h.GetChannelConfig)
    
    // 添加创建接口
    api.POST("/channel/config", h.CreateChannelConfig)
    
    // 添加修改接口
    api.PUT("/channel/config/:id", h.UpdateChannelConfig)
    
    // 添加删除接口
    api.DELETE("/channel/config/:id", h.DeleteChannelConfig)
    
    // 添加启用/禁用接口
    api.PUT("/channel/config/:id/toggle", h.ToggleChannelConfig)
    
    // 添加测试接口
    api.POST("/channel/config/:id/test", h.TestChannelConfig)
    
    // 其他现有接口...
    api.POST("/webhooks/stripe", h.HandleStripeWebhook)
    api.POST("/webhooks/paypal", h.HandlePayPalWebhook)
  }
}

// 实现新处理器方法
// @Summary 创建渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/channel/config [post]
func (h *ChannelHandler) CreateChannelConfig(c *gin.Context) {
  var req CreateChannelConfigRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  
  // 调用 service 创建配置
  config, err := h.channelService.CreateConfig(c.Request.Context(), &req)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusCreated, gin.H{"data": config})
}

// @Summary 更新渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/channel/config/:id [put]
func (h *ChannelHandler) UpdateChannelConfig(c *gin.Context) {
  id := c.Param("id")
  var req UpdateChannelConfigRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  
  config, err := h.channelService.UpdateConfig(c.Request.Context(), id, &req)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusOK, gin.H{"data": config})
}

// @Summary 删除渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/channel/config/:id [delete]
func (h *ChannelHandler) DeleteChannelConfig(c *gin.Context) {
  id := c.Param("id")
  
  err := h.channelService.DeleteConfig(c.Request.Context(), id)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// @Summary 切换渠道配置启用状态
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/channel/config/:id/toggle [put]
func (h *ChannelHandler) ToggleChannelConfig(c *gin.Context) {
  id := c.Param("id")
  var req ToggleConfigRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  
  config, err := h.channelService.ToggleConfig(c.Request.Context(), id, req.IsEnabled)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusOK, gin.H{"data": config})
}

// @Summary 测试渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/channel/config/:id/test [post]
func (h *ChannelHandler) TestChannelConfig(c *gin.Context) {
  id := c.Param("id")
  var req TestChannelConfigRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  
  result, err := h.channelService.TestConfig(c.Request.Context(), id, &req)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusOK, gin.H{"data": result})
}

// 请求结构体
type CreateChannelConfigRequest struct {
  Channel     string                 `json:"channel" binding:"required"`
  Name        string                 `json:"name" binding:"required"`
  Description string                 `json:"description"`
  Config      map[string]interface{} `json:"config" binding:"required"`
  IsEnabled   bool                   `json:"is_enabled"`
}

type UpdateChannelConfigRequest struct {
  Name        string                 `json:"name"`
  Description string                 `json:"description"`
  Config      map[string]interface{} `json:"config"`
  IsEnabled   *bool                  `json:"is_enabled"`
}

type ToggleConfigRequest struct {
  IsEnabled bool `json:"is_enabled" binding:"required"`
}

type TestChannelConfigRequest struct {
  PaymentAmount int64  `json:"payment_amount" binding:"required"`
  Currency      string `json:"currency" binding:"required"`
  // 其他测试参数
}
```

### 前端服务修改 (无需修改，等后端实现)

```typescript
// frontend/admin-portal/src/services/channelService.ts
export const channelService = {
  // ... 现有方法
  
  // 创建渠道 (后端实现后可用)
  create: (data: CreateChannelRequest) => {
    return request.post('/channels', data)  // 已经定义好，等后端支持
  },
  
  // 更新渠道 (后端实现后可用)
  update: (id: string, data: UpdateChannelRequest) => {
    return request.put(`/channels/${id}`, data)  // 已经定义好，等后端支持
  },
  
  // 删除渠道 (后端实现后可用)
  delete: (id: string) => {
    return request.delete(`/channels/${id}`)  // 已经定义好，等后端支持
  },
}
```

---

## 问题 3: Withdrawal/Settlement 操作命名不一致 (🟠 中优先级)

### 问题描述
- 前端: `process`, `complete`
- 后端: `execute`

### 修复方案: 在后端添加路由别名 (推荐)

**文件**: `backend/services/withdrawal-service/internal/handler/withdrawal_handler.go`

```go
// 修改 RegisterRoutes 方法

func (h *WithdrawalHandler) RegisterRoutes(r *gin.Engine) {
  api := r.Group("/api/v1")
  {
    withdrawals := api.Group("/withdrawals")
    {
      withdrawals.POST("", h.CreateWithdrawal)
      withdrawals.GET("", h.ListWithdrawals)
      withdrawals.GET("/:id", h.GetWithdrawal)
      withdrawals.POST("/:id/approve", h.ApproveWithdrawal)
      withdrawals.POST("/:id/reject", h.RejectWithdrawal)
      
      // 原始接口: execute
      withdrawals.POST("/:id/execute", h.ExecuteWithdrawal)
      
      // 添加别名: process -> execute
      withdrawals.POST("/:id/process", h.ExecuteWithdrawal)
      
      withdrawals.POST("/:id/cancel", h.CancelWithdrawal)
      withdrawals.GET("/reports", h.GetWithdrawalReport)
    }
    
    // ... 银行账户路由
  }
}
```

**文件**: `backend/services/settlement-service/internal/handler/settlement_handler.go`

```go
func (h *SettlementHandler) RegisterRoutes(r *gin.Engine) {
  api := r.Group("/api/v1")
  {
    settlements := api.Group("/settlements")
    {
      settlements.POST("", h.CreateSettlement)
      settlements.GET("", h.ListSettlements)
      settlements.GET("/:id", h.GetSettlement)
      settlements.POST("/:id/approve", h.ApproveSettlement)
      settlements.POST("/:id/reject", h.RejectSettlement)
      
      // 原始接口: execute
      settlements.POST("/:id/execute", h.ExecuteSettlement)
      
      // 添加别名: complete -> execute
      settlements.POST("/:id/complete", h.ExecuteSettlement)
      
      settlements.GET("/reports", h.GetSettlementReport)
    }
  }
}
```

### 方案 B: 修改前端路径 (备选)

```typescript
// frontend/admin-portal/src/services/withdrawalService.ts

export const withdrawalService = {
  // 改为调用 execute 而不是 process
  process: (id: string, data: any) => {
    return request.post(`/withdrawals/${id}/execute`, data)  // 修改为 /execute
  },
  
  // 添加 execute 方法
  execute: (id: string, data: any) => {
    return request.post(`/withdrawals/${id}/execute`, data)
  },
  
  // 移除或保留 complete
  complete: (id: string, data: any) => {
    // 方案B: 修改为调用 execute
    return request.post(`/withdrawals/${id}/execute`, data)
  },
}
```

---

## 问题 4: KYC 路径前缀不一致 (🟠 中优先级)

### 问题描述
- 前端: `/kyc/applications`, `/kyc/stats`
- 后端: `/documents`, `/statistics`

### 修复方案 A: 后端添加别名路由 (推荐)

**文件**: `backend/services/kyc-service/internal/handler/kyc_handler.go`

```go
func (h *KYCHandler) RegisterRoutes(r *gin.Engine) {
  api := r.Group("/api/v1")
  {
    // 原始路由（保留向后兼容）
    documents := api.Group("/documents")
    {
      documents.POST("", h.SubmitDocument)
      documents.GET("", h.ListDocuments)
      documents.GET("/:id", h.GetDocument)
      documents.POST("/:id/approve", h.ApproveDocument)
      documents.POST("/:id/reject", h.RejectDocument)
    }
    
    // 添加别名路由（前端期望的路径）
    applications := api.Group("/kyc/applications")
    {
      applications.POST("", h.SubmitDocument)
      applications.GET("", h.ListDocuments)
      applications.GET("/:id", h.GetDocument)
      applications.POST("/:id/approve", h.ApproveDocument)
      applications.POST("/:id/reject", h.RejectDocument)
      
      // 添加前端需要的 reviewing 状态
      applications.POST("/:id/reviewing", h.SetDocumentReviewing)
    }
    
    // 现有资质接口
    qualifications := api.Group("/qualifications")
    {
      qualifications.POST("", h.SubmitQualification)
      // ... 其他资质接口
    }
    
    // ... 其他接口
    
    // 原始 stats 路由
    api.GET("/statistics", h.GetKYCStatistics)
    
    // 添加别名路由
    api.GET("/kyc/stats", h.GetKYCStatistics)
  }
}

// 新方法：设置文档为审核中
func (h *KYCHandler) SetDocumentReviewing(c *gin.Context) {
  id := c.Param("id")
  
  // 调用 service 更新状态为 reviewing
  doc, err := h.kycService.SetDocumentReviewing(c.Request.Context(), id)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  
  c.JSON(http.StatusOK, gin.H{"data": doc})
}
```

### 修复方案 B: 修改前端调用路径

```typescript
// frontend/admin-portal/src/services/kycService.ts

export const kycService = {
  // 改为调用后端实际的路径
  list: (params) => {
    return request.get('/documents', { params })  // 改为 /documents
  },
  
  getById: (id) => {
    return request.get(`/documents/${id}`)  // 改为 /documents/:id
  },
  
  approve: (id, data) => {
    return request.post(`/documents/${id}/approve`, data)  // 改为 /documents/:id/approve
  },
  
  reject: (id, data) => {
    return request.post(`/documents/${id}/reject`, data)  // 改为 /documents/:id/reject
  },
  
  getStats: () => {
    return request.get('/statistics')  // 改为 /statistics
  },
}
```

---

## 问题 5: Merchant Limits 路径完全不匹配 (🟠 中优先级)

### 问题描述
- 前端: `/api/v1/admin/merchant-limits`
- 后端: `/api/v1/limits`

### 修复方案: 后端重新注册路由

**文件**: `backend/services/merchant-limit-service/cmd/main.go`

```go
func main() {
  // ... 其他初始化代码 ...
  
  // 注册路由时添加 /admin 前缀
  api := application.Router.Group("/api/v1")
  {
    // 添加 /admin 前缀
    adminGroup := api.Group("/admin")
    limitHandler.RegisterRoutes(adminGroup)
  }
  
  // 或者直接在 RegisterRoutes 中处理
  // limitHandler.RegisterRoutes(api, "/admin")
  
  if err := application.RunWithGracefulShutdown(); err != nil {
    logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
  }
}
```

**文件**: `backend/services/merchant-limit-service/internal/handler/limit_handler.go`

```go
// 修改 RegisterRoutes 方法以支持前缀

func (h *LimitHandler) RegisterRoutes(router *gin.RouterGroup, prefix ...string) {
  // 支持可选的前缀
  path := ""
  if len(prefix) > 0 {
    path = prefix[0]
  }
  
  // 使用动态路径
  tiers := router.Group(path + "/merchant-limits/tiers")
  {
    tiers.GET("", h.ListTiers)
    tiers.POST("", h.CreateTier)
    tiers.PUT("/:id", h.UpdateTier)
    tiers.DELETE("/:id", h.DeleteTier)
  }
  
  limits := router.Group(path + "/merchant-limits")
  {
    limits.GET("", h.ListLimits)
    limits.GET("/:merchantId", h.GetLimit)
    limits.POST("/:merchantId", h.UpdateLimit)
    limits.PUT("/:merchantId", h.UpdateLimit)
    limits.GET("/:merchantId/usage", h.GetLimitUsage)
    limits.GET("/:merchantId/history", h.GetLimitHistory)
    limits.POST("/:merchantId/reset", h.ResetLimit)
    limits.GET("/alerts", h.ListAlerts)
    limits.GET("/:merchantId/alert-config", h.GetAlertConfig)
    limits.PUT("/:merchantId/alert-config", h.UpdateAlertConfig)
    limits.POST("/batch-update", h.BatchUpdateLimits)
    limits.GET("/export", h.ExportLimits)
    limits.GET("/system-stats", h.GetSystemStats)
    limits.GET("/templates", h.ListTemplates)
    limits.POST("/:merchantId/apply-template", h.ApplyTemplate)
  }
}

// 在 cmd/main.go 中调用时
limitHandler.RegisterRoutes(api)  // 不传前缀
// 或
limitHandler.RegisterRoutes(api, "/admin")  // 传前缀
```

---

## 问题 6: Dispute 和 Reconciliation 路径前缀 (🟠 中优先级)

### 问题描述
- 前端: `/admin/disputes`, `/admin/reconciliation`
- 后端: `/disputes`, `/reconciliation`

### 修复方案: 后端添加别名或前端修改

**选项 A: 后端在 dispute-service 中添加别名**

```go
// backend/services/dispute-service/internal/handler/dispute_handler.go

func (h *DisputeHandler) RegisterRoutes(router *gin.RouterGroup) {
  // 原始路由
  disputes := router.Group("/disputes")
  {
    disputes.GET("", h.ListDisputes)
    disputes.GET("/:id", h.GetDispute)
    disputes.POST("/:id/resolve", h.ResolveDispute)
    disputes.GET("/:disputeId/evidence", h.ListEvidence)
    disputes.POST("/:disputeId/evidence", h.UploadEvidence)
    disputes.GET("/:disputeId/evidence/:evidenceId/download", h.DownloadEvidence)
    disputes.GET("/export", h.ExportDisputes)
    disputes.GET("/stats", h.GetDisputeStats)
  }
  
  // 添加 /admin 前缀的别名路由
  adminDisputes := router.Group("/admin/disputes")
  {
    adminDisputes.GET("", h.ListDisputes)
    adminDisputes.GET("/:id", h.GetDispute)
    adminDisputes.POST("/:id/resolve", h.ResolveDispute)
    adminDisputes.GET("/:disputeId/evidence", h.ListEvidence)
    adminDisputes.POST("/:disputeId/evidence", h.UploadEvidence)
    adminDisputes.GET("/:disputeId/evidence/:evidenceId/download", h.DownloadEvidence)
    adminDisputes.GET("/export", h.ExportDisputes)
    adminDisputes.GET("/stats", h.GetDisputeStats)
  }
}
```

**选项 B: 在 dispute-service 的 cmd/main.go 中修改路由组前缀**

```go
// backend/services/dispute-service/cmd/main.go

func main() {
  // ... 初始化代码 ...
  
  api := application.Router.Group("/api/v1")
  {
    // 添加 /admin 前缀
    adminAPI := api.Group("/admin")
    disputeHandler.RegisterRoutes(adminAPI)
  }
  
  // 或同时支持两种路径
  api.Group("").Group("/disputes").Use()  // 保持兼容性
  
  if err := application.RunWithGracefulShutdown(); err != nil {
    logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
  }
}
```

### 对 Reconciliation Service 应用相同的修复

```go
// backend/services/reconciliation-service/internal/handler/reconciliation_handler.go

func (h *ReconciliationHandler) RegisterRoutes(router *gin.RouterGroup) {
  // 原始路由
  reconciliation := router.Group("/reconciliation")
  {
    reconciliation.GET("", h.ListReconciliation)
    reconciliation.GET("/:id", h.GetReconciliation)
    reconciliation.POST("", h.CreateReconciliation)
    // ... 其他方法
  }
  
  // 添加别名
  adminReconciliation := router.Group("/admin/reconciliation")
  {
    adminReconciliation.GET("", h.ListReconciliation)
    adminReconciliation.GET("/:id", h.GetReconciliation)
    adminReconciliation.POST("", h.CreateReconciliation)
    // ... 其他方法
  }
}
```

---

## 测试修复 (修复完成后)

### 测试 Accounting Service

```bash
# 测试 Accounting 接口
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:40001/api/v1/accounting/entries

curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:40001/api/v1/accounting/balances
```

### 测试 Channel 创建

```bash
curl -X POST http://localhost:40005/api/v1/channel/config \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "stripe",
    "name": "Stripe Production",
    "config": {
      "api_key": "sk_live_xxx",
      "secret": "xxx"
    }
  }'
```

### 测试 Withdrawal 别名

```bash
# 使用 /execute
curl -X POST http://localhost:40014/api/v1/withdrawals/123/execute \
  -H "Authorization: Bearer TOKEN"

# 使用 /process (别名)
curl -X POST http://localhost:40014/api/v1/withdrawals/123/process \
  -H "Authorization: Bearer TOKEN"
```

### 前端集成测试

```typescript
// 在前端添加临时测试
import { accountingService } from '@/services/accountingService'

async function testAccounting() {
  try {
    const response = await accountingService.listEntries({ page: 1, page_size: 10 })
    console.log('✅ Accounting API 工作正常:', response)
  } catch (err) {
    console.error('❌ Accounting API 失败:', err)
  }
}

testAccounting()
```

---

## 修复清单

### 立即执行 (第一阶段)

- [ ] 验证 Accounting Service 路由注册
- [ ] 修改 accountingService.ts 中的路径为 `/api/v1/accounting/...`
- [ ] 在 channel-adapter 中添加 POST/PUT/DELETE `/channel/config` 接口
- [ ] 验证修改是否生效

### 短期执行 (第二阶段)

- [ ] 在 withdrawal-service 中添加 `/process` 别名路由
- [ ] 在 settlement-service 中添加 `/complete` 别名路由
- [ ] 在 kyc-service 中添加 `/kyc/applications` 别名路由
- [ ] 修改 merchant-limit-service 的路由前缀为 `/admin/merchant-limits`
- [ ] 在 dispute-service 和 reconciliation-service 中添加 `/admin/...` 别名

### 可选 (第三阶段)

- [ ] 实现缺失的 API (retry, stats 等)
- [ ] 添加 webhook 管理接口
- [ ] 优化路由设计，避免未来的不一致

---

## 验证清单

修复完成后逐项验证:

### 后端验证

```bash
# 1. 检查 Accounting Service 编译
cd /home/eric/payment/backend/services/accounting-service
go build -o /tmp/accounting-service ./cmd/main.go

# 2. 检查 Channel Adapter 编译
cd /home/eric/payment/backend/services/channel-adapter
go build -o /tmp/channel-adapter ./cmd/main.go

# 3. 检查 Withdrawal Service 编译
cd /home/eric/payment/backend/services/withdrawal-service
go build -o /tmp/withdrawal-service ./cmd/main.go

# 4. 全量编译检查
cd /home/eric/payment/backend
make build
```

### 前端验证

```bash
# 1. 修复 TypeScript 类型错误
cd /home/eric/payment/frontend/admin-portal
npm run build

# 2. 测试开发服务器
npm run dev
```

---

*预计总修复时间: 2-3 小时*  
*建议按照优先级顺序执行，每个修复后进行测试*
