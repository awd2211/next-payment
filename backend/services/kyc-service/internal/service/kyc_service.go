package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/kyc-service/internal/client"
	"payment-platform/kyc-service/internal/model"
	"payment-platform/kyc-service/internal/repository"
)

// KYCService KYC服务接口
type KYCService interface {
	// Document operations
	SubmitDocument(ctx context.Context, input *SubmitDocumentInput) (*model.KYCDocument, error)
	GetDocument(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error)
	ListDocuments(ctx context.Context, query *ListDocumentQuery) (*ListDocumentResponse, error)
	ApproveDocument(ctx context.Context, documentID, reviewerID uuid.UUID, reviewerName, comments string) error
	RejectDocument(ctx context.Context, documentID, reviewerID uuid.UUID, reviewerName, reason string) error

	// Business Qualification operations
	SubmitQualification(ctx context.Context, input *SubmitQualificationInput) (*model.BusinessQualification, error)
	GetQualification(ctx context.Context, merchantID uuid.UUID) (*model.BusinessQualification, error)
	ListQualifications(ctx context.Context, query *ListQualificationQuery) (*ListQualificationResponse, error)
	ApproveQualification(ctx context.Context, qualID, reviewerID uuid.UUID, reviewerName, comments string) error
	RejectQualification(ctx context.Context, qualID, reviewerID uuid.UUID, reviewerName, reason string) error

	// Merchant KYC Level operations
	GetMerchantLevel(ctx context.Context, merchantID uuid.UUID) (*model.MerchantKYCLevel, error)
	UpgradeMerchantLevel(ctx context.Context, merchantID uuid.UUID, newLevel model.KYCLevel) error
	CheckMerchantEligibility(ctx context.Context, merchantID uuid.UUID) (*MerchantEligibility, error)

	// Alert operations
	ListAlerts(ctx context.Context, query *ListAlertQuery) (*ListAlertResponse, error)
	ResolveAlert(ctx context.Context, alertID, resolverID uuid.UUID) error
	CheckExpiringDocuments(ctx context.Context) error

	// Statistics
	GetKYCStatistics(ctx context.Context, merchantID *uuid.UUID) (*KYCStatistics, error)
}

type kycService struct {
	db                 *gorm.DB
	kycRepo            repository.KYCRepository
	notificationClient *client.NotificationClient
	ocrClient          *client.OCRClient
}

// NewKYCService 创建KYC服务
func NewKYCService(db *gorm.DB, kycRepo repository.KYCRepository, notificationClient *client.NotificationClient, ocrClient *client.OCRClient) KYCService {
	return &kycService{
		db:                 db,
		kycRepo:            kycRepo,
		notificationClient: notificationClient,
		ocrClient:          ocrClient,
	}
}

// Document operations

// SubmitDocumentInput 提交文档输入
type SubmitDocumentInput struct {
	MerchantID     uuid.UUID
	DocumentType   model.DocumentType
	DocumentNumber string
	DocumentURL    string
	FrontImageURL  string
	BackImageURL   string
	IssueDate      *time.Time
	ExpiryDate     *time.Time
	IssuingCountry string
}

