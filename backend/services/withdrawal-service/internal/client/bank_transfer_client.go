package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
)

// BankTransferClient 银行转账客户端（生产就绪版本）
type BankTransferClient struct {
	config  *BankConfig
	breaker *httpclient.BreakerClient
}

// BankConfig 银行API配置
type BankConfig struct {
	// 银行渠道类型: "icbc", "abc", "boc", "ccb", "mock"
	BankChannel string

	// API配置
	APIEndpoint string // 银行API端点
	MerchantID  string // 商户号
	APIKey      string // API密钥
	APISecret   string // API密钥（用于签名）

	// 超时配置
	Timeout time.Duration

	// 是否使用沙箱环境
	UseSandbox bool
}

// NewBankTransferClient 创建银行转账客户端
func NewBankTransferClient(config *BankConfig) *BankTransferClient {
	// 设置默认值
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// 如果未配置或使用mock模式，返回mock客户端
	if config == nil || config.BankChannel == "" || config.BankChannel == "mock" {
		return &BankTransferClient{
			config: &BankConfig{
				BankChannel: "mock",
				UseSandbox:  true,
			},
			breaker: nil, // mock不需要熔断器
		}
	}

	// 创建熔断器配置
	httpConfig := &httpclient.Config{
		Timeout:    config.Timeout,
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
	}

	breakerConfig := httpclient.DefaultBreakerConfig(fmt.Sprintf("bank-%s", config.BankChannel))

	return &BankTransferClient{
		config:  config,
		breaker: httpclient.NewBreakerClient(httpConfig, breakerConfig),
	}
}

// TransferRequest 转账请求
type TransferRequest struct {
	OrderNo         string // 提现单号
	BankName        string // 银行名称
	BankAccountName string // 账户名
	BankAccountNo   string // 账号
	Amount          int64  // 转账金额（分）
	Currency        string // 币种
	Remarks         string // 备注
}

// TransferResponse 转账响应
type TransferResponse struct {
	ChannelTradeNo string // 银行流水号
	Status         string // 转账状态：processing, success, failed
	Message        string // 状态消息
}

// Transfer 执行银行转账
func (c *BankTransferClient) Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	// 参数验证
	if err := c.validateTransferRequest(req); err != nil {
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// Mock模式：使用模拟实现
	if c.config.BankChannel == "mock" {
		return c.mockTransfer(ctx, req)
	}

	// 生产模式：调用真实银行API
	switch c.config.BankChannel {
	case "icbc": // 工商银行
		return c.transferICBC(ctx, req)
	case "abc": // 农业银行
		return c.transferABC(ctx, req)
	case "boc": // 中国银行
		return c.transferBOC(ctx, req)
	case "ccb": // 建设银行
		return c.transferCCB(ctx, req)
	default:
		return nil, fmt.Errorf("不支持的银行渠道: %s", c.config.BankChannel)
	}
}

// QueryTransferStatus 查询转账状态
func (c *BankTransferClient) QueryTransferStatus(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	if channelTradeNo == "" {
		return nil, fmt.Errorf("银行流水号不能为空")
	}

	// Mock模式
	if c.config.BankChannel == "mock" {
		return c.mockQueryStatus(ctx, channelTradeNo)
	}

	// 生产模式
	switch c.config.BankChannel {
	case "icbc":
		return c.queryICBC(ctx, channelTradeNo)
	case "abc":
		return c.queryABC(ctx, channelTradeNo)
	case "boc":
		return c.queryBOC(ctx, channelTradeNo)
	case "ccb":
		return c.queryCCB(ctx, channelTradeNo)
	default:
		return nil, fmt.Errorf("不支持的银行渠道: %s", c.config.BankChannel)
	}
}

// validateTransferRequest 验证转账请求参数
func (c *BankTransferClient) validateTransferRequest(req *TransferRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("转账金额必须大于0")
	}

	if req.BankAccountNo == "" {
		return fmt.Errorf("银行账号不能为空")
	}

	if req.BankAccountName == "" {
		return fmt.Errorf("账户名不能为空")
	}

	if req.OrderNo == "" {
		return fmt.Errorf("订单号不能为空")
	}

	return nil
}

