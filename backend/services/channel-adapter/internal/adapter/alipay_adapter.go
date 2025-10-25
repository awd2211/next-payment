package adapter

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"payment-platform/channel-adapter/internal/model"
)

// AlipayAdapter 支付宝支付适配器
type AlipayAdapter struct {
	DefaultPreAuthNotSupported  // 嵌入默认预授权实现
	config      *model.AlipayConfig
	httpClient  *http.Client
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
}

// NewAlipayAdapter 创建支付宝适配器实例
func NewAlipayAdapter(config *model.AlipayConfig) (*AlipayAdapter, error) {
	// 解析私钥
	privateKey, err := parsePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 解析公钥
	publicKey, err := parsePublicKey(config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	return &AlipayAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// GetChannel 获取渠道名称
func (a *AlipayAdapter) GetChannel() string {
	return model.ChannelAlipay
}

// buildCommonParams 构建公共参数
func (a *AlipayAdapter) buildCommonParams() map[string]string {
	return map[string]string{
		"app_id":    a.config.AppID,
		"format":    a.config.Format,
		"charset":   a.config.Charset,
		"sign_type": a.config.SignType,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}
}

// sign 对参数进行签名
func (a *AlipayAdapter) sign(params map[string]string) (string, error) {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接字符串
	var buffer bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buffer.WriteString("&")
		}
		buffer.WriteString(k)
		buffer.WriteString("=")
		buffer.WriteString(params[k])
	}

	// 计算签名
	h := sha256.New()
	h.Write(buffer.Bytes())
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, a.privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verify 验证签名
func (a *AlipayAdapter) verify(params map[string]string, sign string) error {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接字符串
	var buffer bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buffer.WriteString("&")
		}
		buffer.WriteString(k)
		buffer.WriteString("=")
		buffer.WriteString(params[k])
	}

	// 解码签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("解码签名失败: %w", err)
	}

	// 验证签名
	h := sha256.New()
	h.Write(buffer.Bytes())
	hashed := h.Sum(nil)

	if err := rsa.VerifyPKCS1v15(a.publicKey, crypto.SHA256, hashed, signBytes); err != nil {
		return fmt.Errorf("验证签名失败: %w", err)
	}

	return nil
}

// request 发送请求到支付宝
func (a *AlipayAdapter) request(ctx context.Context, method string, bizContent map[string]interface{}) (map[string]interface{}, error) {
	// 构建公共参数
	params := a.buildCommonParams()
	params["method"] = method
	params["notify_url"] = a.config.NotifyURL

	// 业务参数
	bizJSON, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizJSON)

	// 签名
	sign, err := a.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	// 构建请求
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.config.APIGateway, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, nil
}

// CreatePayment 创建支付
func (a *AlipayAdapter) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no": req.PaymentNo,
		"total_amount": fmt.Sprintf("%.2f", float64(req.Amount)/100.0),
		"subject":      req.Description,
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}

	if req.CustomerEmail != "" {
		bizContent["buyer_email"] = req.CustomerEmail
	}

	// 设置页面跳转同步通知页面路径
	if a.config.ReturnURL != "" {
		bizContent["return_url"] = a.config.ReturnURL
	}

	// 调用支付宝接口
	result, err := a.request(ctx, "alipay.trade.page.pay", bizContent)
	if err != nil {
		return nil, err
	}

	// 解析响应
	responseKey := "alipay_trade_page_pay_response"
	respData, ok := result[responseKey].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	code, _ := respData["code"].(string)
	if code != "10000" {
		msg, _ := respData["msg"].(string)
		subMsg, _ := respData["sub_msg"].(string)
		return nil, fmt.Errorf("支付宝接口返回错误: %s - %s", msg, subMsg)
	}

	tradeNo, _ := respData["trade_no"].(string)

	// 构建支付URL（页面跳转）
	params := a.buildCommonParams()
	params["method"] = "alipay.trade.page.pay"
	params["notify_url"] = a.config.NotifyURL
	params["return_url"] = a.config.ReturnURL
	bizJSON, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizJSON)

	sign, _ := a.sign(params)
	params["sign"] = sign

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	paymentURL := a.config.APIGateway + "?" + values.Encode()

	return &CreatePaymentResponse{
		ChannelTradeNo: tradeNo,
		PaymentURL:     paymentURL,
		Status:         PaymentStatusPending,
		Extra: map[string]interface{}{
			"trade_no": tradeNo,
		},
	}, nil
}

