package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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

	// 内部接口（供其他微服务调用）
	GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*model.Merchant, error)
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error
}

type merchantService struct {
	db           *gorm.DB
	merchantRepo repository.MerchantRepository
	apiKeyRepo   repository.APIKeyRepository
	jwtManager   *auth.JWTManager
}

// NewMerchantService 创建商户服务实例
func NewMerchantService(
	db *gorm.DB,
	merchantRepo repository.MerchantRepository,
	apiKeyRepo repository.APIKeyRepository,
	jwtManager *auth.JWTManager,
) MerchantService {
	return &merchantService{
		db:           db,
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
	Token          string          `json:"token,omitempty"`
	Merchant       *model.Merchant `json:"merchant,omitempty"`
	Require2FA     bool            `json:"require_2fa"`
	TempToken      string          `json:"temp_token,omitempty"` // 临时token，用于2FA验证
}

// Create 创建商户（使用事务保证商户和API Key的原子性）
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

	// 准备商户数据
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
		Metadata:     "{}",
	}

	// 在事务中创建商户和默认API Keys
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建商户
		if err := tx.Create(merchant).Error; err != nil {
			return fmt.Errorf("创建商户失败: %w", err)
		}

		// 2. 创建测试环境API Key
		testAPIKey := &model.APIKey{
			MerchantID:  merchant.ID,
			APIKey:      s.generateAPIKey("pk_test"),
			APISecret:   s.generateAPISecret(),
			Name:        "Test API Key",
			Environment: model.EnvironmentTest,
			IsActive:    true,
		}
		if err := tx.Create(testAPIKey).Error; err != nil {
			return fmt.Errorf("创建测试API Key失败: %w", err)
		}

		// 3. 创建生产环境API Key（默认不激活）
		prodAPIKey := &model.APIKey{
			MerchantID:  merchant.ID,
			APIKey:      s.generateAPIKey("pk_live"),
			APISecret:   s.generateAPISecret(),
			Name:        "Production API Key",
			Environment: model.EnvironmentProduction,
			IsActive:    false, // 生产环境默认不激活
		}
		if err := tx.Create(prodAPIKey).Error; err != nil {
			return fmt.Errorf("创建生产API Key失败: %w", err)
		}

		logger.Info("merchant and API keys created successfully",
			zap.String("merchant_id", merchant.ID.String()),
			zap.String("email", merchant.Email))

		return nil
	})

	if err != nil {
		return nil, err
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
	token, err := s.jwtManager.GenerateToken(
		merchant.ID,
		merchant.Email,
		"merchant",
		&merchant.ID, // tenantID 使用 merchant.ID
		[]string{"merchant"},
		[]string{},
	)
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

	// 准备商户数据
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
		IsTestMode:   true,                        // 默认测试模式
		Metadata:     "{}",                        // 空JSON对象
	}

	// 在事务中创建商户和默认API Keys
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建商户
		if err := tx.Create(merchant).Error; err != nil {
			return fmt.Errorf("创建商户失败: %w", err)
		}

		// 2. 创建测试环境API Key
		testAPIKey := &model.APIKey{
			MerchantID:  merchant.ID,
			APIKey:      s.generateAPIKey("pk_test"),
			APISecret:   s.generateAPISecret(),
			Name:        "Test API Key",
			Environment: model.EnvironmentTest,
			IsActive:    true,
		}
		if err := tx.Create(testAPIKey).Error; err != nil {
			return fmt.Errorf("创建测试API Key失败: %w", err)
		}

		// 3. 创建生产环境API Key（默认不激活）
		prodAPIKey := &model.APIKey{
			MerchantID:  merchant.ID,
			APIKey:      s.generateAPIKey("pk_live"),
			APISecret:   s.generateAPISecret(),
			Name:        "Production API Key",
			Environment: model.EnvironmentProduction,
			IsActive:    false,
		}
		if err := tx.Create(prodAPIKey).Error; err != nil {
			return fmt.Errorf("创建生产API Key失败: %w", err)
		}

		logger.Info("merchant registered successfully",
			zap.String("merchant_id", merchant.ID.String()),
			zap.String("email", merchant.Email))

		return nil
	})

	if err != nil {
		return nil, err
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

// GetByIDWithPassword 获取带密码的商户信息（内部接口）
func (s *merchantService) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	merchant, err := s.merchantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取商户失败: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("商户不存在")
	}

	// 注意：这个方法不清除password_hash字段，因为它是供merchant-auth-service使用的内部接口
	return merchant, nil
}

// UpdatePasswordHash 更新商户密码哈希（内部接口）
func (s *merchantService) UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error {
	// 直接调用repository的UpdatePasswordHash方法，只更新password_hash字段
	if err := s.merchantRepo.UpdatePasswordHash(ctx, id, passwordHash); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}
