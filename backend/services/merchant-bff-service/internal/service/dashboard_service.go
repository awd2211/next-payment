package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"payment-platform/merchant-service/internal/client"
)

// DashboardService Dashboard聚合服务接口
type DashboardService interface {
	GetDashboard(ctx context.Context, merchantID uuid.UUID) (*DashboardData, error)
	GetTransactionSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate string) (*TransactionSummary, error)
	GetBalanceInfo(ctx context.Context, merchantID uuid.UUID) (*BalanceInfo, error)
	GetTransactions(ctx context.Context, merchantID uuid.UUID, filter *TransactionFilter) (*TransactionListResult, error)
	GetSettlements(ctx context.Context, merchantID uuid.UUID, page, pageSize int) (*SettlementListResult, error)
}

type dashboardService struct {
	analyticsClient    *client.AnalyticsClient
	accountingClient   *client.AccountingClient
	riskClient         *client.RiskClient
	notificationClient *client.NotificationClient
	paymentClient      *client.PaymentClient
}

// NewDashboardService 创建Dashboard服务实例
func NewDashboardService(
	analyticsClient *client.AnalyticsClient,
	accountingClient *client.AccountingClient,
	riskClient *client.RiskClient,
	notificationClient *client.NotificationClient,
	paymentClient *client.PaymentClient,
) DashboardService {
	return &dashboardService{
		analyticsClient:    analyticsClient,
		accountingClient:   accountingClient,
		riskClient:         riskClient,
		notificationClient: notificationClient,
		paymentClient:      paymentClient,
	}
}

// DashboardData Dashboard数据
type DashboardData struct {
	// 今日数据
	TodayPayments      int     `json:"today_payments"`       // 今日交易笔数
	TodayAmount        int64   `json:"today_amount"`         // 今日交易金额（分）
	TodaySuccessRate   float64 `json:"today_success_rate"`   // 今日成功率

	// 本月数据
	MonthPayments      int     `json:"month_payments"`       // 本月交易笔数
	MonthAmount        int64   `json:"month_amount"`         // 本月交易金额（分）
	MonthSuccessRate   float64 `json:"month_success_rate"`   // 本月成功率

	// 余额信息
	AvailableBalance   int64   `json:"available_balance"`    // 可用余额（分）
	FrozenBalance      int64   `json:"frozen_balance"`       // 冻结余额（分）
	PendingSettlement  int64   `json:"pending_settlement"`   // 待结算金额（分）

	// 风控信息
	RiskLevel          string  `json:"risk_level"`           // 风险等级
	PendingReviews     int     `json:"pending_reviews"`      // 待审核交易数

	// 快捷操作
	PendingWithdrawals int     `json:"pending_withdrawals"`  // 待处理提现数
	UnreadNotifications int    `json:"unread_notifications"` // 未读通知数

	// 趋势数据（最近7天）
	PaymentTrend       []DailyData `json:"payment_trend"`     // 交易趋势
}

// DailyData 每日数据
type DailyData struct {
	Date         string  `json:"date"`          // 日期 YYYY-MM-DD
	Payments     int     `json:"payments"`      // 交易笔数
	Amount       int64   `json:"amount"`        // 交易金额（分）
	SuccessRate  float64 `json:"success_rate"`  // 成功率
}

// TransactionSummary 交易汇总
type TransactionSummary struct {
	TotalPayments      int     `json:"total_payments"`       // 总交易笔数
	SuccessPayments    int     `json:"success_payments"`     // 成功交易笔数
	FailedPayments     int     `json:"failed_payments"`      // 失败交易笔数
	TotalAmount        int64   `json:"total_amount"`         // 总交易金额（分）
	SuccessAmount      int64   `json:"success_amount"`       // 成功金额（分）
	SuccessRate        float64 `json:"success_rate"`         // 成功率
	AverageAmount      int64   `json:"average_amount"`       // 平均交易金额（分）

	// 退款数据
	TotalRefunds       int     `json:"total_refunds"`        // 退款笔数
	TotalRefundAmount  int64   `json:"total_refund_amount"`  // 退款金额（分）
	RefundRate         float64 `json:"refund_rate"`          // 退款率

	// 渠道分布
	ChannelBreakdown   []ChannelData `json:"channel_breakdown"` // 渠道分布
}

