package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/services/merchant-service/internal/model"
	"github.com/payment-platform/services/merchant-service/internal/repository"
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
}

type merchantService struct {
	merchantRepo repository.MerchantRepository
	apiKeyRepo   repository.APIKeyRepository
	jwtManager   *auth.JWTManager
}

// NewMerchantService 创建商户服务实例
func NewMerchantService(
	merchantRepo repository.MerchantRepository,
	apiKeyRepo repository.APIKeyRepository,
	jwtManager *auth.JWTManager,
) MerchantService {
	return &merchantService{
		merchantRepo: merchantRepo,
		apiKeyRepo:   apiKeyRepo,
		jwtManager:   jwtManager,
	}
}

// CreateMerchantInput 创建商户输入
type CreateMerchantInput struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	Phone        string `json:"phone"`
	CompanyName  string `json:"company_name"`
	BusinessType string `json:"business_type"` // individual, company
	Country      string `json:"country"`
	Website      string `json:"website"`
}

// UpdateMerchantInput 更新商户输入
type UpdateMerchantInput struct {
	Name         *string `json:"name"`
	Phone        *string `json:"phone"`
	CompanyName  *string `json:"company_name"`
	BusinessType *string `json:"business_type"`
	Country      *string `json:"country"`
	Website      *string `json:"website"`
	Metadata     *string `json:"metadata"`
}

