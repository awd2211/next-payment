package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payment-platform/reconciliation-service/internal/model"
)

// ReconciliationRepository 对账仓储接口
type ReconciliationRepository interface {
	// 任务管理
	CreateTask(ctx context.Context, task *model.ReconciliationTask) error
	GetTaskByID(ctx context.Context, id uuid.UUID) (*model.ReconciliationTask, error)
	GetTaskByNo(ctx context.Context, taskNo string) (*model.ReconciliationTask, error)
	GetTaskByDateAndChannel(ctx context.Context, taskDate time.Time, channel string) (*model.ReconciliationTask, error)
	UpdateTask(ctx context.Context, task *model.ReconciliationTask) error
	ListTasks(ctx context.Context, filters TaskFilters, page, pageSize int) ([]*model.ReconciliationTask, int64, error)

	// 差异记录管理
	CreateRecord(ctx context.Context, record *model.ReconciliationRecord) error
	BatchCreateRecords(ctx context.Context, records []*model.ReconciliationRecord) error
	GetRecordByID(ctx context.Context, id uuid.UUID) (*model.ReconciliationRecord, error)
	ListRecords(ctx context.Context, filters RecordFilters, page, pageSize int) ([]*model.ReconciliationRecord, int64, error)
	ResolveRecord(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID, note string) error
	CountRecordsByTask(ctx context.Context, taskID uuid.UUID, diffType string) (int, error)

	// 文件管理
	CreateFile(ctx context.Context, file *model.ChannelSettlementFile) error
	GetFileByNo(ctx context.Context, fileNo string) (*model.ChannelSettlementFile, error)
	GetFileByDateAndChannel(ctx context.Context, settlementDate time.Time, channel string) (*model.ChannelSettlementFile, error)
	UpdateFile(ctx context.Context, file *model.ChannelSettlementFile) error
	ListFiles(ctx context.Context, filters FileFilters, page, pageSize int) ([]*model.ChannelSettlementFile, int64, error)
}

// TaskFilters 任务查询过滤条件
type TaskFilters struct {
	TaskDate   *time.Time
	Channel    string
	Status     string
	StartDate  *time.Time
	EndDate    *time.Time
}

// RecordFilters 差异记录查询过滤条件
type RecordFilters struct {
	TaskID     *uuid.UUID
	DiffType   string
	IsResolved *bool
	MerchantID *uuid.UUID
}

// FileFilters 文件查询过滤条件
type FileFilters struct {
	Channel        string
	SettlementDate *time.Time
	Status         string
	StartDate      *time.Time
	EndDate        *time.Time
}

type reconciliationRepository struct {
	db *gorm.DB
}

// NewReconciliationRepository 创建对账仓储实例
func NewReconciliationRepository(db *gorm.DB) ReconciliationRepository {
	return &reconciliationRepository{db: db}
}

