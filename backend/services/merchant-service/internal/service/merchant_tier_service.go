package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// MerchantTierService 商户等级服务接口
type MerchantTierService interface {
	// 等级配置管理
	GetTierConfig(ctx context.Context, tier model.MerchantTier) (*model.MerchantTierConfig, error)
	GetAllTierConfigs(ctx context.Context) ([]*model.MerchantTierConfig, error)
	UpdateTierConfig(ctx context.Context, config *model.MerchantTierConfig) error

	// 商户等级操作
	GetMerchantTier(ctx context.Context, merchantID uuid.UUID) (model.MerchantTier, error)
	UpgradeMerchantTier(ctx context.Context, merchantID uuid.UUID, newTier model.MerchantTier, operator string, reason string) error
	DowngradeMerchantTier(ctx context.Context, merchantID uuid.UUID, newTier model.MerchantTier, operator string, reason string) error

	// 等级权限检查
	CheckTierPermission(ctx context.Context, merchantID uuid.UUID, feature string) (bool, error)
	CalculateMerchantFee(ctx context.Context, merchantID uuid.UUID, amount int64) (int64, error)

	// 等级推荐
	RecommendTierUpgrade(ctx context.Context, merchantID uuid.UUID) (*model.MerchantTier, string, error)

	// 初始化默认等级配置
	InitializeDefaultTiers(ctx context.Context) error
}

// merchantTierService 服务实现
type merchantTierService struct {
	db               *gorm.DB
	tierRepo         repository.MerchantTierRepository
	merchantRepo     repository.MerchantRepository
	merchantLimitSvc MerchantLimitService
	redisClient      *redis.Client
}

// NewMerchantTierService 创建商户等级服务
func NewMerchantTierService(
	db *gorm.DB,
	tierRepo repository.MerchantTierRepository,
	merchantRepo repository.MerchantRepository,
	merchantLimitSvc MerchantLimitService,
	redisClient *redis.Client,
) MerchantTierService {
	return &merchantTierService{
		db:               db,
		tierRepo:         tierRepo,
		merchantRepo:     merchantRepo,
		merchantLimitSvc: merchantLimitSvc,
		redisClient:      redisClient,
	}
}

// GetTierConfig 获取等级配置
func (s *merchantTierService) GetTierConfig(ctx context.Context, tier model.MerchantTier) (*model.MerchantTierConfig, error) {
	config, err := s.tierRepo.GetByTier(ctx, tier)
	if err != nil {
		return nil, fmt.Errorf("获取等级配置失败: %w", err)
	}

	if config == nil {
		// 如果数据库没有配置，返回默认配置
		return model.GetDefaultTierConfig(tier), nil
	}

	return config, nil
}

// GetAllTierConfigs 获取所有等级配置
func (s *merchantTierService) GetAllTierConfigs(ctx context.Context) ([]*model.MerchantTierConfig, error) {
	return s.tierRepo.GetAll(ctx)
}

// UpdateTierConfig 更新等级配置
func (s *merchantTierService) UpdateTierConfig(ctx context.Context, config *model.MerchantTierConfig) error {
	// 检查配置是否存在
	existing, err := s.tierRepo.GetByTier(ctx, config.Tier)
	if err != nil {
		return fmt.Errorf("检查等级配置失败: %w", err)
	}

	if existing == nil {
		return s.tierRepo.Create(ctx, config)
	}

	config.ID = existing.ID
	return s.tierRepo.Update(ctx, config)
}

// GetMerchantTier 获取商户等级
func (s *merchantTierService) GetMerchantTier(ctx context.Context, merchantID uuid.UUID) (model.MerchantTier, error) {
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return "", fmt.Errorf("获取商户信息失败: %w", err)
	}

	if merchant == nil {
		return "", fmt.Errorf("商户不存在")
	}

	// 如果商户没有设置等级，默认为入门版
	if merchant.Tier == "" {
		return model.TierStarter, nil
	}

	return merchant.Tier, nil
}

