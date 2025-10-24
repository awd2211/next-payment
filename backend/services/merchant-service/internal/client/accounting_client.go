package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// AccountingClient Accounting Service HTTP客户端
type AccountingClient struct {
	*ServiceClient
}

// NewAccountingClient 创建Accounting客户端实例（带熔断器）
func NewAccountingClient(baseURL string) *AccountingClient {
	return &AccountingClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "accounting-service"),
	}
}

// BalanceSummaryResponse 余额汇总响应
type BalanceSummaryResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *BalanceSummaryData `json:"data"`
}

// BalanceSummaryData 余额汇总数据
type BalanceSummaryData struct {
	TotalBalance      int64           `json:"total_balance"`
	AvailableBalance  int64           `json:"available_balance"`
	FrozenBalance     int64           `json:"frozen_balance"`
	PendingSettlement int64           `json:"pending_settlement"`
	InTransit         int64           `json:"in_transit"`
	Accounts          []AccountInfo   `json:"accounts"`
}

// AccountInfo 账户信息
type AccountInfo struct {
	ID            string `json:"id"`
	AccountType   string `json:"account_type"`
	Currency      string `json:"currency"`
	Balance       int64  `json:"balance"`
	FrozenBalance int64  `json:"frozen_balance"`
}

// SettlementListResponse 结算列表响应
type SettlementListResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *SettlementListData `json:"data"`
}

// SettlementListData 结算列表数据
type SettlementListData struct {
	List     []SettlementInfo `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// SettlementInfo 结算信息
type SettlementInfo struct {
	ID            string `json:"id"`
	SettlementNo  string `json:"settlement_no"`
	PeriodStart   string `json:"period_start"`
	PeriodEnd     string `json:"period_end"`
	TotalAmount   int64  `json:"total_amount"`
	FeeAmount     int64  `json:"fee_amount"`
	NetAmount     int64  `json:"net_amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	PaymentCount  int    `json:"payment_count"`
	SettledAt     string `json:"settled_at"`
	CreatedAt     string `json:"created_at"`
}

// GetBalanceSummary 获取余额汇总
func (c *AccountingClient) GetBalanceSummary(ctx context.Context, merchantID uuid.UUID) (*BalanceSummaryData, error) {
	path := fmt.Sprintf("/api/v1/balances/merchants/%s/summary", merchantID.String())

	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	var result BalanceSummaryResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

// GetSettlements 获取结算列表
func (c *AccountingClient) GetSettlements(ctx context.Context, merchantID uuid.UUID, page, pageSize int) (*SettlementListData, error) {
	path := fmt.Sprintf("/api/v1/settlements?merchant_id=%s&page=%d&page_size=%d",
		merchantID.String(), page, pageSize)

	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	var result SettlementListResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}
