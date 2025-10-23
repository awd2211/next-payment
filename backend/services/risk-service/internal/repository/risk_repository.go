package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/risk-service/internal/model"
	"gorm.io/gorm"
)

// RiskRepository 风控仓储接口
type RiskRepository interface {
	// 风控规则
	CreateRule(ctx context.Context, rule *model.RiskRule) error
	GetRuleByID(ctx context.Context, id uuid.UUID) (*model.RiskRule, error)
	ListRules(ctx context.Context, query *RuleQuery) ([]*model.RiskRule, int64, error)
	UpdateRule(ctx context.Context, rule *model.RiskRule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error

	// 风控检查
	CreateCheck(ctx context.Context, check *model.RiskCheck) error
	GetCheckByID(ctx context.Context, id uuid.UUID) (*model.RiskCheck, error)
	GetCheckByRelated(ctx context.Context, relatedID uuid.UUID, relatedType string) (*model.RiskCheck, error)
	ListChecks(ctx context.Context, query *CheckQuery) ([]*model.RiskCheck, int64, error)

	// 黑名单
	CreateBlacklist(ctx context.Context, blacklist *model.Blacklist) error
	GetBlacklistByID(ctx context.Context, id uuid.UUID) (*model.Blacklist, error)
	CheckBlacklist(ctx context.Context, entityType, entityValue string) (*model.Blacklist, error)
	ListBlacklist(ctx context.Context, query *BlacklistQuery) ([]*model.Blacklist, int64, error)
	DeleteBlacklist(ctx context.Context, id uuid.UUID) error
}

type riskRepository struct {
	db *gorm.DB
}

// NewRiskRepository 创建风控仓储实例
func NewRiskRepository(db *gorm.DB) RiskRepository {
	return &riskRepository{db: db}
}

// RuleQuery 规则查询条件
type RuleQuery struct {
	RuleType string
	Status   string
	Page     int
	PageSize int
}

// CheckQuery 检查查询条件
type CheckQuery struct {
	RelatedType string
	Decision    string
	RiskLevel   string
	MerchantID  *uuid.UUID
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	PageSize    int
}

// BlacklistQuery 黑名单查询条件
type BlacklistQuery struct {
	EntityType string
	Status     string
	Page       int
	PageSize   int
}

// CreateRule 创建规则
func (r *riskRepository) CreateRule(ctx context.Context, rule *model.RiskRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetRuleByID 根据ID获取规则
func (r *riskRepository) GetRuleByID(ctx context.Context, id uuid.UUID) (*model.RiskRule, error) {
	var rule model.RiskRule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &rule, err
}

// ListRules 规则列表
func (r *riskRepository) ListRules(ctx context.Context, query *RuleQuery) ([]*model.RiskRule, int64, error) {
	var rules []*model.RiskRule
	var total int64

	db := r.db.WithContext(ctx).Model(&model.RiskRule{})

	if query.RuleType != "" {
		db = db.Where("rule_type = ?", query.RuleType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&rules).Error
	return rules, total, err
}

// UpdateRule 更新规则
func (r *riskRepository) UpdateRule(ctx context.Context, rule *model.RiskRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeleteRule 删除规则
func (r *riskRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.RiskRule{}, "id = ?", id).Error
}

// CreateCheck 创建检查记录
func (r *riskRepository) CreateCheck(ctx context.Context, check *model.RiskCheck) error {
	return r.db.WithContext(ctx).Create(check).Error
}

// GetCheckByID 根据ID获取检查记录
func (r *riskRepository) GetCheckByID(ctx context.Context, id uuid.UUID) (*model.RiskCheck, error) {
	var check model.RiskCheck
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&check).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &check, err
}

// GetCheckByRelated 根据关联对象获取检查记录
func (r *riskRepository) GetCheckByRelated(ctx context.Context, relatedID uuid.UUID, relatedType string) (*model.RiskCheck, error) {
	var check model.RiskCheck
	err := r.db.WithContext(ctx).
		Where("related_id = ? AND related_type = ?", relatedID, relatedType).
		First(&check).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &check, err
}

// ListChecks 检查记录列表
func (r *riskRepository) ListChecks(ctx context.Context, query *CheckQuery) ([]*model.RiskCheck, int64, error) {
	var checks []*model.RiskCheck
	var total int64

	db := r.db.WithContext(ctx).Model(&model.RiskCheck{})

	if query.RelatedType != "" {
		db = db.Where("related_type = ?", query.RelatedType)
	}
	if query.Decision != "" {
		db = db.Where("decision = ?", query.Decision)
	}
	if query.RiskLevel != "" {
		db = db.Where("risk_level = ?", query.RiskLevel)
	}
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&checks).Error
	return checks, total, err
}

// CreateBlacklist 创建黑名单
func (r *riskRepository) CreateBlacklist(ctx context.Context, blacklist *model.Blacklist) error {
	return r.db.WithContext(ctx).Create(blacklist).Error
}

// GetBlacklistByID 根据ID获取黑名单
func (r *riskRepository) GetBlacklistByID(ctx context.Context, id uuid.UUID) (*model.Blacklist, error) {
	var blacklist model.Blacklist
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&blacklist).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &blacklist, err
}

// CheckBlacklist 检查黑名单
func (r *riskRepository) CheckBlacklist(ctx context.Context, entityType, entityValue string) (*model.Blacklist, error) {
	var blacklist model.Blacklist
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_value = ? AND status = ?", entityType, entityValue, "active").
		First(&blacklist).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &blacklist, err
}

// ListBlacklist 黑名单列表
func (r *riskRepository) ListBlacklist(ctx context.Context, query *BlacklistQuery) ([]*model.Blacklist, int64, error) {
	var blacklists []*model.Blacklist
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Blacklist{})

	if query.EntityType != "" {
		db = db.Where("entity_type = ?", query.EntityType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&blacklists).Error
	return blacklists, total, err
}

// DeleteBlacklist 删除黑名单
func (r *riskRepository) DeleteBlacklist(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Blacklist{}, "id = ?", id).Error
}
