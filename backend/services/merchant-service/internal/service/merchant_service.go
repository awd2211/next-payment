package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// MerchantService 商户服务接口
type MerchantService interface {
	// 商户管理
	Create(ctx context.Context, input *CreateMerchantInput) (*model.Merchant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error)
	GetByEmail(ctx context.Context, email string) (*model.Merchant, error)
	List(ctx context.Context, page, pageSize int, status, kycStatus, keyword string) ([]*model.Merchant, int64, error)
	Update(ctx context.Context, id uuid.UUID, input *UpdateMerchantInput) (*model.Merchant, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateKYCStatus(ctx context.Context, id uuid.UUID, kycStatus string) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 商户认证
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	Register(ctx context.Context, input *RegisterMerchantInput) (*model.Merchant, error)

	// 内部接口
	GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*model.Merchant, error)
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error
}

type merchantService struct {
	db           *gorm.DB
	merchantRepo repository.MerchantRepository
	jwtManager   *auth.JWTManager
}

// NewMerchantService 创建商户服务实例（Phase 10：移除 apiKeyRepo）
func NewMerchantService(
	db *gorm.DB,
	merchantRepo repository.MerchantRepository,
	jwtManager *auth.JWTManager,
) MerchantService {
	return &merchantService{
		db:           db,
		merchantRepo: merchantRepo,
		jwtManager:   jwtManager,
	}
}

// CreateMerchantInput 创建商户输入
type CreateMerchantInput struct {
	Name         string
	Email        string
	Password     string
	Phone        string
	CompanyName  string
	BusinessType string
	Country      string
	Website      string
}

// UpdateMerchantInput 更新商户输入
type UpdateMerchantInput struct {
	Name         string
	Phone        string
	CompanyName  string
	BusinessType string
	Country      string
	Website      string
	IsTestMode   *bool
	Metadata     string
}

// RegisterMerchantInput 注册商户输入
type RegisterMerchantInput struct {
	Name         string
	Email        string
	Password     string
	CompanyName  string
	BusinessType string
	Country      string
	Website      string
}

// LoginResponse 登录响应
type LoginResponse struct {
	Merchant *model.Merchant `json:"merchant"`
	Token    string          `json:"token"`
}

func (s *merchantService) Create(ctx context.Context, input *CreateMerchantInput) (*model.Merchant, error) {
	// 检查邮箱是否已存在
	existing, _ := s.merchantRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, fmt.Errorf("邮箱已存在")
	}

	// 哈希密码
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建商户
	merchant := &model.Merchant{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: passwordHash,
		Phone:        input.Phone,
		CompanyName:  input.CompanyName,
		BusinessType: input.BusinessType,
		Country:      input.Country,
		Website:      input.Website,
		Status:       model.MerchantStatusPending,
		KYCStatus:    model.KYCStatusPending,
		IsTestMode:   true,
	}

	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		return nil, fmt.Errorf("创建商户失败: %w", err)
	}

	logger.Info("商户创建成功（Phase 10：APIKey 由 merchant-auth-service 管理）",
		zap.String("merchant_id", merchant.ID.String()),
		zap.String("email", merchant.Email))

	// Note: APIKey 创建已移至 merchant-auth-service (port 40011)
	// 前端需调用 merchant-auth-service API 创建 API Keys

	return merchant, nil
}

func (s *merchantService) Register(ctx context.Context, input *RegisterMerchantInput) (*model.Merchant, error) {
	// 检查邮箱是否已存在
	existing, _ := s.merchantRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 哈希密码
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建商户
	merchant := &model.Merchant{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: passwordHash,
		CompanyName:  input.CompanyName,
		BusinessType: input.BusinessType,
		Country:      input.Country,
		Website:      input.Website,
		Status:       model.MerchantStatusPending,
		KYCStatus:    model.KYCStatusPending,
		IsTestMode:   true,
	}

	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		return nil, fmt.Errorf("注册失败: %w", err)
	}

	logger.Info("商户注册成功（Phase 10：APIKey 由 merchant-auth-service 管理）",
		zap.String("merchant_id", merchant.ID.String()),
		zap.String("email", merchant.Email))

	// Note: APIKey 创建已移至 merchant-auth-service
	// 商户注册后需要调用 merchant-auth-service API 创建 test/prod API Keys

	return merchant, nil
}

func (s *merchantService) GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	return s.merchantRepo.GetByID(ctx, id)
}

func (s *merchantService) GetByEmail(ctx context.Context, email string) (*model.Merchant, error) {
	return s.merchantRepo.GetByEmail(ctx, email)
}

func (s *merchantService) List(ctx context.Context, page, pageSize int, status, kycStatus, keyword string) ([]*model.Merchant, int64, error) {
	offset := (page - 1) * pageSize
	merchants, total, err := s.merchantRepo.List(ctx, offset, pageSize, status, kycStatus, keyword)
	if err != nil {
		return nil, 0, err
	}


	return merchants, total, nil
}

func (s *merchantService) Update(ctx context.Context, id uuid.UUID, input *UpdateMerchantInput) (*model.Merchant, error) {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("商户不存在")
	}

	// 更新字段
	if input.Name != "" {
		merchant.Name = input.Name
	}
	if input.Phone != "" {
		merchant.Phone = input.Phone
	}
	if input.CompanyName != "" {
		merchant.CompanyName = input.CompanyName
	}
	if input.BusinessType != "" {
		merchant.BusinessType = input.BusinessType
	}
	if input.Country != "" {
		merchant.Country = input.Country
	}
	if input.Website != "" {
		merchant.Website = input.Website
	}
	if input.IsTestMode != nil {
		merchant.IsTestMode = *input.IsTestMode
	}
	if input.Metadata != "" {
		merchant.Metadata = &input.Metadata
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return nil, fmt.Errorf("更新商户失败: %w", err)
	}

	return merchant, nil
}

func (s *merchantService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("商户不存在")
	}

	merchant.Status = status
	return s.merchantRepo.Update(ctx, merchant)
}

func (s *merchantService) UpdateKYCStatus(ctx context.Context, id uuid.UUID, kycStatus string) error {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("商户不存在")
	}

	merchant.KYCStatus = kycStatus
	return s.merchantRepo.Update(ctx, merchant)
}

func (s *merchantService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.merchantRepo.Delete(ctx, id)
}

func (s *merchantService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// 查找商户
	merchant, err := s.merchantRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 验证密码
	if err := auth.VerifyPassword(password, merchant.PasswordHash); err != nil {
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 检查商户状态
	if merchant.Status != model.MerchantStatusActive {
		return nil, fmt.Errorf("商户状态异常: %s", merchant.Status)
	}

	// 生成JWT Token

	token, err := s.jwtManager.GenerateToken(merchant.ID, merchant.Email, "merchant", &merchant.ID, []string{"merchant"}, nil)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	logger.Info("商户登录成功",
		zap.String("merchant_id", merchant.ID.String()),
		zap.String("email", merchant.Email))

	return &LoginResponse{
		Merchant: merchant,
		Token:    token,
	}, nil
}

func (s *merchantService) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	return s.merchantRepo.GetByID(ctx, id)
}

func (s *merchantService) UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("商户不存在")
	}

	merchant.PasswordHash = passwordHash
	return s.merchantRepo.Update(ctx, merchant)
}