// ChannelData 渠道数据
type ChannelData struct {
	Channel      string  `json:"channel"`       // 渠道名称
	Payments     int     `json:"payments"`      // 交易笔数
	Amount       int64   `json:"amount"`        // 交易金额（分）
	SuccessRate  float64 `json:"success_rate"`  // 成功率
	Percentage   float64 `json:"percentage"`    // 占比
}

// BalanceInfo 余额信息
type BalanceInfo struct {
	Currency          string            `json:"currency"`          // 币种
	AvailableBalance  int64             `json:"available_balance"` // 可用余额（分）
	FrozenBalance     int64             `json:"frozen_balance"`    // 冻结余额（分）
	TotalBalance      int64             `json:"total_balance"`     // 总余额（分）
	PendingSettlement int64             `json:"pending_settlement"`// 待结算金额（分）
	InTransit         int64             `json:"in_transit"`        // 在途金额（分）
	Accounts          []AccountBalance  `json:"accounts"`          // 各账户余额明细
}

// AccountBalance 账户余额
type AccountBalance struct {
	AccountType   string `json:"account_type"`   // 账户类型
	Balance       int64  `json:"balance"`        // 余额（分）
	Currency      string `json:"currency"`       // 币种
}

// GetDashboard 获取Dashboard数据
func (s *dashboardService) GetDashboard(ctx context.Context, merchantID uuid.UUID) (*DashboardData, error) {
	dashboard := &DashboardData{}

	// 调用 analytics-service 获取交易统计数据
	if s.analyticsClient != nil {
		stats, err := s.analyticsClient.GetStatistics(ctx, merchantID)
		if err != nil {
			// 统计数据获取失败不影响整体，记录日志继续
			logger.Error("failed to get analytics statistics",
				zap.Error(err),
				zap.String("merchant_id", merchantID.String()))
		} else {
			dashboard.TodayPayments = stats.TodayPayments
			dashboard.TodayAmount = stats.TodayAmount
			dashboard.TodaySuccessRate = stats.TodaySuccessRate
			dashboard.MonthPayments = stats.MonthPayments
			dashboard.MonthAmount = stats.MonthAmount
			dashboard.MonthSuccessRate = stats.MonthSuccessRate

			// 转换趋势数据
			dashboard.PaymentTrend = make([]DailyData, len(stats.PaymentTrend))
			for i, trend := range stats.PaymentTrend {
				dashboard.PaymentTrend[i] = DailyData{
					Date:        trend.Date,
					Payments:    trend.Payments,
					Amount:      trend.Amount,
					SuccessRate: trend.SuccessRate,
				}
			}
		}
	}

	// 调用 accounting-service 获取余额信息
	if s.accountingClient != nil {
		balance, err := s.accountingClient.GetBalanceSummary(ctx, merchantID)
		if err != nil {
			logger.Error("failed to get balance summary",
				zap.Error(err),
				zap.String("merchant_id", merchantID.String()))
		} else {
			dashboard.AvailableBalance = balance.AvailableBalance
			dashboard.FrozenBalance = balance.FrozenBalance
			dashboard.PendingSettlement = balance.PendingSettlement
		}
	}

	// 调用 risk-service 获取风控信息
	if s.riskClient != nil {
		risk, err := s.riskClient.GetRiskInfo(ctx, merchantID)
		if err != nil {
			logger.Error("failed to get risk info",
				zap.Error(err),
				zap.String("merchant_id", merchantID.String()))
		} else {
			dashboard.RiskLevel = risk.RiskLevel
			dashboard.PendingReviews = risk.PendingReviews
		}
	}

	// 调用 notification-service 获取未读通知数
	if s.notificationClient != nil {
		unread, err := s.notificationClient.GetUnreadCount(ctx, merchantID)
		if err != nil {
			logger.Error("failed to get unread notification count",
				zap.Error(err),
				zap.String("merchant_id", merchantID.String()))
		} else {
			dashboard.UnreadNotifications = unread.Total
		}
	}

	// 待处理提现数（目前设为0，等待 withdrawal-service 实现）
	dashboard.PendingWithdrawals = 0

	return dashboard, nil
}

