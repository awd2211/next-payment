package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ArchiveConfig 归档配置
type ArchiveConfig struct {
	TableName      string        // 表名
	ArchiveTable   string        // 归档表名
	DateColumn     string        // 日期字段名
	RetentionDays  int           // 保留天数
	BatchSize      int           // 每批处理数量
}

// ArchiveTask 数据归档任务
type ArchiveTask struct {
	db      *gorm.DB
	configs []ArchiveConfig
}

// NewArchiveTask 创建归档任务
func NewArchiveTask(db *gorm.DB, configs []ArchiveConfig) *ArchiveTask {
	return &ArchiveTask{
		db:      db,
		configs: configs,
	}
}

// Run 执行归档任务
func (t *ArchiveTask) Run(ctx context.Context) error {
	logger.Info("开始执行数据归档任务")

	totalArchived := 0
	totalDeleted := 0

	for _, config := range t.configs {
		archived, deleted, err := t.archiveTable(ctx, config)
		if err != nil {
			logger.Error("归档表失败",
				zap.String("table", config.TableName),
				zap.Error(err))
			continue
		}

		totalArchived += archived
		totalDeleted += deleted

		logger.Info("表归档完成",
			zap.String("table", config.TableName),
			zap.Int("archived", archived),
			zap.Int("deleted", deleted))
	}

	logger.Info("数据归档任务完成",
		zap.Int("total_archived", totalArchived),
		zap.Int("total_deleted", totalDeleted))

	return nil
}

// archiveTable 归档单个表
func (t *ArchiveTask) archiveTable(ctx context.Context, config ArchiveConfig) (int, int, error) {
	cutoffDate := time.Now().AddDate(0, 0, -config.RetentionDays)

	archivedCount := 0
	deletedCount := 0

	// 1. 复制数据到归档表
	if config.ArchiveTable != "" {
		query := fmt.Sprintf(`
			INSERT INTO %s
			SELECT * FROM %s
			WHERE %s < ?
			AND NOT EXISTS (
				SELECT 1 FROM %s arch
				WHERE arch.id = %s.id
			)
			LIMIT ?
		`, config.ArchiveTable, config.TableName, config.DateColumn,
			config.ArchiveTable, config.TableName)

		for {
			result := t.db.WithContext(ctx).Exec(query, cutoffDate, config.BatchSize)
			if result.Error != nil {
				return archivedCount, deletedCount, fmt.Errorf("归档数据失败: %w", result.Error)
			}

			rowsAffected := int(result.RowsAffected)
			archivedCount += rowsAffected

			if rowsAffected < config.BatchSize {
				break // 全部归档完成
			}

			// 避免长时间占用资源
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 2. 删除已归档的旧数据
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE %s < ?
		LIMIT ?
	`, config.TableName, config.DateColumn)

	for {
		result := t.db.WithContext(ctx).Exec(deleteQuery, cutoffDate, config.BatchSize)
		if result.Error != nil {
			return archivedCount, deletedCount, fmt.Errorf("删除旧数据失败: %w", result.Error)
		}

		rowsAffected := int(result.RowsAffected)
		deletedCount += rowsAffected

		if rowsAffected < config.BatchSize {
			break // 全部删除完成
		}

		time.Sleep(100 * time.Millisecond)
	}

	return archivedCount, deletedCount, nil
}

// RunArchiveTask 运行归档任务（供调度器使用）
func RunArchiveTask(db *gorm.DB) func(context.Context) error {
	// 配置需要归档的表
	configs := []ArchiveConfig{
		{
			TableName:     "payment_callbacks",
			ArchiveTable:  "payment_callbacks_archive",
			DateColumn:    "created_at",
			RetentionDays: 90, // 保留90天
			BatchSize:     1000,
		},
		{
			TableName:     "notifications",
			ArchiveTable:  "notifications_archive",
			DateColumn:    "created_at",
			RetentionDays: 30, // 保留30天
			BatchSize:     1000,
		},
		{
			TableName:     "audit_logs",
			ArchiveTable:  "audit_logs_archive",
			DateColumn:    "created_at",
			RetentionDays: 180, // 保留180天
			BatchSize:     1000,
		},
		{
			TableName:     "export_tasks",
			ArchiveTable:  "", // 不归档，直接删除
			DateColumn:    "created_at",
			RetentionDays: 7, // 保留7天
			BatchSize:     100,
		},
	}

	task := NewArchiveTask(db, configs)
	return task.Run
}
