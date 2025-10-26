package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-quota-service/internal/model"
	"payment-platform/merchant-quota-service/internal/repository"
)

// UsageLogService 配额使用日志服务接口
type UsageLogService interface {
	// 获取日志详情
	GetLogByID(ctx context.Context, logID uuid.UUID) (*model.QuotaUsageLog, error)

	// 查询订单的配额操作日志
	GetLogsByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) ([]*model.QuotaUsageLog, error)

	// 查询商户的使用日志
	ListMerchantLogs(ctx context.Context, input *ListMerchantLogsInput) (*UsageLogListOutput, error)

	// 查询时间范围内的日志（审计用）
	ListLogsByTimeRange(ctx context.Context, input *ListLogsByTimeRangeInput) (*UsageLogListOutput, error)

	// 统计配额操作次数
	GetActionStatistics(ctx context.Context, merchantID uuid.UUID, actionType string, startTime, endTime time.Time) (int64, error)
}

type usageLogService struct {
	usageLogRepo repository.UsageLogRepository
}

// NewUsageLogService 创建使用日志服务实例
func NewUsageLogService(usageLogRepo repository.UsageLogRepository) UsageLogService {
	return &usageLogService{
		usageLogRepo: usageLogRepo,
	}
}

// ListMerchantLogsInput 查询商户日志输入
type ListMerchantLogsInput struct {
	MerchantID uuid.UUID  `json:"merchant_id" binding:"required"`
	ActionType string     `json:"action_type"` // consume, release, reset, adjust
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

// ListLogsByTimeRangeInput 按时间范围查询日志输入
type ListLogsByTimeRangeInput struct {
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}

// UsageLogListOutput 使用日志列表输出
type UsageLogListOutput struct {
	Logs       []*model.QuotaUsageLog `json:"logs"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

func (s *usageLogService) GetLogByID(ctx context.Context, logID uuid.UUID) (*model.QuotaUsageLog, error) {
	log, err := s.usageLogRepo.GetByID(ctx, logID)
	if err != nil {
		return nil, fmt.Errorf("查询日志失败: %w", err)
	}
	if log == nil {
		return nil, fmt.Errorf("日志不存在")
	}
	return log, nil
}

func (s *usageLogService) GetLogsByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) ([]*model.QuotaUsageLog, error) {
	logs, err := s.usageLogRepo.GetByOrderNo(ctx, merchantID, orderNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单日志失败: %w", err)
	}
	return logs, nil
}

func (s *usageLogService) ListMerchantLogs(ctx context.Context, input *ListMerchantLogsInput) (*UsageLogListOutput, error) {
	// 参数校验
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	// 时间范围校验
	if input.StartTime != nil && input.EndTime != nil {
		if input.EndTime.Before(*input.StartTime) {
			return nil, fmt.Errorf("结束时间不能早于开始时间")
		}
	}

	offset := (input.Page - 1) * input.PageSize
	logs, total, err := s.usageLogRepo.ListByMerchant(
		ctx,
		input.MerchantID,
		input.ActionType,
		input.StartTime,
		input.EndTime,
		offset,
		input.PageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("查询商户日志失败: %w", err)
	}

	totalPages := int(total) / input.PageSize
	if int(total)%input.PageSize > 0 {
		totalPages++
	}

	return &UsageLogListOutput{
		Logs:       logs,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *usageLogService) ListLogsByTimeRange(ctx context.Context, input *ListLogsByTimeRangeInput) (*UsageLogListOutput, error) {
	// 参数校验
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	if input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("结束时间不能早于开始时间")
	}

	// 限制最大时间范围为90天
	if input.EndTime.Sub(input.StartTime) > 90*24*time.Hour {
		return nil, fmt.Errorf("查询时间范围不能超过90天")
	}

	offset := (input.Page - 1) * input.PageSize
	logs, total, err := s.usageLogRepo.ListByTimeRange(
		ctx,
		input.StartTime,
		input.EndTime,
		offset,
		input.PageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("查询时间范围日志失败: %w", err)
	}

	totalPages := int(total) / input.PageSize
	if int(total)%input.PageSize > 0 {
		totalPages++
	}

	return &UsageLogListOutput{
		Logs:       logs,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *usageLogService) GetActionStatistics(ctx context.Context, merchantID uuid.UUID, actionType string, startTime, endTime time.Time) (int64, error) {
	if endTime.Before(startTime) {
		return 0, fmt.Errorf("结束时间不能早于开始时间")
	}

	count, err := s.usageLogRepo.CountByAction(ctx, merchantID, actionType, startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("统计配额操作失败: %w", err)
	}

	return count, nil
}
