package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SagaStatus Saga 状态
type SagaStatus string

const (
	SagaStatusPending     SagaStatus = "pending"      // 等待执行
	SagaStatusInProgress  SagaStatus = "in_progress"  // 执行中
	SagaStatusCompleted   SagaStatus = "completed"    // 已完成
	SagaStatusCompensated SagaStatus = "compensated"  // 已补偿
	SagaStatusFailed      SagaStatus = "failed"       // 失败（补偿也失败）
)

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending     StepStatus = "pending"      // 等待执行
	StepStatusCompleted   StepStatus = "completed"    // 已完成
	StepStatusCompensated StepStatus = "compensated"  // 已补偿
	StepStatusFailed      StepStatus = "failed"       // 失败
)

// SagaOrchestrator Saga 编排器
type SagaOrchestrator struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewSagaOrchestrator 创建 Saga 编排器
func NewSagaOrchestrator(db *gorm.DB, redis *redis.Client) *SagaOrchestrator {
	return &SagaOrchestrator{
		db:    db,
		redis: redis,
	}
}

// Saga Saga 实例
type Saga struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	BusinessID    string         `json:"business_id" gorm:"index;not null"` // 业务ID（如 payment_no）
	BusinessType  string         `json:"business_type" gorm:"index"`         // 业务类型（payment, refund等）
	Status        SagaStatus     `json:"status" gorm:"index"`
	Steps         []SagaStep     `json:"steps" gorm:"foreignKey:SagaID"`
	CurrentStep   int            `json:"current_step"`
	ErrorMessage  string         `json:"error_message"`
	Metadata      string         `json:"metadata" gorm:"type:text"` // JSON格式元数据
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	CompensatedAt *time.Time     `json:"compensated_at"`
}