// SubmitDocument 提交KYC文档
func (s *kycService) SubmitDocument(ctx context.Context, input *SubmitDocumentInput) (*model.KYCDocument, error) {
	document := &model.KYCDocument{
		MerchantID:     input.MerchantID,
		DocumentType:   input.DocumentType,
		DocumentNumber: input.DocumentNumber,
		DocumentURL:    input.DocumentURL,
		FrontImageURL:  input.FrontImageURL,
		BackImageURL:   input.BackImageURL,
		IssueDate:      input.IssueDate,
		ExpiryDate:     input.ExpiryDate,
		IssuingCountry: input.IssuingCountry,
		Status:         model.KYCStatusPending,
	}

	// 调用OCR服务识别文档信息（异步处理，不阻塞主流程）
	if s.ocrClient != nil && input.FrontImageURL != "" {
		go func(doc *model.KYCDocument, imageURL string) {
			ocrCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// 根据文档类型选择OCR方法
			var ocrData *client.OCRData
			var err error

			switch doc.DocumentType {
			case model.DocumentTypePassport:
				ocrData, err = s.ocrClient.ExtractPassport(ocrCtx, imageURL)
			case model.DocumentTypeIDCard:
				ocrData, err = s.ocrClient.ExtractIDCard(ocrCtx, imageURL)
			case model.DocumentTypeBusinessLicense:
				ocrData, err = s.ocrClient.ExtractBusinessLicense(ocrCtx, imageURL)
			default:
				ocrData, err = s.ocrClient.ExtractDocument(ocrCtx, &client.OCRExtractRequest{
					ImageURL:     imageURL,
					DocumentType: string(doc.DocumentType),
					Language:     "auto",
				})
			}

			if err != nil {
				logger.Warn("OCR识别失败（非致命）",
					zap.Error(err),
					zap.String("document_id", doc.ID.String()),
					zap.String("document_type", string(doc.DocumentType)))
				return
			}

			// 验证OCR质量
			isValid, validationMsg := client.ValidateOCRQuality(ocrData)
			if !isValid {
				logger.Warn("OCR质量验证失败",
					zap.String("document_id", doc.ID.String()),
					zap.String("reason", validationMsg),
					zap.Float64("confidence", ocrData.Confidence))
			}

			// 更新文档OCR数据（序列化为JSON字符串）
			ocrDataBytes, jsonErr := json.Marshal(ocrData)
			if jsonErr != nil {
				logger.Error("序列化OCR数据失败",
					zap.Error(jsonErr),
					zap.String("document_id", doc.ID.String()))
				return
			}
			doc.OCRData = string(ocrDataBytes)

			// 如果OCR识别出的文档号与用户输入不一致，记录警告
			if ocrData.DocumentNumber != "" && doc.DocumentNumber != "" && ocrData.DocumentNumber != doc.DocumentNumber {
				logger.Warn("OCR识别的文档号与用户输入不一致",
					zap.String("document_id", doc.ID.String()),
					zap.String("user_input", doc.DocumentNumber),
					zap.String("ocr_result", ocrData.DocumentNumber))
			}

			// 异步更新数据库
			if err := s.kycRepo.UpdateDocument(context.Background(), doc); err != nil {
				logger.Error("更新文档OCR数据失败",
					zap.Error(err),
					zap.String("document_id", doc.ID.String()))
			}
		}(document, input.FrontImageURL)
	}

	if err := s.kycRepo.CreateDocument(ctx, document); err != nil {
		return nil, fmt.Errorf("创建文档失败: %w", err)
	}

	// 发送审核通知（异步）
	if s.notificationClient != nil {
		go func(doc *model.KYCDocument) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := s.notificationClient.SendKYCNotification(notifyCtx, &client.SendNotificationRequest{
				MerchantID: doc.MerchantID,
				Type:       "kyc_submitted",
				Title:      "KYC 文档已提交",
				Content:    fmt.Sprintf("您的 %s 已提交，正在审核中", doc.DocumentType),
				Priority:   "medium",
				Data: map[string]interface{}{
					"document_id":   doc.ID.String(),
					"document_type": doc.DocumentType,
					"status":        doc.Status,
				},
			})
			if err != nil {
				logger.Warn("发送KYC提交通知失败（非致命）",
					zap.Error(err),
					zap.String("merchant_id", doc.MerchantID.String()),
					zap.String("document_id", doc.ID.String()))
			}
		}(document)
	}

	return document, nil
}

// GetDocument 获取文档详情
func (s *kycService) GetDocument(ctx context.Context, id uuid.UUID) (*model.KYCDocument, error) {
	return s.kycRepo.GetDocumentByID(ctx, id)
}

// ListDocumentQuery 文档列表查询
type ListDocumentQuery struct {
	MerchantID   *uuid.UUID
	DocumentType *model.DocumentType
	Status       *model.KYCStatus
	Page         int
	PageSize     int
}

