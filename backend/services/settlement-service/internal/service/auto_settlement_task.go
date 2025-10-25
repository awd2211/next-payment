package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/settlement-service/internal/client"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/repository"
)

// AutoSettlementConfig 自动结算配置
type AutoSettlementConfig struct {
	MinSettlementAmount int64  // 最小结算金额（分）
	AutoApproveThreshold int64  // 自动审批阈值（分）
	RequireKYC          bool   // 是否要求KYC验证
}

// DefaultAutoSettlementConfig 默认配置
var DefaultAutoSettlementConfig = AutoSettlementConfig{
	MinSettlementAmount:  10000,  // 100元
	AutoApproveThreshold: 1000000, // 10000元以下自动审批
	RequireKYC:          true,
}

// AutoSettlementTask 自动结算任务
type AutoSettlementTask struct {
	db                 *gorm.DB
	settlementRepo     repository.SettlementRepository
	accountingClient   *client.AccountingClient
	merchantClient     *client.MerchantClient
	notificationClient *client.NotificationClient
	config             AutoSettlementConfig
}

// NewAutoSettlementTask 创建自动结算任务
func NewAutoSettlementTask(
	db *gorm.DB,
	settlementRepo repository.SettlementRepository,
	accountingClient *client.AccountingClient,
	merchantClient *client.MerchantClient,
	notificationClient *client.NotificationClient,
) *AutoSettlementTask {
	return &AutoSettlementTask{
		db:                 db,
		settlementRepo:     settlementRepo,
		accountingClient:   accountingClient,
		merchantClient:     merchantClient,
		notificationClient: notificationClient,
		config:             DefaultAutoSettlementConfig,
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
	// 查询昨天有交易但今天还没生成结算单的商户
	var merchantIDs []uuid.UUID

	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)

	// 方法1: 从本地数据库查询昨天未结算的商户（如果settlement_items表有merchant_id）
	// 由于settlement_items没有merchant_id，我们需要通过accounting service获取

	// 方法2: 硬编码测试商户（生产环境应该从merchant_config表查询启用自动结算的商户）
	// TODO: 实现merchant_config_service后从配置表查询启用自动结算的商户列表

	// 临时方案：查询昨天有交易的所有商户
	query := `
		SELECT DISTINCT merchant_id
		FROM settlements
		WHERE start_date >= ?
		AND start_date < ?
		LIMIT 1000
	`

	err := t.db.WithContext(ctx).
		Raw(query, yesterday.AddDate(0, 0, -7), yesterday). // 查询过去7天有过结算的商户
		Scan(&merchantIDs).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询商户失败: %w", err)
	}

	// 如果没有历史记录，返回空列表（生产环境应该从merchant service获取）
	if len(merchantIDs) == 0 {
		logger.Info("没有找到历史结算商户，自动结算任务跳过")
		return []uuid.UUID{}, nil
	}

	logger.Info(fmt.Sprintf("找到 %d 个候选商户进行自动结算检查", len(merchantIDs)))
	return merchantIDs, nil
}

