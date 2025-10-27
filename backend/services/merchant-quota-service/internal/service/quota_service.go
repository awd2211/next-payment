package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"payment-platform/merchant-quota-service/internal/client"
	"payment-platform/merchant-quota-service/internal/model"
	"payment-platform/merchant-quota-service/internal/repository"
	"github.com/payment-platform/pkg/logger"
)

// QuotaService 配额服务接口
type QuotaService interface {
	// 初始化商户配额
	InitializeQuota(ctx context.Context, input *InitializeQuotaInput) (*model.MerchantQuota, error)

	// 消耗配额（交易时调用）
	ConsumeQuota(ctx context.Context, input *ConsumeQuotaInput) (*QuotaOperationResult, error)

	// 释放配额（退款时调用）
	ReleaseQuota(ctx context.Context, input *ReleaseQuotaInput) (*QuotaOperationResult, error)

	// 调整配额（管理员操作）
	AdjustQuota(ctx context.Context, input *AdjustQuotaInput) (*model.MerchantQuota, error)

	// 重置日配额（定时任务）
	ResetDailyQuotas(ctx context.Context) error

	// 重置月配额（定时任务）
	ResetMonthlyQuotas(ctx context.Context) error

	// 暂停/恢复商户配额
	SuspendQuota(ctx context.Context, merchantID uuid.UUID, currency string) error
	ResumeQuota(ctx context.Context, merchantID uuid.UUID, currency string) error

	// 查询配额
	GetQuota(ctx context.Context, merchantID uuid.UUID, currency string) (*model.MerchantQuota, error)
	ListQuotas(ctx context.Context, merchantID *uuid.UUID, currency string, isSuspended *bool, page, pageSize int) (*QuotaListOutput, error)
}

type quotaService struct {
	quotaRepo    repository.QuotaRepository
	usageLogRepo repository.UsageLogRepository
	policyClient client.PolicyClient
}

// NewQuotaService 创建配额服务实例
func NewQuotaService(
	quotaRepo repository.QuotaRepository,
	usageLogRepo repository.UsageLogRepository,
	policyClient client.PolicyClient,
) QuotaService {
	return &quotaService{
		quotaRepo:    quotaRepo,
		usageLogRepo: usageLogRepo,
		policyClient: policyClient,
	}
}

// InitializeQuotaInput 初始化配额输入
type InitializeQuotaInput struct {
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
	Currency   string    `json:"currency" binding:"required"`
}

// ConsumeQuotaInput 消耗配额输入
type ConsumeQuotaInput struct {
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
	Currency   string    `json:"currency" binding:"required"`
	Amount     int64     `json:"amount" binding:"required,min=1"`
	OrderNo    string    `json:"order_no" binding:"required"`
}

// ReleaseQuotaInput 释放配额输入
type ReleaseQuotaInput struct {
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
	Currency   string    `json:"currency" binding:"required"`
	Amount     int64     `json:"amount" binding:"required,min=1"`
	OrderNo    string    `json:"order_no" binding:"required"`
}

// AdjustQuotaInput 调整配额输入
type AdjustQuotaInput struct {
	MerchantID     uuid.UUID `json:"merchant_id" binding:"required"`
	Currency       string    `json:"currency" binding:"required"`
	DailyAdjust    int64     `json:"daily_adjust"`    // 可为负数
	MonthlyAdjust  int64     `json:"monthly_adjust"`  // 可为负数
	YearlyAdjust   int64     `json:"yearly_adjust"`   // 可为负数
	AdjustedBy     uuid.UUID `json:"adjusted_by" binding:"required"`
	AdjustReason   string    `json:"adjust_reason" binding:"required"`
}

// QuotaOperationResult 配额操作结果
type QuotaOperationResult struct {
	Success       bool                 `json:"success"`
	Message       string               `json:"message"`
	Quota         *model.MerchantQuota `json:"quota"`
	UsageLog      *model.QuotaUsageLog `json:"usage_log"`
}