// UpgradeMerchantTier 升级商户等级
func (s *merchantTierService) UpgradeMerchantTier(ctx context.Context, merchantID uuid.UUID, newTier model.MerchantTier, operator string, reason string) error {
	// 获取当前等级
	currentTier, err := s.GetMerchantTier(ctx, merchantID)
	if err != nil {
		return err
	}

	// 获取配置验证是否可升级
	currentConfig, err := s.GetTierConfig(ctx, currentTier)
	if err != nil {
		return err
	}

	if !currentConfig.CanUpgradeTo(newTier) {
		return fmt.Errorf("无法从 %s 升级到 %s", currentTier, newTier)
	}

	// 获取新等级配置
	newConfig, err := s.GetTierConfig(ctx, newTier)
	if err != nil {
		return err
	}

	// 开始事务
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新商户等级
		err := tx.Model(&model.Merchant{}).
			Where("id = ?", merchantID).
			Updates(map[string]interface{}{
				"tier":       newTier,
				"updated_at": time.Now(),
			}).Error

		if err != nil {
			return fmt.Errorf("更新商户等级失败: %w", err)
		}

		// 更新商户限额配置
		err = tx.Model(&model.MerchantLimit{}).
			Where("merchant_id = ?", merchantID).
			Updates(map[string]interface{}{
				"daily_limit":   newConfig.DailyLimit,
				"monthly_limit": newConfig.MonthlyLimit,
				"single_limit":  newConfig.SingleLimit,
				"updated_at":    time.Now(),
			}).Error

		if err != nil {
			return fmt.Errorf("更新商户限额失败: %w", err)
		}

		logger.Info("商户等级升级成功",
			zap.String("merchant_id", merchantID.String()),
			zap.String("from_tier", string(currentTier)),
			zap.String("to_tier", string(newTier)),
			zap.String("operator", operator),
			zap.String("reason", reason))

		return nil
	})
}

// DowngradeMerchantTier 降级商户等级
func (s *merchantTierService) DowngradeMerchantTier(ctx context.Context, merchantID uuid.UUID, newTier model.MerchantTier, operator string, reason string) error {
	currentTier, err := s.GetMerchantTier(ctx, merchantID)
	if err != nil {
		return err
	}

	// 检查是否为降级
	tierLevels := map[model.MerchantTier]int{
		model.TierStarter:    1,
		model.TierBusiness:   2,
		model.TierEnterprise: 3,
		model.TierPremium:    4,
	}

	if tierLevels[newTier] >= tierLevels[currentTier] {
		return fmt.Errorf("目标等级 %s 不低于当前等级 %s", newTier, currentTier)
	}

	// 获取新等级配置
	newConfig, err := s.GetTierConfig(ctx, newTier)
	if err != nil {
		return err
	}

	// 开始事务
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新商户等级
		err := tx.Model(&model.Merchant{}).
			Where("id = ?", merchantID).
			Updates(map[string]interface{}{
				"tier":       newTier,
				"updated_at": time.Now(),
			}).Error

		if err != nil {
			return fmt.Errorf("更新商户等级失败: %w", err)
		}

		// 更新商户限额配置
		err = tx.Model(&model.MerchantLimit{}).
			Where("merchant_id = ?", merchantID).
			Updates(map[string]interface{}{
				"daily_limit":   newConfig.DailyLimit,
				"monthly_limit": newConfig.MonthlyLimit,
				"single_limit":  newConfig.SingleLimit,
				"updated_at":    time.Now(),
			}).Error

		if err != nil {
			return fmt.Errorf("更新商户限额失败: %w", err)
		}

		logger.Warn("商户等级降级",
			zap.String("merchant_id", merchantID.String()),
			zap.String("from_tier", string(currentTier)),
			zap.String("to_tier", string(newTier)),
			zap.String("operator", operator),
			zap.String("reason", reason))

		return nil
	})
}