// QueryPayment 查询支付状态
func (a *AlipayAdapter) QueryPayment(ctx context.Context, channelTradeNo string) (*QueryPaymentResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"trade_no": channelTradeNo,
	}

	// 调用支付宝接口
	result, err := a.request(ctx, "alipay.trade.query", bizContent)
	if err != nil {
		return nil, err
	}

	// 解析响应
	responseKey := "alipay_trade_query_response"
	respData, ok := result[responseKey].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	code, _ := respData["code"].(string)
	if code != "10000" {
		msg, _ := respData["msg"].(string)
		subMsg, _ := respData["sub_msg"].(string)
		return nil, fmt.Errorf("支付宝接口返回错误: %s - %s", msg, subMsg)
	}

	tradeNo, _ := respData["trade_no"].(string)
	tradeStatus, _ := respData["trade_status"].(string)
	totalAmount, _ := respData["total_amount"].(string)

	// 将金额字符串转换为分
	var amountFloat float64
	fmt.Sscanf(totalAmount, "%f", &amountFloat)

	response := &QueryPaymentResponse{
		ChannelTradeNo: tradeNo,
		Status:         convertAlipayStatus(tradeStatus),
		Amount:         int64(amountFloat * 100),
		Currency:       "CNY",
		PaymentMethod:  "alipay",
	}

	// 支付时间
	if sendPayDate, ok := respData["send_pay_date"].(string); ok {
		if t, err := time.Parse("2006-01-02 15:04:05", sendPayDate); err == nil {
			paidAt := t.Unix()
			response.PaidAt = &paidAt
		}
	}

	return response, nil
}

// CancelPayment 取消支付
func (a *AlipayAdapter) CancelPayment(ctx context.Context, channelTradeNo string) error {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"trade_no": channelTradeNo,
	}

	// 调用支付宝接口
	result, err := a.request(ctx, "alipay.trade.cancel", bizContent)
	if err != nil {
		return err
	}

	// 解析响应
	responseKey := "alipay_trade_cancel_response"
	respData, ok := result[responseKey].(map[string]interface{})
	if !ok {
		return fmt.Errorf("响应格式错误")
	}

	code, _ := respData["code"].(string)
	if code != "10000" {
		msg, _ := respData["msg"].(string)
		subMsg, _ := respData["sub_msg"].(string)
		return fmt.Errorf("支付宝接口返回错误: %s - %s", msg, subMsg)
	}

	return nil
}

// CreateRefund 创建退款
func (a *AlipayAdapter) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"trade_no":      req.ChannelTradeNo,
		"refund_amount": fmt.Sprintf("%.2f", float64(req.Amount)/100.0),
		"refund_reason": req.Reason,
		"out_request_no": req.RefundNo,
	}

	// 调用支付宝接口
	result, err := a.request(ctx, "alipay.trade.refund", bizContent)
	if err != nil {
		return nil, err
	}

	// 解析响应
	responseKey := "alipay_trade_refund_response"
	respData, ok := result[responseKey].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	code, _ := respData["code"].(string)
	if code != "10000" {
		msg, _ := respData["msg"].(string)
		subMsg, _ := respData["sub_msg"].(string)
		return nil, fmt.Errorf("支付宝接口返回错误: %s - %s", msg, subMsg)
	}

	return &CreateRefundResponse{
		RefundNo:        req.RefundNo,
		ChannelRefundNo: req.RefundNo, // 支付宝使用 out_request_no 作为退款号
		Status:          PaymentStatusRefunded,
		Extra: map[string]interface{}{
			"fund_change": respData["fund_change"],
		},
	}, nil
}

// QueryRefund 查询退款状态
func (a *AlipayAdapter) QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_request_no": refundNo,
	}

	// 调用支付宝接口
	result, err := a.request(ctx, "alipay.trade.fastpay.refund.query", bizContent)
	if err != nil {
		return nil, err
	}

	// 解析响应
	responseKey := "alipay_trade_fastpay_refund_query_response"
	respData, ok := result[responseKey].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	code, _ := respData["code"].(string)
	if code != "10000" {
		msg, _ := respData["msg"].(string)
		subMsg, _ := respData["sub_msg"].(string)
		return nil, fmt.Errorf("支付宝接口返回错误: %s - %s", msg, subMsg)
	}

	refundAmount, _ := respData["refund_amount"].(string)

	// 将金额字符串转换为分
	var amountFloat float64
	fmt.Sscanf(refundAmount, "%f", &amountFloat)

	response := &QueryRefundResponse{
		ChannelRefundNo: refundNo,
		Status:          PaymentStatusRefunded,
		Amount:          int64(amountFloat * 100),
		Currency:        "CNY",
	}

	return response, nil
}