// ListDocumentResponse 文档列表响应
type ListDocumentResponse struct {
	Documents []*model.KYCDocument `json:"documents"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"page_size"`
}

// ListDocuments 文档列表
func (s *kycService) ListDocuments(ctx context.Context, query *ListDocumentQuery) (*ListDocumentResponse, error) {
	repoQuery := &repository.DocumentQuery{
		MerchantID:   query.MerchantID,
		DocumentType: query.DocumentType,
		Status:       query.Status,
		Page:         query.Page,
		PageSize:     query.PageSize,
	}

	documents, total, err := s.kycRepo.ListDocuments(ctx, repoQuery)
	if err != nil {
		return nil, fmt.Errorf("查询文档列表失败: %w", err)
	}

	return &ListDocumentResponse{
		Documents: documents,
		Total:     total,
		Page:      query.Page,
		PageSize:  query.PageSize,
	}, nil
}

// ApproveDocument 审批通过文档
func (s *kycService) ApproveDocument(ctx context.Context, documentID, reviewerID uuid.UUID, reviewerName, comments string) error {
	document, err := s.kycRepo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("文档不存在: %w", err)
	}

	if document.Status != model.KYCStatusPending {
		return fmt.Errorf("文档状态不是待审核")
	}

	now := time.Now()
	document.Status = model.KYCStatusApproved
	document.ReviewerID = &reviewerID
	document.ReviewerName = reviewerName
	document.ReviewComments = comments
	document.ReviewedAt = &now

	review := &model.KYCReview{
		MerchantID:   document.MerchantID,
		DocumentID:   &documentID,
		ReviewerID:   reviewerID,
		ReviewerName: reviewerName,
		Action:       "approve",
		Status:       model.KYCStatusApproved,
		Comments:     comments,
		ReviewedAt:   now,
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.kycRepo.UpdateDocument(ctx, document); err != nil {
			return fmt.Errorf("更新文档失败: %w", err)
		}
		if err := s.kycRepo.CreateReview(ctx, review); err != nil {
			return fmt.Errorf("创建审核记录失败: %w", err)
		}

		// 检查并更新KYC级别
		if err := s.updateMerchantLevelAfterApproval(ctx, document.MerchantID, document.DocumentType); err != nil {
			return fmt.Errorf("更新商户KYC级别失败: %w", err)
		}

		// 发送KYC审核通过通知（异步）
		if s.notificationClient != nil {
			go func(doc *model.KYCDocument) {
				notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				err := s.notificationClient.SendKYCNotification(notifyCtx, &client.SendNotificationRequest{
					MerchantID: doc.MerchantID,
					Type:       "kyc_approved",
					Title:      "KYC 认证通过",
					Content:    fmt.Sprintf("您的 %s 已审核通过", doc.DocumentType),
					Priority:   "high",
					Data: map[string]interface{}{
						"document_id":   doc.ID.String(),
						"document_type": doc.DocumentType,
						"reviewer_name": reviewerName,
						"comments":      comments,
					},
				})
				if err != nil {
					logger.Warn("发送KYC通过通知失败（非致命）",
						zap.Error(err),
						zap.String("merchant_id", doc.MerchantID.String()),
						zap.String("document_id", doc.ID.String()))
				}
			}(document)
		}

		return nil
	})
}

