package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"gorm.io/gorm"
)

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error)
	List(ctx context.Context, filter *AuditLogFilter) ([]*model.AuditLog, int64, error)
}

// AuditLogFilter 审计日志查询过滤器
type AuditLogFilter struct {
	AdminID      *uuid.UUID
	Action       string
	Resource     string
	Method       string
	StartTime    *time.Time
	EndTime      *time.Time
	IP           string
	ResponseCode *int
	Page         int
	PageSize     int
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository 创建审计日志仓储实例
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create 创建审计日志
func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByID 根据ID获取审计日志
func (r *auditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	var log model.AuditLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List 分页查询审计日志列表
func (r *auditLogRepository) List(ctx context.Context, filter *AuditLogFilter) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	// 应用过滤条件
	if filter.AdminID != nil {
		query = query.Where("admin_id = ?", *filter.AdminID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}
	if filter.Method != "" {
		query = query.Where("method = ?", filter.Method)
	}
	if filter.IP != "" {
		query = query.Where("ip = ?", filter.IP)
	}
	if filter.ResponseCode != nil {
		query = query.Where("response_code = ?", *filter.ResponseCode)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
