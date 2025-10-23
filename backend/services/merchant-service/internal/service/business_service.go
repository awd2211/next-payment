package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// BusinessService 业务服务接口（聚合所有业务功能）
type BusinessService interface {
	// 结算账户相关
	CreateSettlementAccount(ctx context.Context, input *CreateSettlementAccountInput) (*model.SettlementAccount, error)
	GetSettlementAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.SettlementAccount, error)
	UpdateSettlementAccount(ctx context.Context, id uuid.UUID, input *UpdateSettlementAccountInput) (*model.SettlementAccount, error)
	DeleteSettlementAccount(ctx context.Context, id uuid.UUID) error
	SetDefaultAccount(ctx context.Context, merchantID, accountID uuid.UUID) error
	VerifySettlementAccount(ctx context.Context, id uuid.UUID, status string, reason string) error

	// KYC文档相关
	UploadKYCDocument(ctx context.Context, input *UploadKYCDocumentInput) (*model.KYCDocument, error)
	GetKYCDocuments(ctx context.Context, merchantID uuid.UUID, documentType string) ([]*model.KYCDocument, error)
	ReviewKYCDocument(ctx context.Context, id uuid.UUID, status, reviewNotes string, reviewedBy uuid.UUID) error
	DeleteKYCDocument(ctx context.Context, id uuid.UUID) error

	// 费率配置相关
	CreateFeeConfig(ctx context.Context, input *CreateFeeConfigInput) (*model.MerchantFeeConfig, error)
	GetFeeConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error)
	UpdateFeeConfig(ctx context.Context, id uuid.UUID, input *UpdateFeeConfigInput) (*model.MerchantFeeConfig, error)
	DeleteFeeConfig(ctx context.Context, id uuid.UUID) error

	// 子账户相关
	InviteUser(ctx context.Context, input *InviteUserInput) (*model.MerchantUser, error)
	GetMerchantUsers(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantUser, error)
	UpdateMerchantUser(ctx context.Context, id uuid.UUID, input *UpdateMerchantUserInput) (*model.MerchantUser, error)
	DeleteMerchantUser(ctx context.Context, id uuid.UUID) error

	// 交易限额相关
	CreateTransactionLimit(ctx context.Context, input *CreateTransactionLimitInput) (*model.MerchantTransactionLimit, error)
	GetTransactionLimits(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error)
	UpdateTransactionLimit(ctx context.Context, id uuid.UUID, input *UpdateTransactionLimitInput) (*model.MerchantTransactionLimit, error)
	DeleteTransactionLimit(ctx context.Context, id uuid.UUID) error

	// 业务资质相关
	CreateQualification(ctx context.Context, input *CreateQualificationInput) (*model.BusinessQualification, error)
	GetQualifications(ctx context.Context, merchantID uuid.UUID) ([]*model.BusinessQualification, error)
	VerifyQualification(ctx context.Context, id uuid.UUID, status string, verifiedBy uuid.UUID) error
	DeleteQualification(ctx context.Context, id uuid.UUID) error
}

type businessService struct {
	settlementAccountRepo repository.SettlementAccountRepository
	kycDocRepo            repository.KYCDocumentRepository
	feeConfigRepo         repository.MerchantFeeConfigRepository
	merchantUserRepo      repository.MerchantUserRepository
	transactionLimitRepo  repository.MerchantTransactionLimitRepository
	qualificationRepo     repository.BusinessQualificationRepository
	merchantRepo          repository.MerchantRepository
}

// NewBusinessService 创建业务服务实例
func NewBusinessService(
	settlementAccountRepo repository.SettlementAccountRepository,
	kycDocRepo repository.KYCDocumentRepository,
	feeConfigRepo repository.MerchantFeeConfigRepository,
	merchantUserRepo repository.MerchantUserRepository,
	transactionLimitRepo repository.MerchantTransactionLimitRepository,
	qualificationRepo repository.BusinessQualificationRepository,
	merchantRepo repository.MerchantRepository,
) BusinessService {
	return &businessService{
		settlementAccountRepo: settlementAccountRepo,
		kycDocRepo:            kycDocRepo,
		feeConfigRepo:         feeConfigRepo,
		merchantUserRepo:      merchantUserRepo,
		transactionLimitRepo:  transactionLimitRepo,
		qualificationRepo:     qualificationRepo,
		merchantRepo:          merchantRepo,
	}
}

// ==================== 结算账户相关 ====================