// GetTransactionSummary 获取交易汇总
func (s *dashboardService) GetTransactionSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate string) (*TransactionSummary, error) {
	summary := &TransactionSummary{}

	// 调用 analytics-service 获取交易汇总数据
	if s.analyticsClient != nil {
		data, err := s.analyticsClient.GetTransactionSummary(ctx, merchantID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("获取交易汇总失败: %w", err)
		}

		summary.TotalPayments = data.TotalPayments
		summary.SuccessPayments = data.SuccessPayments
		summary.FailedPayments = data.FailedPayments
		summary.TotalAmount = data.TotalAmount
		summary.SuccessAmount = data.SuccessAmount
		summary.SuccessRate = data.SuccessRate
		summary.AverageAmount = data.AverageAmount
		summary.TotalRefunds = data.TotalRefunds
		summary.TotalRefundAmount = data.TotalRefundAmount
		summary.RefundRate = data.RefundRate

		// 转换渠道分布数据
		summary.ChannelBreakdown = make([]ChannelData, len(data.ChannelBreakdown))
		for i, ch := range data.ChannelBreakdown {
			summary.ChannelBreakdown[i] = ChannelData{
				Channel:     ch.Channel,
				Payments:    ch.Payments,
				Amount:      ch.Amount,
				SuccessRate: ch.SuccessRate,
				Percentage:  ch.Percentage,
			}
		}
	} else {
		return nil, fmt.Errorf("analytics客户端未初始化")
	}

	return summary, nil
}

// GetBalanceInfo 获取余额信息
func (s *dashboardService) GetBalanceInfo(ctx context.Context, merchantID uuid.UUID) (*BalanceInfo, error) {
	if s.accountingClient == nil {
		return nil, fmt.Errorf("accounting客户端未初始化")
	}

	// 调用 accounting-service 获取余额汇总和各账户明细
	balance, err := s.accountingClient.GetBalanceSummary(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取余额信息失败: %w", err)
	}

	balanceInfo := &BalanceInfo{
		Currency:          "USD", // 默认币种，后续可以从配置获取
		AvailableBalance:  balance.AvailableBalance,
		FrozenBalance:     balance.FrozenBalance,
		TotalBalance:      balance.TotalBalance,
		PendingSettlement: balance.PendingSettlement,
		InTransit:         balance.InTransit,
		Accounts:          make([]AccountBalance, len(balance.Accounts)),
	}

	// 转换账户明细
	for i, acc := range balance.Accounts {
		balanceInfo.Accounts[i] = AccountBalance{
			AccountType: acc.AccountType,
			Balance:     acc.Balance,
			Currency:    acc.Currency,
		}
	}

	return balanceInfo, nil
}

// TransactionFilter 交易查询过滤器
type TransactionFilter struct {
	StartDate     string `json:"start_date"`     // 开始日期
	EndDate       string `json:"end_date"`       // 结束日期
	Status        string `json:"status"`         // 状态
	Channel       string `json:"channel"`        // 渠道
	PaymentMethod string `json:"payment_method"` // 支付方式
	MinAmount     int64  `json:"min_amount"`     // 最小金额（分）
	MaxAmount     int64  `json:"max_amount"`     // 最大金额（分）
	Page          int    `json:"page"`           // 页码
	PageSize      int    `json:"page_size"`      // 每页数量
}

// Transaction 交易记录（简化版，用于列表展示）
type Transaction struct {
	ID            string `json:"id"`
	OrderNo       string `json:"order_no"`       // 订单号
	PaymentNo     string `json:"payment_no"`     // 支付流水号
	Amount        int64  `json:"amount"`         // 金额（分）
	Currency      string `json:"currency"`       // 币种
	Status        string `json:"status"`         // 状态
	Channel       string `json:"channel"`        // 渠道
	PaymentMethod string `json:"payment_method"` // 支付方式
	CustomerEmail string `json:"customer_email"` // 客户邮箱
	CreatedAt     string `json:"created_at"`     // 创建时间
}

