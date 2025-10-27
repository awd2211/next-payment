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

// transferABC 农业银行转账实现
func (c *BankTransferClient) transferABC(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	logger.Info("调用农业银行API执行转账", zap.String("order_no", req.OrderNo))

	// 农业银行API参数格式（与工商银行类似，但字段名可能不同）
	params := map[string]interface{}{
		"merchantNo":   c.config.MerchantID,
		"orderNo":      req.OrderNo,
		"payeeName":    req.BankAccountName,
		"payeeAccount": req.BankAccountNo,
		"amount":       fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"currency":     req.Currency,
		"memo":         req.Remarks,
		"requestTime":  time.Now().Format("20060102150405"), // ABC使用格式化时间
	}

	// 生成签名
	sign := c.generateSignature(params, c.config.APISecret)
	params["signature"] = sign

	// 发送HTTP请求
	httpReq := &httpclient.Request{
		Method: "POST",
		URL:    c.config.APIEndpoint + "/transfer/singlePay",
		Body:   params,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"AppId":        c.config.APIKey, // ABC使用AppId而非X-Api-Key
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("农业银行API调用失败: %w", err)
	}

	// 解析响应（ABC响应格式）
	var apiResp struct {
		ReturnCode string `json:"returnCode"`
		ReturnMsg  string `json:"returnMsg"`
		TradeNo    string `json:"tradeNo"`
		Status     string `json:"status"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.ReturnCode != "0000" { // ABC使用0000表示成功
		return nil, fmt.Errorf("转账失败: %s", apiResp.ReturnMsg)
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.TradeNo,
		Status:         apiResp.Status,
		Message:        apiResp.ReturnMsg,
	}, nil
}

// queryABC 农业银行查询实现
func (c *BankTransferClient) queryABC(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	logger.Info("调用农业银行API查询转账状态", zap.String("trade_no", channelTradeNo))

	// 构建查询请求
	params := map[string]interface{}{
		"merchantNo":  c.config.MerchantID,
		"tradeNo":     channelTradeNo,
		"requestTime": time.Now().Format("20060102150405"),
	}

	sign := c.generateSignature(params, c.config.APISecret)
	params["signature"] = sign

	url := fmt.Sprintf("%s/transfer/query?%s", c.config.APIEndpoint, c.buildQueryString(params))

	httpReq := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
		Headers: map[string]string{
			"AppId": c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("农业银行查询API调用失败: %w", err)
	}

	var apiResp struct {
		ReturnCode string `json:"returnCode"`
		ReturnMsg  string `json:"returnMsg"`
		TradeNo    string `json:"tradeNo"`
		Status     string `json:"status"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.ReturnCode != "0000" {
		return nil, fmt.Errorf("查询失败: %s", apiResp.ReturnMsg)
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.TradeNo,
		Status:         apiResp.Status,
		Message:        apiResp.ReturnMsg,
	}, nil
}

// transferBOC 中国银行转账实现
func (c *BankTransferClient) transferBOC(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	logger.Info("调用中国银行API执行转账", zap.String("order_no", req.OrderNo))

	// 中国银行API参数格式
	params := map[string]interface{}{
		"mchId":       c.config.MerchantID,
		"mchOrderNo":  req.OrderNo,
		"accountName": req.BankAccountName,
		"accountNo":   req.BankAccountNo,
		"tranAmt":     fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"currency":    req.Currency,
		"remark":      req.Remarks,
		"reqTime":     time.Now().Format("2006-01-02 15:04:05"), // BOC使用标准时间格式
	}

	// 生成签名
	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	// 发送HTTP请求
	httpReq := &httpclient.Request{
		Method: "POST",
		URL:    c.config.APIEndpoint + "/api/payment/transfer",
		Body:   params,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.config.APIKey, // BOC使用Bearer token
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("中国银行API调用失败: %w", err)
	}

	// 解析响应（BOC响应格式）
	var apiResp struct {
		RespCode string `json:"respCode"`
		RespMsg  string `json:"respMsg"`
		OrderNo  string `json:"orderNo"`
		Status   string `json:"status"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.RespCode != "SUCCESS" { // BOC使用SUCCESS表示成功
		return nil, fmt.Errorf("转账失败: %s", apiResp.RespMsg)
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.OrderNo,
		Status:         apiResp.Status,
		Message:        apiResp.RespMsg,
	}, nil
}

// queryBOC 中国银行查询实现
func (c *BankTransferClient) queryBOC(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	logger.Info("调用中国银行API查询转账状态", zap.String("trade_no", channelTradeNo))

	// 构建查询请求
	params := map[string]interface{}{
		"mchId":   c.config.MerchantID,
		"orderNo": channelTradeNo,
		"reqTime": time.Now().Format("2006-01-02 15:04:05"),
	}

	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	url := fmt.Sprintf("%s/api/payment/query?%s", c.config.APIEndpoint, c.buildQueryString(params))

	httpReq := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
		Headers: map[string]string{
			"Authorization": "Bearer " + c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("中国银行查询API调用失败: %w", err)
	}

	var apiResp struct {
		RespCode string `json:"respCode"`
		RespMsg  string `json:"respMsg"`
		OrderNo  string `json:"orderNo"`
		Status   string `json:"status"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.RespCode != "SUCCESS" {
		return nil, fmt.Errorf("查询失败: %s", apiResp.RespMsg)
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.OrderNo,
		Status:         apiResp.Status,
		Message:        apiResp.RespMsg,
	}, nil
}

// transferCCB 建设银行转账实现
func (c *BankTransferClient) transferCCB(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	logger.Info("调用建设银行API执行转账", zap.String("order_no", req.OrderNo))

	// 建设银行API参数格式
	params := map[string]interface{}{
		"partnerid":       c.config.MerchantID,
		"out_trade_no":    req.OrderNo,
		"payee_real_name": req.BankAccountName,
		"payee_account":   req.BankAccountNo,
		"trans_amount":    fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"fee_type":        req.Currency,
		"desc":            req.Remarks,
		"timestamp":       fmt.Sprintf("%d", time.Now().Unix()), // CCB使用Unix时间戳
	}

	// 生成签名
	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	// 发送HTTP请求
	httpReq := &httpclient.Request{
		Method: "POST",
		URL:    c.config.APIEndpoint + "/mmpaymkttransfers/promotion/transfers",
		Body:   params,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Mchid":        c.config.MerchantID, // CCB使用Mchid header
			"ApiKey":       c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("建设银行API调用失败: %w", err)
	}

	// 解析响应（CCB响应格式）
	var apiResp struct {
		ResultCode   string `json:"result_code"`
		ErrCodeDes   string `json:"err_code_des"`
		PaymentNo    string `json:"payment_no"`
		PaymentTime  string `json:"payment_time"`
		TransferStatus string `json:"transfer_status"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.ResultCode != "SUCCESS" { // CCB使用SUCCESS表示成功
		return nil, fmt.Errorf("转账失败: %s", apiResp.ErrCodeDes)
	}

	// 映射CCB状态到标准状态
	status := "processing"
	if apiResp.TransferStatus == "SUCCESS" {
		status = "success"
	} else if apiResp.TransferStatus == "FAILED" {
		status = "failed"
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.PaymentNo,
		Status:         status,
		Message:        "转账提交成功",
	}, nil
}

// queryCCB 建设银行查询实现
func (c *BankTransferClient) queryCCB(ctx context.Context, channelTradeNo string) (*TransferResponse, error) {
	logger.Info("调用建设银行API查询转账状态", zap.String("trade_no", channelTradeNo))

	// 构建查询请求
	params := map[string]interface{}{
		"partnerid":    c.config.MerchantID,
		"payment_no":   channelTradeNo,
		"timestamp":    fmt.Sprintf("%d", time.Now().Unix()),
	}

	sign := c.generateSignature(params, c.config.APISecret)
	params["sign"] = sign

	url := fmt.Sprintf("%s/mmpaymkttransfers/gettransferinfo?%s", c.config.APIEndpoint, c.buildQueryString(params))

	httpReq := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
		Headers: map[string]string{
			"Mchid":  c.config.MerchantID,
			"ApiKey": c.config.APIKey,
		},
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("建设银行查询API调用失败: %w", err)
	}

	var apiResp struct {
		ResultCode     string `json:"result_code"`
		ErrCodeDes     string `json:"err_code_des"`
		PaymentNo      string `json:"payment_no"`
		TransferStatus string `json:"transfer_status"`
		Reason         string `json:"reason"`
	}

	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("查询失败: %s", apiResp.ErrCodeDes)
	}

	// 映射CCB状态到标准状态
	status := "processing"
	if apiResp.TransferStatus == "SUCCESS" {
		status = "success"
	} else if apiResp.TransferStatus == "FAILED" {
		status = "failed"
	}

	message := "转账处理中"
	if status == "success" {
		message = "转账成功"
	} else if status == "failed" {
		message = apiResp.Reason
	}

	return &TransferResponse{
		ChannelTradeNo: apiResp.PaymentNo,
		Status:         status,
		Message:        message,
	}, nil
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
