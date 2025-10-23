package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/kyc-service/internal/model"
)

// KYCRepository KYC仓储接口
type KYCRepository interface {
	// Document operations
	CreateDocument(ctx context.Context, doc *model.KYCDocument) error
	UpdateDocument(ctx context.Context, doc *model.KYCDocument) error
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error)
	ListDocuments(ctx context.Context, query *DocumentQuery) ([]*model.KYCDocument, int64, error)
	GetMerchantDocuments(ctx context.Context, merchantID uuid.UUID, docType *model.DocumentType) ([]*model.KYCDocument, error)

	// Business Qualification operations
	CreateQualification(ctx context.Context, qual *model.BusinessQualification) error
	UpdateQualification(ctx context.Context, qual *model.BusinessQualification) error
	GetQualificationByID(ctx context.Context, id uuid.UUID) (*model.BusinessQualification, error)
	GetQualificationByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.BusinessQualification, error)
	ListQualifications(ctx context.Context, query *QualificationQuery) ([]*model.BusinessQualification, int64, error)

	// Review operations
	CreateReview(ctx context.Context, review *model.KYCReview) error
	GetReviews(ctx context.Context, merchantID uuid.UUID, documentID *uuid.UUID) ([]*model.KYCReview, error)

	// Merchant KYC Level operations
	CreateMerchantLevel(ctx context.Context, level *model.MerchantKYCLevel) error
	UpdateMerchantLevel(ctx context.Context, level *model.MerchantKYCLevel) error
	GetMerchantLevel(ctx context.Context, merchantID uuid.UUID) (*model.MerchantKYCLevel, error)

	// Alert operations
	CreateAlert(ctx context.Context, alert *model.KYCAlert) error
	UpdateAlert(ctx context.Context, alert *model.KYCAlert) error
	ListAlerts(ctx context.Context, query *AlertQuery) ([]*model.KYCAlert, int64, error)
	GetExpiringDocuments(ctx context.Context, days int) ([]*model.KYCDocument, error)
}

// DocumentQuery 文档查询条件
type DocumentQuery struct {
	MerchantID   *uuid.UUID
	DocumentType *model.DocumentType
	Status       *model.KYCStatus
	Page         int
	PageSize     int
}

// QualificationQuery 企业资质查询条件
type QualificationQuery struct {
	Status   *model.KYCStatus
	Industry *string
	Page     int
	PageSize int
}

// AlertQuery 预警查询条件
type AlertQuery struct {
	MerchantID *uuid.UUID
	AlertType  *string
	Severity   *string
	Status     *string
	Page       int
	PageSize   int
}

type kycRepository struct {
	db *gorm.DB
}

// NewKYCRepository 创建KYC仓储
func NewKYCRepository(db *gorm.DB) KYCRepository {
	return &kycRepository{
		db: db,
	}
}

// Document operations
func (r *kycRepository) CreateDocument(ctx context.Context, doc *model.KYCDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *kycRepository) UpdateDocument(ctx context.Context, doc *model.KYCDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *kycRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error) {
	var doc model.KYCDocument
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *kycRepository) ListDocuments(ctx context.Context, query *DocumentQuery) ([]*model.KYCDocument, int64, error) {
	var documents []*model.KYCDocument
	var total int64

	db := r.db.WithContext(ctx).Model(&model.KYCDocument{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.DocumentType != nil {
		db = db.Where("document_type = ?", *query.DocumentType)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&documents).Error
	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (r *kycRepository) GetMerchantDocuments(ctx context.Context, merchantID uuid.UUID, docType *model.DocumentType) ([]*model.KYCDocument, error) {
	var documents []*model.KYCDocument
	db := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	if docType != nil {
		db = db.Where("document_type = ?", *docType)
	}

	err := db.Order("created_at DESC").Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

// Business Qualification operations
func (r *kycRepository) CreateQualification(ctx context.Context, qual *model.BusinessQualification) error {
	return r.db.WithContext(ctx).Create(qual).Error
}

func (r *kycRepository) UpdateQualification(ctx context.Context, qual *model.BusinessQualification) error {
	return r.db.WithContext(ctx).Save(qual).Error
}

func (r *kycRepository) GetQualificationByID(ctx context.Context, id uuid.UUID) (*model.BusinessQualification, error) {
	var qual model.BusinessQualification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&qual).Error
	if err != nil {
		return nil, err
	}
	return &qual, nil
}

func (r *kycRepository) GetQualificationByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.BusinessQualification, error) {
	var qual model.BusinessQualification
	err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&qual).Error
	if err != nil {
		return nil, err
	}
	return &qual, nil
}

func (r *kycRepository) ListQualifications(ctx context.Context, query *QualificationQuery) ([]*model.BusinessQualification, int64, error) {
	var qualifications []*model.BusinessQualification
	var total int64

	db := r.db.WithContext(ctx).Model(&model.BusinessQualification{})

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Industry != nil {
		db = db.Where("industry = ?", *query.Industry)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&qualifications).Error
	if err != nil {
		return nil, 0, err
	}

	return qualifications, total, nil
}

// Review operations
func (r *kycRepository) CreateReview(ctx context.Context, review *model.KYCReview) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *kycRepository) GetReviews(ctx context.Context, merchantID uuid.UUID, documentID *uuid.UUID) ([]*model.KYCReview, error) {
	var reviews []*model.KYCReview
	db := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	if documentID != nil {
		db = db.Where("document_id = ?", *documentID)
	}

	err := db.Order("created_at DESC").Find(&reviews).Error
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

// Merchant KYC Level operations
func (r *kycRepository) CreateMerchantLevel(ctx context.Context, level *model.MerchantKYCLevel) error {
	return r.db.WithContext(ctx).Create(level).Error
}

func (r *kycRepository) UpdateMerchantLevel(ctx context.Context, level *model.MerchantKYCLevel) error {
	return r.db.WithContext(ctx).Save(level).Error
}

func (r *kycRepository) GetMerchantLevel(ctx context.Context, merchantID uuid.UUID) (*model.MerchantKYCLevel, error) {
	var level model.MerchantKYCLevel
	err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&level).Error
	if err != nil {
		// If not found, return a default level
		if err == gorm.ErrRecordNotFound {
			return &model.MerchantKYCLevel{
				MerchantID:   merchantID,
				CurrentLevel: model.KYCLevelBasic,
			}, nil
		}
		return nil, err
	}
	return &level, nil
}

// Alert operations
func (r *kycRepository) CreateAlert(ctx context.Context, alert *model.KYCAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *kycRepository) UpdateAlert(ctx context.Context, alert *model.KYCAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *kycRepository) ListAlerts(ctx context.Context, query *AlertQuery) ([]*model.KYCAlert, int64, error) {
	var alerts []*model.KYCAlert
	var total int64

	db := r.db.WithContext(ctx).Model(&model.KYCAlert{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.AlertType != nil {
		db = db.Where("alert_type = ?", *query.AlertType)
	}
	if query.Severity != nil {
		db = db.Where("severity = ?", *query.Severity)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&alerts).Error
	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

func (r *kycRepository) GetExpiringDocuments(ctx context.Context, days int) ([]*model.KYCDocument, error) {
	var documents []*model.KYCDocument
	expiryThreshold := time.Now().AddDate(0, 0, days)

	err := r.db.WithContext(ctx).
		Where("expiry_date IS NOT NULL AND expiry_date <= ? AND expiry_date > ? AND status = ?",
			expiryThreshold, time.Now(), model.KYCStatusApproved).
		Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}
