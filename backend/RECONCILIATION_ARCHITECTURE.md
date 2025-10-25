# 对账系统分布式架构设计

## 📋 总体架构

对账系统是一个**跨服务的业务流程**,需要在多个微服务中协同实现:

```
┌─────────────────────────────────────────────────────────────────┐
│                      对账系统整体架构                              │
└─────────────────────────────────────────────────────────────────┘

1. accounting-service (核心对账引擎) ⭐
   ├─ 三方对账逻辑 (内部账 vs 渠道账单 vs 银行流水)
   ├─ 差异检测算法
   ├─ 对账报告生成
   └─ 调账执行

2. channel-adapter (渠道对账文件管理)
   ├─ 渠道账单下载 (Stripe/PayPal/Alipay)
   ├─ 账单文件解析
   ├─ 标准化数据格式
   └─ 渠道差异确认

3. settlement-service (结算对账)
   ├─ 结算单对账
   ├─ 提现对账
   └─ 商户账户余额核对

4. admin-service (差异处理工作流) ⭐
   ├─ 差异工单管理
   ├─ 工作流状态机
   ├─ SLA管理
   ├─ 审批流程
   └─ 后台管理界面

5. notification-service (对账通知)
   ├─ 对账完成通知
   ├─ 差异告警
   └─ SLA超时提醒
```

---

## 1️⃣ accounting-service (核心对账引擎)

**职责**: 对账核心逻辑、差异检测、调账执行

### 数据模型

```go
// internal/model/reconciliation.go
package model

import (
    "time"
    "github.com/google/uuid"
)

// 对账任务
type ReconciliationTask struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TaskDate        time.Time `gorm:"index;not null"` // 对账日期
    TaskType        string    `gorm:"size:50;not null"` // "daily", "weekly", "monthly"
    Status          string    `gorm:"size:50;not null"` // "pending", "running", "completed", "failed"

    // 对账范围
    StartTime       time.Time
    EndTime         time.Time

    // 统计信息
    TotalRecords    int64
    MatchedRecords  int64
    MismatchRecords int64
    MatchRate       float64

    // 差异汇总
    TotalDiscrepancies int
    CriticalCount      int
    HighCount          int
    MediumCount        int
    LowCount           int

    // 任务执行信息
    StartedAt       *time.Time
    CompletedAt     *time.Time
    ExecutionTime   int64  // 毫秒
    ErrorMessage    string `gorm:"type:text"`

    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// 对账记录(每笔交易的对账结果)
type ReconciliationRecord struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TaskID          uuid.UUID `gorm:"type:uuid;index;not null"`

    // 交易标识
    InternalOrderNo string    `gorm:"size:100;index;not null"`
    ChannelOrderNo  string    `gorm:"size:100;index"`
    MerchantID      uuid.UUID `gorm:"type:uuid;index"`

    // 对账结果
    Status          string    `gorm:"size:50;not null"` // "matched", "missing_channel", "missing_internal", "amount_mismatch", "status_mismatch"

    // 内部账信息
    InternalAmount  int64
    InternalStatus  string
    InternalCurrency string
    InternalTime    time.Time

    // 渠道账信息
    ChannelAmount   *int64
    ChannelStatus   *string
    ChannelCurrency *string
    ChannelTime     *time.Time
    ChannelName     string

    // 差异信息
    AmountDiff      int64
    HasDiscrepancy  bool      `gorm:"index"`
    DiscrepancyID   *uuid.UUID `gorm:"type:uuid;index"` // 关联到差异工单

    CreatedAt       time.Time
}

// 调账记录
type Adjustment struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    AdjustmentNo    string    `gorm:"size:100;unique;not null"`

    // 关联信息
    DiscrepancyID   uuid.UUID `gorm:"type:uuid;index;not null"`
    RecordID        uuid.UUID `gorm:"type:uuid;index;not null"`

    // 调账类型
    AdjustmentType  string    `gorm:"size:50;not null"` // "add_transaction", "reverse_transaction", "adjust_amount", "write_off"
    Reason          string    `gorm:"type:text;not null"`

    // 调账金额
    Amount          int64
    Currency        string

    // 会计分录
    DebitAccount    string    // 借方账户
    CreditAccount   string    // 贷方账户

    // 审批信息
    Status          string    `gorm:"size:50;not null"` // "pending", "approved", "rejected", "executed"
    RequestedBy     uuid.UUID `gorm:"type:uuid;not null"`
    ApprovedBy      *uuid.UUID `gorm:"type:uuid"`
    ApprovedAt      *time.Time
    RejectReason    string    `gorm:"type:text"`

    // 执行信息
    ExecutedAt      *time.Time
    ExecutedBy      *uuid.UUID
    EntryID         *uuid.UUID `gorm:"type:uuid"` // 关联到会计分录

    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 服务接口

```go
// internal/service/reconciliation_service.go
package service

