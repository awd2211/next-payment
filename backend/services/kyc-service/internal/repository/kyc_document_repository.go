package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/kyc-service/internal/model"
)

// KYCDocumentRepository KYC文档仓储接口
type KYCDocumentRepository interface {
	Create(ctx context.Context, doc *model.KYCDocument) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.KYCDocument, error)
	GetByMerchantIDAndType(ctx context.Context, merchantID uuid.UUID, docType string) ([]*model.KYCDocument, error)
	Update(ctx context.Context, doc *model.KYCDocument) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status string, limit, offset int) ([]*model.KYCDocument, int64, error)
}

type kycDocumentRepository struct {
	db *gorm.DB
}

// NewKYCDocumentRepository 创建KYC文档仓储
func NewKYCDocumentRepository(db *gorm.DB) KYCDocumentRepository {
	return &kycDocumentRepository{db: db}
}

// Create 创建KYC文档
func (r *kycDocumentRepository) Create(ctx context.Context, doc *model.KYCDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetByID 根据ID获取KYC文档
func (r *kycDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error) {
	var doc model.KYCDocument
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetByMerchantID 获取商户的所有KYC文档
func (r *kycDocumentRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.KYCDocument, error) {
	var docs []*model.KYCDocument
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&docs).Error
	return docs, err
}

// GetByMerchantIDAndType 获取商户指定类型的KYC文档
func (r *kycDocumentRepository) GetByMerchantIDAndType(ctx context.Context, merchantID uuid.UUID, docType string) ([]*model.KYCDocument, error) {
	var docs []*model.KYCDocument
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND document_type = ?", merchantID, docType).
		Order("created_at DESC").
		Find(&docs).Error
	return docs, err
}

// Update 更新KYC文档
func (r *kycDocumentRepository) Update(ctx context.Context, doc *model.KYCDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// Delete 删除KYC文档（软删除）
func (r *kycDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.KYCDocument{}, "id = ?", id).Error
}

// List 分页列出KYC文档
func (r *kycDocumentRepository) List(ctx context.Context, status string, limit, offset int) ([]*model.KYCDocument, int64, error) {
	var docs []*model.KYCDocument
	var total int64

	query := r.db.WithContext(ctx).Model(&model.KYCDocument{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&docs).Error

	return docs, total, err
}
