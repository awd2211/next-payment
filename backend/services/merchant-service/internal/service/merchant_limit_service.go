package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// MerchantLimitService 商户额度服务接口
type MerchantLimitService interface {
	// 初始化商户额度（注册时调用）
	InitializeMerchantLimit(ctx context.Context, merchantID uuid.UUID) error

	// 检查是否可以处理交易
	CheckLimit(ctx context.Context, merchantID uuid.UUID, amount int64) (bool, string, error)

	// 增加使用额度（交易成功后调用）
	IncreaseUsage(ctx context.Context, merchantID uuid.UUID, amount int64) error

	// 减少使用额度（退款时调用）
	DecreaseUsage(ctx context.Context, merchantID uuid.UUID, amount int64) error

	// 获取商户额度信息
	GetMerchantLimit(ctx context.Context, merchantID uuid.UUID) (*model.MerchantLimit, error)

	// 更新额度配置
	UpdateLimitConfig(ctx context.Context, merchantID uuid.UUID, dailyLimit, monthlyLimit, singleLimit int64) error

	// 设置限额状态（风控使用）
	SetLimitStatus(ctx context.Context, merchantID uuid.UUID, isLimited bool, reason string) error
}

// merchantLimitService 商户额度服务实现
type merchantLimitService struct {
	db          *gorm.DB
	limitRepo   repository.MerchantLimitRepository
	redisClient *redis.Client
}

// NewMerchantLimitService 创建商户额度服务
func NewMerchantLimitService(db *gorm.DB, limitRepo repository.MerchantLimitRepository, redisClient *redis.Client) MerchantLimitService {
	return &merchantLimitService{
		db:          db,
		limitRepo:   limitRepo,
		redisClient: redisClient,
	}
}

// InitializeMerchantLimit 初始化商户额度
func (s *merchantLimitService) InitializeMerchantLimit(ctx context.Context, merchantID uuid.UUID) error {
	// 检查是否已存在
	existing, err := s.limitRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("检查商户额度失败: %w", err)
	}

	if existing != nil {
		return nil // 已存在，不重复创建
	}

	// 创建默认额度配置
	limit := &model.MerchantLimit{
		MerchantID:   merchantID,
		DailyLimit:   100000000,  // 100万（分）
		MonthlyLimit: 3000000000, // 3000万（分）
		SingleLimit:  10000000,   // 10万（分）
		UsedToday:    0,
		UsedMonth:    0,
		IsLimited:    false,
		LastResetDay: time.Now(),
	}

	if err := s.limitRepo.Create(ctx, limit); err != nil {
		return fmt.Errorf("创建商户额度失败: %w", err)
	}

	logger.Info("商户额度初始化成功", zap.String("merchant_id", merchantID.String()))
	return nil
}

// CheckLimit 检查是否可以处理交易
func (s *merchantLimitService) CheckLimit(ctx context.Context, merchantID uuid.UUID, amount int64) (bool, string, error) {
	// 1. 从DB获取额度配置
	limit, err := s.limitRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return false, "", fmt.Errorf("获取商户额度失败: %w", err)
	}

	if limit == nil {
		return false, "商户额度未初始化", nil
	}

	// 检查是否被限额
	if limit.IsLimited {
		return false, limit.LimitReason, nil
	}

	// 检查单笔限额
	if amount > limit.SingleLimit {
		return false, fmt.Sprintf("超过单笔限额（限额: %d, 请求: %d）", limit.SingleLimit, amount), nil
	}

	// 2. 从Redis获取实时使用量
	usedToday, err := s.getUsedTodayFromRedis(ctx, merchantID)
	if err != nil {
		logger.Warn("从Redis获取今日使用量失败，降级到DB",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		usedToday = limit.UsedToday
	}

	usedMonth, err := s.getUsedMonthFromRedis(ctx, merchantID)
	if err != nil {
		logger.Warn("从Redis获取本月使用量失败，降级到DB",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		usedMonth = limit.UsedMonth
	}

	// 3. 检查日限额
	if usedToday+amount > limit.DailyLimit {
		return false, fmt.Sprintf("超过日限额（限额: %d, 已用: %d, 请求: %d）",
			limit.DailyLimit, usedToday, amount), nil
	}

	// 4. 检查月限额
	if usedMonth+amount > limit.MonthlyLimit {
		return false, fmt.Sprintf("超过月限额（限额: %d, 已用: %d, 请求: %d）",
			limit.MonthlyLimit, usedMonth, amount), nil
	}

	return true, "", nil
}

// IncreaseUsage 增加使用额度
func (s *merchantLimitService) IncreaseUsage(ctx context.Context, merchantID uuid.UUID, amount int64) error {
	// 使用Redis INCRBY原子操作增加使用量
	today := time.Now().Format("20060102")
	month := time.Now().Format("200601")

	dailyKey := fmt.Sprintf("merchant_limit:daily:%s:%s", merchantID.String(), today)
	monthlyKey := fmt.Sprintf("merchant_limit:monthly:%s:%s", merchantID.String(), month)

	pipe := s.redisClient.Pipeline()
	pipe.IncrBy(ctx, dailyKey, amount)
	pipe.Expire(ctx, dailyKey, 48*time.Hour) // 保留2天
	pipe.IncrBy(ctx, monthlyKey, amount)
	pipe.Expire(ctx, monthlyKey, 32*24*time.Hour) // 保留32天

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Redis增加额度使用量失败",
			zap.String("merchant_id", merchantID.String()),
			zap.Int64("amount", amount),
			zap.Error(err))
		return fmt.Errorf("增加额度使用量失败: %w", err)
	}

	logger.Info("额度使用量已增加",
		zap.String("merchant_id", merchantID.String()),
		zap.Int64("amount", amount))

	// 异步同步到DB
	go s.syncUsageToDB(context.Background(), merchantID)

	return nil
}