// SagaStep Saga 步骤
type SagaStep struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	SagaID          uuid.UUID  `json:"saga_id" gorm:"type:uuid;index;not null"`
	StepOrder       int        `json:"step_order" gorm:"not null"` // 步骤顺序（从0开始）
	StepName        string     `json:"step_name" gorm:"not null"`  // 步骤名称
	Status          StepStatus `json:"status"`
	ExecuteData     string     `json:"execute_data" gorm:"type:text"`     // 执行参数（JSON）
	CompensateData  string     `json:"compensate_data" gorm:"type:text"`  // 补偿参数（JSON）
	Result          string     `json:"result" gorm:"type:text"`           // 执行结果（JSON）
	ErrorMessage    string     `json:"error_message"`
	ExecutedAt      *time.Time `json:"executed_at"`
	CompensatedAt   *time.Time `json:"compensated_at"`
	RetryCount      int        `json:"retry_count"`
	MaxRetryCount   int        `json:"max_retry_count"`
	NextRetryAt     *time.Time `json:"next_retry_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Saga) TableName() string {
	return "saga_instances"
}

// TableName 指定表名
func (SagaStep) TableName() string {
	return "saga_steps"
}

// StepFunc 步骤执行函数
type StepFunc func(ctx context.Context, executeData string) (result string, err error)

// CompensateFunc 补偿函数
type CompensateFunc func(ctx context.Context, compensateData string, executeResult string) error

// StepDefinition 步骤定义
type StepDefinition struct {
	Name           string
	Execute        StepFunc
	Compensate     CompensateFunc
	MaxRetryCount  int
}

// SagaBuilder Saga 构建器
type SagaBuilder struct {
	orchestrator *SagaOrchestrator
	businessID   string
	businessType string
	metadata     map[string]interface{}
	steps        []StepDefinition
}

// NewSagaBuilder 创建 Saga 构建器
func (o *SagaOrchestrator) NewSagaBuilder(businessID, businessType string) *SagaBuilder {
	return &SagaBuilder{
		orchestrator: o,
		businessID:   businessID,
		businessType: businessType,
		metadata:     make(map[string]interface{}),
		steps:        []StepDefinition{},
	}
}

// AddStep 添加步骤
func (b *SagaBuilder) AddStep(name string, execute StepFunc, compensate CompensateFunc, maxRetry int) *SagaBuilder {
	if maxRetry <= 0 {
		maxRetry = 3 // 默认重试3次
	}
	b.steps = append(b.steps, StepDefinition{
		Name:          name,
		Execute:       execute,
		Compensate:    compensate,
		MaxRetryCount: maxRetry,
	})
	return b
}

// SetMetadata 设置元数据
func (b *SagaBuilder) SetMetadata(metadata map[string]interface{}) *SagaBuilder {
	b.metadata = metadata
	return b
}

// Build 构建 Saga 实例
func (b *SagaBuilder) Build(ctx context.Context) (*Saga, error) {
	sagaID := uuid.New()

	metadataJSON := "{}"
	if len(b.metadata) > 0 {
		metaBytes, err := json.Marshal(b.metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata failed: %w", err)
		}
		metadataJSON = string(metaBytes)
	}

	saga := &Saga{
		ID:           sagaID,
		BusinessID:   b.businessID,
		BusinessType: b.businessType,
		Status:       SagaStatusPending,
		CurrentStep:  0,
		Metadata:     metadataJSON,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 创建 Saga 实例
	if err := b.orchestrator.db.Create(saga).Error; err != nil {
		return nil, fmt.Errorf("create saga failed: %w", err)
	}

	// 创建步骤记录（不保存执行函数）
	for i := range b.steps {
		step := &SagaStep{
			ID:            uuid.New(),
			SagaID:        sagaID,
			StepOrder:     i,
			StepName:      b.steps[i].Name,
			Status:        StepStatusPending,
			MaxRetryCount: b.steps[i].MaxRetryCount,
			RetryCount:    0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := b.orchestrator.db.Create(step).Error; err != nil {
			return nil, fmt.Errorf("create saga step failed: %w", err)
		}
	}

	// 重新加载 Saga（包含 Steps）
	if err := b.orchestrator.db.Preload("Steps").First(saga, "id = ?", sagaID).Error; err != nil {
		return nil, fmt.Errorf("reload saga failed: %w", err)
	}

	return saga, nil
}

// Execute 执行 Saga
func (o *SagaOrchestrator) Execute(ctx context.Context, saga *Saga, stepDefs []StepDefinition) error {
	// 更新状态为执行中
	saga.Status = SagaStatusInProgress
	saga.UpdatedAt = time.Now()
	if err := o.db.Save(saga).Error; err != nil {
		return fmt.Errorf("update saga status failed: %w", err)
	}

	logger.Info("saga execution started",
		zap.String("saga_id", saga.ID.String()),
		zap.String("business_id", saga.BusinessID),
		zap.String("business_type", saga.BusinessType))

	// 执行每个步骤
	for i := saga.CurrentStep; i < len(saga.Steps); i++ {
		step := &saga.Steps[i]
		stepDef := stepDefs[i]

		logger.Info("executing saga step",
			zap.String("saga_id", saga.ID.String()),
			zap.Int("step", i),
			zap.String("step_name", step.StepName))

		// 执行步骤
		result, err := stepDef.Execute(ctx, step.ExecuteData)
		now := time.Now()

		if err != nil {
			// 步骤执行失败
			step.Status = StepStatusFailed
			step.ErrorMessage = err.Error()
			step.RetryCount++
			step.UpdatedAt = now

			if step.RetryCount < step.MaxRetryCount {
				// 计算下次重试时间（指数退避）
				nextRetry := now.Add(time.Duration(1<<uint(step.RetryCount)) * time.Second)
				step.NextRetryAt = &nextRetry
				logger.Warn("saga step failed, will retry",
					zap.String("saga_id", saga.ID.String()),
					zap.String("step_name", step.StepName),
					zap.Int("retry_count", step.RetryCount),
					zap.Time("next_retry", nextRetry),
					zap.Error(err))
			} else {
				// 达到最大重试次数，开始补偿
				logger.Error("saga step failed after max retries, starting compensation",
					zap.String("saga_id", saga.ID.String()),
					zap.String("step_name", step.StepName),
					zap.Error(err))

				saga.ErrorMessage = fmt.Sprintf("步骤 %s 失败: %v", step.StepName, err)
				if err := o.db.Save(step).Error; err != nil {
					logger.Error("failed to save step", zap.Error(err))
				}

				// 开始补偿流程
				return o.Compensate(ctx, saga, stepDefs)
			}

			if err := o.db.Save(step).Error; err != nil {
				return fmt.Errorf("save step failed: %w", err)
			}
			return fmt.Errorf("step %s failed: %w", step.StepName, err)
		}

		// 步骤执行成功
		step.Status = StepStatusCompleted
		step.Result = result
		step.ExecutedAt = &now
		step.UpdatedAt = now

		if err := o.db.Save(step).Error; err != nil {
			return fmt.Errorf("save step failed: %w", err)
		}

		// 更新当前步骤
		saga.CurrentStep = i + 1
		saga.UpdatedAt = now
		if err := o.db.Save(saga).Error; err != nil {
			return fmt.Errorf("update saga current step failed: %w", err)
		}

		logger.Info("saga step completed",
			zap.String("saga_id", saga.ID.String()),
			zap.String("step_name", step.StepName))
	}

	// 所有步骤执行成功
	now := time.Now()
	saga.Status = SagaStatusCompleted
	saga.CompletedAt = &now
	saga.UpdatedAt = now

	if err := o.db.Save(saga).Error; err != nil {
		return fmt.Errorf("update saga status failed: %w", err)
	}

	logger.Info("saga execution completed",
		zap.String("saga_id", saga.ID.String()),
		zap.String("business_id", saga.BusinessID))

	return nil
}

// Compensate 执行补偿
func (o *SagaOrchestrator) Compensate(ctx context.Context, saga *Saga, stepDefs []StepDefinition) error {
	logger.Info("saga compensation started",
		zap.String("saga_id", saga.ID.String()),
		zap.String("business_id", saga.BusinessID),
		zap.Int("steps_to_compensate", saga.CurrentStep))

	// 按相反顺序补偿已完成的步骤
	for i := saga.CurrentStep - 1; i >= 0; i-- {
		step := &saga.Steps[i]
		if step.Status != StepStatusCompleted {
			continue // 只补偿已完成的步骤
		}

		stepDef := stepDefs[i]
		if stepDef.Compensate == nil {
			logger.Warn("no compensation function for step",
				zap.String("saga_id", saga.ID.String()),
				zap.String("step_name", step.StepName))
			continue
		}

		logger.Info("compensating saga step",
			zap.String("saga_id", saga.ID.String()),
			zap.String("step_name", step.StepName))

		// 执行补偿
		err := stepDef.Compensate(ctx, step.CompensateData, step.Result)
		now := time.Now()

		if err != nil {
			logger.Error("saga step compensation failed",
				zap.String("saga_id", saga.ID.String()),
				zap.String("step_name", step.StepName),
				zap.Error(err))

			step.ErrorMessage = fmt.Sprintf("补偿失败: %v", err)
			step.UpdatedAt = now
			o.db.Save(step)
			// 补偿失败，但继续补偿其他步骤
			continue
		}

		// 补偿成功
		step.Status = StepStatusCompensated
		step.CompensatedAt = &now
		step.UpdatedAt = now

		if err := o.db.Save(step).Error; err != nil {
			logger.Error("failed to save compensated step", zap.Error(err))
		}

		logger.Info("saga step compensated",
			zap.String("saga_id", saga.ID.String()),
			zap.String("step_name", step.StepName))
	}

	// 更新 Saga 状态为已补偿
	now := time.Now()
	saga.Status = SagaStatusCompensated
	saga.CompensatedAt = &now
	saga.UpdatedAt = now

	if err := o.db.Save(saga).Error; err != nil {
		return fmt.Errorf("update saga status failed: %w", err)
	}

	logger.Info("saga compensation completed",
		zap.String("saga_id", saga.ID.String()),
		zap.String("business_id", saga.BusinessID))

	return nil
}

// GetSaga 获取 Saga
func (o *SagaOrchestrator) GetSaga(ctx context.Context, sagaID uuid.UUID) (*Saga, error) {
	var saga Saga
	if err := o.db.Preload("Steps").First(&saga, "id = ?", sagaID).Error; err != nil {
		return nil, err
	}
	return &saga, nil
}

// GetSagaByBusinessID 根据业务ID获取 Saga
func (o *SagaOrchestrator) GetSagaByBusinessID(ctx context.Context, businessID string) (*Saga, error) {
	var saga Saga
	if err := o.db.Preload("Steps").Where("business_id = ?", businessID).Order("created_at DESC").First(&saga).Error; err != nil {
		return nil, err
	}
	return &saga, nil
}

// ListPendingRetries 列出待重试的步骤
func (o *SagaOrchestrator) ListPendingRetries(ctx context.Context, limit int) ([]*SagaStep, error) {
	var steps []*SagaStep
	now := time.Now()

	err := o.db.Where("status = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ?",
		StepStatusFailed, now).
		Limit(limit).
		Find(&steps).Error

	return steps, err
}