// RejectDocument 拒绝文档
func (s *kycService) RejectDocument(ctx context.Context, documentID, reviewerID uuid.UUID, reviewerName, reason string) error {
	document, err := s.kycRepo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("文档不存在: %w", err)
	}

	if document.Status != model.KYCStatusPending {
		return fmt.Errorf("文档状态不是待审核")
	}

	now := time.Now()
	document.Status = model.KYCStatusRejected
	document.ReviewerID = &reviewerID
	document.ReviewerName = reviewerName
	document.RejectionReason = reason
	document.ReviewedAt = &now

	review := &model.KYCReview{
		MerchantID:      document.MerchantID,
		DocumentID:      &documentID,
		ReviewerID:      reviewerID,
		ReviewerName:    reviewerName,
		Action:          "reject",
		Status:          model.KYCStatusRejected,
		RejectionReason: reason,
		ReviewedAt:      now,
	}

	// 使用事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.kycRepo.UpdateDocument(ctx, document); err != nil {
			return fmt.Errorf("更新文档失败: %w", err)
		}
		if err := s.kycRepo.CreateReview(ctx, review); err != nil {
			return fmt.Errorf("创建审核记录失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 发送KYC审核拒绝通知（异步）
	if s.notificationClient != nil {
		go func(doc *model.KYCDocument) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := s.notificationClient.SendKYCNotification(notifyCtx, &client.SendNotificationRequest{
				MerchantID: doc.MerchantID,
				Type:       "kyc_rejected",
				Title:      "KYC 认证未通过",
				Content:    fmt.Sprintf("您的 %s 审核未通过：%s", doc.DocumentType, reason),
				Priority:   "high",
				Data: map[string]interface{}{
					"document_id":      doc.ID.String(),
					"document_type":    doc.DocumentType,
					"reviewer_name":    reviewerName,
					"rejection_reason": reason,
				},
			})
			if err != nil {
				logger.Warn("发送KYC拒绝通知失败（非致命）",
					zap.Error(err),
					zap.String("merchant_id", doc.MerchantID.String()),
					zap.String("document_id", doc.ID.String()))
			}
		}(document)
	}

	return nil
}

// Business Qualification operations

// SubmitQualificationInput 提交企业资质输入
type SubmitQualificationInput struct {
	MerchantID                    uuid.UUID
	CompanyName                   string
	BusinessLicenseNo             string
	BusinessLicenseURL            string
	LegalPersonName               string
	LegalPersonIDCard             string
	LegalPersonIDCardFrontURL     string
	LegalPersonIDCardBackURL      string
	RegisteredAddress             string
	RegisteredCapital             int64
	EstablishedDate               *time.Time
	BusinessScope                 string
	Industry                      string
	TaxRegistrationNo             string
	TaxRegistrationURL            string
	OrganizationCode              string
}

// SubmitQualification 提交企业资质
func (s *kycService) SubmitQualification(ctx context.Context, input *SubmitQualificationInput) (*model.BusinessQualification, error) {
	// 检查是否已存在
	existing, _ := s.kycRepo.GetQualificationByMerchantID(ctx, input.MerchantID)
	if existing != nil {
		return nil, fmt.Errorf("该商户已提交企业资质")
	}

	qualification := &model.BusinessQualification{
		MerchantID:                    input.MerchantID,
		CompanyName:                   input.CompanyName,
		BusinessLicenseNo:             input.BusinessLicenseNo,
		BusinessLicenseURL:            input.BusinessLicenseURL,
		LegalPersonName:               input.LegalPersonName,
		LegalPersonIDCard:             input.LegalPersonIDCard,
		LegalPersonIDCardFrontURL:     input.LegalPersonIDCardFrontURL,
		LegalPersonIDCardBackURL:      input.LegalPersonIDCardBackURL,
		RegisteredAddress:             input.RegisteredAddress,
		RegisteredCapital:             input.RegisteredCapital,
		EstablishedDate:               input.EstablishedDate,
		BusinessScope:                 input.BusinessScope,
		Industry:                      input.Industry,
		TaxRegistrationNo:             input.TaxRegistrationNo,
		TaxRegistrationURL:            input.TaxRegistrationURL,
		OrganizationCode:              input.OrganizationCode,
		Status:                        model.KYCStatusPending,
	}

	if err := s.kycRepo.CreateQualification(ctx, qualification); err != nil {
		return nil, fmt.Errorf("创建企业资质失败: %w", err)
	}

	return qualification, nil
}