// DecreaseUsage 减少使用额度（退款时）
func (s *merchantLimitService) DecreaseUsage(ctx context.Context, merchantID uuid.UUID, amount int64) error {
	today := time.Now().Format("20060102")
	month := time.Now().Format("200601")

	dailyKey := fmt.Sprintf("merchant_limit:daily:%s:%s", merchantID.String(), today)
	monthlyKey := fmt.Sprintf("merchant_limit:monthly:%s:%s", merchantID.String(), month)

	pipe := s.redisClient.Pipeline()
	pipe.DecrBy(ctx, dailyKey, amount)
	pipe.DecrBy(ctx, monthlyKey, amount)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Redis减少额度使用量失败",
			zap.String("merchant_id", merchantID.String()),
			zap.Int64("amount", amount),
			zap.Error(err))
		return fmt.Errorf("减少额度使用量失败: %w", err)
	}

	logger.Info("额度使用量已减少（退款）",
		zap.String("merchant_id", merchantID.String()),
		zap.Int64("amount", amount))

	// 异步同步到DB
	go s.syncUsageToDB(context.Background(), merchantID)

	return nil
}

// getUsedTodayFromRedis 从Redis获取今日使用量
func (s *merchantLimitService) getUsedTodayFromRedis(ctx context.Context, merchantID uuid.UUID) (int64, error) {
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("merchant_limit:daily:%s:%s", merchantID.String(), today)

	val, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	used, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return used, nil
}

// getUsedMonthFromRedis 从Redis获取本月使用量
func (s *merchantLimitService) getUsedMonthFromRedis(ctx context.Context, merchantID uuid.UUID) (int64, error) {
	month := time.Now().Format("200601")
	key := fmt.Sprintf("merchant_limit:monthly:%s:%s", merchantID.String(), month)

	val, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	used, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return used, nil
}

// syncUsageToDB 同步使用量到数据库
func (s *merchantLimitService) syncUsageToDB(ctx context.Context, merchantID uuid.UUID) {
	usedToday, err := s.getUsedTodayFromRedis(ctx, merchantID)
	if err != nil {
		logger.Error("同步额度到DB失败：获取今日使用量",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		return
	}

	usedMonth, err := s.getUsedMonthFromRedis(ctx, merchantID)
	if err != nil {
		logger.Error("同步额度到DB失败：获取本月使用量",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		return
	}

	if err := s.limitRepo.UpdateUsedAmount(ctx, merchantID, usedToday, usedMonth); err != nil {
		logger.Error("同步额度到DB失败",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
	}
}

// GetMerchantLimit 获取商户额度信息
func (s *merchantLimitService) GetMerchantLimit(ctx context.Context, merchantID uuid.UUID) (*model.MerchantLimit, error) {
	// 从DB获取配置
	limit, err := s.limitRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取商户额度失败: %w", err)
	}

	if limit == nil {
		return nil, fmt.Errorf("商户额度不存在")
	}

	// 从Redis获取实时使用量
	usedToday, _ := s.getUsedTodayFromRedis(ctx, merchantID)
	usedMonth, _ := s.getUsedMonthFromRedis(ctx, merchantID)

	limit.UsedToday = usedToday
	limit.UsedMonth = usedMonth

	return limit, nil
}

// UpdateLimitConfig 更新额度配置
func (s *merchantLimitService) UpdateLimitConfig(ctx context.Context, merchantID uuid.UUID, dailyLimit, monthlyLimit, singleLimit int64) error {
	if err := s.limitRepo.UpdateLimitConfig(ctx, merchantID, dailyLimit, monthlyLimit, singleLimit); err != nil {
		return fmt.Errorf("更新商户额度配置失败: %w", err)
	}

	logger.Info("商户额度配置已更新",
		zap.String("merchant_id", merchantID.String()),
		zap.Int64("daily_limit", dailyLimit),
		zap.Int64("monthly_limit", monthlyLimit),
		zap.Int64("single_limit", singleLimit))

	return nil
}

// SetLimitStatus 设置限额状态
func (s *merchantLimitService) SetLimitStatus(ctx context.Context, merchantID uuid.UUID, isLimited bool, reason string) error {
	if err := s.limitRepo.SetLimitStatus(ctx, merchantID, isLimited, reason); err != nil {
		return fmt.Errorf("设置商户限额状态失败: %w", err)
	}

	logger.Info("商户限额状态已更新",
		zap.String("merchant_id", merchantID.String()),
		zap.Bool("is_limited", isLimited),
		zap.String("reason", reason))

	return nil
}