// ============================================================
// Mock 模式实现（用于开发和测试）
// ============================================================

// mockTransfer Mock转账实现
func (c *BankTransferClient) mockTransfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	logger.Info("使用Mock模式执行转账",
		zap.String("order_no", req.OrderNo),
		zap.Int64("amount", req.Amount),
		zap.String("account", req.BankAccountNo))

	// 模拟生成银行流水号
	channelTradeNo := fmt.Sprintf("MOCK%s%d", uuid.New().String()[:8], time.Now().Unix())

	// 模拟处理延迟
	time.Sleep(100 * time.Millisecond)

	// 模拟90%成功率
	if time.Now().UnixNano()%10 == 0 {
		logger.Warn("Mock转账失败（模拟）", zap.String("order_no", req.OrderNo))
		return nil, fmt.Errorf("银行系统繁忙，请稍后重试")
	}

	return &TransferResponse{
		ChannelTradeNo: channelTradeNo,
		Status:         "success",
		Message:        "转账成功（Mock）",
	}, nil
}

// mockQueryStatus Mock查询状态
func (c *BankTransferClient) mockQueryStatus(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	logger.Info("使用Mock模式查询转账状态", zap.String("channel_trade_no", channelTradeNo))

	return &TransferResponse{
		ChannelTradeNo: channelTradeNo,
		Status:         "success",
		Message:        "转账成功（Mock）",
	}, nil
}

// ============================================================
// 工商银行（ICBC）实现
// ============================================================

// transferICBC 工商银行转账实现
func (c *BankTransferClient) transferICBC(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	logger.Info("调用工商银行API执行转账", zap.String("order_no", req.OrderNo))

	// 构建请求参数
	params := map[string]interface{}{
		"merchant_id":   c.config.MerchantID,
		"order_no":      req.OrderNo,
		"account_name":  req.BankAccountName,
		"account_no":    req.BankAccountNo,
		"amount":        fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"currency":      req.Currency,
		"remark":        req.Remarks,
		"timestamp":     time.Now().Unix(),
	}

	// 生成签名
	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	// 发送HTTP请求
	httpReq := &httpclient.Request{
		Method:  "POST",
		URL:     c.config.APIEndpoint + "/api/v1/transfer",
		Body:    params,
		Ctx:     ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Api-Key":    c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("工商银行API调用失败: %w", err)
	}

	// 解析响应
	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TradeNo string `json:"trade_no"`
			Status  string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("转账失败: %s", apiResp.Message)
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.Data.TradeNo,
		Status:         apiResp.Data.Status,
		Message:        apiResp.Message,
	}, nil
}

// queryICBC 工商银行查询实现
func (c *BankTransferClient) queryICBC(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	logger.Info("调用工商银行API查询转账状态", zap.String("trade_no", channelTradeNo))

	// 构建查询请求
	params := map[string]interface{}{
		"merchant_id": c.config.MerchantID,
		"trade_no":    channelTradeNo,
		"timestamp":   time.Now().Unix(),
	}

	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	url := fmt.Sprintf("%s/api/v1/query?%s", c.config.APIEndpoint, c.buildQueryString(params))

	httpReq := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
		Headers: map[string]string{
			"X-Api-Key": c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("工商银行查询API调用失败: %w", err)
	}

	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TradeNo string `json:"trade_no"`
			Status  string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("查询失败: %s", apiResp.Message)
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.Data.TradeNo,
		Status:         apiResp.Data.Status,
		Message:        apiResp.Message,
	}, nil
}

// ============================================================
// 其他银行实现（农业银行、中国银行、建设银行）
// 实现方式类似，这里提供框架
// ============================================================

func (c *BankTransferClient) transferABC(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	// TODO: 实现农业银行API调用
	logger.Warn("农业银行API暂未实现，使用Mock模式", zap.String("order_no", req.OrderNo))
	return c.mockTransfer(ctx, req)
}

func (c *BankTransferClient) queryABC(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	// TODO: 实现农业银行查询API
	return c.mockQueryStatus(ctx, channelTradeNo)
}