// GetQualification 获取企业资质
func (s *kycService) GetQualification(ctx context.Context, merchantID uuid.UUID) (*model.BusinessQualification, error) {
	return s.kycRepo.GetQualificationByMerchantID(ctx, merchantID)
}

// ListQualificationQuery 企业资质列表查询
type ListQualificationQuery struct {
	Status   *model.KYCStatus
	Industry *string
	Page     int
	PageSize int
}

// ListQualificationResponse 企业资质列表响应
type ListQualificationResponse struct {
	Qualifications []*model.BusinessQualification `json:"qualifications"`
	Total          int64                          `json:"total"`
	Page           int                            `json:"page"`
	PageSize       int                            `json:"page_size"`
}

// ListQualifications 企业资质列表
func (s *kycService) ListQualifications(ctx context.Context, query *ListQualificationQuery) (*ListQualificationResponse, error) {
	repoQuery := &repository.QualificationQuery{
		Status:   query.Status,
		Industry: query.Industry,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	qualifications, total, err := s.kycRepo.ListQualifications(ctx, repoQuery)
	if err != nil {
		return nil, fmt.Errorf("查询企业资质列表失败: %w", err)
	}

	return &ListQualificationResponse{
		Qualifications: qualifications,
		Total:          total,
		Page:           query.Page,
		PageSize:       query.PageSize,
	}, nil
}

// ApproveQualification 审批通过企业资质
func (s *kycService) ApproveQualification(ctx context.Context, qualID, reviewerID uuid.UUID, reviewerName, comments string) error {
	qualification, err := s.kycRepo.GetQualificationByID(ctx, qualID)
	if err != nil {
		return fmt.Errorf("企业资质不存在: %w", err)
	}

	if qualification.Status != model.KYCStatusPending {
		return fmt.Errorf("企业资质状态不是待审核")
	}

	now := time.Now()
	qualification.Status = model.KYCStatusApproved
	qualification.ReviewerID = &reviewerID
	qualification.ReviewerName = reviewerName
	qualification.ReviewComments = comments
	qualification.ReviewedAt = &now

	review := &model.KYCReview{
		MerchantID:      qualification.MerchantID,
		QualificationID: &qualID,
		ReviewerID:      reviewerID,
		ReviewerName:    reviewerName,
		Action:          "approve",
		Status:          model.KYCStatusApproved,
		Comments:        comments,
		ReviewedAt:      now,
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.kycRepo.UpdateQualification(ctx, qualification); err != nil {
			return fmt.Errorf("更新企业资质失败: %w", err)
		}
		if err := s.kycRepo.CreateReview(ctx, review); err != nil {
			return fmt.Errorf("创建审核记录失败: %w", err)
		}

		// 升级到企业级KYC
		if err := s.UpgradeMerchantLevel(ctx, qualification.MerchantID, model.KYCLevelEnterprise); err != nil {
			return fmt.Errorf("升级KYC级别失败: %w", err)
		}

		return nil
	})
}

// RejectQualification 拒绝企业资质
func (s *kycService) RejectQualification(ctx context.Context, qualID, reviewerID uuid.UUID, reviewerName, reason string) error {
	qualification, err := s.kycRepo.GetQualificationByID(ctx, qualID)
	if err != nil {
		return fmt.Errorf("企业资质不存在: %w", err)
	}

	if qualification.Status != model.KYCStatusPending {
		return fmt.Errorf("企业资质状态不是待审核")
	}

	now := time.Now()
	qualification.Status = model.KYCStatusRejected
	qualification.ReviewerID = &reviewerID
	qualification.ReviewerName = reviewerName
	qualification.RejectionReason = reason
	qualification.ReviewedAt = &now

	review := &model.KYCReview{
		MerchantID:      qualification.MerchantID,
		QualificationID: &qualID,
		ReviewerID:      reviewerID,
		ReviewerName:    reviewerName,
		Action:          "reject",
		Status:          model.KYCStatusRejected,
		RejectionReason: reason,
		ReviewedAt:      now,
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.kycRepo.UpdateQualification(ctx, qualification); err != nil {
			return fmt.Errorf("更新企业资质失败: %w", err)
		}
		if err := s.kycRepo.CreateReview(ctx, review); err != nil {
			return fmt.Errorf("创建审核记录失败: %w", err)
		}
		return nil
	})
}

