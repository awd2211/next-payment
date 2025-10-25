package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payment-platform/merchant-limit-service/internal/model"
	"payment-platform/merchant-limit-service/internal/repository"
)

// LimitService 额度服务接口
type LimitService interface {
	// Tier management
	CreateTier(ctx context.Context, input *CreateTierInput) (*model.MerchantTier, error)
	GetTier(ctx context.Context, tierID uuid.UUID) (*model.MerchantTier, error)
	ListTiers(ctx context.Context) ([]*model.MerchantTier, error)
	UpdateTier(ctx context.Context, tierID uuid.UUID, input *UpdateTierInput) (*model.MerchantTier, error)
	DeleteTier(ctx context.Context, tierID uuid.UUID) error

	// Limit management
	InitializeMerchantLimit(ctx context.Context, merchantID, tierID uuid.UUID) (*model.MerchantLimit, error)
	GetMerchantLimit(ctx context.Context, merchantID uuid.UUID) (*MerchantLimitDetails, error)
	UpdateMerchantLimit(ctx context.Context, merchantID uuid.UUID, input *UpdateLimitInput) (*model.MerchantLimit, error)
	ChangeMerchantTier(ctx context.Context, merchantID, newTierID uuid.UUID) error
	SuspendMerchant(ctx context.Context, merchantID uuid.UUID, reason string) error
	UnsuspendMerchant(ctx context.Context, merchantID uuid.UUID) error

	// Limit enforcement (核心功能)
	CheckLimit(ctx context.Context, merchantID uuid.UUID, amount int64) (*CheckLimitResult, error)
	ConsumeLimit(ctx context.Context, input *ConsumeLimitInput) error
	ReleaseLimit(ctx context.Context, input *ReleaseLimitInput) error

	// Usage logs
	GetUsageHistory(ctx context.Context, merchantID uuid.UUID, startTime, endTime *time.Time, page, pageSize int) (*UsageHistoryResult, error)

	// Statistics
	GetStatistics(ctx context.Context, merchantID uuid.UUID) (*repository.MerchantStatistics, error)
}

// Input/Output DTOs

type CreateTierInput struct {
	TierCode           string  `json:"tier_code" binding:"required"`
	TierName           string  `json:"tier_name" binding:"required"`
	TierLevel          int     `json:"tier_level" binding:"required"`
	Description        string  `json:"description"`
	DailyLimit         int64   `json:"daily_limit" binding:"required"`
	MonthlyLimit       int64   `json:"monthly_limit" binding:"required"`
	SingleTransLimit   int64   `json:"single_trans_limit" binding:"required"`
	TransactionFeeRate float64 `json:"transaction_fee_rate" binding:"required"`
	WithdrawalFeeRate  float64 `json:"withdrawal_fee_rate" binding:"required"`
}

type UpdateTierInput struct {
	TierName           *string  `json:"tier_name"`
	Description        *string  `json:"description"`
	DailyLimit         *int64   `json:"daily_limit"`
	MonthlyLimit       *int64   `json:"monthly_limit"`
	SingleTransLimit   *int64   `json:"single_trans_limit"`
	TransactionFeeRate *float64 `json:"transaction_fee_rate"`
	WithdrawalFeeRate  *float64 `json:"withdrawal_fee_rate"`
}

type UpdateLimitInput struct {
	CustomDailyLimit       *int64 `json:"custom_daily_limit"`
	CustomMonthlyLimit     *int64 `json:"custom_monthly_limit"`
	CustomSingleTransLimit *int64 `json:"custom_single_trans_limit"`
}

type ConsumeLimitInput struct {
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
	PaymentNo  string    `json:"payment_no"`
	OrderNo    string    `json:"order_no"`
	Amount     int64     `json:"amount" binding:"required"`
	Currency   string    `json:"currency" binding:"required"`
}

type ReleaseLimitInput struct {
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
	PaymentNo  string    `json:"payment_no"`
	OrderNo    string    `json:"order_no"`
	Amount     int64     `json:"amount" binding:"required"`
	Currency   string    `json:"currency" binding:"required"`
	Reason     string    `json:"reason"`
}

type CheckLimitResult struct {
	Allowed          bool   `json:"allowed"`
	Reason           string `json:"reason,omitempty"`
	DailyRemaining   int64  `json:"daily_remaining"`
	MonthlyRemaining int64  `json:"monthly_remaining"`
	SingleTransLimit int64  `json:"single_trans_limit"`
}

type MerchantLimitDetails struct {
	Limit      *model.MerchantLimit `json:"limit"`
	Tier       *model.MerchantTier  `json:"tier"`
	Statistics *repository.MerchantStatistics `json:"statistics"`
}

