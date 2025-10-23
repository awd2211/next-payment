package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DBChecker 数据库健康检查器
type DBChecker struct {
	name    string
	db      *gorm.DB
	timeout time.Duration
}

// NewDBChecker 创建数据库健康检查器
func NewDBChecker(name string, db *gorm.DB) *DBChecker {
	return &DBChecker{
		name:    name,
		db:      db,
		timeout: 5 * time.Second, // 默认5秒超时
	}
}

// WithTimeout 设置超时时间
func (c *DBChecker) WithTimeout(timeout time.Duration) *DBChecker {
	c.timeout = timeout
	return c
}

// Name 返回检查器名称
func (c *DBChecker) Name() string {
	return c.name
}

// Check 执行数据库健康检查
func (c *DBChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// 设置超时上下文
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 获取底层SQL DB
	sqlDB, err := c.db.DB()
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "无法获取数据库连接"
		return result
	}

	// 1. 检查数据库连接
	err = sqlDB.PingContext(checkCtx)
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "数据库连接失败"
		return result
	}

	// 2. 获取连接池统计信息
	stats := sqlDB.Stats()
	result.Metadata["max_open_connections"] = stats.MaxOpenConnections
	result.Metadata["open_connections"] = stats.OpenConnections
	result.Metadata["in_use"] = stats.InUse
	result.Metadata["idle"] = stats.Idle
	result.Metadata["wait_count"] = stats.WaitCount
	result.Metadata["wait_duration"] = stats.WaitDuration.String()

	// 3. 执行简单查询验证
	var dummy int
	err = c.db.WithContext(checkCtx).Raw("SELECT 1").Scan(&dummy).Error
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "数据库查询失败"
		return result
	}

	result.Duration = time.Since(startTime)

	// 4. 检查连接池状态，判断是否降级
	if stats.WaitCount > 100 {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("数据库连接池等待次数过高 (%d)", stats.WaitCount)
	} else if float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.9 {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("数据库连接池使用率过高 (%.1f%%)",
			float64(stats.InUse)/float64(stats.MaxOpenConnections)*100)
	} else {
		result.Status = StatusHealthy
		result.Message = "数据库正常"
	}

	return result
}

// SQLDBChecker 原生SQL数据库健康检查器（用于不使用GORM的场景）
type SQLDBChecker struct {
	name    string
	db      *sql.DB
	timeout time.Duration
}

// NewSQLDBChecker 创建SQL数据库健康检查器
func NewSQLDBChecker(name string, db *sql.DB) *SQLDBChecker {
	return &SQLDBChecker{
		name:    name,
		db:      db,
		timeout: 5 * time.Second,
	}
}

// WithTimeout 设置超时时间
func (c *SQLDBChecker) WithTimeout(timeout time.Duration) *SQLDBChecker {
	c.timeout = timeout
	return c
}

// Name 返回检查器名称
func (c *SQLDBChecker) Name() string {
	return c.name
}

// Check 执行数据库健康检查
func (c *SQLDBChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// 设置超时上下文
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 检查连接
	err := c.db.PingContext(checkCtx)
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "数据库连接失败"
		return result
	}

	// 获取统计信息
	stats := c.db.Stats()
	result.Metadata["max_open_connections"] = stats.MaxOpenConnections
	result.Metadata["open_connections"] = stats.OpenConnections
	result.Metadata["in_use"] = stats.InUse
	result.Metadata["idle"] = stats.Idle

	result.Duration = time.Since(startTime)
	result.Status = StatusHealthy
	result.Message = "数据库正常"

	return result
}
