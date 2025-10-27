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

// AlertService 配额预警服务接口
type AlertService interface {
	// 检查所有商户的配额预警（定时任务）
	CheckQuotaAlerts(ctx context.Context) error

	// 检查单个商户的配额预警
	CheckMerchantQuotaAlert(ctx context.Context, merchantID uuid.UUID, currency string) error

	// 标记预警为已处理
	ResolveAlert(ctx context.Context, alertID uuid.UUID, resolvedBy uuid.UUID) error

	// 获取商户的活跃预警
	GetActiveAlerts(ctx context.Context, merchantID uuid.UUID, alertLevel string) ([]*model.QuotaAlert, error)

	// 清理过期预警
	CleanupExpiredAlerts(ctx context.Context) error

	// 列表查询
	ListAlerts(ctx context.Context, merchantID *uuid.UUID, alertLevel, alertType string, isResolved *bool, page, pageSize int) (*AlertListOutput, error)
}

type alertService struct {
	alertRepo          repository.AlertRepository
	quotaRepo          repository.QuotaRepository
	policyClient       client.PolicyClient
	notificationClient *client.NotificationClient
}

// NewAlertService 创建预警服务实例
func NewAlertService(
	alertRepo repository.AlertRepository,
	quotaRepo repository.QuotaRepository,
	policyClient client.PolicyClient,
	notificationClient *client.NotificationClient,
) AlertService {
	return &alertService{
		alertRepo:          alertRepo,
		quotaRepo:          quotaRepo,
		policyClient:       policyClient,
		notificationClient: notificationClient,
	}
}

