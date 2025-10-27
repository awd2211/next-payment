package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"payment-platform/reconciliation-service/internal/model"
	"payment-platform/reconciliation-service/internal/notifier"
	"payment-platform/reconciliation-service/internal/service"
)

// DailyScheduler 每日自动对账调度器
type DailyScheduler struct {
	reconService service.ReconciliationService
	alerter      *notifier.AlertNotifier
	logger       *zap.Logger
	stopChan     chan struct{}
}

// NewDailyScheduler 创建每日调度器
func NewDailyScheduler(
	reconService service.ReconciliationService,
	alerter *notifier.AlertNotifier,
	logger *zap.Logger,
) *DailyScheduler {
	return &DailyScheduler{
		reconService: reconService,
		alerter:      alerter,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动调度器（每天凌晨2点执行）
func (s *DailyScheduler) Start() {
	s.logger.Info("Starting daily reconciliation scheduler")

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// 立即执行一次检查（用于测试）
		// s.runDailyReconciliation()

		// 等待下一个凌晨2点
		s.waitUntil(2, 0)

		for {
			select {
			case <-ticker.C:
				s.runDailyReconciliation()
			case <-s.stopChan:
				s.logger.Info("Stopping daily reconciliation scheduler")
				return
			}
		}
	}()
}

// Stop 停止调度器
func (s *DailyScheduler) Stop() {
	close(s.stopChan)
}

// runDailyReconciliation 执行每日对账任务
func (s *DailyScheduler) runDailyReconciliation() {
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)

	s.logger.Info("Starting daily reconciliation",
		zap.Time("reconciliation_date", yesterday))

	// 自动创建对账任务（针对所有渠道）
	channels := []string{"stripe", "paypal", "alipay", "wechat"}

	for _, channel := range channels {
		s.reconcileChannel(ctx, channel, yesterday)
	}

	s.logger.Info("Daily reconciliation completed")
}

// reconcileChannel 对单个渠道执行对账
func (s *DailyScheduler) reconcileChannel(ctx context.Context, channel string, date time.Time) {
	s.logger.Info("Reconciling channel",
		zap.String("channel", channel),
		zap.Time("date", date))

	// 创建对账任务
	input := &service.CreateTaskInput{
		TaskDate:    date,
		Channel:     channel,
		TaskType:    "automated_daily",
		Description: "自动每日对账",
	}

	task, err := s.reconService.CreateTask(ctx, input)
	if err != nil {
		s.logger.Error("Failed to create reconciliation task",
			zap.String("channel", channel),
			zap.Error(err))
		return
	}

	// 执行对账任务
	if err := s.reconService.ExecuteTask(ctx, task.ID); err != nil {
		s.logger.Error("Failed to execute reconciliation task",
			zap.String("task_id", task.ID.String()),
			zap.Error(err))

		// 更新任务状态为失败
		s.updateTaskStatus(ctx, task.ID, model.ReconciliationStatusFailed)
		return
	}

	// 查询任务结果
	completedTask, err := s.reconService.GetTask(ctx, task.ID)
	if err != nil {
		s.logger.Error("Failed to get task result",
			zap.String("task_id", task.ID.String()),
			zap.Error(err))
		return
	}

	// 如果有差异，发送告警
	if completedTask.DifferenceCount > 0 {
		differences, err := s.reconService.GetTaskDifferences(ctx, task.ID, 0, 100)
		if err != nil {
			s.logger.Error("Failed to get task differences",
				zap.String("task_id", task.ID.String()),
				zap.Error(err))
			return
		}

		// 发送差异告警
		if err := s.alerter.SendDifferenceAlert(ctx, completedTask, differences); err != nil {
			s.logger.Error("Failed to send difference alert",
				zap.String("task_id", task.ID.String()),
				zap.Error(err))
		}

		// 如果有严重差异，发送紧急告警
		criticalDiffs := s.filterCriticalDifferences(differences)
		if len(criticalDiffs) > 0 {
			if err := s.alerter.SendCriticalAlert(ctx, completedTask, criticalDiffs); err != nil {
				s.logger.Error("Failed to send critical alert",
					zap.String("task_id", task.ID.String()),
					zap.Error(err))
			}
		}
	}

	s.logger.Info("Channel reconciliation completed",
		zap.String("channel", channel),
		zap.String("status", completedTask.Status),
		zap.Int("differences", completedTask.DifferenceCount))
}

// updateTaskStatus 更新任务状态
func (s *DailyScheduler) updateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) {
	if err := s.reconService.UpdateTaskStatus(ctx, taskID, status); err != nil {
		s.logger.Error("Failed to update task status",
			zap.String("task_id", taskID.String()),
			zap.String("status", status),
			zap.Error(err))
	}
}

// filterCriticalDifferences 过滤严重差异
func (s *DailyScheduler) filterCriticalDifferences(differences []*model.ReconciliationDifference) []*model.ReconciliationDifference {
	critical := make([]*model.ReconciliationDifference, 0)
	for _, diff := range differences {
		if diff.Severity == "critical" {
			critical = append(critical, diff)
		}
	}
	return critical
}

// waitUntil 等待到指定的小时和分钟
func (s *DailyScheduler) waitUntil(hour, minute int) {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// 如果今天的目标时间已过，等到明天
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	duration := time.Until(next)
	s.logger.Info("Waiting until next execution",
		zap.Time("next_execution", next),
		zap.Duration("wait_duration", duration))

	time.Sleep(duration)
}
