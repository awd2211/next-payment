package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/repository"
)

// SettlementAccountService 结算账户服务接口
type SettlementAccountService interface {
	CreateAccount(ctx context.Context, account *model.SettlementAccount) error
	GetAccount(ctx context.Context, id uuid.UUID) (*model.SettlementAccount, error)
	GetMerchantAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.SettlementAccount, error)
	GetDefaultAccount(ctx context.Context, merchantID uuid.UUID) (*model.SettlementAccount, error)
	UpdateAccount(ctx context.Context, account *model.SettlementAccount) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	SetDefaultAccount(ctx context.Context, merchantID, accountID uuid.UUID) error
	VerifyAccount(ctx context.Context, id uuid.UUID, method string) error
	RejectAccount(ctx context.Context, id uuid.UUID, reason string) error
	ListPendingAccounts(ctx context.Context, limit, offset int) ([]*model.SettlementAccount, int64, error)
}

type settlementAccountService struct {
	repo repository.SettlementAccountRepository
}

// NewSettlementAccountService 创建结算账户服务
func NewSettlementAccountService(repo repository.SettlementAccountRepository) SettlementAccountService {
	return &settlementAccountService{repo: repo}
}

// CreateAccount 创建结算账户
func (s *settlementAccountService) CreateAccount(ctx context.Context, account *model.SettlementAccount) error {
	// 验证账户类型
	if !isValidAccountType(account.AccountType) {
		return errors.New("invalid account type")
	}

	// 设置初始状态
	account.Status = model.AccountStatusPendingVerify

	return s.repo.Create(ctx, account)
}

// GetAccount 获取结算账户
func (s *settlementAccountService) GetAccount(ctx context.Context, id uuid.UUID) (*model.SettlementAccount, error) {
	return s.repo.GetByID(ctx, id)
}

// GetMerchantAccounts 获取商户的所有结算账户
func (s *settlementAccountService) GetMerchantAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.SettlementAccount, error) {
	return s.repo.GetByMerchantID(ctx, merchantID)
}

// GetDefaultAccount 获取商户的默认结算账户
func (s *settlementAccountService) GetDefaultAccount(ctx context.Context, merchantID uuid.UUID) (*model.SettlementAccount, error) {
	return s.repo.GetDefaultByMerchantID(ctx, merchantID)
}

// UpdateAccount 更新结算账户
func (s *settlementAccountService) UpdateAccount(ctx context.Context, account *model.SettlementAccount) error {
	// 验证账户存在
	existing, err := s.repo.GetByID(ctx, account.ID)
	if err != nil {
		return err
	}

	// 不允许修改商户ID
	if existing.MerchantID != account.MerchantID {
		return errors.New("cannot change merchant_id")
	}

	return s.repo.Update(ctx, account)
}

// DeleteAccount 删除结算账户
func (s *settlementAccountService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 不允许删除默认账户
	if account.IsDefault {
		return errors.New("cannot delete default account")
	}

	return s.repo.Delete(ctx, id)
}

// SetDefaultAccount 设置默认结算账户
func (s *settlementAccountService) SetDefaultAccount(ctx context.Context, merchantID, accountID uuid.UUID) error {
	// 验证账户存在且属于该商户
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}

	if account.MerchantID != merchantID {
		return errors.New("account does not belong to merchant")
	}

	// 只有已验证的账户才能设为默认
	if account.Status != model.AccountStatusVerified {
		return errors.New("only verified accounts can be set as default")
	}

	return s.repo.SetDefault(ctx, merchantID, accountID)
}

// VerifyAccount 验证结算账户
func (s *settlementAccountService) VerifyAccount(ctx context.Context, id uuid.UUID, method string) error {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if account.Status != model.AccountStatusPendingVerify {
		return errors.New("account is not in pending status")
	}

	now := time.Now()
	account.Status = model.AccountStatusVerified
	account.VerificationMethod = method
	account.VerifiedAt = &now

	// 如果是第一个账户，自动设为默认
	accounts, err := s.repo.GetByMerchantID(ctx, account.MerchantID)
	if err == nil && len(accounts) == 1 {
		account.IsDefault = true
	}

	return s.repo.Update(ctx, account)
}

// RejectAccount 拒绝结算账户
func (s *settlementAccountService) RejectAccount(ctx context.Context, id uuid.UUID, reason string) error {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if account.Status != model.AccountStatusPendingVerify {
		return errors.New("account is not in pending status")
	}

	account.Status = model.AccountStatusRejected
	account.RejectReason = reason

	return s.repo.Update(ctx, account)
}

// ListPendingAccounts 列出待验证的结算账户
func (s *settlementAccountService) ListPendingAccounts(ctx context.Context, limit, offset int) ([]*model.SettlementAccount, int64, error) {
	return s.repo.List(ctx, model.AccountStatusPendingVerify, limit, offset)
}

// isValidAccountType 验证账户类型
func isValidAccountType(accountType string) bool {
	validTypes := map[string]bool{
		model.AccountTypeBankAccount:  true,
		model.AccountTypePayPal:       true,
		model.AccountTypeCryptoWallet: true,
		model.AccountTypeAlipay:       true,
		model.AccountTypeWechat:       true,
	}
	return validTypes[accountType]
}
