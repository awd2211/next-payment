package saga

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockRedisClient 模拟 Redis 客户端
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*MockStatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *MockIntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*MockIntCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *MockIntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*MockIntCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*MockStatusCmd)
}

type MockStatusCmd struct {
	mock.Mock
	result bool
	err    error
}

func (m *MockStatusCmd) Result() (bool, error) {
	return m.result, m.err
}

func (m *MockStatusCmd) Err() error {
	return m.err
}

type MockIntCmd struct {
	mock.Mock
	result int64
	err    error
}

func (m *MockIntCmd) Result() (int64, error) {
	return m.result, m.err
}

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&Saga{}, &SagaStep{})
	assert.NoError(t, err)

	return db
}

// TestSagaBuilder 测试 Saga 构建器
func TestSagaBuilder(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()
	businessID := "TEST-001"
	businessType := "payment"

	// 构建 Saga
	sagaBuilder := orchestrator.NewSagaBuilder(businessID, businessType)
	sagaBuilder.SetMetadata(map[string]interface{}{
		"amount":   100,
		"currency": "USD",
	})

	sagaBuilder.AddStep(
		"Step1",
		func(ctx context.Context, data string) (string, error) {
			return "step1_result", nil
		},
		func(ctx context.Context, compensateData string, executeResult string) error {
			return nil
		},
		3,
	)

	saga, err := sagaBuilder.Build(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, saga)
	assert.Equal(t, businessID, saga.BusinessID)
	assert.Equal(t, businessType, saga.BusinessType)
	assert.Equal(t, SagaStatusPending, saga.Status)
	assert.Equal(t, 1, len(saga.Steps))
	assert.Equal(t, "Step1", saga.Steps[0].StepName)
}

// TestSagaExecutionSuccess 测试 Saga 成功执行
func TestSagaExecutionSuccess(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()
	businessID := "TEST-002"

	// 构建 Saga
	sagaBuilder := orchestrator.NewSagaBuilder(businessID, "payment")

	step1Executed := false
	step2Executed := false

	stepDefs := []StepDefinition{
		{
			Name: "Step1",
			Execute: func(ctx context.Context, data string) (string, error) {
				step1Executed = true
				return "step1_result", nil
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return nil
			},
			MaxRetryCount: 3,
			Timeout:       5 * time.Second,
		},
		{
			Name: "Step2",
			Execute: func(ctx context.Context, data string) (string, error) {
				step2Executed = true
				return "step2_result", nil
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return nil
			},
			MaxRetryCount: 3,
			Timeout:       5 * time.Second,
		},
	}

	for _, def := range stepDefs {
		sagaBuilder.AddStepWithTimeout(def.Name, def.Execute, def.Compensate, def.MaxRetryCount, def.Timeout)
	}

	saga, err := sagaBuilder.Build(ctx)
	assert.NoError(t, err)

	// 执行 Saga
	err = orchestrator.Execute(ctx, saga, stepDefs)
	assert.NoError(t, err)
	assert.True(t, step1Executed)
	assert.True(t, step2Executed)

	// 验证状态
	updatedSaga, err := orchestrator.GetSaga(ctx, saga.ID)
	assert.NoError(t, err)
	assert.Equal(t, SagaStatusCompleted, updatedSaga.Status)
	assert.Equal(t, 2, updatedSaga.CurrentStep)
}

// TestSagaExecutionWithCompensation 测试 Saga 执行失败并触发补偿
func TestSagaExecutionWithCompensation(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()
	businessID := "TEST-003"

	step1Executed := false
	step1Compensated := false
	step2Executed := false

	stepDefs := []StepDefinition{
		{
			Name: "Step1",
			Execute: func(ctx context.Context, data string) (string, error) {
				step1Executed = true
				return "step1_result", nil
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				step1Compensated = true
				return nil
			},
			MaxRetryCount: 1, // 只重试1次，快速失败
			Timeout:       5 * time.Second,
		},
		{
			Name: "Step2",
			Execute: func(ctx context.Context, data string) (string, error) {
				step2Executed = true
				return "", errors.New("step2 failed")
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return nil
			},
			MaxRetryCount: 1,
			Timeout:       5 * time.Second,
		},
	}

	sagaBuilder := orchestrator.NewSagaBuilder(businessID, "payment")
	for _, def := range stepDefs {
		sagaBuilder.AddStepWithTimeout(def.Name, def.Execute, def.Compensate, def.MaxRetryCount, def.Timeout)
	}

	saga, err := sagaBuilder.Build(ctx)
	assert.NoError(t, err)

	// 执行 Saga（应该失败并触发补偿）
	err = orchestrator.Execute(ctx, saga, stepDefs)
	assert.Error(t, err)
	assert.True(t, step1Executed)
	assert.True(t, step2Executed)
	assert.True(t, step1Compensated) // Step1 应该被补偿

	// 验证状态
	updatedSaga, err := orchestrator.GetSaga(ctx, saga.ID)
	assert.NoError(t, err)
	assert.Equal(t, SagaStatusCompensated, updatedSaga.Status)
}