type CreateSettlementAccountInput struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	AccountType   string    `json:"account_type" binding:"required"`
	BankName      string    `json:"bank_name"`
	AccountNumber string    `json:"account_number" binding:"required"`
	AccountName   string    `json:"account_name" binding:"required"`
	SwiftCode     string    `json:"swift_code"`
	IBAN          string    `json:"iban"`
	Currency      string    `json:"currency"`
	Country       string    `json:"country"`
}

type UpdateSettlementAccountInput struct {
	BankName    *string `json:"bank_name"`
	AccountName *string `json:"account_name"`
	SwiftCode   *string `json:"swift_code"`
	IBAN        *string `json:"iban"`
}

func (s *businessService) CreateSettlementAccount(ctx context.Context, input *CreateSettlementAccountInput) (*model.SettlementAccount, error) {
	// 验证商户是否存在
	_, err := s.merchantRepo.GetByID(ctx, input.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("商户不存在")
	}

	account := &model.SettlementAccount{
		MerchantID:    input.MerchantID,
		AccountType:   input.AccountType,
		BankName:      input.BankName,
		AccountNumber: input.AccountNumber, // TODO: 应该加密存储
		AccountName:   input.AccountName,
		SwiftCode:     input.SwiftCode,
		IBAN:          input.IBAN,
		Currency:      input.Currency,
		Country:       input.Country,
		Status:        model.AccountStatusPendingVerify,
	}

	if err := s.settlementAccountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("创建结算账户失败: %w", err)
	}

	return account, nil
}

func (s *businessService) GetSettlementAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.SettlementAccount, error) {
	return s.settlementAccountRepo.GetByMerchantID(ctx, merchantID)
}

func (s *businessService) UpdateSettlementAccount(ctx context.Context, id uuid.UUID, input *UpdateSettlementAccountInput) (*model.SettlementAccount, error) {
	account, err := s.settlementAccountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("结算账户不存在")
	}

	if input.BankName != nil {
		account.BankName = *input.BankName
	}
	if input.AccountName != nil {
		account.AccountName = *input.AccountName
	}
	if input.SwiftCode != nil {
		account.SwiftCode = *input.SwiftCode
	}
	if input.IBAN != nil {
		account.IBAN = *input.IBAN
	}

	if err := s.settlementAccountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("更新结算账户失败: %w", err)
	}

	return account, nil
}

func (s *businessService) DeleteSettlementAccount(ctx context.Context, id uuid.UUID) error {
	return s.settlementAccountRepo.Delete(ctx, id)
}

func (s *businessService) SetDefaultAccount(ctx context.Context, merchantID, accountID uuid.UUID) error {
	return s.settlementAccountRepo.SetDefault(ctx, merchantID, accountID)
}

func (s *businessService) VerifySettlementAccount(ctx context.Context, id uuid.UUID, status string, reason string) error {
	account, err := s.settlementAccountRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("结算账户不存在")
	}

	account.Status = status
	if status == model.AccountStatusRejected {
		account.RejectReason = reason
	}

	return s.settlementAccountRepo.Update(ctx, account)
}

// ==================== KYC文档相关 ====================

