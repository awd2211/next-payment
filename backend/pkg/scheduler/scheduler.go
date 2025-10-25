package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待执行
	TaskStatusRunning   TaskStatus = "running"   // 正在执行
	TaskStatusCompleted TaskStatus = "completed" // 执行成功
	TaskStatusFailed    TaskStatus = "failed"    // 执行失败
	TaskStatusSkipped   TaskStatus = "skipped"   // 被跳过（上次未完成）
)

// ScheduledTask 定时任务记录
type ScheduledTask struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskName    string     `gorm:"type:varchar(100);not null;index" json:"task_name"`
	Schedule    string     `gorm:"type:varchar(50);not null" json:"schedule"` // cron表达式或间隔时间
	Status      TaskStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	LastRunAt   *time.Time `gorm:"type:timestamptz" json:"last_run_at"`
	NextRunAt   *time.Time `gorm:"type:timestamptz;index" json:"next_run_at"`
	Duration    int64      `gorm:"type:bigint" json:"duration"` // 执行时长（毫秒）
	ErrorMsg    string     `gorm:"type:text" json:"error_msg"`
	RunCount    int        `gorm:"type:integer;default:0" json:"run_count"`
	SuccessCount int       `gorm:"type:integer;default:0" json:"success_count"`
	FailedCount int        `gorm:"type:integer;default:0" json:"failed_count"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

// TaskFunc 任务执行函数
type TaskFunc func(ctx context.Context) error

// Task 任务定义
type Task struct {
	Name        string        // 任务名称
	Interval    time.Duration // 执行间隔
	Func        TaskFunc      // 执行函数
	Description string        // 任务描述
}

// Scheduler 定时任务调度器
type Scheduler struct {
	db          *gorm.DB
	redisClient *redis.Client
	tasks       map[string]*Task
	tasksMu     sync.RWMutex
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewScheduler 创建调度器
func NewScheduler(db *gorm.DB, redisClient *redis.Client) *Scheduler {
	return &Scheduler{
		db:          db,
		redisClient: redisClient,
		tasks:       make(map[string]*Task),
		stopCh:      make(chan struct{}),
	}
}

// RegisterTask 注册任务
func (s *Scheduler) RegisterTask(task *Task) error {
	s.tasksMu.Lock()
	defer s.tasksMu.Unlock()

	if _, exists := s.tasks[task.Name]; exists {
		return fmt.Errorf("任务已存在: %s", task.Name)
	}

	s.tasks[task.Name] = task

	// 在数据库中初始化任务记录
	var dbTask ScheduledTask
	err := s.db.Where("task_name = ?", task.Name).First(&dbTask).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新任务记录
		nextRun := time.Now().Add(task.Interval)
		dbTask = ScheduledTask{
			TaskName:  task.Name,
			Schedule:  task.Interval.String(),
			Status:    TaskStatusPending,
			NextRunAt: &nextRun,
		}
		if err := s.db.Create(&dbTask).Error; err != nil {
			return fmt.Errorf("创建任务记录失败: %w", err)
		}
	}

	logger.Info("定时任务已注册",
		zap.String("task_name", task.Name),
		zap.Duration("interval", task.Interval),
		zap.String("description", task.Description))

	return nil
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	logger.Info("定时任务调度器启动")

	s.tasksMu.RLock()
	taskCount := len(s.tasks)
	s.tasksMu.RUnlock()

	logger.Info(fmt.Sprintf("已注册 %d 个定时任务", taskCount))

	// 为每个任务启动独立的goroutine
	s.tasksMu.RLock()
	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runTask(ctx, task)
	}
	s.tasksMu.RUnlock()

	// 等待停止信号
	<-s.stopCh
	s.wg.Wait()

	logger.Info("定时任务调度器已停止")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// runTask 运行单个任务
func (s *Scheduler) runTask(ctx context.Context, task *Task) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	logger.Info("定时任务已启动",
		zap.String("task_name", task.Name),
		zap.Duration("interval", task.Interval))

	for {
		select {
		case <-ticker.C:
			s.executeTask(ctx, task)

		case <-s.stopCh:
			logger.Info("定时任务收到停止信号", zap.String("task_name", task.Name))
			return

		case <-ctx.Done():
			logger.Info("定时任务上下文取消", zap.String("task_name", task.Name))
			return
		}
	}
}

// executeTask 执行任务
func (s *Scheduler) executeTask(ctx context.Context, task *Task) {
	// 使用Redis分布式锁防止重复执行
	lockKey := fmt.Sprintf("scheduler:lock:%s", task.Name)
	locked, err := s.acquireLock(ctx, lockKey, task.Interval)
	if err != nil {
		logger.Error("获取任务锁失败",
			zap.String("task_name", task.Name),
			zap.Error(err))
		return
	}

	if !locked {
		logger.Info("任务正在其他节点执行，跳过",
			zap.String("task_name", task.Name))
		s.updateTaskStatus(ctx, task.Name, TaskStatusSkipped, "")
		return
	}

	defer s.releaseLock(ctx, lockKey)

	// 更新任务状态为运行中
	s.updateTaskStatus(ctx, task.Name, TaskStatusRunning, "")

	startTime := time.Now()
	logger.Info("开始执行定时任务", zap.String("task_name", task.Name))

	// 执行任务
	err = task.Func(ctx)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("定时任务执行失败",
			zap.String("task_name", task.Name),
			zap.Duration("duration", duration),
			zap.Error(err))
		s.updateTaskResult(ctx, task.Name, TaskStatusFailed, duration, err.Error())
	} else {
		logger.Info("定时任务执行成功",
			zap.String("task_name", task.Name),
			zap.Duration("duration", duration))
		s.updateTaskResult(ctx, task.Name, TaskStatusCompleted, duration, "")
	}
}

// acquireLock 获取分布式锁
func (s *Scheduler) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if s.redisClient == nil {
		return true, nil // 无Redis时降级为单机模式
	}

	result, err := s.redisClient.SetNX(ctx, key, "1", ttl).Result()
	return result, err
}

// releaseLock 释放分布式锁
func (s *Scheduler) releaseLock(ctx context.Context, key string) {
	if s.redisClient == nil {
		return
	}

	s.redisClient.Del(ctx, key)
}

// updateTaskStatus 更新任务状态
func (s *Scheduler) updateTaskStatus(ctx context.Context, taskName string, status TaskStatus, errorMsg string) {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == TaskStatusRunning {
		now := time.Now()
		updates["last_run_at"] = &now
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	s.db.WithContext(ctx).
		Model(&ScheduledTask{}).
		Where("task_name = ?", taskName).
		Updates(updates)
}

// updateTaskResult 更新任务执行结果
func (s *Scheduler) updateTaskResult(ctx context.Context, taskName string, status TaskStatus, duration time.Duration, errorMsg string) {
	// 获取任务以计算下次执行时间
	s.tasksMu.RLock()
	task, exists := s.tasks[taskName]
	s.tasksMu.RUnlock()

	if !exists {
		return
	}

	nextRun := time.Now().Add(task.Interval)

	updates := map[string]interface{}{
		"status":      status,
		"duration":    duration.Milliseconds(),
		"next_run_at": &nextRun,
		"updated_at":  time.Now(),
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
		s.db.WithContext(ctx).
			Model(&ScheduledTask{}).
			Where("task_name = ?", taskName).
			Updates(updates).
			UpdateColumn("failed_count", gorm.Expr("failed_count + 1"))
	} else {
		updates["error_msg"] = ""
		s.db.WithContext(ctx).
			Model(&ScheduledTask{}).
			Where("task_name = ?", taskName).
			Updates(updates).
			UpdateColumn("success_count", gorm.Expr("success_count + 1"))
	}

	// 更新运行次数
	s.db.WithContext(ctx).
		Model(&ScheduledTask{}).
		Where("task_name = ?", taskName).
		UpdateColumn("run_count", gorm.Expr("run_count + 1"))
}

// GetTaskStatus 获取任务状态
func (s *Scheduler) GetTaskStatus(ctx context.Context, taskName string) (*ScheduledTask, error) {
	var task ScheduledTask
	err := s.db.WithContext(ctx).
		Where("task_name = ?", taskName).
		First(&task).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("任务不存在")
	}

	return &task, err
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks(ctx context.Context) ([]ScheduledTask, error) {
	var tasks []ScheduledTask
	err := s.db.WithContext(ctx).
		Order("task_name ASC").
		Find(&tasks).Error

	return tasks, err
}