type ReconciliationService interface {
    // 创建对账任务
    CreateReconciliationTask(ctx context.Context, date time.Time, taskType string) (*model.ReconciliationTask, error)

    // 执行对账
    ExecuteReconciliation(ctx context.Context, taskID uuid.UUID) error

    // 三方对账(内部 vs 渠道 vs 银行)
    ThreeWayReconcile(ctx context.Context, date time.Time) (*ReconcileReport, error)

    // 获取对账报告
    GetReconciliationReport(ctx context.Context, taskID uuid.UUID) (*ReconcileReport, error)

    // 调账相关
    CreateAdjustment(ctx context.Context, input *CreateAdjustmentInput) (*model.Adjustment, error)
    ApproveAdjustment(ctx context.Context, adjustmentID uuid.UUID, approverID uuid.UUID) error
    ExecuteAdjustment(ctx context.Context, adjustmentID uuid.UUID) error
}

type ReconcileReport struct {
    TaskID              uuid.UUID
    Date                time.Time
    TotalTransactions   int64
    MatchedTransactions int64
    MatchRate           float64

    Discrepancies struct {
        Total              int
        ByType             map[string]int     // missing_channel: 5, amount_mismatch: 3
        BySeverity         map[string]int     // critical: 2, high: 3, medium: 10
        TotalAmount        int64              // 差异总金额
        ResolvedCount      int
        PendingCount       int
        AverageResolveTime time.Duration
    }

    FinancialImpact struct {
        WriteOffAmount     int64  // 核销金额
        AdjustmentAmount   int64  // 调账金额
        RecoveredAmount    int64  // 追回金额
        NetLoss            int64  // 净损失
    }

    ChannelBreakdown map[string]*ChannelReconcileStats // 按渠道统计
}

type ChannelReconcileStats struct {
    ChannelName    string
    TotalCount     int64
    MatchedCount   int64
    MismatchCount  int64
    MatchRate      float64
}
```

### 定时任务

```go
// internal/worker/reconciliation_worker.go
package worker

type ReconciliationWorker struct {
    reconcileService service.ReconciliationService
    discrepancyClient client.DiscrepancyClient // 调用admin-service
}

// 每日自动对账 (凌晨3点执行)
func (w *ReconciliationWorker) DailyReconcile() {
    yesterday := time.Now().AddDate(0, 0, -1)

    // 1. 创建对账任务
    task, err := w.reconcileService.CreateReconciliationTask(ctx, yesterday, "daily")

    // 2. 执行对账
    err = w.reconcileService.ExecuteReconciliation(ctx, task.ID)

    // 3. 获取对账报告
    report, err := w.reconcileService.GetReconciliationReport(ctx, task.ID)

    // 4. 发现差异,自动创建工单(调用admin-service)
    if report.Discrepancies.Total > 0 {
        for _, discrepancy := range report.GetDiscrepancies() {
            w.discrepancyClient.CreateTicket(ctx, discrepancy)
        }
    }

    // 5. 发送对账报告通知
    w.notifyReconcileComplete(report)
}
```

---

## 2️⃣ channel-adapter (渠道对账文件管理)

**职责**: 渠道账单下载、解析、标准化

### 新增功能

```go
// internal/service/channel_reconcile_service.go
package service

type ChannelReconcileService interface {
    // 下载渠道对账文件
    DownloadChannelFile(ctx context.Context, channel string, date time.Time) (*ChannelFile, error)

    // 解析渠道账单
    ParseChannelFile(ctx context.Context, fileID uuid.UUID) ([]*ChannelTransaction, error)

    // 标准化交易数据
    NormalizeTransactions(ctx context.Context, channel string, rawData []byte) ([]*StandardTransaction, error)
}

