package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status 健康状态
type Status string

const (
	StatusHealthy   Status = "healthy"   // 健康
	StatusDegraded  Status = "degraded"  // 降级（部分功能不可用）
	StatusUnhealthy Status = "unhealthy" // 不健康
)

// CheckResult 健康检查结果
type CheckResult struct {
	Name      string                 `json:"name"`                // 检查项名称
	Status    Status                 `json:"status"`              // 状态
	Message   string                 `json:"message,omitempty"`   // 消息
	Error     string                 `json:"error,omitempty"`     // 错误信息
	Timestamp time.Time              `json:"timestamp"`           // 检查时间
	Duration  time.Duration          `json:"duration"`            // 检查耗时
	Metadata  map[string]interface{} `json:"metadata,omitempty"`  // 额外元数据
}

// Checker 健康检查接口
type Checker interface {
	// Name 返回检查器名称
	Name() string

	// Check 执行健康检查
	Check(ctx context.Context) *CheckResult
}

// HealthChecker 聚合健康检查器
type HealthChecker struct {
	checkers []Checker
	mu       sync.RWMutex
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checkers: make([]Checker, 0),
	}
}

// Register 注册检查器
func (h *HealthChecker) Register(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

// Check 执行所有健康检查
func (h *HealthChecker) Check(ctx context.Context) *HealthReport {
	h.mu.RLock()
	checkers := make([]Checker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

	startTime := time.Now()
	results := make([]*CheckResult, 0, len(checkers))

	// 并发执行所有检查
	var wg sync.WaitGroup
	resultChan := make(chan *CheckResult, len(checkers))

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			result := c.Check(ctx)
			resultChan <- result
		}(checker)
	}

	// 等待所有检查完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for result := range resultChan {
		results = append(results, result)
	}

	// 计算总体状态
	overallStatus := h.calculateOverallStatus(results)

	return &HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
		Checks:    results,
	}
}

// calculateOverallStatus 计算总体健康状态
func (h *HealthChecker) calculateOverallStatus(results []*CheckResult) Status {
	if len(results) == 0 {
		return StatusHealthy
	}

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, result := range results {
		switch result.Status {
		case StatusHealthy:
			healthyCount++
		case StatusDegraded:
			degradedCount++
		case StatusUnhealthy:
			unhealthyCount++
		}
	}

	// 如果有任何不健康的检查，整体状态为不健康
	if unhealthyCount > 0 {
		return StatusUnhealthy
	}

	// 如果有降级的检查，整体状态为降级
	if degradedCount > 0 {
		return StatusDegraded
	}

	// 所有检查都健康
	return StatusHealthy
}

// HealthReport 健康报告
type HealthReport struct {
	Status    Status         `json:"status"`    // 总体状态
	Timestamp time.Time      `json:"timestamp"` // 报告时间
	Duration  time.Duration  `json:"duration"`  // 总检查时长
	Checks    []*CheckResult `json:"checks"`    // 各项检查结果
}

// IsHealthy 判断是否健康
func (r *HealthReport) IsHealthy() bool {
	return r.Status == StatusHealthy
}

// GetStatusCode 获取HTTP状态码
func (r *HealthReport) GetStatusCode() int {
	switch r.Status {
	case StatusHealthy:
		return 200 // OK
	case StatusDegraded:
		return 200 // OK (仍可服务)
	case StatusUnhealthy:
		return 503 // Service Unavailable
	default:
		return 500 // Internal Server Error
	}
}

// SimpleChecker 简单的函数式检查器
type SimpleChecker struct {
	name     string
	checkFn  func(ctx context.Context) error
	metadata map[string]interface{}
}

// NewSimpleChecker 创建简单检查器
func NewSimpleChecker(name string, checkFn func(ctx context.Context) error) *SimpleChecker {
	return &SimpleChecker{
		name:     name,
		checkFn:  checkFn,
		metadata: make(map[string]interface{}),
	}
}

// Name 返回检查器名称
func (c *SimpleChecker) Name() string {
	return c.name
}

// Check 执行检查
func (c *SimpleChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  c.metadata,
	}

	// 执行检查
	err := c.checkFn(ctx)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("检查失败: %v", err)
	} else {
		result.Status = StatusHealthy
		result.Message = "正常"
	}

	return result
}

// WithMetadata 添加元数据
func (c *SimpleChecker) WithMetadata(key string, value interface{}) *SimpleChecker {
	c.metadata[key] = value
	return c
}
