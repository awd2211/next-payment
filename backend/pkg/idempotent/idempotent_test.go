package idempotent

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	// 使用 miniredis 进行单元测试
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestIdempotentService_CheckAndStore(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	svc := NewService(client)
	ctx := context.Background()

	// 测试数据
	key := "payment:test-merchant:order-123"
	type TestResult struct {
		PaymentNo string `json:"payment_no"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount"`
	}

	originalResult := &TestResult{
		PaymentNo: "PAY-001",
		Status:    "success",
		Amount:    10000,
	}

	// 1. 首次检查 - 应该不存在
	exists, err := svc.Check(ctx, key, nil)
	assert.NoError(t, err)
	assert.False(t, exists, "首次检查应该返回不存在")

	// 2. 存储结果
	err = svc.Store(ctx, key, originalResult, 24*time.Hour)
	assert.NoError(t, err)

	// 3. 再次检查 - 应该存在并返回相同结果
	var cachedResult TestResult
	exists, err = svc.Check(ctx, key, &cachedResult)
	assert.NoError(t, err)
	assert.True(t, exists, "第二次检查应该返回存在")
	assert.Equal(t, originalResult.PaymentNo, cachedResult.PaymentNo)
	assert.Equal(t, originalResult.Status, cachedResult.Status)
	assert.Equal(t, originalResult.Amount, cachedResult.Amount)

	// 4. 删除记录
	err = svc.Delete(ctx, key)
	assert.NoError(t, err)

	// 5. 删除后再检查 - 应该不存在
	exists, err = svc.Check(ctx, key, nil)
	assert.NoError(t, err)
	assert.False(t, exists, "删除后应该返回不存在")
}

func TestIdempotentService_DistributedLock(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	svc := NewService(client)
	ctx := context.Background()

	lockKey := "payment:test-merchant:order-456"

	// 1. 首次获取锁 - 应该成功
	acquired, err := svc.Try(ctx, lockKey, 10*time.Second)
	assert.NoError(t, err)
	assert.True(t, acquired, "首次获取锁应该成功")

	// 2. 再次获取同一个锁 - 应该失败
	acquired, err = svc.Try(ctx, lockKey, 10*time.Second)
	assert.NoError(t, err)
	assert.False(t, acquired, "重复获取同一个锁应该失败")

	// 3. 释放锁
	err = svc.Release(ctx, lockKey)
	assert.NoError(t, err)

	// 4. 释放后再次获取 - 应该成功
	acquired, err = svc.Try(ctx, lockKey, 10*time.Second)
	assert.NoError(t, err)
	assert.True(t, acquired, "释放后再次获取锁应该成功")
}

func TestIdempotentService_TTL(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	svc := NewService(client)
	ctx := context.Background()

	key := "payment:test-ttl"
	result := map[string]interface{}{
		"test": "data",
	}

	// 存储结果，设置 1 秒过期
	err := svc.Store(ctx, key, result, 1*time.Second)
	assert.NoError(t, err)

	// 立即检查 - 应该存在
	exists, err := svc.Check(ctx, key, nil)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 使用 miniredis 快进时间
	mr.FastForward(2 * time.Second)

	// 过期后检查 - 应该不存在
	exists, err = svc.Check(ctx, key, nil)
	assert.NoError(t, err)
	assert.False(t, exists, "过期后应该返回不存在")
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "single part",
			parts:    []string{"payment"},
			expected: "payment",
		},
		{
			name:     "multiple parts",
			parts:    []string{"payment", "merchant-123", "order-456"},
			expected: "payment:merchant-123:order-456",
		},
		{
			name:     "empty parts",
			parts:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateKey(tt.parts...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkIdempotentService_Check(b *testing.B) {
	client, mr := setupTestRedis(&testing.T{})
	defer mr.Close()
	defer client.Close()

	svc := NewService(client)
	ctx := context.Background()

	// 预先存储一些数据
	for i := 0; i < 100; i++ {
		key := GenerateKey("payment", "bench", string(rune(i)))
		svc.Store(ctx, key, map[string]interface{}{"test": i}, time.Hour)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := GenerateKey("payment", "bench", string(rune(i%100)))
		svc.Check(ctx, key, nil)
	}
}