// Merchant KYC Level operations

// GetMerchantLevel 获取商户KYC级别
func (s *kycService) GetMerchantLevel(ctx context.Context, merchantID uuid.UUID) (*model.MerchantKYCLevel, error) {
	return s.kycRepo.GetMerchantLevel(ctx, merchantID)
}

// UpgradeMerchantLevel 升级商户KYC级别
func (s *kycService) UpgradeMerchantLevel(ctx context.Context, merchantID uuid.UUID, newLevel model.KYCLevel) error {
	level, err := s.kycRepo.GetMerchantLevel(ctx, merchantID)
	if err != nil {
		return err
	}

	// If level doesn't exist, create it
	if level.ID == uuid.Nil {
		level.ID = uuid.New()
		level.MerchantID = merchantID
	}

	now := time.Now()
	level.CurrentLevel = newLevel
	level.ApprovedLevel = newLevel

	// Update level flags and timestamps
	switch newLevel {
	case model.KYCLevelBasic:
		level.HasBasic = true
		level.BasicApprovedAt = &now
		level.TransactionLimit = 1000000    // 10,000元
		level.DailyLimit = 5000000          // 50,000元
		level.MonthlyLimit = 100000000      // 1,000,000元
	case model.KYCLevelIntermediate:
		level.HasIntermediate = true
		level.IntermediateApprovedAt = &now
		level.TransactionLimit = 10000000   // 100,000元
		level.DailyLimit = 50000000         // 500,000元
		level.MonthlyLimit = 1000000000     // 10,000,000元
	case model.KYCLevelAdvanced:
		level.HasAdvanced = true
		level.AdvancedApprovedAt = &now
		level.TransactionLimit = 100000000  // 1,000,000元
		level.DailyLimit = 500000000        // 5,000,000元
		level.MonthlyLimit = 10000000000    // 100,000,000元
	case model.KYCLevelEnterprise:
		level.HasEnterprise = true
		level.EnterpriseApprovedAt = &now
		level.TransactionLimit = 1000000000 // 10,000,000元
		level.DailyLimit = 5000000000       // 50,000,000元
		level.MonthlyLimit = 100000000000   // 1,000,000,000元
	}

	// Set next review date (annual)
	nextReview := now.AddDate(1, 0, 0)
	level.NextReviewDate = &nextReview

	// Update or create
	if level.CreatedAt.IsZero() {
		return s.kycRepo.CreateMerchantLevel(ctx, level)
	}
	return s.kycRepo.UpdateMerchantLevel(ctx, level)
}

// updateMerchantLevelAfterApproval 文档审批后更新KYC级别
func (s *kycService) updateMerchantLevelAfterApproval(ctx context.Context, merchantID uuid.UUID, docType model.DocumentType) error {
	// Get all approved documents
	approvedStatus := model.KYCStatusApproved
	documents, _, err := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: &merchantID,
		Status:     &approvedStatus,
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		return err
	}

	// Determine level based on documents
	hasIDCard := false
	hasBusinessLicense := false

	for _, doc := range documents {
		if doc.DocumentType == model.DocumentTypeIDCard || doc.DocumentType == model.DocumentTypePassport {
			hasIDCard = true
		}
		if doc.DocumentType == model.DocumentTypeBusinessLicense {
			hasBusinessLicense = true
		}
	}

	// Upgrade level
	currentLevel, _ := s.kycRepo.GetMerchantLevel(ctx, merchantID)
	if currentLevel.CurrentLevel == model.KYCLevelBasic && hasIDCard {
		return s.UpgradeMerchantLevel(ctx, merchantID, model.KYCLevelIntermediate)
	}
	if currentLevel.CurrentLevel == model.KYCLevelIntermediate && hasBusinessLicense {
		return s.UpgradeMerchantLevel(ctx, merchantID, model.KYCLevelAdvanced)
	}

	return nil
}