// TestStepTimeout 测试步骤超时
func TestStepTimeout(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()
	businessID := "TEST-004"

	stepDefs := []StepDefinition{
		{
			Name: "SlowStep",
			Execute: func(ctx context.Context, data string) (string, error) {
				// 模拟耗时操作
				select {
				case <-time.After(10 * time.Second):
					return "result", nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return nil
			},
			MaxRetryCount: 1,
			Timeout:       1 * time.Second, // 1秒超时
		},
	}

	sagaBuilder := orchestrator.NewSagaBuilder(businessID, "payment")
	for _, def := range stepDefs {
		sagaBuilder.AddStepWithTimeout(def.Name, def.Execute, def.Compensate, def.MaxRetryCount, def.Timeout)
	}

	saga, err := sagaBuilder.Build(ctx)
	assert.NoError(t, err)

	// 执行 Saga（应该超时失败）
	err = orchestrator.Execute(ctx, saga, stepDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestCompensationRetry 测试补偿重试逻辑
func TestCompensationRetry(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()
	businessID := "TEST-005"

	step1Executed := false
	compensationAttempts := 0

	stepDefs := []StepDefinition{
		{
			Name: "Step1",
			Execute: func(ctx context.Context, data string) (string, error) {
				step1Executed = true
				return "step1_result", nil
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				compensationAttempts++
				if compensationAttempts < 2 {
					return errors.New("compensation failed temporarily")
				}
				return nil // 第二次成功
			},
			MaxRetryCount: 1,
			Timeout:       5 * time.Second,
		},
		{
			Name: "Step2",
			Execute: func(ctx context.Context, data string) (string, error) {
				return "", errors.New("step2 failed")
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return nil
			},
			MaxRetryCount: 1,
			Timeout:       5 * time.Second,
		},
	}

	sagaBuilder := orchestrator.NewSagaBuilder(businessID, "payment")
	for _, def := range stepDefs {
		sagaBuilder.AddStepWithTimeout(def.Name, def.Execute, def.Compensate, def.MaxRetryCount, def.Timeout)
	}

	saga, err := sagaBuilder.Build(ctx)
	assert.NoError(t, err)

	// 执行 Saga
	err = orchestrator.Execute(ctx, saga, stepDefs)
	assert.Error(t, err)
	assert.True(t, step1Executed)
	assert.GreaterOrEqual(t, compensationAttempts, 2) // 补偿应该至少尝试2次

	// 验证最终状态
	updatedSaga, err := orchestrator.GetSaga(ctx, saga.ID)
	assert.NoError(t, err)
	assert.Equal(t, SagaStatusCompensated, updatedSaga.Status)
}

// TestGetSagaByBusinessID 测试根据业务ID获取Saga
func TestGetSagaByBusinessID(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()
	businessID := "TEST-006"

	// 创建 Saga
	sagaBuilder := orchestrator.NewSagaBuilder(businessID, "payment")
	sagaBuilder.AddStep("Step1", nil, nil, 3)

	saga, err := sagaBuilder.Build(ctx)
	assert.NoError(t, err)

	// 根据 businessID 查询
	foundSaga, err := orchestrator.GetSagaByBusinessID(ctx, businessID)
	assert.NoError(t, err)
	assert.NotNil(t, foundSaga)
	assert.Equal(t, saga.ID, foundSaga.ID)
	assert.Equal(t, businessID, foundSaga.BusinessID)
}

// TestListFailedSagas 测试列出失败的Saga
func TestListFailedSagas(t *testing.T) {
	db := setupTestDB(t)
	orchestrator := NewSagaOrchestrator(db, nil)

	ctx := context.Background()

	// 创建一个失败的 Saga
	failedSaga := &Saga{
		ID:           uuid.New(),
		BusinessID:   "FAILED-001",
		BusinessType: "payment",
		Status:       SagaStatusFailed,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := db.Create(failedSaga).Error
	assert.NoError(t, err)

	// 查询失败的 Saga
	failedSagas, err := orchestrator.ListFailedSagas(ctx, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(failedSagas), 1)

	found := false
	for _, s := range failedSagas {
		if s.ID == failedSaga.ID {
			found = true
			break
		}
	}
	assert.True(t, found)
}
