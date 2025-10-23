package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// KYCDocumentRepository KYC文档仓储接口
type KYCDocumentRepository interface {
	Create(ctx context.Context, doc *model.KYCDocument) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID, documentType string) ([]*model.KYCDocument, error)
	Update(ctx context.Context, doc *model.KYCDocument) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status, reviewNotes string, reviewedBy uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type kycDocumentRepository struct {
	db *gorm.DB
}

// NewKYCDocumentRepository 创建KYC文档仓储实例
func NewKYCDocumentRepository(db *gorm.DB) KYCDocumentRepository {
	return &kycDocumentRepository{db: db}
}

func (r *kycDocumentRepository) Create(ctx context.Context, doc *model.KYCDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *kycDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error) {
	var doc model.KYCDocument
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *kycDocumentRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID, documentType string) ([]*model.KYCDocument, error) {
	var docs []*model.KYCDocument
	query := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	if documentType != "" {
		query = query.Where("document_type = ?", documentType)
	}

	err := query.Order("created_at DESC").Find(&docs).Error
	return docs, err
}

func (r *kycDocumentRepository) Update(ctx context.Context, doc *model.KYCDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *kycDocumentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, reviewNotes string, reviewedBy uuid.UUID) error {
	updates := map[string]interface{}{
		"status":       status,
		"review_notes": reviewNotes,
		"reviewed_by":  reviewedBy,
		"reviewed_at":  gorm.Expr("NOW()"),
	}
	return r.db.WithContext(ctx).Model(&model.KYCDocument{}).Where("id = ?", id).Updates(updates).Error
}

func (r *kycDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.KYCDocument{}, "id = ?", id).Error
}