// CheckTierPermission 检查等级权限
func (s *merchantTierService) CheckTierPermission(ctx context.Context, merchantID uuid.UUID, feature string) (bool, error) {
	tier, err := s.GetMerchantTier(ctx, merchantID)
	if err != nil {
		return false, err
	}

	config, err := s.GetTierConfig(ctx, tier)
	if err != nil {
		return false, err
	}

	// 根据功能检查权限
	switch feature {
	case "multi_currency":
		return config.EnableMultiCurrency, nil
	case "refund":
		return config.EnableRefund, nil
	case "partial_refund":
		return config.EnablePartialRefund, nil
	case "pre_auth":
		return config.EnablePreAuth, nil
	case "recurring":
		return config.EnableRecurring, nil
	case "split":
		return config.EnableSplit, nil
	case "webhook":
		return config.EnableWebhook, nil
	case "custom_branding":
		return config.CustomBranding, nil
	default:
		return false, fmt.Errorf("未知功能: %s", feature)
	}
}

// CalculateMerchantFee 计算商户手续费
func (s *merchantTierService) CalculateMerchantFee(ctx context.Context, merchantID uuid.UUID, amount int64) (int64, error) {
	tier, err := s.GetMerchantTier(ctx, merchantID)
	if err != nil {
		return 0, err
	}

	config, err := s.GetTierConfig(ctx, tier)
	if err != nil {
		return 0, err
	}

	return config.CalculateFee(amount), nil
}

// RecommendTierUpgrade 推荐等级升级
func (s *merchantTierService) RecommendTierUpgrade(ctx context.Context, merchantID uuid.UUID) (*model.MerchantTier, string, error) {
	// 获取当前等级
	currentTier, err := s.GetMerchantTier(ctx, merchantID)
	if err != nil {
		return nil, "", err
	}

	// 如果已经是最高等级，无需推荐
	if currentTier == model.TierPremium {
		return nil, "已经是最高等级", nil
	}

	// 获取商户交易数据（最近30天）
	limit, err := s.merchantLimitSvc.GetMerchantLimit(ctx, merchantID)
	if err != nil {
		return nil, "", err
	}

	// 获取当前等级配置
	currentConfig, err := s.GetTierConfig(ctx, currentTier)
	if err != nil {
		return nil, "", err
	}

	// 判断是否需要升级
	// 规则1: 如果月使用量超过当前限额的80%，推荐升级
	if limit.UsedMonth > currentConfig.MonthlyLimit*80/100 {
		nextTier := s.getNextTier(currentTier)
		return &nextTier, fmt.Sprintf("月交易量已达限额的%.1f%%，建议升级以获得更高限额和更低费率",
			float64(limit.UsedMonth)/float64(currentConfig.MonthlyLimit)*100), nil
	}

	// 规则2: 如果日使用量频繁接近限额（超过70%），推荐升级
	if limit.UsedToday > currentConfig.DailyLimit*70/100 {
		nextTier := s.getNextTier(currentTier)
		return &nextTier, fmt.Sprintf("日交易量已达限额的%.1f%%，建议升级以避免限额影响业务",
			float64(limit.UsedToday)/float64(currentConfig.DailyLimit)*100), nil
	}

	return nil, "当前等级适合您的业务规模", nil
}

// getNextTier 获取下一个等级
func (s *merchantTierService) getNextTier(current model.MerchantTier) model.MerchantTier {
	switch current {
	case model.TierStarter:
		return model.TierBusiness
	case model.TierBusiness:
		return model.TierEnterprise
	case model.TierEnterprise:
		return model.TierPremium
	default:
		return model.TierPremium
	}
}

// InitializeDefaultTiers 初始化默认等级配置
func (s *merchantTierService) InitializeDefaultTiers(ctx context.Context) error {
	tiers := []model.MerchantTier{
		model.TierStarter,
		model.TierBusiness,
		model.TierEnterprise,
		model.TierPremium,
	}

	for _, tier := range tiers {
		// 检查是否已存在
		existing, err := s.tierRepo.GetByTier(ctx, tier)
		if err != nil {
			return err
		}

		if existing != nil {
			logger.Info("等级配置已存在，跳过", zap.String("tier", string(tier)))
			continue
		}

		// 创建默认配置
		config := model.GetDefaultTierConfig(tier)
		err = s.tierRepo.Create(ctx, config)
		if err != nil {
			return fmt.Errorf("创建等级配置失败 %s: %w", tier, err)
		}

		logger.Info("等级配置已创建", zap.String("tier", string(tier)))
	}

	return nil
}
