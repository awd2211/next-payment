package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/settlement-service/internal/model"
)

// SettlementRepository 结算仓储接口
type SettlementRepository interface {
	Create(ctx context.Context, settlement *model.Settlement) error
	Update(ctx context.Context, settlement *model.Settlement) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Settlement, error)
	GetBySettlementNo(ctx context.Context, settlementNo string) (*model.Settlement, error)
	List(ctx context.Context, query *SettlementQuery) ([]*model.Settlement, int64, error)
	CreateItem(ctx context.Context, item *model.SettlementItem) error
	CreateItems(ctx context.Context, items []*model.SettlementItem) error
	GetItems(ctx context.Context, settlementID uuid.UUID) ([]*model.SettlementItem, error)
	CreateApproval(ctx context.Context, approval *model.SettlementApproval) error
	GetApprovals(ctx context.Context, settlementID uuid.UUID) ([]*model.SettlementApproval, error)
	GetPendingSettlements(ctx context.Context, merchantID uuid.UUID) ([]*model.Settlement, error)
}

// SettlementQuery 查询条件
type SettlementQuery struct {
	MerchantID *uuid.UUID
	Status     *model.SettlementStatus
	Cycle      *model.SettlementCycle
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

type settlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository 创建结算仓储
func NewSettlementRepository(db *gorm.DB) SettlementRepository {
	return &settlementRepository{
		db: db,
	}
}

func (r *settlementRepository) Create(ctx context.Context, settlement *model.Settlement) error {
	return r.db.WithContext(ctx).Create(settlement).Error
}

func (r *settlementRepository) Update(ctx context.Context, settlement *model.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

func (r *settlementRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&settlement).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) GetBySettlementNo(ctx context.Context, settlementNo string) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.WithContext(ctx).Where("settlement_no = ?", settlementNo).First(&settlement).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) List(ctx context.Context, query *SettlementQuery) ([]*model.Settlement, int64, error) {
	var settlements []*model.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Settlement{})

	// 应用过滤条件
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Cycle != nil {
		db = db.Where("cycle = ?", *query.Cycle)
	}
	if query.StartDate != nil {
		db = db.Where("start_date >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("end_date <= ?", *query.EndDate)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&settlements).Error
	if err != nil {
		return nil, 0, err
	}

	return settlements, total, nil
}

func (r *settlementRepository) CreateItem(ctx context.Context, item *model.SettlementItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *settlementRepository) CreateItems(ctx context.Context, items []*model.SettlementItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *settlementRepository) GetItems(ctx context.Context, settlementID uuid.UUID) ([]*model.SettlementItem, error) {
	var items []*model.SettlementItem
	err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Order("transaction_at DESC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *settlementRepository) CreateApproval(ctx context.Context, approval *model.SettlementApproval) error {
	return r.db.WithContext(ctx).Create(approval).Error
}

func (r *settlementRepository) GetApprovals(ctx context.Context, settlementID uuid.UUID) ([]*model.SettlementApproval, error) {
	var approvals []*model.SettlementApproval
	err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Order("created_at DESC").Find(&approvals).Error
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *settlementRepository) GetPendingSettlements(ctx context.Context, merchantID uuid.UUID) ([]*model.Settlement, error) {
	var settlements []*model.Settlement
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND status = ?", merchantID, model.SettlementStatusPending).
		Order("created_at ASC").
		Find(&settlements).Error
	if err != nil {
		return nil, err
	}
	return settlements, nil
}
