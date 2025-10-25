package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/reconciliation-service/internal/model"
)

// ChannelDownloader 渠道账单文件下载器接口
type ChannelDownloader interface {
	// Download 下载渠道账单文件
	Download(ctx context.Context, channel string, settlementDate time.Time) (*model.ChannelSettlementFile, error)

	// Parse 解析渠道账单文件
	Parse(ctx context.Context, fileURL string) ([]*ChannelPayment, error)
}

// PlatformDataFetcher 平台数据获取器接口
type PlatformDataFetcher interface {
	// FetchPayments 获取平台支付记录
	FetchPayments(ctx context.Context, date time.Time, channel string) ([]*PlatformPayment, error)
}

// ReportGenerator 报告生成器接口
type ReportGenerator interface {
	// Generate 生成对账报告
	Generate(ctx context.Context, task *model.ReconciliationTask) (string, error)
}

// PlatformPayment 平台支付记录
type PlatformPayment struct {
	PaymentNo      string
	ChannelTradeNo string
	OrderNo        string
	MerchantID     *uuid.UUID
	Amount         int64
	Currency       string
	Status         string
	PaymentTime    time.Time
}

// ChannelPayment 渠道支付记录
type ChannelPayment struct {
	ChannelTradeNo string
	Amount         int64
	Currency       string
	Status         string
	SettlementTime time.Time
}