// AlertListOutput 预警列表输出
type AlertListOutput struct {
	Alerts     []*model.QuotaAlert `json:"alerts"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// 预警阈值配置
const (
	DailyWarningThreshold  = 0.8  // 日配额80%预警
	DailyFull              = 1.0  // 日配额100%预警
	MonthlyWarningThreshold = 0.8  // 月配额80%预警
	MonthlyFull             = 1.0  // 月配额100%预警
)

func (s *alertService) CheckQuotaAlerts(ctx context.Context) error {
	logger.Info("开始检查所有商户的配额预警...")

	// 查询所有未暂停的配额
	quotas, _, err := s.quotaRepo.List(ctx, nil, "", boolPtr(false), 0, 10000)
	if err != nil {
		return fmt.Errorf("查询配额列表失败: %w", err)
	}

	totalChecked := 0
	totalAlerts := 0

	for _, quota := range quotas {
		// ✅ FIXED: 调用 policy-service 获取限额策略
		limitPolicy, err := s.policyClient.GetEffectiveLimitPolicy(ctx, quota.MerchantID, "all", quota.Currency)
		if err != nil {
			logger.Error("获取限额策略失败，使用默认限额",
				zap.String("merchant_id", quota.MerchantID.String()),
				zap.Error(err))
			// 使用默认限额作为降级方案
			limitPolicy = &client.LimitPolicy{
				DailyLimit:   500000000,  // $5,000,000
				MonthlyLimit: 10000000000, // $100,000,000
			}
		}

		dailyLimit := limitPolicy.DailyLimit
		monthlyLimit := limitPolicy.MonthlyLimit

		alerts := s.checkQuotaThresholds(quota, dailyLimit, monthlyLimit)
		for _, alert := range alerts {
			// 检查是否已存在相同预警（防止重复发送）
			exists, err := s.alertRepo.ExistsByMerchantAndType(ctx, quota.MerchantID, alert.AlertType, time.Now().Add(-24*time.Hour))
			if err != nil {
				logger.Error("检查预警是否存在失败", zap.Error(err))
				continue
			}
			if exists {
				continue // 24小时内已发送过相同预警，跳过
			}

			// 创建预警
			if err := s.alertRepo.Create(ctx, alert); err != nil {
				logger.Error("创建预警失败", zap.Error(err))
				continue
			}

			// ✅ FIXED: 发送预警通知
			if s.notificationClient != nil {
				notifReq := &client.SendQuotaAlertRequest{
					MerchantID: alert.MerchantID,
					Type:       getNotificationType(alert.AlertLevel),
					Title:      "配额预警通知",
					Content:    alert.Message,
					Data: map[string]interface{}{
						"alert_id":      alert.ID.String(),
						"alert_type":    alert.AlertType,
						"alert_level":   alert.AlertLevel,
						"currency":      alert.Currency,
						"current_used":  alert.CurrentUsed,
						"limit":         alert.Limit,
						"usage_percent": alert.UsagePercent,
					},
					Priority: getPriority(alert.AlertLevel),
				}

				if err := s.notificationClient.SendQuotaAlert(ctx, notifReq); err != nil {
					logger.Error("发送配额预警通知失败",
						zap.String("alert_id", alert.ID.String()),
						zap.Error(err))
					// 不阻止流程继续，通知发送失败不影响预警创建
				}
			}

			logger.Warn("配额预警",
				zap.String("merchant_id", alert.MerchantID.String()),
				zap.String("alert_type", alert.AlertType),
				zap.String("alert_level", alert.AlertLevel),
				zap.Float64("usage_percent", alert.UsagePercent),
			)

			totalAlerts++
		}

		totalChecked++
	}

	logger.Info("配额预警检查完成",
		zap.Int("total_checked", totalChecked),
		zap.Int("total_alerts", totalAlerts),
	)

	return nil
}

func (s *alertService) CheckMerchantQuotaAlert(ctx context.Context, merchantID uuid.UUID, currency string) error {
	quota, err := s.quotaRepo.GetByMerchantAndCurrency(ctx, merchantID, currency)
	if err != nil {
		return fmt.Errorf("查询配额失败: %w", err)
	}
	if quota == nil {
		return fmt.Errorf("配额不存在")
	}

	// ✅ FIXED: 调用 policy-service 获取限额策略
	limitPolicy, err := s.policyClient.GetEffectiveLimitPolicy(ctx, merchantID, "all", currency)
	if err != nil {
		logger.Error("获取限额策略失败，使用默认限额",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		// 使用默认限额作为降级方案
		limitPolicy = &client.LimitPolicy{
			DailyLimit:   500000000,  // $5,000,000
			MonthlyLimit: 10000000000, // $100,000,000
		}
	}

	dailyLimit := limitPolicy.DailyLimit
	monthlyLimit := limitPolicy.MonthlyLimit

	alerts := s.checkQuotaThresholds(quota, dailyLimit, monthlyLimit)
	for _, alert := range alerts {
		if err := s.alertRepo.Create(ctx, alert); err != nil {
			return fmt.Errorf("创建预警失败: %w", err)
		}
	}

	return nil
}

// checkQuotaThresholds 检查配额阈值并生成预警
func (s *alertService) checkQuotaThresholds(quota *model.MerchantQuota, dailyLimit, monthlyLimit int64) []*model.QuotaAlert {
	var alerts []*model.QuotaAlert

	// 检查日配额
	if dailyLimit > 0 {
		dailyUsagePercent := float64(quota.DailyUsed) / float64(dailyLimit)

		if dailyUsagePercent >= DailyFull {
			alerts = append(alerts, &model.QuotaAlert{
				MerchantID:   quota.MerchantID,
				Currency:     quota.Currency,
				AlertType:    "daily_100",
				AlertLevel:   "critical",
				CurrentUsed:  quota.DailyUsed,
				Limit:        dailyLimit,
				UsagePercent: dailyUsagePercent * 100,
				Message:      fmt.Sprintf("日配额已用尽: %d/%d (%.2f%%)", quota.DailyUsed, dailyLimit, dailyUsagePercent*100),
				IsResolved:   false,
			})
		} else if dailyUsagePercent >= DailyWarningThreshold {
			alerts = append(alerts, &model.QuotaAlert{
				MerchantID:   quota.MerchantID,
				Currency:     quota.Currency,
				AlertType:    "daily_80",
				AlertLevel:   "warning",
				CurrentUsed:  quota.DailyUsed,
				Limit:        dailyLimit,
				UsagePercent: dailyUsagePercent * 100,
				Message:      fmt.Sprintf("日配额即将用尽: %d/%d (%.2f%%)", quota.DailyUsed, dailyLimit, dailyUsagePercent*100),
				IsResolved:   false,
			})
		}
	}

	// 检查月配额
	if monthlyLimit > 0 {
		monthlyUsagePercent := float64(quota.MonthlyUsed) / float64(monthlyLimit)

		if monthlyUsagePercent >= MonthlyFull {
			alerts = append(alerts, &model.QuotaAlert{
				MerchantID:   quota.MerchantID,
				Currency:     quota.Currency,
				AlertType:    "monthly_100",
				AlertLevel:   "critical",
				CurrentUsed:  quota.MonthlyUsed,
				Limit:        monthlyLimit,
				UsagePercent: monthlyUsagePercent * 100,
				Message:      fmt.Sprintf("月配额已用尽: %d/%d (%.2f%%)", quota.MonthlyUsed, monthlyLimit, monthlyUsagePercent*100),
				IsResolved:   false,
			})
		} else if monthlyUsagePercent >= MonthlyWarningThreshold {
			alerts = append(alerts, &model.QuotaAlert{
				MerchantID:   quota.MerchantID,
				Currency:     quota.Currency,
				AlertType:    "monthly_80",
				AlertLevel:   "warning",
				CurrentUsed:  quota.MonthlyUsed,
				Limit:        monthlyLimit,
				UsagePercent: monthlyUsagePercent * 100,
				Message:      fmt.Sprintf("月配额即将用尽: %d/%d (%.2f%%)", quota.MonthlyUsed, monthlyLimit, monthlyUsagePercent*100),
				IsResolved:   false,
			})
		}
	}

	return alerts
}

func (s *alertService) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolvedBy uuid.UUID) error {
	alert, err := s.alertRepo.GetByID(ctx, alertID)
	if err != nil {
		return fmt.Errorf("查询预警失败: %w", err)
	}
	if alert == nil {
		return fmt.Errorf("预警不存在")
	}

	if alert.IsResolved {
		return fmt.Errorf("预警已处理")
	}

	if err := s.alertRepo.MarkAsResolved(ctx, alertID, resolvedBy); err != nil {
		return fmt.Errorf("标记预警为已处理失败: %w", err)
	}

	logger.Info("预警已处理",
		zap.String("alert_id", alertID.String()),
		zap.String("resolved_by", resolvedBy.String()),
	)

	return nil
}

func (s *alertService) GetActiveAlerts(ctx context.Context, merchantID uuid.UUID, alertLevel string) ([]*model.QuotaAlert, error) {
	alerts, err := s.alertRepo.GetActiveAlerts(ctx, merchantID, alertLevel)
	if err != nil {
		return nil, fmt.Errorf("查询活跃预警失败: %w", err)
	}
	return alerts, nil
}

func (s *alertService) CleanupExpiredAlerts(ctx context.Context) error {
	// 清理24小时前的warning级别预警
	expiredBefore := time.Now().Add(-24 * time.Hour)
	if err := s.alertRepo.CleanupExpiredAlerts(ctx, expiredBefore); err != nil {
		return fmt.Errorf("清理过期预警失败: %w", err)
	}
	logger.Info("清理过期预警完成", zap.Time("expired_before", expiredBefore))
	return nil
}

func (s *alertService) ListAlerts(ctx context.Context, merchantID *uuid.UUID, alertLevel, alertType string, isResolved *bool, page, pageSize int) (*AlertListOutput, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	alerts, total, err := s.alertRepo.List(ctx, merchantID, alertLevel, alertType, isResolved, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询预警列表失败: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &AlertListOutput{
		Alerts:     alerts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

// getNotificationType 根据预警级别返回通知类型
func getNotificationType(alertLevel string) string {
	switch alertLevel {
	case "critical":
		return "quota_critical"
	case "warning":
		return "quota_warning"
	default:
		return "quota_info"
	}
}

// getPriority 根据预警级别返回通知优先级
func getPriority(alertLevel string) string {
	switch alertLevel {
	case "critical":
		return "high"
	case "warning":
		return "medium"
	default:
		return "low"
	}
}