// CreateTask 创建对账任务
func (r *reconciliationRepository) CreateTask(ctx context.Context, task *model.ReconciliationTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetTaskByID 根据ID查询任务
func (r *reconciliationRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.ReconciliationTask, error) {
	var task model.ReconciliationTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &task, err
}

// GetTaskByNo 根据任务号查询
func (r *reconciliationRepository) GetTaskByNo(ctx context.Context, taskNo string) (*model.ReconciliationTask, error) {
	var task model.ReconciliationTask
	err := r.db.WithContext(ctx).Where("task_no = ?", taskNo).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &task, err
}

// GetTaskByDateAndChannel 根据日期和渠道查询任务
func (r *reconciliationRepository) GetTaskByDateAndChannel(ctx context.Context, taskDate time.Time, channel string) (*model.ReconciliationTask, error) {
	var task model.ReconciliationTask
	err := r.db.WithContext(ctx).
		Where("task_date = ? AND channel = ?", taskDate.Format("2006-01-02"), channel).
		First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &task, err
}

// UpdateTask 更新任务
func (r *reconciliationRepository) UpdateTask(ctx context.Context, task *model.ReconciliationTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// ListTasks 查询任务列表
func (r *reconciliationRepository) ListTasks(ctx context.Context, filters TaskFilters, page, pageSize int) ([]*model.ReconciliationTask, int64, error) {
	var tasks []*model.ReconciliationTask
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ReconciliationTask{})

	// 应用过滤条件
	if filters.TaskDate != nil {
		query = query.Where("task_date = ?", filters.TaskDate.Format("2006-01-02"))
	}
	if filters.Channel != "" {
		query = query.Where("channel = ?", filters.Channel)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.StartDate != nil {
		query = query.Where("task_date >= ?", filters.StartDate.Format("2006-01-02"))
	}
	if filters.EndDate != nil {
		query = query.Where("task_date <= ?", filters.EndDate.Format("2006-01-02"))
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count tasks failed: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.
		Order("task_date DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("list tasks failed: %w", err)
	}

	return tasks, total, nil
}

// CreateRecord 创建差异记录
func (r *reconciliationRepository) CreateRecord(ctx context.Context, record *model.ReconciliationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// BatchCreateRecords 批量创建差异记录
func (r *reconciliationRepository) BatchCreateRecords(ctx context.Context, records []*model.ReconciliationRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

// GetRecordByID 根据ID查询差异记录
func (r *reconciliationRepository) GetRecordByID(ctx context.Context, id uuid.UUID) (*model.ReconciliationRecord, error) {
	var record model.ReconciliationRecord
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &record, err
}

// ListRecords 查询差异记录列表
func (r *reconciliationRepository) ListRecords(ctx context.Context, filters RecordFilters, page, pageSize int) ([]*model.ReconciliationRecord, int64, error) {
	var records []*model.ReconciliationRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ReconciliationRecord{})

	// 应用过滤条件
	if filters.TaskID != nil {
		query = query.Where("task_id = ?", *filters.TaskID)
	}
	if filters.DiffType != "" {
		query = query.Where("diff_type = ?", filters.DiffType)
	}
	if filters.IsResolved != nil {
		query = query.Where("is_resolved = ?", *filters.IsResolved)
	}
	if filters.MerchantID != nil {
		query = query.Where("merchant_id = ?", *filters.MerchantID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count records failed: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list records failed: %w", err)
	}

	return records, total, nil
}

// ResolveRecord 标记差异已解决
func (r *reconciliationRepository) ResolveRecord(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID, note string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.ReconciliationRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_resolved":     true,
			"resolved_by":     resolvedBy,
			"resolved_at":     now,
			"resolution_note": note,
			"updated_at":      now,
		}).Error
}

// CountRecordsByTask 统计任务的差异记录数
func (r *reconciliationRepository) CountRecordsByTask(ctx context.Context, taskID uuid.UUID, diffType string) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.ReconciliationRecord{}).
		Where("task_id = ?", taskID)

	if diffType != "" {
		query = query.Where("diff_type = ?", diffType)
	}

	err := query.Count(&count).Error
	return int(count), err
}

// CreateFile 创建渠道账单文件记录
func (r *reconciliationRepository) CreateFile(ctx context.Context, file *model.ChannelSettlementFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetFileByNo 根据文件号查询
func (r *reconciliationRepository) GetFileByNo(ctx context.Context, fileNo string) (*model.ChannelSettlementFile, error) {
	var file model.ChannelSettlementFile
	err := r.db.WithContext(ctx).Where("file_no = ?", fileNo).First(&file).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &file, err
}

// GetFileByDateAndChannel 根据日期和渠道查询文件
func (r *reconciliationRepository) GetFileByDateAndChannel(ctx context.Context, settlementDate time.Time, channel string) (*model.ChannelSettlementFile, error) {
	var file model.ChannelSettlementFile
	err := r.db.WithContext(ctx).
		Where("settlement_date = ? AND channel = ?", settlementDate.Format("2006-01-02"), channel).
		First(&file).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &file, err
}

// UpdateFile 更新文件记录
func (r *reconciliationRepository) UpdateFile(ctx context.Context, file *model.ChannelSettlementFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// ListFiles 查询文件列表
func (r *reconciliationRepository) ListFiles(ctx context.Context, filters FileFilters, page, pageSize int) ([]*model.ChannelSettlementFile, int64, error) {
	var files []*model.ChannelSettlementFile
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ChannelSettlementFile{})

	// 应用过滤条件
	if filters.Channel != "" {
		query = query.Where("channel = ?", filters.Channel)
	}
	if filters.SettlementDate != nil {
		query = query.Where("settlement_date = ?", filters.SettlementDate.Format("2006-01-02"))
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.StartDate != nil {
		query = query.Where("settlement_date >= ?", filters.StartDate.Format("2006-01-02"))
	}
	if filters.EndDate != nil {
		query = query.Where("settlement_date <= ?", filters.EndDate.Format("2006-01-02"))
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count files failed: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.
		Order("settlement_date DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&files).Error; err != nil {
		return nil, 0, fmt.Errorf("list files failed: %w", err)
	}

	return files, total, nil
}