// MerchantEligibility 商户资格
type MerchantEligibility struct {
	CurrentLevel      model.KYCLevel `json:"current_level"`
	CanUpgradeToLevel model.KYCLevel `json:"can_upgrade_to_level"`
	MissingDocuments  []string       `json:"missing_documents"`
	TransactionLimit  int64          `json:"transaction_limit"`
}

// CheckMerchantEligibility 检查商户资格
func (s *kycService) CheckMerchantEligibility(ctx context.Context, merchantID uuid.UUID) (*MerchantEligibility, error) {
	level, err := s.kycRepo.GetMerchantLevel(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	approvedStatus := model.KYCStatusApproved
	documents, _, err := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: &merchantID,
		Status:     &approvedStatus,
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		return nil, err
	}

	eligibility := &MerchantEligibility{
		CurrentLevel:     level.CurrentLevel,
		TransactionLimit: level.TransactionLimit,
		MissingDocuments: []string{},
	}

	// Check what's missing for next level
	hasIDCard := false
	hasBusinessLicense := false

	for _, doc := range documents {
		if doc.DocumentType == model.DocumentTypeIDCard || doc.DocumentType == model.DocumentTypePassport {
			hasIDCard = true
		}
		if doc.DocumentType == model.DocumentTypeBusinessLicense {
			hasBusinessLicense = true
		}
	}

	switch level.CurrentLevel {
	case model.KYCLevelBasic:
		if !hasIDCard {
			eligibility.MissingDocuments = append(eligibility.MissingDocuments, "身份证或护照")
		} else {
			eligibility.CanUpgradeToLevel = model.KYCLevelIntermediate
		}
	case model.KYCLevelIntermediate:
		if !hasBusinessLicense {
			eligibility.MissingDocuments = append(eligibility.MissingDocuments, "营业执照")
		} else {
			eligibility.CanUpgradeToLevel = model.KYCLevelAdvanced
		}
	case model.KYCLevelAdvanced:
		eligibility.CanUpgradeToLevel = model.KYCLevelEnterprise
	}

	return eligibility, nil
}

// Alert operations

// ListAlertQuery 预警列表查询
type ListAlertQuery struct {
	MerchantID *uuid.UUID
	AlertType  *string
	Severity   *string
	Status     *string
	Page       int
	PageSize   int
}