// RegisterMerchantInput 商户注册输入
type RegisterMerchantInput struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	CompanyName  string `json:"company_name" binding:"required"`
	BusinessType string `json:"business_type" binding:"required"`
	Country      string `json:"country" binding:"required"`
	Website      string `json:"website"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string           `json:"token"`
	Merchant *model.Merchant  `json:"merchant"`
}

// Create 创建商户
func (s *merchantService) Create(ctx context.Context, input *CreateMerchantInput) (*model.Merchant, error) {
	// 检查邮箱是否已存在
	existing, err := s.merchantRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("邮箱已被使用")
	}

	// 加密密码
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
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

	// 创建默认测试API Key
	if err := s.createDefaultAPIKeys(ctx, merchant.ID); err != nil {
		return nil, fmt.Errorf("创建默认API Key失败: %w", err)
	}

	return merchant, nil
}

// GetByID 根据ID获取商户
func (s *merchantService) GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取商户失败: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("商户不存在")
	}
	return merchant, nil
}

// GetByEmail 根据邮箱获取商户
func (s *merchantService) GetByEmail(ctx context.Context, email string) (*model.Merchant, error) {
	return s.merchantRepo.GetByEmail(ctx, email)
}

// List 分页查询商户列表
func (s *merchantService) List(ctx context.Context, page, pageSize int, status, kycStatus, keyword string) ([]*model.Merchant, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.merchantRepo.List(ctx, page, pageSize, status, kycStatus, keyword)
}

// Update 更新商户
func (s *merchantService) Update(ctx context.Context, id uuid.UUID, input *UpdateMerchantInput) (*model.Merchant, error) {
	merchant, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		merchant.Name = *input.Name
	}
	if input.Phone != nil {
		merchant.Phone = *input.Phone
	}
	if input.CompanyName != nil {
		merchant.CompanyName = *input.CompanyName
	}
	if input.BusinessType != nil {
		merchant.BusinessType = *input.BusinessType
	}
	if input.Country != nil {
		merchant.Country = *input.Country
	}
	if input.Website != nil {
		merchant.Website = *input.Website
	}
	if input.Metadata != nil {
		merchant.Metadata = *input.Metadata
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return nil, fmt.Errorf("更新商户失败: %w", err)
	}

	return merchant, nil
}

// UpdateStatus 更新商户状态
func (s *merchantService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	// 验证状态
	validStatuses := []string{
		model.MerchantStatusPending,
		model.MerchantStatusActive,
		model.MerchantStatusSuspended,
		model.MerchantStatusRejected,
	}

	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("无效的状态: %s", status)
	}

	return s.merchantRepo.UpdateStatus(ctx, id, status)
}

// UpdateKYCStatus 更新KYC状态
func (s *merchantService) UpdateKYCStatus(ctx context.Context, id uuid.UUID, kycStatus string) error {
	// 验证KYC状态
	validStatuses := []string{
		model.KYCStatusPending,
		model.KYCStatusVerified,
		model.KYCStatusRejected,
	}

	isValid := false
	for _, s := range validStatuses {
		if kycStatus == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("无效的KYC状态: %s", kycStatus)
	}

	return s.merchantRepo.UpdateKYCStatus(ctx, id, kycStatus)
}

// Delete 删除商户
func (s *merchantService) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查商户是否存在
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.merchantRepo.Delete(ctx, id)
}

// Login 商户登录
func (s *merchantService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// 查找商户
	merchant, err := s.merchantRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 验证密码
	if err := auth.VerifyPassword(password, merchant.PasswordHash); err != nil {
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 检查商户状态
	if merchant.Status == model.MerchantStatusSuspended {
		return nil, fmt.Errorf("账户已被暂停")
	}
	if merchant.Status == model.MerchantStatusRejected {
		return nil, fmt.Errorf("账户已被拒绝")
	}

	// 生成JWT Token
	claims := &auth.Claims{
		UserID:   merchant.ID,
		Username: merchant.Email,
		UserType: "merchant",
		Roles:    []string{"merchant"},
	}

	token, err := s.jwtManager.GenerateToken(claims, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 清除密码字段
	merchant.PasswordHash = ""

	return &LoginResponse{
		Token:    token,
		Merchant: merchant,
	}, nil
}

// Register 商户注册
func (s *merchantService) Register(ctx context.Context, input *RegisterMerchantInput) (*model.Merchant, error) {
	// 检查邮箱是否已存在
	existing, err := s.merchantRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("邮箱已被使用")
	}

	// 验证业务类型
	if input.BusinessType != model.BusinessTypeIndividual && input.BusinessType != model.BusinessTypeCompany {
		return nil, fmt.Errorf("无效的业务类型")
	}

	// 加密密码
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
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
		Status:       model.MerchantStatusPending, // 待审核
		KYCStatus:    model.KYCStatusPending,      // 待KYC验证
		IsTestMode:   true,                         // 默认测试模式
	}

	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		return nil, fmt.Errorf("创建商户失败: %w", err)
	}

	// 创建默认测试API Keys
	if err := s.createDefaultAPIKeys(ctx, merchant.ID); err != nil {
		// 不影响注册流程，只记录错误
		fmt.Printf("创建默认API Keys失败: %v\n", err)
	}

	// 清除密码字段
	merchant.PasswordHash = ""

	return merchant, nil
}

// createDefaultAPIKeys 创建默认API Keys
func (s *merchantService) createDefaultAPIKeys(ctx context.Context, merchantID uuid.UUID) error {
	// 创建测试环境API Key
	testAPIKey := &model.APIKey{
		MerchantID:  merchantID,
		APIKey:      s.generateAPIKey("pk_test"),
		APISecret:   s.generateAPISecret(),
		Name:        "Test API Key",
		Environment: model.EnvironmentTest,
		IsActive:    true,
	}

	if err := s.apiKeyRepo.Create(ctx, testAPIKey); err != nil {
		return err
	}

	// 创建生产环境API Key（默认不激活）
	prodAPIKey := &model.APIKey{
		MerchantID:  merchantID,
		APIKey:      s.generateAPIKey("pk_live"),
		APISecret:   s.generateAPISecret(),
		Name:        "Production API Key",
		Environment: model.EnvironmentProduction,
		IsActive:    false, // 生产环境默认不激活，需要通过KYC后手动激活
	}

	return s.apiKeyRepo.Create(ctx, prodAPIKey)
}

// generateAPIKey 生成API Key
func (s *merchantService) generateAPIKey(prefix string) string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, base64.URLEncoding.EncodeToString(b)[:43])
}

// generateAPISecret 生成API Secret
func (s *merchantService) generateAPISecret() string {
	b := make([]byte, 64)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