// TransactionListResult 交易列表结果
type TransactionListResult struct {
	List     []Transaction `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// GetTransactions 获取交易列表（聚合查询）
func (s *dashboardService) GetTransactions(ctx context.Context, merchantID uuid.UUID, filter *TransactionFilter) (*TransactionListResult, error) {
	if s.paymentClient == nil {
		return nil, fmt.Errorf("payment客户端未初始化")
	}

	// 构建查询参数
	params := make(map[string]string)
	if filter.StartDate != "" {
		params["start_date"] = filter.StartDate
	}
	if filter.EndDate != "" {
		params["end_date"] = filter.EndDate
	}
	if filter.Status != "" {
		params["status"] = filter.Status
	}
	if filter.Channel != "" {
		params["channel"] = filter.Channel
	}
	if filter.PaymentMethod != "" {
		params["payment_method"] = filter.PaymentMethod
	}
	if filter.MinAmount > 0 {
		params["min_amount"] = strconv.FormatInt(filter.MinAmount, 10)
	}
	if filter.MaxAmount > 0 {
		params["max_amount"] = strconv.FormatInt(filter.MaxAmount, 10)
	}
	if filter.Page > 0 {
		params["page"] = strconv.Itoa(filter.Page)
	}
	if filter.PageSize > 0 {
		params["page_size"] = strconv.Itoa(filter.PageSize)
	}

	// 调用 payment-gateway 获取交易列表
	data, err := s.paymentClient.GetPayments(ctx, merchantID, params)
	if err != nil {
		return nil, fmt.Errorf("获取交易列表失败: %w", err)
	}

	// 转换数据格式
	result := &TransactionListResult{
		Total:    data.Total,
		Page:     data.Page,
		PageSize: data.PageSize,
		List:     make([]Transaction, len(data.List)),
	}

	for i, p := range data.List {
		result.List[i] = Transaction{
			ID:            p.ID,
			OrderNo:       p.OrderNo,
			PaymentNo:     p.PaymentNo,
			Amount:        p.Amount,
			Currency:      p.Currency,
			Status:        p.Status,
			Channel:       p.Channel,
			PaymentMethod: p.PayMethod,
			CustomerEmail: p.CustomerEmail,
			CreatedAt:     p.CreatedAt,
		}
	}

	return result, nil
}

// Settlement 结算记录
type Settlement struct {
	ID            string `json:"id"`
	SettlementNo  string `json:"settlement_no"`  // 结算单号
	PeriodStart   string `json:"period_start"`   // 结算周期开始
	PeriodEnd     string `json:"period_end"`     // 结算周期结束
	TotalAmount   int64  `json:"total_amount"`   // 结算总额（分）
	FeeAmount     int64  `json:"fee_amount"`     // 手续费（分）
	NetAmount     int64  `json:"net_amount"`     // 净额（分）
	Currency      string `json:"currency"`       // 币种
	Status        string `json:"status"`         // 状态
	PaymentCount  int    `json:"payment_count"`  // 支付笔数
	SettledAt     string `json:"settled_at"`     // 结算时间
}

// SettlementListResult 结算列表结果
type SettlementListResult struct {
	List     []Settlement `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// GetSettlements 获取结算列表（聚合查询）
func (s *dashboardService) GetSettlements(ctx context.Context, merchantID uuid.UUID, page, pageSize int) (*SettlementListResult, error) {
	if s.accountingClient == nil {
		return nil, fmt.Errorf("accounting客户端未初始化")
	}

	// 调用 accounting-service 获取结算列表
	data, err := s.accountingClient.GetSettlements(ctx, merchantID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取结算列表失败: %w", err)
	}

	// 转换数据格式
	result := &SettlementListResult{
		Total:    data.Total,
		Page:     data.Page,
		PageSize: data.PageSize,
		List:     make([]Settlement, len(data.List)),
	}

	for i, s := range data.List {
		result.List[i] = Settlement{
			ID:           s.ID,
			SettlementNo: s.SettlementNo,
			PeriodStart:  s.PeriodStart,
			PeriodEnd:    s.PeriodEnd,
			TotalAmount:  s.TotalAmount,
			FeeAmount:    s.FeeAmount,
			NetAmount:    s.NetAmount,
			Currency:     s.Currency,
			Status:       s.Status,
			PaymentCount: s.PaymentCount,
			SettledAt:    s.SettledAt,
		}
	}

	return result, nil
}
