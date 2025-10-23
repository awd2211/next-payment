package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"gorm.io/gorm"
)

// EmailTemplateRepository 邮件模板仓储接口
type EmailTemplateRepository interface {
	Create(ctx context.Context, template *model.EmailTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.EmailTemplate, error)
	GetByCode(ctx context.Context, code string) (*model.EmailTemplate, error)
	List(ctx context.Context, page, pageSize int, category string, isActive *bool) ([]*model.EmailTemplate, int64, error)
	Update(ctx context.Context, template *model.EmailTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 邮件日志
	CreateLog(ctx context.Context, log *model.EmailLog) error
	UpdateLog(ctx context.Context, log *model.EmailLog) error
	GetLogByID(ctx context.Context, id uuid.UUID) (*model.EmailLog, error)
	ListLogs(ctx context.Context, page, pageSize int, status, to string) ([]*model.EmailLog, int64, error)
}

type emailTemplateRepository struct {
	db *gorm.DB
}

// NewEmailTemplateRepository 创建邮件模板仓储实例
func NewEmailTemplateRepository(db *gorm.DB) EmailTemplateRepository {
	return &emailTemplateRepository{db: db}
}

// Create 创建邮件模板
func (r *emailTemplateRepository) Create(ctx context.Context, template *model.EmailTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID 根据ID获取模板
func (r *emailTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// GetByCode 根据代码获取模板
func (r *emailTemplateRepository) GetByCode(ctx context.Context, code string) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	err := r.db.WithContext(ctx).First(&template, "code = ? AND is_active = true", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// List 分页查询模板列表
func (r *emailTemplateRepository) List(ctx context.Context, page, pageSize int, category string, isActive *bool) ([]*model.EmailTemplate, int64, error) {
	var templates []*model.EmailTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&model.EmailTemplate{})

	// 分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 状态筛选
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&templates).Error

	return templates, total, err
}

// Update 更新模板
func (r *emailTemplateRepository) Update(ctx context.Context, template *model.EmailTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete 删除模板（软删除）
func (r *emailTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否为系统模板
	var template model.EmailTemplate
	if err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		return err
	}

	if template.IsSystem {
		return errors.New("系统内置模板不可删除")
	}

	return r.db.WithContext(ctx).Delete(&model.EmailTemplate{}, "id = ?", id).Error
}

// CreateLog 创建邮件发送日志
func (r *emailTemplateRepository) CreateLog(ctx context.Context, log *model.EmailLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// UpdateLog 更新邮件发送日志
func (r *emailTemplateRepository) UpdateLog(ctx context.Context, log *model.EmailLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// GetLogByID 根据ID获取日志
func (r *emailTemplateRepository) GetLogByID(ctx context.Context, id uuid.UUID) (*model.EmailLog, error) {
	var log model.EmailLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// ListLogs 分页查询邮件日志
func (r *emailTemplateRepository) ListLogs(ctx context.Context, page, pageSize int, status, to string) ([]*model.EmailLog, int64, error) {
	var logs []*model.EmailLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.EmailLog{})

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 收件人筛选
	if to != "" {
		query = query.Where("to LIKE ?", "%"+to+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}
