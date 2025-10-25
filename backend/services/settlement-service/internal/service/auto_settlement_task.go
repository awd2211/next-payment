package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/repository"
)

// AutoSettlementTask 自动结算任务
type AutoSettlementTask struct {
	db             *gorm.DB
	settlementRepo repository.SettlementRepository
}

// NewAutoSettlementTask 创建自动结算任务
func NewAutoSettlementTask(db *gorm.DB, settlementRepo repository.SettlementRepository) *AutoSettlementTask {
	return &AutoSettlementTask{
		db:             db,
		settlementRepo: settlementRepo,
	}
}

// Run 执行自动结算任务
func (t *AutoSettlementTask) Run(ctx context.Context) error {
	logger.Info("开始执行自动结算任务")

	// 1. 查询所有启用自动结算的商户
	merchants, err := t.getAutoSettlementMerchants(ctx)
	if err != nil {
		return fmt.Errorf("查询自动结算商户失败: %w", err)
	}

	if len(merchants) == 0 {
		logger.Info("没有需要自动结算的商户")
		return nil
	}

	logger.Info(fmt.Sprintf("找到 %d 个需要自动结算的商户", len(merchants)))

	successCount := 0
	failedCount := 0

	// 2. 为每个商户执行结算
	for _, merchantID := range merchants {
		if err := t.settleMerchant(ctx, merchantID); err != nil {
			logger.Error("商户自动结算失败",
				zap.String("merchant_id", merchantID.String()),
				zap.Error(err))
			failedCount++
		} else {
			successCount++
		}
	}

	logger.Info("自动结算任务完成",
		zap.Int("total", len(merchants)),
		zap.Int("success", successCount),
		zap.Int("failed", failedCount))

	return nil
}

// getAutoSettlementMerchants 获取启用自动结算的商户列表
func (t *AutoSettlementTask) getAutoSettlementMerchants(ctx context.Context) ([]uuid.UUID, error) {
	// 这里简化处理，实际应该从merchant表或配置表查询
	// 查询今天还没有结算的商户
	var merchantIDs []uuid.UUID

	// 示例SQL：查询昨天有交易但今天还没结算的商户
	yesterday := time.Now().AddDate(0, 0, -1)
	today := time.Now()

	query := `
		SELECT DISTINCT merchant_id
		FROM settlements
		WHERE created_at >= ?
		AND created_at < ?
		AND status = 'pending'
		LIMIT 100
	`

	err := t.db.WithContext(ctx).
		Raw(query, yesterday, today).
		Scan(&merchantIDs).Error

	return merchantIDs, err
}

// settleMerchant 为单个商户执行结算
func (t *AutoSettlementTask) settleMerchant(ctx context.Context, merchantID uuid.UUID) error {
	logger.Info("开始商户自动结算", zap.String("merchant_id", merchantID.String()))

	// 1. 查询待结算金额
	var totalAmount int64
	var count int64

	// 查询昨天的成功交易金额
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	today := yesterday.Add(24 * time.Hour)

	query := `
		SELECT COALESCE(SUM(amount), 0) as total, COUNT(*) as count
		FROM settlements
		WHERE merchant_id = ?
		AND status = 'pending'
		AND created_at >= ?
		AND created_at < ?
	`

	type Result struct {
		Total int64
		Count int64
	}

	var result Result
	err := t.db.WithContext(ctx).
		Raw(query, merchantID, yesterday, today).
		Scan(&result).Error

	if err != nil {
		return fmt.Errorf("查询待结算金额失败: %w", err)
	}

	totalAmount = result.Total
	count = result.Count

	if totalAmount == 0 {
		logger.Info("商户无待结算金额，跳过",
			zap.String("merchant_id", merchantID.String()))
		return nil
	}

	// 2. 创建结算单
	feeAmount := calculateFee(totalAmount)
	settlement := &model.Settlement{
		MerchantID:       merchantID,
		SettlementNo:     generateSettlementNo(),
		TotalAmount:      totalAmount,
		FeeAmount:        feeAmount,
		SettlementAmount: totalAmount - feeAmount,
		TotalCount:       int(count),
		Status:           model.SettlementStatusPending,
		Cycle:            model.SettlementCycleDaily, // 每日结算
		StartDate:        yesterday,
		EndDate:          today,
	}

	if err := t.settlementRepo.Create(ctx, settlement); err != nil {
		return fmt.Errorf("创建结算单失败: %w", err)
	}

	logger.Info("商户自动结算成功",
		zap.String("merchant_id", merchantID.String()),
		zap.String("settlement_no", settlement.SettlementNo),
		zap.Int64("total_amount", totalAmount),
		zap.Int64("fee_amount", feeAmount),
		zap.Int64("settlement_amount", settlement.SettlementAmount),
		zap.Int("transaction_count", int(count)))

	return nil
}

// generateSettlementNo 生成结算单号
func generateSettlementNo() string {
	return fmt.Sprintf("STL%s%d",
		time.Now().Format("20060102"),
		time.Now().UnixNano()%1000000)
}

// calculateFee 计算手续费（简化版本，实际应该根据商户费率配置）
func calculateFee(amount int64) int64 {
	// 默认费率 0.6%
	return amount * 6 / 1000
}

// RunDailySettlement 每日结算任务（供定时调度器调用）
func RunDailySettlement(db *gorm.DB, settlementRepo repository.SettlementRepository) func(context.Context) error {
	task := NewAutoSettlementTask(db, settlementRepo)
	return task.Run
}
