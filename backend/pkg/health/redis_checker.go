package health

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisChecker Redis健康检查器
type RedisChecker struct {
	name    string
	client  *redis.Client
	timeout time.Duration
}

// NewRedisChecker 创建Redis健康检查器
func NewRedisChecker(name string, client *redis.Client) *RedisChecker {
	return &RedisChecker{
		name:    name,
		client:  client,
		timeout: 5 * time.Second, // 默认5秒超时
	}
}

// WithTimeout 设置超时时间
func (c *RedisChecker) WithTimeout(timeout time.Duration) *RedisChecker {
	c.timeout = timeout
	return c
}

// Name 返回检查器名称
func (c *RedisChecker) Name() string {
	return c.name
}

// Check 执行Redis健康检查
func (c *RedisChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// 设置超时上下文
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 1. 执行PING命令
	pong, err := c.client.Ping(checkCtx).Result()
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis连接失败"
		return result
	}

	if pong != "PONG" {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Redis PING响应异常: %s", pong)
		return result
	}

	// 2. 获取连接池统计信息
	stats := c.client.PoolStats()
	result.Metadata["hits"] = stats.Hits
	result.Metadata["misses"] = stats.Misses
	result.Metadata["timeouts"] = stats.Timeouts
	result.Metadata["total_conns"] = stats.TotalConns
	result.Metadata["idle_conns"] = stats.IdleConns
	result.Metadata["stale_conns"] = stats.StaleConns

	// 3. 尝试简单的SET/GET操作
	testKey := "__health_check__"
	testValue := fmt.Sprintf("%d", time.Now().Unix())

	err = c.client.Set(checkCtx, testKey, testValue, time.Second*10).Err()
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis写入失败"
		return result
	}

	val, err := c.client.Get(checkCtx, testKey).Result()
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis读取失败"
		return result
	}

	if val != testValue {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Message = "Redis读写数据不一致"
		return result
	}

	// 清理测试键
	c.client.Del(checkCtx, testKey)

	result.Duration = time.Since(startTime)

	// 4. 检查连接池状态
	if stats.Timeouts > 100 {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("Redis连接超时次数过多 (%d)", stats.Timeouts)
	} else if stats.StaleConns > 50 {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("Redis过期连接数过多 (%d)", stats.StaleConns)
	} else {
		result.Status = StatusHealthy
		result.Message = "Redis正常"
	}

	return result
}

// RedisClusterChecker Redis集群健康检查器
type RedisClusterChecker struct {
	name    string
	client  *redis.ClusterClient
	timeout time.Duration
}

// NewRedisClusterChecker 创建Redis集群健康检查器
func NewRedisClusterChecker(name string, client *redis.ClusterClient) *RedisClusterChecker {
	return &RedisClusterChecker{
		name:    name,
		client:  client,
		timeout: 5 * time.Second,
	}
}

// WithTimeout 设置超时时间
func (c *RedisClusterChecker) WithTimeout(timeout time.Duration) *RedisClusterChecker {
	c.timeout = timeout
	return c
}

// Name 返回检查器名称
func (c *RedisClusterChecker) Name() string {
	return c.name
}

// Check 执行Redis集群健康检查
func (c *RedisClusterChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// 设置超时上下文
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 执行PING命令
	err := c.client.ForEachShard(checkCtx, func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})

	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis集群连接失败"
		return result
	}

	result.Duration = time.Since(startTime)
	result.Status = StatusHealthy
	result.Message = "Redis集群正常"

	return result
}