type UsageHistoryResult struct {
	Logs       []*model.LimitUsageLog `json:"logs"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// limitService 额度服务实现
type limitService struct {
	repo repository.LimitRepository
	db   *gorm.DB
}

// NewLimitService 创建额度服务实例
func NewLimitService(repo repository.LimitRepository, db *gorm.DB) LimitService {
	return &limitService{
		repo: repo,
		db:   db,
	}
}

// CreateTier 创建等级
func (s *limitService) CreateTier(ctx context.Context, input *CreateTierInput) (*model.MerchantTier, error) {
	// Check if tier code already exists
	existing, err := s.repo.GetTierByCode(ctx, input.TierCode)
	if err != nil {
		return nil, fmt.Errorf("check existing tier failed: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("tier code already exists: %s", input.TierCode)
	}

	tier := &model.MerchantTier{
		TierCode:           input.TierCode,
		TierName:           input.TierName,
		TierLevel:          input.TierLevel,
		Description:        input.Description,
		DailyLimit:         input.DailyLimit,
		MonthlyLimit:       input.MonthlyLimit,
		SingleTransLimit:   input.SingleTransLimit,
		TransactionFeeRate: input.TransactionFeeRate,
		WithdrawalFeeRate:  input.WithdrawalFeeRate,
	}

	if err := s.repo.CreateTier(ctx, tier); err != nil {
		return nil, fmt.Errorf("create tier failed: %w", err)
	}

	return tier, nil
}

// GetTier 获取等级
func (s *limitService) GetTier(ctx context.Context, tierID uuid.UUID) (*model.MerchantTier, error) {
	tier, err := s.repo.GetTierByID(ctx, tierID)
	if err != nil {
		return nil, fmt.Errorf("get tier failed: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("tier not found")
	}
	return tier, nil
}

// ListTiers 查询所有等级
func (s *limitService) ListTiers(ctx context.Context) ([]*model.MerchantTier, error) {
	tiers, err := s.repo.ListTiers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tiers failed: %w", err)
	}
	return tiers, nil
}

// UpdateTier 更新等级
func (s *limitService) UpdateTier(ctx context.Context, tierID uuid.UUID, input *UpdateTierInput) (*model.MerchantTier, error) {
	tier, err := s.repo.GetTierByID(ctx, tierID)
	if err != nil {
		return nil, fmt.Errorf("get tier failed: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("tier not found")
	}

	// Apply updates
	if input.TierName != nil {
		tier.TierName = *input.TierName
	}
	if input.Description != nil {
		tier.Description = *input.Description
	}
	if input.DailyLimit != nil {
		tier.DailyLimit = *input.DailyLimit
	}
	if input.MonthlyLimit != nil {
		tier.MonthlyLimit = *input.MonthlyLimit
	}
	if input.SingleTransLimit != nil {
		tier.SingleTransLimit = *input.SingleTransLimit
	}
	if input.TransactionFeeRate != nil {
		tier.TransactionFeeRate = *input.TransactionFeeRate
	}
	if input.WithdrawalFeeRate != nil {
		tier.WithdrawalFeeRate = *input.WithdrawalFeeRate
	}

	if err := s.repo.UpdateTier(ctx, tier); err != nil {
		return nil, fmt.Errorf("update tier failed: %w", err)
	}

	return tier, nil
}

// DeleteTier 删除等级
func (s *limitService) DeleteTier(ctx context.Context, tierID uuid.UUID) error {
	// TODO: Check if any merchant is using this tier
	if err := s.repo.DeleteTier(ctx, tierID); err != nil {
		return fmt.Errorf("delete tier failed: %w", err)
	}
	return nil
}

// InitializeMerchantLimit 初始化商户额度
func (s *limitService) InitializeMerchantLimit(ctx context.Context, merchantID, tierID uuid.UUID) (*model.MerchantLimit, error) {
	// Check if limit already exists
	existing, err := s.repo.GetLimitByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("check existing limit failed: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("merchant limit already exists")
	}

	// Verify tier exists
	tier, err := s.repo.GetTierByID(ctx, tierID)
	if err != nil {
		return nil, fmt.Errorf("get tier failed: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("tier not found")
	}

	// Create limit
	now := time.Now()
	limit := &model.MerchantLimit{
		MerchantID:     merchantID,
		TierID:         tierID,
		DailyUsed:      0,
		MonthlyUsed:    0,
		DailyResetAt:   now.Add(24 * time.Hour),
		MonthlyResetAt: now.AddDate(0, 1, 0),
		IsSuspended:    false,
	}

	if err := s.repo.CreateLimit(ctx, limit); err != nil {
		return nil, fmt.Errorf("create limit failed: %w", err)
	}

	return limit, nil
}

// GetMerchantLimit 获取商户额度详情
func (s *limitService) GetMerchantLimit(ctx context.Context, merchantID uuid.UUID) (*MerchantLimitDetails, error) {
	limit, err := s.repo.GetLimitByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("get limit failed: %w", err)
	}
	if limit == nil {
		return nil, fmt.Errorf("merchant limit not found")
	}

	tier, err := s.repo.GetTierByID(ctx, limit.TierID)
	if err != nil {
		return nil, fmt.Errorf("get tier failed: %w", err)
	}

	stats, err := s.repo.GetMerchantStatistics(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("get statistics failed: %w", err)
	}

	return &MerchantLimitDetails{
		Limit:      limit,
		Tier:       tier,
		Statistics: stats,
	}, nil
}

// UpdateMerchantLimit 更新商户额度
func (s *limitService) UpdateMerchantLimit(ctx context.Context, merchantID uuid.UUID, input *UpdateLimitInput) (*model.MerchantLimit, error) {
	limit, err := s.repo.GetLimitByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("get limit failed: %w", err)
	}
	if limit == nil {
		return nil, fmt.Errorf("merchant limit not found")
	}

	// Apply custom limits
	limit.CustomDailyLimit = input.CustomDailyLimit
	limit.CustomMonthlyLimit = input.CustomMonthlyLimit
	limit.CustomSingleTransLimit = input.CustomSingleTransLimit

	if err := s.repo.UpdateLimit(ctx, limit); err != nil {
		return nil, fmt.Errorf("update limit failed: %w", err)
	}

	return limit, nil
}

// ChangeMerchantTier 更改商户等级
func (s *limitService) ChangeMerchantTier(ctx context.Context, merchantID, newTierID uuid.UUID) error {
	limit, err := s.repo.GetLimitByMerchantID(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("get limit failed: %w", err)
	}
	if limit == nil {
		return fmt.Errorf("merchant limit not found")
	}

	// Verify new tier exists
	tier, err := s.repo.GetTierByID(ctx, newTierID)
	if err != nil {
		return fmt.Errorf("get tier failed: %w", err)
	}
	if tier == nil {
		return fmt.Errorf("tier not found")
	}

	limit.TierID = newTierID
	if err := s.repo.UpdateLimit(ctx, limit); err != nil {
		return fmt.Errorf("update limit failed: %w", err)
	}

	return nil
}

// SuspendMerchant 暂停商户
func (s *limitService) SuspendMerchant(ctx context.Context, merchantID uuid.UUID, reason string) error {
	if err := s.repo.SuspendMerchant(ctx, merchantID, reason); err != nil {
		return fmt.Errorf("suspend merchant failed: %w", err)
	}
	return nil
}

// UnsuspendMerchant 恢复商户
func (s *limitService) UnsuspendMerchant(ctx context.Context, merchantID uuid.UUID) error {
	if err := s.repo.UnsuspendMerchant(ctx, merchantID); err != nil {
		return fmt.Errorf("unsuspend merchant failed: %w", err)
	}
	return nil
}

// CheckLimit 检查额度 (核心功能)
func (s *limitService) CheckLimit(ctx context.Context, merchantID uuid.UUID, amount int64) (*CheckLimitResult, error) {
	limit, err := s.repo.GetLimitByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("get limit failed: %w", err)
	}
	if limit == nil {
		return &CheckLimitResult{
			Allowed: false,
			Reason:  "Merchant limit not initialized",
		}, nil
	}

	// Check if suspended
	if limit.IsSuspended {
		return &CheckLimitResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Merchant is suspended: %s", limit.SuspendedReason),
		}, nil
	}

	// Get tier
	tier, err := s.repo.GetTierByID(ctx, limit.TierID)
	if err != nil {
		return nil, fmt.Errorf("get tier failed: %w", err)
	}
	if tier == nil {
		return &CheckLimitResult{
			Allowed: false,
			Reason:  "Tier not found",
		}, nil
	}

	// Calculate effective limits (custom overrides tier)
	dailyLimit := tier.DailyLimit
	if limit.CustomDailyLimit != nil {
		dailyLimit = *limit.CustomDailyLimit
	}

	monthlyLimit := tier.MonthlyLimit
	if limit.CustomMonthlyLimit != nil {
		monthlyLimit = *limit.CustomMonthlyLimit
	}

	singleTransLimit := tier.SingleTransLimit
	if limit.CustomSingleTransLimit != nil {
		singleTransLimit = *limit.CustomSingleTransLimit
	}

	// Check single transaction limit
	if amount > singleTransLimit {
		return &CheckLimitResult{
			Allowed:          false,
			Reason:           fmt.Sprintf("Amount exceeds single transaction limit: %d > %d", amount, singleTransLimit),
			SingleTransLimit: singleTransLimit,
		}, nil
	}

	// Check daily limit
	dailyRemaining := dailyLimit - limit.DailyUsed
	if amount > dailyRemaining {
		return &CheckLimitResult{
			Allowed:        false,
			Reason:         fmt.Sprintf("Amount exceeds daily remaining limit: %d > %d", amount, dailyRemaining),
			DailyRemaining: dailyRemaining,
		}, nil
	}

	// Check monthly limit
	monthlyRemaining := monthlyLimit - limit.MonthlyUsed
	if amount > monthlyRemaining {
		return &CheckLimitResult{
			Allowed:          false,
			Reason:           fmt.Sprintf("Amount exceeds monthly remaining limit: %d > %d", amount, monthlyRemaining),
			MonthlyRemaining: monthlyRemaining,
		}, nil
	}

	// All checks passed
	return &CheckLimitResult{
		Allowed:          true,
		DailyRemaining:   dailyRemaining - amount,
		MonthlyRemaining: monthlyRemaining - amount,
		SingleTransLimit: singleTransLimit,
	}, nil
}

// ConsumeLimit 消费额度 (核心功能)
func (s *limitService) ConsumeLimit(ctx context.Context, input *ConsumeLimitInput) error {
	// Check limit first
	checkResult, err := s.CheckLimit(ctx, input.MerchantID, input.Amount)
	if err != nil {
		return fmt.Errorf("check limit failed: %w", err)
	}

	if !checkResult.Allowed {
		// Log failed consumption
		log := &model.LimitUsageLog{
			MerchantID:    input.MerchantID,
			PaymentNo:     input.PaymentNo,
			OrderNo:       input.OrderNo,
			ActionType:    model.ActionTypeConsume,
			Amount:        input.Amount,
			Currency:      input.Currency,
			Success:       false,
			FailureReason: checkResult.Reason,
		}
		s.repo.CreateUsageLog(ctx, log)

		return fmt.Errorf("limit check failed: %s", checkResult.Reason)
	}

	// Get current usage for logging
	limit, _ := s.repo.GetLimitByMerchantID(ctx, input.MerchantID)

	// Consume limit (atomic operation)
	if err := s.repo.UpdateUsage(ctx, input.MerchantID, input.Amount, input.Amount); err != nil {
		return fmt.Errorf("update usage failed: %w", err)
	}

	// Log successful consumption
	log := &model.LimitUsageLog{
		MerchantID:        input.MerchantID,
		PaymentNo:         input.PaymentNo,
		OrderNo:           input.OrderNo,
		ActionType:        model.ActionTypeConsume,
		Amount:            input.Amount,
		Currency:          input.Currency,
		DailyUsedBefore:   limit.DailyUsed,
		DailyUsedAfter:    limit.DailyUsed + input.Amount,
		MonthlyUsedBefore: limit.MonthlyUsed,
		MonthlyUsedAfter:  limit.MonthlyUsed + input.Amount,
		Success:           true,
	}
	s.repo.CreateUsageLog(ctx, log)

	return nil
}

// ReleaseLimit 释放额度 (核心功能)
func (s *limitService) ReleaseLimit(ctx context.Context, input *ReleaseLimitInput) error {
	// Get current usage for logging
	limit, err := s.repo.GetLimitByMerchantID(ctx, input.MerchantID)
	if err != nil {
		return fmt.Errorf("get limit failed: %w", err)
	}
	if limit == nil {
		return fmt.Errorf("merchant limit not found")
	}

	// Release limit (atomic operation, negative delta)
	if err := s.repo.UpdateUsage(ctx, input.MerchantID, -input.Amount, -input.Amount); err != nil {
		return fmt.Errorf("update usage failed: %w", err)
	}

	// Log release
	log := &model.LimitUsageLog{
		MerchantID:        input.MerchantID,
		PaymentNo:         input.PaymentNo,
		OrderNo:           input.OrderNo,
		ActionType:        model.ActionTypeRelease,
		Amount:            input.Amount,
		Currency:          input.Currency,
		DailyUsedBefore:   limit.DailyUsed,
		DailyUsedAfter:    limit.DailyUsed - input.Amount,
		MonthlyUsedBefore: limit.MonthlyUsed,
		MonthlyUsedAfter:  limit.MonthlyUsed - input.Amount,
		Success:           true,
		FailureReason:     input.Reason,
	}
	s.repo.CreateUsageLog(ctx, log)

	return nil
}

// GetUsageHistory 获取使用历史
func (s *limitService) GetUsageHistory(ctx context.Context, merchantID uuid.UUID, startTime, endTime *time.Time, page, pageSize int) (*UsageHistoryResult, error) {
	logs, total, err := s.repo.ListUsageLogs(ctx, merchantID, startTime, endTime, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list usage logs failed: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &UsageHistoryResult{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStatistics 获取统计信息
func (s *limitService) GetStatistics(ctx context.Context, merchantID uuid.UUID) (*repository.MerchantStatistics, error) {
	stats, err := s.repo.GetMerchantStatistics(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("get statistics failed: %w", err)
	}
	return stats, nil
}