// settleMerchant 为单个商户执行结算
func (t *AutoSettlementTask) settleMerchant(ctx context.Context, merchantID uuid.UUID) error {
	logger.Info("开始商户自动结算", zap.String("merchant_id", merchantID.String()))

	// 1. 检查是否已经存在今日的结算单（避免重复）
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	today := yesterday.Add(24 * time.Hour)

	existingSettlement, err := t.settlementRepo.GetByMerchantAndDate(ctx, merchantID, yesterday, today)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查现有结算单失败: %w", err)
	}
	if existingSettlement != nil {
		logger.Info("商户今日已有结算单，跳过",
			zap.String("merchant_id", merchantID.String()),
			zap.String("settlement_no", existingSettlement.SettlementNo))
		return nil
	}

	// 2. 从accounting service获取昨天的交易数据
	transactions, err := t.accountingClient.GetTransactions(ctx, merchantID, yesterday, today)
	if err != nil {
		logger.Error("获取交易数据失败",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		return fmt.Errorf("获取交易数据失败: %w", err)
	}

	if len(transactions) == 0 {
		logger.Info("商户昨日无交易，跳过结算",
			zap.String("merchant_id", merchantID.String()))
		return nil
	}

	// 3. 计算结算金额
	var totalAmount int64
	var totalFee int64
	for _, tx := range transactions {
		totalAmount += tx.Amount
		totalFee += tx.Fee
	}

	settlementAmount := totalAmount - totalFee

	// 4. 检查最小结算金额
	if settlementAmount < t.config.MinSettlementAmount {
		logger.Info("结算金额低于最小值，跳过",
			zap.String("merchant_id", merchantID.String()),
			zap.Int64("settlement_amount", settlementAmount),
			zap.Int64("min_amount", t.config.MinSettlementAmount))
		return nil
	}

	// 5. 创建结算单
	settlement := &model.Settlement{
		MerchantID:       merchantID,
		SettlementNo:     generateSettlementNo(),
		TotalAmount:      totalAmount,
		FeeAmount:        totalFee,
		SettlementAmount: settlementAmount,
		TotalCount:       len(transactions),
		Status:           model.SettlementStatusPending,
		Cycle:            model.SettlementCycleDaily,
		StartDate:        yesterday,
		EndDate:          today,
		RefundAmount:     0, // TODO: 从accounting获取退款数据
		RefundCount:      0,
	}

	// 开始数据库事务
	err = t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 5.1 创建结算单
		if err := t.settlementRepo.Create(ctx, settlement); err != nil {
			return fmt.Errorf("创建结算单失败: %w", err)
		}

		// 5.2 创建结算明细
		for _, tx := range transactions {
			item := &model.SettlementItem{
				SettlementID:  settlement.ID,
				TransactionID: uuid.MustParse(tx.ID),
				OrderNo:       tx.OrderNo,
				PaymentNo:     tx.PaymentNo,
				Amount:        tx.Amount,
				Fee:           tx.Fee,
				SettleAmount:  tx.Amount - tx.Fee,
				TransactionAt: parseTransactionTime(tx.TransactionAt),
			}

			if err := t.db.WithContext(ctx).Create(item).Error; err != nil {
				return fmt.Errorf("创建结算明细失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("保存结算数据失败: %w", err)
	}

	// 6. 自动审批逻辑
	if settlementAmount <= t.config.AutoApproveThreshold {
		if err := t.autoApproveSettlement(ctx, settlement); err != nil {
			logger.Error("自动审批失败",
				zap.String("settlement_no", settlement.SettlementNo),
				zap.Error(err))
			// 自动审批失败不影响结算单创建，只记录日志
		} else {
			logger.Info("结算单已自动审批",
				zap.String("settlement_no", settlement.SettlementNo))
		}
	}

	// 7. 发送通知
	if err := t.sendSettlementNotification(ctx, merchantID, settlement); err != nil {
		logger.Error("发送结算通知失败",
			zap.String("settlement_no", settlement.SettlementNo),
			zap.Error(err))
		// 通知失败不影响结算流程
	}

	logger.Info("商户自动结算成功",
		zap.String("merchant_id", merchantID.String()),
		zap.String("settlement_no", settlement.SettlementNo),
		zap.Int64("total_amount", totalAmount),
		zap.Int64("fee_amount", totalFee),
		zap.Int64("settlement_amount", settlementAmount),
		zap.Int("transaction_count", len(transactions)))

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

// autoApproveSettlement 自动审批结算单
func (t *AutoSettlementTask) autoApproveSettlement(ctx context.Context, settlement *model.Settlement) error {
	now := time.Now()
	systemUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // 系统用户ID

	settlement.Status = model.SettlementStatusApproved
	settlement.ApprovedAt = &now
	settlement.ApprovedBy = &systemUserID

	// 创建审批记录
	approval := &model.SettlementApproval{
		SettlementID: settlement.ID,
		ApproverID:   systemUserID,
		ApproverName: "System Auto-Approve",
		Action:       "approve",
		Comments:     fmt.Sprintf("自动审批（金额 %.2f 元低于阈值 %.2f 元）",
			float64(settlement.SettlementAmount)/100,
			float64(t.config.AutoApproveThreshold)/100),
		ApprovedAt:   now,
	}

	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新结算单状态
		if err := tx.Model(&model.Settlement{}).Where("id = ?", settlement.ID).Updates(map[string]interface{}{
			"status":      model.SettlementStatusApproved,
			"approved_at": now,
			"approved_by": systemUserID,
		}).Error; err != nil {
			return err
		}

		// 创建审批记录
		if err := tx.Create(approval).Error; err != nil {
			return err
		}

		return nil
	})
}

// sendSettlementNotification 发送结算通知
func (t *AutoSettlementTask) sendSettlementNotification(ctx context.Context, merchantID uuid.UUID, settlement *model.Settlement) error {
	if t.notificationClient == nil {
		logger.Warn("通知客户端未初始化，跳过发送通知")
		return nil
	}

	// 获取商户结算账户信息（用于通知）
	account, err := t.merchantClient.GetDefaultSettlementAccount(ctx, merchantID)
	if err != nil {
		logger.Warn("获取商户结算账户失败，仍然发送通知",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
	}

	// 构建通知内容
	var accountInfo string
	if account != nil {
		accountInfo = fmt.Sprintf("\n结算账户: %s (%s)", account.BankName, account.AccountName)
	}

	statusText := "待审批"
	if settlement.Status == model.SettlementStatusApproved {
		statusText = "已自动审批"
	}

	message := fmt.Sprintf(
		"结算单生成通知\n\n"+
			"结算单号: %s\n"+
			"结算周期: 每日结算\n"+
			"结算期间: %s 至 %s\n"+
			"交易笔数: %d 笔\n"+
			"交易总额: %.2f 元\n"+
			"手续费: %.2f 元\n"+
			"结算金额: %.2f 元\n"+
			"状态: %s%s\n\n"+
			"请登录商户平台查看详情。",
		settlement.SettlementNo,
		settlement.StartDate.Format("2006-01-02"),
		settlement.EndDate.Format("2006-01-02"),
		settlement.TotalCount,
		float64(settlement.TotalAmount)/100,
		float64(settlement.FeeAmount)/100,
		float64(settlement.SettlementAmount)/100,
		statusText,
		accountInfo,
	)

	// 发送通知（简化版本，实际应该调用notification service）
	logger.Info("发送结算通知",
		zap.String("merchant_id", merchantID.String()),
		zap.String("settlement_no", settlement.SettlementNo),
		zap.String("message", message))

	// TODO: 实际调用notification client发送邮件/短信
	// err = t.notificationClient.SendEmail(ctx, &client.SendEmailRequest{
	// 	To:      merchantEmail,
	// 	Subject: "结算单生成通知",
	// 	Body:    message,
	// })

	return nil
}

// parseTransactionTime 解析交易时间
func parseTransactionTime(timeStr string) time.Time {
	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	// 解析失败返回当前时间
	logger.Warn("解析交易时间失败，使用当前时间", zap.String("time_str", timeStr))
	return time.Now()
}

// RunDailySettlement 每日结算任务（供定时调度器调用）
func RunDailySettlement(
	db *gorm.DB,
	settlementRepo repository.SettlementRepository,
	accountingClient *client.AccountingClient,
	merchantClient *client.MerchantClient,
	notificationClient *client.NotificationClient,
) func(context.Context) error {
	task := NewAutoSettlementTask(db, settlementRepo, accountingClient, merchantClient, notificationClient)
	return task.Run
}