// 渠道账单文件
type ChannelFile struct {
    ID           uuid.UUID
    Channel      string    // "stripe", "paypal"
    FileType     string    // "settlement", "transaction", "payout"
    Date         time.Time
    FilePath     string    // S3/本地路径
    FileSize     int64
    Status       string    // "downloaded", "parsed", "imported"
    RecordCount  int
    ParsedAt     *time.Time
    CreatedAt    time.Time
}

// 标准化交易格式
type StandardTransaction struct {
    ChannelOrderNo   string
    ChannelName      string
    TransactionType  string    // "payment", "refund", "payout"
    Amount           int64
    Currency         string
    Fee              int64
    Net              int64
    Status           string
    TransactionTime  time.Time
    SettlementTime   *time.Time
    MerchantRef      string    // 商户订单号
    RawData          map[string]interface{} // 原始数据
}
```

### 渠道适配器实现

```go
// internal/adapter/stripe_reconcile_adapter.go
package adapter

type StripeReconcileAdapter struct {
    stripeClient *stripe.Client
}

// 下载Stripe账单 (通过Stripe API)
func (a *StripeReconcileAdapter) DownloadDailyReport(date time.Time) ([]byte, error) {
    // 调用 Stripe Balance Transaction API
    params := &stripe.BalanceTransactionListParams{
        Created: &stripe.RangeQueryParams{
            GreaterThanOrEqual: date.Unix(),
            LessThan:           date.AddDate(0, 0, 1).Unix(),
        },
    }

    transactions := []*stripe.BalanceTransaction{}
    i := balancetransaction.List(params)
    for i.Next() {
        transactions = append(transactions, i.BalanceTransaction())
    }

    // 转换为标准格式
    return a.normalizeTransactions(transactions)
}

// PayPal适配器类似实现
type PayPalReconcileAdapter struct {
    // PayPal Transaction Search API
}
```

---

## 3️⃣ settlement-service (结算对账)

**职责**: 结算单对账、提现对账

### 新增功能

```go
// internal/service/settlement_reconcile_service.go
package service

type SettlementReconcileService interface {
    // 结算单对账
    ReconcileSettlement(ctx context.Context, settlementID uuid.UUID) error

    // 提现对账
    ReconcileWithdrawal(ctx context.Context, withdrawalID uuid.UUID) error

    // 商户账户余额核对
    ReconcileMerchantBalance(ctx context.Context, merchantID uuid.UUID, date time.Time) error
}

// 结算对账记录
type SettlementReconcile struct {
    ID              uuid.UUID
    SettlementID    uuid.UUID

    // 对账结果
    Status          string    // "matched", "mismatch"

    // 金额对比
    SettlementAmount int64    // 结算单金额
    ActualAmount     int64    // 实际支付金额
    Difference       int64    // 差异

    // 明细对比
    ExpectedCount    int      // 预期笔数
    ActualCount      int      // 实际笔数
    MissingOrders    []string // 缺失订单
    ExtraOrders      []string // 多余订单

    ReconcileAt      time.Time
}
```

---

## 4️⃣ admin-service (差异处理工作流) ⭐

**职责**: 工单管理、工作流引擎、SLA、审批

### 数据模型

```go
// internal/model/discrepancy.go
package model

// 差异工单 (完整的工单系统)
type DiscrepancyTicket struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TicketNo        string    `gorm:"size:100;unique;not null;index"`

    // 差异信息
    ReconcileRecordID uuid.UUID `gorm:"type:uuid;index;not null"` // 关联accounting-service的对账记录
    DiscrepancyType   string    `gorm:"size:50;not null"` // "missing_payment", "duplicate_payment", "amount_mismatch"
    Severity          string    `gorm:"size:20;not null"` // "low", "medium", "high", "critical"

    // 交易信息
    InternalOrderNo string    `gorm:"size:100;index"`
    ChannelOrderNo  string    `gorm:"size:100;index"`
    MerchantID      uuid.UUID `gorm:"type:uuid;index"`
    PaymentChannel  string    `gorm:"size:50"`

    // 差异详情
    InternalAmount  int64
    ChannelAmount   int64
    AmountDiff      int64
    Currency        string
    TransactionDate time.Time
    ReconcileDate   time.Time

    // 工作流状态
    Status          string    `gorm:"size:50;not null;index"` // "open", "investigating", "pending_channel", "resolved", "closed"
    Priority        int       `gorm:"default:3"` // 1-5

    // 分配信息
    AssignedTo      *uuid.UUID `gorm:"type:uuid;index"`
    AssignedAt      *time.Time
    DueDate         time.Time  `gorm:"index"` // SLA截止时间

    // 处理结果
    ResolutionType  *string   // "channel_error", "internal_error", "timing_difference", "write_off"
    ResolutionNote  string    `gorm:"type:text"`
    ResolvedAt      *time.Time
    ResolvedBy      *uuid.UUID `gorm:"type:uuid"`

    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// 工单操作记录