// ListAlertResponse 预警列表响应
type ListAlertResponse struct {
	Alerts   []*model.KYCAlert `json:"alerts"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// ListAlerts 预警列表
func (s *kycService) ListAlerts(ctx context.Context, query *ListAlertQuery) (*ListAlertResponse, error) {
	repoQuery := &repository.AlertQuery{
		MerchantID: query.MerchantID,
		AlertType:  query.AlertType,
		Severity:   query.Severity,
		Status:     query.Status,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}

	alerts, total, err := s.kycRepo.ListAlerts(ctx, repoQuery)
	if err != nil {
		return nil, fmt.Errorf("查询预警列表失败: %w", err)
	}

	return &ListAlertResponse{
		Alerts:   alerts,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// ResolveAlert 处理预警
func (s *kycService) ResolveAlert(ctx context.Context, alertID, resolverID uuid.UUID) error {
	alert, err := s.kycRepo.GetDocumentByID(ctx, alertID)
	if err != nil {
		return fmt.Errorf("预警不存在: %w", err)
	}

	if alert.Status != "open" {
		return fmt.Errorf("预警已处理")
	}

	return nil
}

// CheckExpiringDocuments 检查即将过期的文档
func (s *kycService) CheckExpiringDocuments(ctx context.Context) error {
	// Check documents expiring in 30 days
	documents, err := s.kycRepo.GetExpiringDocuments(ctx, 30)
	if err != nil {
		return err
	}

	for _, doc := range documents {
		alert := &model.KYCAlert{
			MerchantID:  doc.MerchantID,
			AlertType:   "expiry",
			Severity:    "medium",
			Title:       "KYC文档即将过期",
			Description: fmt.Sprintf("文档 %s 将在 %s 过期", doc.DocumentType, doc.ExpiryDate.Format("2006-01-02")),
			Status:      "open",
		}
		s.kycRepo.CreateAlert(ctx, alert)
	}

	return nil
}

// Statistics

// KYCStatistics KYC统计
type KYCStatistics struct {
	TotalDocuments       int   `json:"total_documents"`
	PendingDocuments     int   `json:"pending_documents"`
	ApprovedDocuments    int   `json:"approved_documents"`
	RejectedDocuments    int   `json:"rejected_documents"`
	ExpiredDocuments     int   `json:"expired_documents"`
	TotalQualifications  int   `json:"total_qualifications"`
	PendingQualifications int   `json:"pending_qualifications"`
	ApprovedQualifications int   `json:"approved_qualifications"`
	TotalAlerts          int   `json:"total_alerts"`
	OpenAlerts           int   `json:"open_alerts"`
}

// GetKYCStatistics 获取KYC统计
func (s *kycService) GetKYCStatistics(ctx context.Context, merchantID *uuid.UUID) (*KYCStatistics, error) {
	stats := &KYCStatistics{}

	// Get document statistics
	pendingStatus := model.KYCStatusPending
	approvedStatus := model.KYCStatusApproved
	rejectedStatus := model.KYCStatusRejected
	expiredStatus := model.KYCStatusExpired

	_, totalDocs, _ := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: merchantID,
		Page:       1,
		PageSize:   1,
	})
	stats.TotalDocuments = int(totalDocs)

	_, pendingCount, _ := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: merchantID,
		Status:     &pendingStatus,
		Page:       1,
		PageSize:   1,
	})
	stats.PendingDocuments = int(pendingCount)

	_, approvedCount, _ := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: merchantID,
		Status:     &approvedStatus,
		Page:       1,
		PageSize:   1,
	})
	stats.ApprovedDocuments = int(approvedCount)

	_, rejectedCount, _ := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: merchantID,
		Status:     &rejectedStatus,
		Page:       1,
		PageSize:   1,
	})
	stats.RejectedDocuments = int(rejectedCount)

	_, expiredCount, _ := s.kycRepo.ListDocuments(ctx, &repository.DocumentQuery{
		MerchantID: merchantID,
		Status:     &expiredStatus,
		Page:       1,
		PageSize:   1,
	})
	stats.ExpiredDocuments = int(expiredCount)

	// Get qualification statistics
	_, totalQual, _ := s.kycRepo.ListQualifications(ctx, &repository.QualificationQuery{
		Page:     1,
		PageSize: 1,
	})
	stats.TotalQualifications = int(totalQual)

	_, pendingQual, _ := s.kycRepo.ListQualifications(ctx, &repository.QualificationQuery{
		Status:   &pendingStatus,
		Page:     1,
		PageSize: 1,
	})
	stats.PendingQualifications = int(pendingQual)

	_, approvedQual, _ := s.kycRepo.ListQualifications(ctx, &repository.QualificationQuery{
		Status:   &approvedStatus,
		Page:     1,
		PageSize: 1,
	})
	stats.ApprovedQualifications = int(approvedQual)

	// Get alert statistics
	openStatus := "open"
	_, totalAlerts, _ := s.kycRepo.ListAlerts(ctx, &repository.AlertQuery{
		MerchantID: merchantID,
		Page:       1,
		PageSize:   1,
	})
	stats.TotalAlerts = int(totalAlerts)

	_, openAlerts, _ := s.kycRepo.ListAlerts(ctx, &repository.AlertQuery{
		MerchantID: merchantID,
		Status:     &openStatus,
		Page:       1,
		PageSize:   1,
	})
	stats.OpenAlerts = int(openAlerts)

	return stats, nil
}