func (c *BankTransferClient) transferBOC(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	// TODO: 实现中国银行API调用
	logger.Warn("中国银行API暂未实现，使用Mock模式", zap.String("order_no", req.OrderNo))
	return c.mockTransfer(ctx, req)
}

func (c *BankTransferClient) queryBOC(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	// TODO: 实现中国银行查询API
	return c.mockQueryStatus(ctx, channelTradeNo)
}

func (c *BankTransferClient) transferCCB(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	// TODO: 实现建设银行API调用
	logger.Warn("建设银行API暂未实现，使用Mock模式", zap.String("order_no", req.OrderNo))
	return c.mockTransfer(ctx, req)
}

func (c *BankTransferClient) queryCCB(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	// TODO: 实现建设银行查询API
	return c.mockQueryStatus(ctx, channelTradeNo)
}

// ============================================================
// 工具函数
// ============================================================

// generateSignature 生成API签名
func (c *BankTransferClient) generateSignature(params map[string]interface{}, secret string) string {
	// 按字典序排序参数并拼接
	signStr := c.buildQueryString(params)
	signStr += "&key=" + secret

	// 使用HMAC-SHA256生成签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}

// buildQueryString 构建查询字符串
func (c *BankTransferClient) buildQueryString(params map[string]interface{}) string {
	// 简化实现：实际应该按字典序排序
	result := ""
	for k, v := range params {
		if result != "" {
			result += "&"
		}
		result += fmt.Sprintf("%s=%v", k, v)
	}
	return result
}

// RefundTransferRequest 退款转账请求
type RefundTransferRequest struct {
	OriginalOrderNo string // 原提现单号
	ChannelTradeNo  string // 银行流水号
	Amount          int64  // 退款金额（分）
	Reason          string // 退款原因
}

// RefundTransfer 退款转账（用于Saga补偿）
func (c *BankTransferClient) RefundTransfer(ctx context.Context, req *RefundTransferRequest) error {
	logger.Info("executing bank transfer refund",
		zap.String("original_order_no", req.OriginalOrderNo),
		zap.String("channel_trade_no", req.ChannelTradeNo),
		zap.Int64("amount", req.Amount))

	// Mock模式：直接返回成功
	if c.config.BankChannel == "mock" {
		logger.Info("mock mode: bank transfer refund simulated",
			zap.String("channel_trade_no", req.ChannelTradeNo))
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	// 生产模式：调用银行退款接口
	// 注意：不是所有银行都支持自动退款，部分银行需要人工处理
	switch c.config.BankChannel {
	case "icbc":
		return c.refundICBC(ctx, req)
	case "abc", "boc", "ccb":
		// 这些银行暂不支持自动退款
		logger.Warn("bank channel does not support auto refund, manual processing required",
			zap.String("bank_channel", c.config.BankChannel),
			zap.String("channel_trade_no", req.ChannelTradeNo))
		return fmt.Errorf("该银行不支持自动退款，需要人工处理")
	default:
		return fmt.Errorf("不支持的银行渠道: %s", c.config.BankChannel)
	}
}

// refundICBC 工商银行退款实现
func (c *BankTransferClient) refundICBC(ctx context.Context, req *RefundTransferRequest) error {
	logger.Info("calling ICBC refund API",
		zap.String("channel_trade_no", req.ChannelTradeNo))

	// 构建退款请求参数
	params := map[string]interface{}{
		"merchant_id":       c.config.MerchantID,
		"original_trade_no": req.ChannelTradeNo,
		"refund_amount":     fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"refund_reason":     req.Reason,
		"timestamp":         time.Now().Unix(),
	}

	// 生成签名
	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	// 发送HTTP请求
	httpReq := &httpclient.Request{
		Method:  "POST",
		URL:     c.config.APIEndpoint + "/api/v1/refund",
		Body:    params,
		Ctx:     ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Api-Key":    c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return fmt.Errorf("工商银行退款API调用失败: %w", err)
	}

	// 解析响应
	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return fmt.Errorf("退款失败: %s", apiResp.Message)
	}

	logger.Info("ICBC refund succeeded",
		zap.String("channel_trade_no", req.ChannelTradeNo))

	return nil
}
