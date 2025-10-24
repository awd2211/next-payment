package saga

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RecoveryWorker Saga 恢复工作器（处理失败的补偿）
type RecoveryWorker struct {
	orchestrator *SagaOrchestrator
	interval     time.Duration
	batchSize    int
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewRecoveryWorker 创建恢复工作器
func NewRecoveryWorker(orchestrator *SagaOrchestrator, interval time.Duration, batchSize int) *RecoveryWorker {
	if interval <= 0 {
		interval = 5 * time.Minute // 默认5分钟检查一次
	}
	if batchSize <= 0 {
		batchSize = 10 // 默认每次处理10个
	}

	return &RecoveryWorker{
		orchestrator: orchestrator,
		interval:     interval,
		batchSize:    batchSize,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动恢复工作器
func (w *RecoveryWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
	logger.Info("saga recovery worker started",
		zap.Duration("interval", w.interval),
		zap.Int("batch_size", w.batchSize))
}

// Stop 停止恢复工作器
func (w *RecoveryWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	logger.Info("saga recovery worker stopped")
}

// run 运行工作器主循环
func (w *RecoveryWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 启动时立即执行一次
	w.processFailedSagas(ctx)

	for {
		select {
		case <-ticker.C:
			w.processFailedSagas(ctx)
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// processFailedSagas 处理失败的 Saga
func (w *RecoveryWorker) processFailedSagas(ctx context.Context) {
	logger.Info("processing failed sagas", zap.Int("batch_size", w.batchSize))

	// 获取失败的 Saga 列表
	failedSagas, err := w.orchestrator.ListFailedSagas(ctx, w.batchSize)
	if err != nil {
		logger.Error("failed to list failed sagas", zap.Error(err))
		return
	}

	if len(failedSagas) == 0 {
		logger.Info("no failed sagas to process")
		return
	}

	logger.Info("found failed sagas to process", zap.Int("count", len(failedSagas)))

	successCount := 0
	failureCount := 0

	for _, saga := range failedSagas {
		// 检查是否需要进入死信队列（DLQ）
		if w.shouldMoveToDLQ(saga) {
			if err := w.moveToDLQ(ctx, saga); err != nil {
				logger.Error("failed to move saga to DLQ",
					zap.String("saga_id", saga.ID.String()),
					zap.Error(err))
			}
			failureCount++
			continue
		}

		// 尝试重试补偿（注意：这里需要stepDefs，实际使用时需要从业务层传入）
		// 这是一个示例，实际实现需要根据业务类型获取对应的stepDefs
		logger.Warn("saga recovery requires business-specific step definitions",
			zap.String("saga_id", saga.ID.String()),
			zap.String("business_id", saga.BusinessID),
			zap.String("business_type", saga.BusinessType))

		// 实际项目中，可以：
		// 1. 发送到 Kafka 队列，由对应的业务服务消费
		// 2. 调用业务服务的 API 触发重试
		// 3. 使用注册表模式，根据 business_type 获取对应的 stepDefs

		failureCount++
	}

	logger.Info("failed saga processing completed",
		zap.Int("success", successCount),
		zap.Int("failure", failureCount))
}

// shouldMoveToDLQ 判断是否应该移动到死信队列
func (w *RecoveryWorker) shouldMoveToDLQ(saga *Saga) bool {
	// 失败超过3天，或者补偿失败次数过多
	if time.Since(saga.UpdatedAt) > 3*24*time.Hour {
		return true
	}

	// 统计补偿失败的步骤数
	failedCompensationCount := 0
	for _, step := range saga.Steps {
		if step.Status == StepStatusFailed && step.RetryCount >= 3 {
			failedCompensationCount++
		}
	}

	// 如果所有需要补偿的步骤都失败了，移动到 DLQ
	return failedCompensationCount >= saga.CurrentStep
}

// moveToDLQ 移动到死信队列
func (w *RecoveryWorker) moveToDLQ(ctx context.Context, saga *Saga) error {
	logger.Info("moving saga to dead letter queue",
		zap.String("saga_id", saga.ID.String()),
		zap.String("business_id", saga.BusinessID),
		zap.String("business_type", saga.BusinessType))

	// 使用 Redis 保存 DLQ 记录
	if w.orchestrator.redis != nil {
		dlqKey := fmt.Sprintf("saga:dlq:%s", saga.ID.String())
		dlqData := map[string]interface{}{
			"saga_id":       saga.ID.String(),
			"business_id":   saga.BusinessID,
			"business_type": saga.BusinessType,
			"error_message": saga.ErrorMessage,
			"moved_at":      time.Now().Format(time.RFC3339),
		}

		// 保存到 Redis Hash，永久保留（需要人工处理）
		if err := w.orchestrator.redis.HSet(ctx, dlqKey, dlqData).Err(); err != nil {
			return fmt.Errorf("failed to save to DLQ: %w", err)
		}

		// 添加到 DLQ 集合，方便查询
		dlqSetKey := "saga:dlq:set"
		if err := w.orchestrator.redis.ZAdd(ctx, dlqSetKey,
			redis.Z{
				Score:  float64(time.Now().Unix()),
				Member: saga.ID.String(),
			},
		).Err(); err != nil {
			return fmt.Errorf("failed to add to DLQ set: %w", err)
		}
	}

	// 更新数据库记录的错误信息
	saga.ErrorMessage = fmt.Sprintf("[DLQ] %s (moved to dead letter queue at %s)",
		saga.ErrorMessage, time.Now().Format(time.RFC3339))
	if err := w.orchestrator.db.Save(saga).Error; err != nil {
		return fmt.Errorf("failed to update saga: %w", err)
	}

	logger.Info("saga moved to DLQ successfully",
		zap.String("saga_id", saga.ID.String()),
		zap.String("business_id", saga.BusinessID))

	return nil
}

// GetDLQSagas 获取死信队列中的 Saga 列表
func (w *RecoveryWorker) GetDLQSagas(ctx context.Context, limit int) ([]string, error) {
	if w.orchestrator.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	dlqSetKey := "saga:dlq:set"
	// 按时间倒序获取（最新的在前）
	sagaIDs, err := w.orchestrator.redis.ZRevRange(ctx, dlqSetKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ sagas: %w", err)
	}

	return sagaIDs, nil
}

// RemoveFromDLQ 从死信队列中移除（人工处理后）
func (w *RecoveryWorker) RemoveFromDLQ(ctx context.Context, sagaID string) error {
	if w.orchestrator.redis == nil {
		return fmt.Errorf("redis not available")
	}

	dlqKey := fmt.Sprintf("saga:dlq:%s", sagaID)
	dlqSetKey := "saga:dlq:set"

	// 删除 Hash 数据
	if err := w.orchestrator.redis.Del(ctx, dlqKey).Err(); err != nil {
		return fmt.Errorf("failed to delete DLQ record: %w", err)
	}

	// 从集合中移除
	if err := w.orchestrator.redis.ZRem(ctx, dlqSetKey, sagaID).Err(); err != nil {
		return fmt.Errorf("failed to remove from DLQ set: %w", err)
	}

	logger.Info("saga removed from DLQ", zap.String("saga_id", sagaID))
	return nil
}