type DiscrepancyAction struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TicketID    uuid.UUID `gorm:"type:uuid;index;not null"`
    ActionType  string    `gorm:"size:50;not null"` // "assign", "comment", "status_change", "escalate"
    ActionBy    uuid.UUID `gorm:"type:uuid;not null"`
    ActionByName string   `gorm:"size:100"`
    Description string    `gorm:"type:text"`

    // 状态变更
    OldStatus   *string   `gorm:"size:50"`
    NewStatus   *string   `gorm:"size:50"`

    // 元数据
    Metadata    string    `gorm:"type:jsonb"` // JSON格式

    CreatedAt   time.Time
}

// 工单评论
type DiscrepancyComment struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TicketID   uuid.UUID `gorm:"type:uuid;index;not null"`
    UserID     uuid.UUID `gorm:"type:uuid;not null"`
    UserName   string    `gorm:"size:100"`
    Comment    string    `gorm:"type:text;not null"`
    IsInternal bool      `gorm:"default:false"` // 内部备注
    CreatedAt  time.Time
}

// 工单附件
type DiscrepancyAttachment struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TicketID   uuid.UUID `gorm:"type:uuid;index;not null"`
    FileName   string    `gorm:"size:255;not null"`
    FileURL    string    `gorm:"type:text;not null"`
    FileSize   int64
    FileType   string    `gorm:"size:50"` // "image", "pdf", "excel"
    UploadedBy uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt  time.Time
}
```

### 工作流引擎

```go
// internal/service/discrepancy_workflow_service.go
package service

type DiscrepancyWorkflowService interface {
    // 创建工单(由accounting-service调用)
    CreateTicket(ctx context.Context, input *CreateTicketInput) (*model.DiscrepancyTicket, error)

    // 分配工单
    AssignTicket(ctx context.Context, ticketID, assigneeID uuid.UUID) error

    // 更新状态
    UpdateStatus(ctx context.Context, ticketID uuid.UUID, newStatus string, userID uuid.UUID) error

    // 升级工单
    EscalateTicket(ctx context.Context, ticketID uuid.UUID, reason string) error

    // 解决工单
    ResolveTicket(ctx context.Context, ticketID uuid.UUID, resolution *Resolution) error

    // 关闭工单
    CloseTicket(ctx context.Context, ticketID uuid.UUID) error

    // SLA检查(定时任务)
    CheckSLA(ctx context.Context) error
}

// 工作流状态机
type WorkflowStateMachine struct {
    allowedTransitions map[string][]string
}

func NewWorkflowStateMachine() *WorkflowStateMachine {
    return &WorkflowStateMachine{
        allowedTransitions: map[string][]string{
            "open":             {"investigating", "closed"},
            "investigating":    {"pending_channel", "resolved", "escalated"},
            "pending_channel":  {"investigating", "resolved", "escalated"},
            "escalated":        {"investigating", "resolved"},
            "resolved":         {"closed", "investigating"}, // 可重开
            "closed":           {},
        },
    }
}

func (sm *WorkflowStateMachine) CanTransition(from, to string) bool {
    allowed, exists := sm.allowedTransitions[from]
    if !exists {
        return false
    }
    for _, state := range allowed {
        if state == to {
            return true
        }
    }
    return false
}
```

### SLA管理

```go
// internal/worker/sla_worker.go
package worker

// SLA规则
var SLADurations = map[string]time.Duration{
    "critical": 2 * time.Hour,    // 2小时内响应
    "high":     24 * time.Hour,   // 24小时
    "medium":   3 * 24 * time.Hour,  // 3天
    "low":      7 * 24 * time.Hour,  // 7天
}