// QuotaListOutput 配额列表输出
type QuotaListOutput struct {
	Quotas     []*model.MerchantQuota `json:"quotas"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

func (s *quotaService) InitializeQuota(ctx context.Context, input *InitializeQuotaInput) (*model.MerchantQuota, error) {
	// 检查配额是否已存在
	existing, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询配额失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("配额已存在")
	}

	// 调用 policy-service 获取商户的限额策略
	limitPolicy, err := s.policyClient.GetEffectiveLimitPolicy(ctx, input.MerchantID, "all", input.Currency)
	if err != nil {
		logger.Warn("获取限额策略失败，使用默认值",
			zap.String("merchant_id", input.MerchantID.String()),
			zap.Error(err))
		// 继续使用默认策略
	}

	// 创建配额记录
	now := time.Now()
	quota := &model.MerchantQuota{
		MerchantID:     input.MerchantID,
		Currency:       input.Currency,
		DailyUsed:      0,
		MonthlyUsed:    0,
		YearlyUsed:     0,
		PendingAmount:  0,
		DailyResetAt:   now,
		MonthlyResetAt: now,
		IsSuspended:    false,
		Version:        1,
	}

	// 如果获取到策略,记录在日志中
	if limitPolicy != nil {
		logger.Info("已应用限额策略",
			zap.String("merchant_id", input.MerchantID.String()),
			zap.Int64("daily_limit", limitPolicy.DailyLimit),
			zap.Int64("monthly_limit", limitPolicy.MonthlyLimit))
	}

	if err := s.quotaRepo.Create(ctx, quota); err != nil {
		return nil, fmt.Errorf("创建配额失败: %w", err)
	}

	logger.Info("初始化商户配额成功",
		zap.String("merchant_id", input.MerchantID.String()),
		zap.String("currency", input.Currency),
	)

	return quota, nil
}

func (s *quotaService) ConsumeQuota(ctx context.Context, input *ConsumeQuotaInput) (*QuotaOperationResult, error) {
	// 1. 获取当前配额
	quota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询配额失败: %w", err)
	}
	if quota == nil {
		return nil, fmt.Errorf("配额不存在，请先初始化")
	}

	// 2. 检查配额是否暂停
	if quota.IsSuspended {
		return &QuotaOperationResult{
			Success: false,
			Message: "配额已暂停",
			Quota:   quota,
		}, nil
	}

	// 3. 调用 policy-service 检查限额
	limitCheckResult, err := s.policyClient.CheckLimit(ctx, input.MerchantID, "all", input.Currency, input.Amount, quota.DailyUsed, quota.MonthlyUsed)
	if err != nil {
		logger.Warn("检查限额策略失败，允许交易继续",
			zap.String("merchant_id", input.MerchantID.String()),
			zap.Error(err))
		// 降级策略：如果无法获取限额策略，允许交易继续
	} else if !limitCheckResult.IsAllowed {
		return &QuotaOperationResult{
			Success: false,
			Message: limitCheckResult.RejectionReason,
			Quota:   quota,
		}, nil
	}

	// 4. 保存消耗前的快照
	dailyBefore := quota.DailyUsed
	monthlyBefore := quota.MonthlyUsed
	// yearlyBefore := quota.YearlyUsed  // 暂未使用

	// 5. 执行配额消耗（使用乐观锁）
	if err := s.quotaRepo.ConsumeQuota(ctx, input.MerchantID, input.Currency, input.Amount, input.OrderNo); err != nil {
		return nil, fmt.Errorf("消耗配额失败: %w", err)
	}

	// 6. 重新查询配额获取最新值
	updatedQuota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询更新后配额失败: %w", err)
	}

	// 7. 记录使用日志
	usageLog := &model.QuotaUsageLog{
		MerchantID:        input.MerchantID,
		OrderNo:           input.OrderNo,
		Currency:          input.Currency,
		Amount:            input.Amount,
		ActionType:        "consume",
		DailyUsedBefore:   dailyBefore,
		DailyUsedAfter:    updatedQuota.DailyUsed,
		MonthlyUsedBefore: monthlyBefore,
		MonthlyUsedAfter:  updatedQuota.MonthlyUsed,
	}

	if err := s.usageLogRepo.Create(ctx, usageLog); err != nil {
		logger.Error("记录配额使用日志失败", zap.Error(err))
	}

	logger.Info("消耗配额成功",
		zap.String("merchant_id", input.MerchantID.String()),
		zap.String("currency", input.Currency),
		zap.Int64("amount", input.Amount),
		zap.String("order_no", input.OrderNo),
	)

	return &QuotaOperationResult{
		Success:  true,
		Message:  "配额消耗成功",
		Quota:    updatedQuota,
		UsageLog: usageLog,
	}, nil
}

func (s *quotaService) ReleaseQuota(ctx context.Context, input *ReleaseQuotaInput) (*QuotaOperationResult, error) {
	// 1. 获取当前配额
	quota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询配额失败: %w", err)
	}
	if quota == nil {
		return nil, fmt.Errorf("配额不存在")
	}

	// 2. 保存释放前的快照
	// pendingBefore := quota.PendingAmount  // 暂未使用

	// 3. 执行配额释放
	if err := s.quotaRepo.ReleaseQuota(ctx, input.MerchantID, input.Currency, input.Amount, input.OrderNo); err != nil {
		return nil, fmt.Errorf("释放配额失败: %w", err)
	}

	// 4. 重新查询配额
	updatedQuota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询更新后配额失败: %w", err)
	}

	// 5. 记录使用日志
	usageLog := &model.QuotaUsageLog{
		MerchantID:        input.MerchantID,
		OrderNo:           input.OrderNo,
		Currency:          input.Currency,
		Amount:            input.Amount,
		ActionType:        "release",
		DailyUsedBefore:   quota.DailyUsed,
		DailyUsedAfter:    quota.DailyUsed, // 释放不影响已使用量
		MonthlyUsedBefore: quota.MonthlyUsed,
		MonthlyUsedAfter:  quota.MonthlyUsed,
	}

	if err := s.usageLogRepo.Create(ctx, usageLog); err != nil {
		logger.Error("记录配额释放日志失败", zap.Error(err))
	}

	logger.Info("释放配额成功",
		zap.String("merchant_id", input.MerchantID.String()),
		zap.String("currency", input.Currency),
		zap.Int64("amount", input.Amount),
		zap.String("order_no", input.OrderNo),
	)

	return &QuotaOperationResult{
		Success:  true,
		Message:  "配额释放成功",
		Quota:    updatedQuota,
		UsageLog: usageLog,
	}, nil
}

func (s *quotaService) AdjustQuota(ctx context.Context, input *AdjustQuotaInput) (*model.MerchantQuota, error) {
	// 获取当前配额
	quota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询配额失败: %w", err)
	}
	if quota == nil {
		return nil, fmt.Errorf("配额不存在")
	}

	// 保存调整前的快照
	dailyBefore := quota.DailyUsed
	monthlyBefore := quota.MonthlyUsed
	// yearlyBefore := quota.YearlyUsed  // 暂未使用

	// 执行调整
	if err := s.quotaRepo.AdjustQuota(ctx, input.MerchantID, input.Currency, input.DailyAdjust, input.MonthlyAdjust, input.YearlyAdjust); err != nil {
		return nil, fmt.Errorf("调整配额失败: %w", err)
	}

	// 重新查询配额
	updatedQuota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, input.MerchantID, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("查询更新后配额失败: %w", err)
	}

	// 记录使用日志
	usageLog := &model.QuotaUsageLog{
		MerchantID:        input.MerchantID,
		OrderNo:           "", // 管理员操作无订单号
		Currency:          input.Currency,
		Amount:            input.DailyAdjust, // 记录日调整量
		ActionType:        "adjust",
		DailyUsedBefore:   dailyBefore,
		DailyUsedAfter:    updatedQuota.DailyUsed,
		MonthlyUsedBefore: monthlyBefore,
		MonthlyUsedAfter:  updatedQuota.MonthlyUsed,
		OperatorID:        &input.AdjustedBy,
		Remarks:           input.AdjustReason,
	}

	if err := s.usageLogRepo.Create(ctx, usageLog); err != nil {
		logger.Error("记录配额调整日志失败", zap.Error(err))
	}

	logger.Info("调整配额成功",
		zap.String("merchant_id", input.MerchantID.String()),
		zap.String("currency", input.Currency),
		zap.Int64("daily_adjust", input.DailyAdjust),
		zap.Int64("monthly_adjust", input.MonthlyAdjust),
		zap.String("adjusted_by", input.AdjustedBy.String()),
	)

	return updatedQuota, nil
}

func (s *quotaService) ResetDailyQuotas(ctx context.Context) error {
	logger.Info("开始重置所有商户的日配额...")
	if err := s.quotaRepo.ResetDailyQuotas(ctx); err != nil {
		logger.Error("重置日配额失败", zap.Error(err))
		return fmt.Errorf("重置日配额失败: %w", err)
	}
	logger.Info("日配额重置完成")
	return nil
}

func (s *quotaService) ResetMonthlyQuotas(ctx context.Context) error {
	logger.Info("开始重置所有商户的月配额...")
	if err := s.quotaRepo.ResetMonthlyQuotas(ctx); err != nil {
		logger.Error("重置月配额失败", zap.Error(err))
		return fmt.Errorf("重置月配额失败: %w", err)
	}
	logger.Info("月配额重置完成")
	return nil
}

func (s *quotaService) SuspendQuota(ctx context.Context, merchantID uuid.UUID, currency string) error {
	if err := s.quotaRepo.SuspendQuota(ctx, merchantID, currency); err != nil {
		return fmt.Errorf("暂停配额失败: %w", err)
	}
	logger.Info("暂停配额成功", zap.String("merchant_id", merchantID.String()), zap.String("currency", currency))
	return nil
}

func (s *quotaService) ResumeQuota(ctx context.Context, merchantID uuid.UUID, currency string) error {
	if err := s.quotaRepo.ResumeQuota(ctx, merchantID, currency); err != nil {
		return fmt.Errorf("恢复配额失败: %w", err)
	}
	logger.Info("恢复配额成功", zap.String("merchant_id", merchantID.String()), zap.String("currency", currency))
	return nil
}

func (s *quotaService) GetQuota(ctx context.Context, merchantID uuid.UUID, currency string) (*model.MerchantQuota, error) {
	quota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, merchantID, currency)
	if err != nil {
		return nil, fmt.Errorf("查询配额失败: %w", err)
	}
	if quota == nil {
		return nil, fmt.Errorf("配额不存在")
	}
	return quota, nil
}

func (s *quotaService) ListQuotas(ctx context.Context, merchantID *uuid.UUID, currency string, isSuspended *bool, page, pageSize int) (*QuotaListOutput, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	quotas, total, err := s.quotaRepo.List(ctx, merchantID, currency, isSuspended, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询配额列表失败: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &QuotaListOutput{
		Quotas:     quotas,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
