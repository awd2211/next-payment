package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ExportTask 导出任务
type ExportTask struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	MerchantID uuid.UUID `json:"merchant_id" gorm:"type:uuid;not null;index"`
	Type       string    `json:"type" gorm:"size:50;not null"` // payment, refund, settlement, etc.
	Format     string    `json:"format" gorm:"size:20;not null"` // csv, excel, pdf
	Status     string    `json:"status" gorm:"size:20;not null;index"` // pending, processing, completed, failed
	FileName   string    `json:"file_name" gorm:"size:255"`
	FilePath   string    `json:"file_path" gorm:"size:500"`
	FileSize   int64     `json:"file_size"`
	RowCount   int       `json:"row_count"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	ErrorMsg   string    `json:"error_msg" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// ExportService 导出服务
type ExportService struct {
	db          *gorm.DB
	redisClient *redis.Client
	storageDir  string
}

// NewExportService 创建导出服务
func NewExportService(db *gorm.DB, redisClient *redis.Client, storageDir string) *ExportService {
	// 确保存储目录存在
	os.MkdirAll(storageDir, 0755)

	return &ExportService{
		db:          db,
		redisClient: redisClient,
		storageDir:  storageDir,
	}
}

// CreateExportTask 创建导出任务
func (s *ExportService) CreateExportTask(ctx context.Context, merchantID uuid.UUID, exportType, format string, startDate, endDate time.Time) (*ExportTask, error) {
	task := &ExportTask{
		ID:         uuid.New(),
		MerchantID: merchantID,
		Type:       exportType,
		Format:     format,
		Status:     "pending",
		StartDate:  startDate,
		EndDate:    endDate,
		CreatedAt:  time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建导出任务失败: %w", err)
	}

	logger.Info("导出任务已创建",
		zap.String("task_id", task.ID.String()),
		zap.String("type", exportType),
		zap.String("format", format))

	return task, nil
}

// GetExportTask 获取导出任务
func (s *ExportService) GetExportTask(ctx context.Context, taskID uuid.UUID, merchantID uuid.UUID) (*ExportTask, error) {
	var task ExportTask
	err := s.db.WithContext(ctx).
		Where("id = ? AND merchant_id = ?", taskID, merchantID).
		First(&task).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("导出任务不存在")
	}

	return &task, err
}

// UpdateTaskStatus 更新任务状态
func (s *ExportService) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "completed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	return s.db.WithContext(ctx).
		Model(&ExportTask{}).
		Where("id = ?", taskID).
		Updates(updates).Error
}

// UpdateTaskFile 更新任务文件信息
func (s *ExportService) UpdateTaskFile(ctx context.Context, taskID uuid.UUID, fileName, filePath string, fileSize int64, rowCount int) error {
	return s.db.WithContext(ctx).
		Model(&ExportTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"file_name": fileName,
			"file_path": filePath,
			"file_size": fileSize,
			"row_count": rowCount,
		}).Error
}

// ExportToCSV 导出数据到CSV文件
// data: 要导出的数据（二维数组）
// headers: CSV表头
func (s *ExportService) ExportToCSV(ctx context.Context, taskID uuid.UUID, headers []string, data [][]string) error {
	// 生成文件名
	fileName := fmt.Sprintf("export_%s_%s.csv", taskID.String(), time.Now().Format("20060102_150405"))
	filePath := filepath.Join(s.storageDir, fileName)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建CSV文件失败: %w", err)
	}
	defer file.Close()

	// 写入UTF-8 BOM（Excel兼容性）
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	// 创建CSV写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("写入CSV表头失败: %w", err)
	}

	// 写入数据
	rowCount := 0
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入CSV数据失败: %w", err)
		}
		rowCount++
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("CSV写入器错误: %w", err)
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 更新任务信息
	if err := s.UpdateTaskFile(ctx, taskID, fileName, filePath, fileInfo.Size(), rowCount); err != nil {
		return fmt.Errorf("更新任务文件信息失败: %w", err)
	}

	logger.Info("CSV导出完成",
		zap.String("task_id", taskID.String()),
		zap.String("file_path", filePath),
		zap.Int("row_count", rowCount),
		zap.Int64("file_size", fileInfo.Size()))

	return nil
}

// CleanupExpiredTasks 清理过期任务（建议定时调用）
func (s *ExportService) CleanupExpiredTasks(ctx context.Context, expireDays int) error {
	expireDate := time.Now().AddDate(0, 0, -expireDays)

	var tasks []ExportTask
	if err := s.db.WithContext(ctx).
		Where("created_at < ? AND status = ?", expireDate, "completed").
		Find(&tasks).Error; err != nil {
		return err
	}

	for _, task := range tasks {
		// 删除文件
		if task.FilePath != "" {
			os.Remove(task.FilePath)
		}

		// 删除数据库记录
		s.db.WithContext(ctx).Delete(&task)

		logger.Info("已清理过期导出任务",
			zap.String("task_id", task.ID.String()),
			zap.String("file_path", task.FilePath))
	}

	return nil
}

// ListExportTasks 查询导出任务列表
func (s *ExportService) ListExportTasks(ctx context.Context, merchantID uuid.UUID, page, pageSize int) ([]ExportTask, int64, error) {
	var tasks []ExportTask
	var total int64

	query := s.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC")

	// 计算总数
	if err := query.Model(&ExportTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}