// VerifyWebhook 验证 Webhook 签名
func (a *AlipayAdapter) VerifyWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	// 解析请求参数
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return false, fmt.Errorf("解析参数失败: %w", err)
	}

	// 构建参数 map
	params := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 获取签名
	sign := params["sign"]
	if sign == "" {
		return false, fmt.Errorf("签名为空")
	}

	// 验证签名
	if err := a.verify(params, sign); err != nil {
		return false, err
	}

	return true, nil
}

// ParseWebhook 解析 Webhook 数据
func (a *AlipayAdapter) ParseWebhook(ctx context.Context, body []byte) (*WebhookEvent, error) {
	// 解析请求参数
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("解析参数失败: %w", err)
	}

	// 构建参数 map
	params := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 构造 Webhook 事件
	webhookEvent := &WebhookEvent{
		EventID:        params["notify_id"],
		ChannelTradeNo: params["trade_no"],
		PaymentNo:      params["out_trade_no"],
		RawData:        params,
	}

	// 根据交易状态设置事件类型和状态
	tradeStatus := params["trade_status"]
	webhookEvent.Status = convertAlipayStatus(tradeStatus)
	webhookEvent.EventType = convertAlipayEventType(tradeStatus)

	// 提取金额信息
	if totalAmount := params["total_amount"]; totalAmount != "" {
		var amountFloat float64
		fmt.Sscanf(totalAmount, "%f", &amountFloat)
		webhookEvent.Amount = int64(amountFloat * 100)
	}

	webhookEvent.Currency = "CNY"

	return webhookEvent, nil
}

// convertAlipayStatus 转换支付宝交易状态为统一状态
func convertAlipayStatus(status string) string {
	switch status {
	case "WAIT_BUYER_PAY":
		return PaymentStatusPending
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return PaymentStatusSuccess
	case "TRADE_CLOSED":
		return PaymentStatusCancelled
	default:
		return PaymentStatusFailed
	}
}

// convertAlipayEventType 转换支付宝事件类型为统一事件类型
func convertAlipayEventType(status string) string {
	switch status {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return EventTypePaymentSuccess
	case "TRADE_CLOSED":
		return EventTypePaymentCancelled
	default:
		return status
	}
}

// parsePrivateKey 解析私钥
func parsePrivateKey(privateKeyStr string) (*rsa.PrivateKey, error) {
	// 移除头尾标记
	privateKeyStr = strings.TrimSpace(privateKeyStr)
	privateKeyStr = strings.Replace(privateKeyStr, "-----BEGIN PRIVATE KEY-----", "", 1)
	privateKeyStr = strings.Replace(privateKeyStr, "-----END PRIVATE KEY-----", "", 1)
	privateKeyStr = strings.Replace(privateKeyStr, "-----BEGIN RSA PRIVATE KEY-----", "", 1)
	privateKeyStr = strings.Replace(privateKeyStr, "-----END RSA PRIVATE KEY-----", "", 1)
	privateKeyStr = strings.Replace(privateKeyStr, "\n", "", -1)
	privateKeyStr = strings.Replace(privateKeyStr, "\r", "", -1)

	// 解码 Base64
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("解码私钥失败: %w", err)
	}

	// 解析私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		// 尝试 PKCS1 格式
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥类型错误")
	}

	return rsaPrivateKey, nil
}

// parsePublicKey 解析公钥
func parsePublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	// 移除头尾标记
	publicKeyStr = strings.TrimSpace(publicKeyStr)
	publicKeyStr = strings.Replace(publicKeyStr, "-----BEGIN PUBLIC KEY-----", "", 1)
	publicKeyStr = strings.Replace(publicKeyStr, "-----END PUBLIC KEY-----", "", 1)
	publicKeyStr = strings.Replace(publicKeyStr, "\n", "", -1)
	publicKeyStr = strings.Replace(publicKeyStr, "\r", "", -1)

	// 解码 Base64
	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, fmt.Errorf("解码公钥失败: %w", err)
	}

	// 解析公钥
	publicKey, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥类型错误")
	}

	return rsaPublicKey, nil
}
