package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/email"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// MerchantUserService 商户子账户服务接口
// Phase 7-8评估决定：保留在merchant-service（属于Merchant聚合根）
type MerchantUserService interface {
	InviteUser(ctx context.Context, input *InviteUserInput) (*model.MerchantUser, error)
	GetMerchantUsers(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantUser, error)
	UpdateMerchantUser(ctx context.Context, id uuid.UUID, input *UpdateMerchantUserInput) (*model.MerchantUser, error)
	DeleteMerchantUser(ctx context.Context, id uuid.UUID) error
}

type merchantUserService struct {
	merchantUserRepo repository.MerchantUserRepository
	merchantRepo     repository.MerchantRepository
	emailProvider    email.EmailProvider
}

func NewMerchantUserService(
	merchantUserRepo repository.MerchantUserRepository,
	merchantRepo repository.MerchantRepository,
	emailProvider email.EmailProvider,
) MerchantUserService {
	return &merchantUserService{
		merchantUserRepo: merchantUserRepo,
		merchantRepo:     merchantRepo,
		emailProvider:    emailProvider,
	}
}

// InviteUserInput 邀请用户输入
type InviteUserInput struct {
	MerchantID  uuid.UUID
	Email       string
	Name        string
	Phone       string
	Role        string
	Permissions string
	InvitedBy   uuid.UUID
}

// UpdateMerchantUserInput 更新用户输入
type UpdateMerchantUserInput struct {
	Name        string
	Phone       string
	Role        string
	Permissions string
	Status      string
}

func (s *merchantUserService) InviteUser(ctx context.Context, input *InviteUserInput) (*model.MerchantUser, error) {
	// 验证商户存在
	merchant, err := s.merchantRepo.GetByID(ctx, input.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("商户不存在: %w", err)
	}

	// 检查邮箱是否已存在
	existing, _ := s.merchantUserRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, errors.New("该邮箱已被邀请")
	}

	// 创建子账户
	user := &model.MerchantUser{
		MerchantID:  input.MerchantID,
		Email:       input.Email,
		Name:        input.Name,
		Phone:       input.Phone,
		Role:        input.Role,
		Permissions: input.Permissions,
		Status:      model.UserStatusPending,
		InvitedBy:   &input.InvitedBy,
		InvitedAt:   time.Now(),
	}

	if err := s.merchantUserRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建子账户失败: %w", err)
	}

	// 发送邀请邮件（异步）
	if s.emailProvider != nil {
		_ = s.emailProvider.Send(
			[]string{input.Email},
			fmt.Sprintf("您已被邀请加入 %s", merchant.CompanyName),
			fmt.Sprintf("您好 %s，您已被邀请加入 %s 的团队，角色为 %s。", input.Name, merchant.CompanyName, input.Role),
			"",
			nil,
		)
	}

	return user, nil
}

func (s *merchantUserService) GetMerchantUsers(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantUser, error) {
	return s.merchantUserRepo.GetByMerchantID(ctx, merchantID)
}

func (s *merchantUserService) UpdateMerchantUser(ctx context.Context, id uuid.UUID, input *UpdateMerchantUserInput) (*model.MerchantUser, error) {
	user, err := s.merchantUserRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("子账户不存在: %w", err)
	}

	// 更新字段
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}
	if input.Role != "" {
		user.Role = input.Role
	}
	if input.Permissions != "" {
		user.Permissions = input.Permissions
	}
	if input.Status != "" {
		user.Status = input.Status
	}

	if err := s.merchantUserRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("更新子账户失败: %w", err)
	}

	return user, nil
}

func (s *merchantUserService) DeleteMerchantUser(ctx context.Context, id uuid.UUID) error {
	return s.merchantUserRepo.Delete(ctx, id)
}