type UploadKYCDocumentInput struct {
	MerchantID   uuid.UUID `json:"merchant_id"`
	DocumentType string    `json:"document_type" binding:"required"`
	FileURL      string    `json:"file_url" binding:"required"`
	FileHash     string    `json:"file_hash"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	ExpiryDate   *string   `json:"expiry_date"`
}

func (s *businessService) UploadKYCDocument(ctx context.Context, input *UploadKYCDocumentInput) (*model.KYCDocument, error) {
	doc := &model.KYCDocument{
		MerchantID:   input.MerchantID,
		DocumentType: input.DocumentType,
		FileURL:      input.FileURL,
		FileHash:     input.FileHash,
		FileSize:     input.FileSize,
		MimeType:     input.MimeType,
		Status:       model.DocumentStatusPending,
	}

	if err := s.kycDocRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("上传KYC文档失败: %w", err)
	}

	return doc, nil
}

func (s *businessService) GetKYCDocuments(ctx context.Context, merchantID uuid.UUID, documentType string) ([]*model.KYCDocument, error) {
	return s.kycDocRepo.GetByMerchantID(ctx, merchantID, documentType)
}

func (s *businessService) ReviewKYCDocument(ctx context.Context, id uuid.UUID, status, reviewNotes string, reviewedBy uuid.UUID) error {
	return s.kycDocRepo.UpdateStatus(ctx, id, status, reviewNotes, reviewedBy)
}

func (s *businessService) DeleteKYCDocument(ctx context.Context, id uuid.UUID) error {
	return s.kycDocRepo.Delete(ctx, id)
}

// ==================== 费率配置相关 ====================

type CreateFeeConfigInput struct {
	MerchantID     uuid.UUID `json:"merchant_id"`
	Channel        string    `json:"channel" binding:"required"`
	PaymentMethod  string    `json:"payment_method"`
	FeeType        string    `json:"fee_type" binding:"required"`
	FeePercentage  float64   `json:"fee_percentage"`
	FeeFixed       int64     `json:"fee_fixed"`
	MinFee         int64     `json:"min_fee"`
	MaxFee         int64     `json:"max_fee"`
	Currency       string    `json:"currency"`
}

type UpdateFeeConfigInput struct {
	FeePercentage *float64 `json:"fee_percentage"`
	FeeFixed      *int64   `json:"fee_fixed"`
	MinFee        *int64   `json:"min_fee"`
	MaxFee        *int64   `json:"max_fee"`
	Status        *string  `json:"status"`
}

func (s *businessService) CreateFeeConfig(ctx context.Context, input *CreateFeeConfigInput) (*model.MerchantFeeConfig, error) {
	config := &model.MerchantFeeConfig{
		MerchantID:    input.MerchantID,
		Channel:       input.Channel,
		PaymentMethod: input.PaymentMethod,
		FeeType:       input.FeeType,
		FeePercentage: input.FeePercentage,
		FeeFixed:      input.FeeFixed,
		MinFee:        input.MinFee,
		MaxFee:        input.MaxFee,
		Currency:      input.Currency,
		Status:        "active",
	}

	if err := s.feeConfigRepo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("创建费率配置失败: %w", err)
	}

	return config, nil
}

func (s *businessService) GetFeeConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error) {
	return s.feeConfigRepo.GetByMerchantID(ctx, merchantID)
}

func (s *businessService) UpdateFeeConfig(ctx context.Context, id uuid.UUID, input *UpdateFeeConfigInput) (*model.MerchantFeeConfig, error) {
	config, err := s.feeConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("费率配置不存在")
	}

	if input.FeePercentage != nil {
		config.FeePercentage = *input.FeePercentage
	}
	if input.FeeFixed != nil {
		config.FeeFixed = *input.FeeFixed
	}
	if input.MinFee != nil {
		config.MinFee = *input.MinFee
	}
	if input.MaxFee != nil {
		config.MaxFee = *input.MaxFee
	}
	if input.Status != nil {
		config.Status = *input.Status
	}

	if err := s.feeConfigRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("更新费率配置失败: %w", err)
	}

	return config, nil
}

func (s *businessService) DeleteFeeConfig(ctx context.Context, id uuid.UUID) error {
	return s.feeConfigRepo.Delete(ctx, id)
}

// ==================== 子账户相关 ====================

type InviteUserInput struct {
	MerchantID  uuid.UUID `json:"merchant_id"`
	Email       string    `json:"email" binding:"required,email"`
	Name        string    `json:"name" binding:"required"`
	Role        string    `json:"role" binding:"required"`
	Permissions string    `json:"permissions"`
	InvitedBy   uuid.UUID `json:"invited_by"`
}

type UpdateMerchantUserInput struct {
	Name        *string `json:"name"`
	Role        *string `json:"role"`
	Permissions *string `json:"permissions"`
	Status      *string `json:"status"`
}

func (s *businessService) InviteUser(ctx context.Context, input *InviteUserInput) (*model.MerchantUser, error) {
	// 检查邮箱是否已存在
	existing, _ := s.merchantUserRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, fmt.Errorf("该邮箱已被邀请")
	}

	user := &model.MerchantUser{
		MerchantID:  input.MerchantID,
		Email:       input.Email,
		Name:        input.Name,
		Role:        input.Role,
		Permissions: input.Permissions,
		Status:      model.UserStatusPending,
		InvitedBy:   &input.InvitedBy,
	}

	if err := s.merchantUserRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("邀请用户失败: %w", err)
	}

	// TODO: 发送邀请邮件

	return user, nil
}

func (s *businessService) GetMerchantUsers(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantUser, error) {
	return s.merchantUserRepo.GetByMerchantID(ctx, merchantID)
}

func (s *businessService) UpdateMerchantUser(ctx context.Context, id uuid.UUID, input *UpdateMerchantUserInput) (*model.MerchantUser, error) {
	user, err := s.merchantUserRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.Permissions != nil {
		user.Permissions = *input.Permissions
	}
	if input.Status != nil {
		user.Status = *input.Status
	}

	if err := s.merchantUserRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return user, nil
}

func (s *businessService) DeleteMerchantUser(ctx context.Context, id uuid.UUID) error {
	return s.merchantUserRepo.Delete(ctx, id)
}

// ==================== 交易限额相关 ====================

type CreateTransactionLimitInput struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	LimitType     string    `json:"limit_type" binding:"required"`
	PaymentMethod string    `json:"payment_method"`
	Channel       string    `json:"channel"`
	Currency      string    `json:"currency"`
	MinAmount     int64     `json:"min_amount"`
	MaxAmount     int64     `json:"max_amount"`
	MaxCount      int       `json:"max_count"`
}

type UpdateTransactionLimitInput struct {
	MinAmount *int64  `json:"min_amount"`
	MaxAmount *int64  `json:"max_amount"`
	MaxCount  *int    `json:"max_count"`
	Status    *string `json:"status"`
}

func (s *businessService) CreateTransactionLimit(ctx context.Context, input *CreateTransactionLimitInput) (*model.MerchantTransactionLimit, error) {
	limit := &model.MerchantTransactionLimit{
		MerchantID:    input.MerchantID,
		LimitType:     input.LimitType,
		PaymentMethod: input.PaymentMethod,
		Channel:       input.Channel,
		Currency:      input.Currency,
		MinAmount:     input.MinAmount,
		MaxAmount:     input.MaxAmount,
		MaxCount:      input.MaxCount,
		Status:        "active",
	}

	if err := s.transactionLimitRepo.Create(ctx, limit); err != nil {
		return nil, fmt.Errorf("创建交易限额失败: %w", err)
	}

	return limit, nil
}

func (s *businessService) GetTransactionLimits(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error) {
	return s.transactionLimitRepo.GetByMerchantID(ctx, merchantID)
}

func (s *businessService) UpdateTransactionLimit(ctx context.Context, id uuid.UUID, input *UpdateTransactionLimitInput) (*model.MerchantTransactionLimit, error) {
	limit, err := s.transactionLimitRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("交易限额不存在")
	}

	if input.MinAmount != nil {
		limit.MinAmount = *input.MinAmount
	}
	if input.MaxAmount != nil {
		limit.MaxAmount = *input.MaxAmount
	}
	if input.MaxCount != nil {
		limit.MaxCount = *input.MaxCount
	}
	if input.Status != nil {
		limit.Status = *input.Status
	}

	if err := s.transactionLimitRepo.Update(ctx, limit); err != nil {
		return nil, fmt.Errorf("更新交易限额失败: %w", err)
	}

	return limit, nil
}

func (s *businessService) DeleteTransactionLimit(ctx context.Context, id uuid.UUID) error {
	return s.transactionLimitRepo.Delete(ctx, id)
}

// ==================== 业务资质相关 ====================

type CreateQualificationInput struct {
	MerchantID        uuid.UUID `json:"merchant_id"`
	QualificationType string    `json:"qualification_type" binding:"required"`
	LicenseNumber     string    `json:"license_number" binding:"required"`
	LicenseName       string    `json:"license_name"`
	IssuedBy          string    `json:"issued_by"`
	FileURL           string    `json:"file_url"`
}

func (s *businessService) CreateQualification(ctx context.Context, input *CreateQualificationInput) (*model.BusinessQualification, error) {
	qualification := &model.BusinessQualification{
		MerchantID:        input.MerchantID,
		QualificationType: input.QualificationType,
		LicenseNumber:     input.LicenseNumber,
		LicenseName:       input.LicenseName,
		IssuedBy:          input.IssuedBy,
		FileURL:           input.FileURL,
		Status:            model.QualificationStatusPending,
	}

	if err := s.qualificationRepo.Create(ctx, qualification); err != nil {
		return nil, fmt.Errorf("创建业务资质失败: %w", err)
	}

	return qualification, nil
}

func (s *businessService) GetQualifications(ctx context.Context, merchantID uuid.UUID) ([]*model.BusinessQualification, error) {
	return s.qualificationRepo.GetByMerchantID(ctx, merchantID)
}

func (s *businessService) VerifyQualification(ctx context.Context, id uuid.UUID, status string, verifiedBy uuid.UUID) error {
	return s.qualificationRepo.UpdateStatus(ctx, id, status, verifiedBy)
}

func (s *businessService) DeleteQualification(ctx context.Context, id uuid.UUID) error {
	return s.qualificationRepo.Delete(ctx, id)
}