type SLAWorker struct {
    workflowService service.DiscrepancyWorkflowService
    notifyClient    client.NotificationClient
}

// 每小时检查SLA
func (w *SLAWorker) CheckOverdueSLA() {
    now := time.Now()

    // 查询超时工单
    overdueTickets, err := w.repository.FindOverdueTickets(now)

    for _, ticket := range overdueTickets {
        // 自动升级
        if ticket.Status != "escalated" {
            w.workflowService.EscalateTicket(ctx, ticket.ID, "SLA timeout")
        }

        // 发送告警通知
        w.notifyClient.SendSLAAlert(ticket)
    }
}
```

---

## 5️⃣ notification-service (对账通知)

**职责**: 发送对账相关通知

### 新增通知类型

```go
// internal/service/reconcile_notification_service.go
package service

// 对账完成通知
func (s *NotificationService) SendReconcileCompleteNotification(report *ReconcileReport) {
    // 发送给财务团队
}

// 差异告警通知
func (s *NotificationService) SendDiscrepancyAlert(ticket *DiscrepancyTicket) {
    // Critical: 短信 + 邮件 + 企业微信
    // High: 邮件 + 企业微信
    // Medium/Low: 仅邮件
}

// SLA超时提醒
func (s *NotificationService) SendSLAOverdueAlert(ticket *DiscrepancyTicket) {
    // 通知处理人和上级
}
```

---

## 🔄 完整对账流程

```
[每日凌晨3点]
1. accounting-service 启动定时任务
   ├─ 调用 channel-adapter.DownloadChannelFile() 下载Stripe/PayPal账单
   ├─ 调用 settlement-service.GetSettlementRecords() 获取结算记录
   ├─ 执行三方对账算法
   └─ 生成对账报告

2. 发现差异后
   └─ 调用 admin-service.CreateTicket() 创建差异工单

3. admin-service 自动分配工单
   ├─ 根据渠道分配给渠道负责人
   ├─ 计算SLA截止时间
   └─ 调用 notification-service 发送通知

4. 处理人员处理工单 (Admin后台)
   ├─ 查看交易详情
   ├─ 添加评论和附件
   ├─ 联系渠道/商户
   └─ 提交解决方案

5. 需要调账时
   ├─ 在 admin-service 创建调账申请
   ├─ 提交到审批流程
   ├─ 审批通过后,调用 accounting-service.ExecuteAdjustment()
   └─ accounting-service 执行会计分录

6. SLA检查 (每小时)
   └─ admin-service SLA Worker 自动检查超时工单并升级
```

---

## 📊 服务间API调用关系

```
accounting-service:
  → channel-adapter.DownloadChannelFile()
  → settlement-service.GetSettlementRecords()
  → admin-service.CreateTicket()
  → notification-service.SendNotification()

channel-adapter:
  → 无对外调用(被动提供API)

settlement-service:
  → 无对外调用(被动提供API)

admin-service:
  → accounting-service.CreateAdjustment()
  → accounting-service.ExecuteAdjustment()
  → notification-service.SendNotification()

notification-service:
  → 无对外调用(被动提供API)
```

---

## 🎯 实施优先级

### Phase 1: 核心对账能力 (Week 1-2)
- accounting-service: 对账引擎 + 差异检测
- channel-adapter: Stripe账单下载和解析

### Phase 2: 工单系统 (Week 2-3)
- admin-service: 工单CRUD + 工作流状态机
- notification-service: 差异告警通知

### Phase 3: 高级功能 (Week 3-4)
- admin-service: SLA管理 + 审批流程
- accounting-service: 调账执行
- settlement-service: 结算对账

### Phase 4: 前端界面 (Week 4)
- admin-portal: 对账报告页面
- admin-portal: 差异工单管理页面

---

## 📝 数据库设计总结

| 服务 | 新增表 | 说明 |
|------|-------|------|
| accounting-service | reconciliation_tasks, reconciliation_records, adjustments | 对账核心数据 |
| channel-adapter | channel_files, channel_transactions | 渠道账单数据 |
| settlement-service | settlement_reconciles | 结算对账数据 |
| admin-service | discrepancy_tickets, discrepancy_actions, discrepancy_comments, discrepancy_attachments | 工单系统数据 |

---

**总结**: 采用**分布式对账架构**,每个服务专注自己的领域,通过HTTP API协同完成完整的对账流程。✅
