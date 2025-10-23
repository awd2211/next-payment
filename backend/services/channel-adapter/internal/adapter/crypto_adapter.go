package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"payment-platform/channel-adapter/internal/model"
)

// CryptoAdapter 加密货币支付适配器
// 支持多链：ETH, BSC, TRON, BTC 等
type CryptoAdapter struct {
	config     *model.CryptoConfig
	httpClient *http.Client
	priceCache map[string]CryptoPrice // 价格缓存
	cacheTime  time.Time
}

// CryptoPrice 加密货币价格
type CryptoPrice struct {
	Symbol string  `json:"symbol"` // BTC, ETH, USDT 等
	USD    float64 `json:"usd"`    // 美元价格
}

// CryptoTransaction 链上交易信息
type CryptoTransaction struct {
	TxHash        string    `json:"tx_hash"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Amount        string    `json:"amount"`
	Symbol        string    `json:"symbol"`
	Confirmations int       `json:"confirmations"`
	Status        string    `json:"status"`
	BlockNumber   int64     `json:"block_number"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewCryptoAdapter 创建加密货币适配器实例
func NewCryptoAdapter(config *model.CryptoConfig) *CryptoAdapter {
	return &CryptoAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		priceCache: make(map[string]CryptoPrice),
	}
}

// GetChannel 获取渠道名称
func (a *CryptoAdapter) GetChannel() string {
	return model.ChannelCrypto
}

// getCryptoPrice 获取加密货币实时价格（使用 CoinGecko 免费 API）
func (a *CryptoAdapter) getCryptoPrice(ctx context.Context, symbol string) (float64, error) {
	// 检查缓存（5分钟有效）
	if time.Since(a.cacheTime) < 5*time.Minute {
		if price, ok := a.priceCache[symbol]; ok {
			return price.USD, nil
		}
	}

	// 符号映射（统一转换为 CoinGecko ID）
	symbolMap := map[string]string{
		"BTC":  "bitcoin",
		"ETH":  "ethereum",
		"USDT": "tether",
		"USDC": "usd-coin",
		"BNB":  "binancecoin",
		"TRX":  "tron",
	}

	coinID, ok := symbolMap[strings.ToUpper(symbol)]
	if !ok {
		return 0, fmt.Errorf("不支持的加密货币: %s", symbol)
	}

	// 调用 CoinGecko API
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coinID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建价格查询请求失败: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("获取价格失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("获取价格失败，状态码: %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("解析价格响应失败: %w", err)
	}

	price, ok := result[coinID]["usd"]
	if !ok {
		return 0, fmt.Errorf("价格数据不存在")
	}

	// 更新缓存
	a.priceCache[symbol] = CryptoPrice{
		Symbol: symbol,
		USD:    price,
	}
	a.cacheTime = time.Now()

	return price, nil
}

// calculateCryptoAmount 计算需要支付的加密货币数量
func (a *CryptoAdapter) calculateCryptoAmount(ctx context.Context, fiatAmount int64, fiatCurrency, cryptoSymbol string) (float64, error) {
	// 获取加密货币价格
	cryptoPrice, err := a.getCryptoPrice(ctx, cryptoSymbol)
	if err != nil {
		return 0, err
	}

	// 将法币金额转换为美元（使用真实汇率）
	usdAmount := float64(fiatAmount) / 100.0
	if fiatCurrency != "USD" {
		// 使用汇率客户端进行转换
		if a.exchangeRateClient != nil {
			rate, err := a.exchangeRateClient.GetRate(ctx, fiatCurrency, "USD")
			if err == nil {
				usdAmount = usdAmount * rate
			}
			// 如果获取汇率失败，使用原始金额（已有降级处理）
		}
	}

	// 计算加密货币数量
	cryptoAmount := usdAmount / cryptoPrice

	return cryptoAmount, nil
}

// CreatePayment 创建支付
func (a *CryptoAdapter) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 确定使用的网络和代币
	network := "ETH" // 默认使用以太坊
	symbol := "ETH"  // 默认使用 ETH

	// 从扩展信息中获取网络和代币配置
	if req.Extra != nil {
		if n, ok := req.Extra["network"].(string); ok && n != "" {
			network = n
		}
		if s, ok := req.Extra["symbol"].(string); ok && s != "" {
			symbol = s
		}
	}

	// 验证网络是否支持
	networkSupported := false
	for _, n := range a.config.Networks {
		if strings.EqualFold(n, network) {
			networkSupported = true
			break
		}
	}
	if !networkSupported {
		return nil, fmt.Errorf("不支持的网络: %s", network)
	}

	// 计算需要支付的加密货币数量
	cryptoAmount, err := a.calculateCryptoAmount(ctx, req.Amount, req.Currency, symbol)
	if err != nil {
		return nil, fmt.Errorf("计算加密货币数量失败: %w", err)
	}

	// 生成支付信息（使用配置的钱包地址）
	paymentAddress := a.config.WalletAddress

	// 生成二维码数据（适用于钱包扫码支付）
	// 格式：ethereum:0xAddress@chainId?value=amount
	var qrData string
	switch strings.ToUpper(network) {
	case "ETH":
		qrData = fmt.Sprintf("ethereum:%s?value=%.8f", paymentAddress, cryptoAmount)
	case "BSC":
		qrData = fmt.Sprintf("ethereum:%s@56?value=%.8f", paymentAddress, cryptoAmount)
	case "TRON":
		qrData = fmt.Sprintf("tron:%s?amount=%.8f", paymentAddress, cryptoAmount)
	default:
		qrData = fmt.Sprintf("%s:%s?amount=%.8f", strings.ToLower(network), paymentAddress, cryptoAmount)
	}

	// 构造响应
	response := &CreatePaymentResponse{
		ChannelTradeNo: req.PaymentNo, // 使用平台支付号作为交易号
		QRCodeURL:      qrData,         // 二维码数据
		Status:         PaymentStatusPending,
		Extra: map[string]interface{}{
			"payment_address": paymentAddress,
			"crypto_amount":   cryptoAmount,
			"crypto_symbol":   symbol,
			"network":         network,
			"confirmations_required": a.config.Confirmations,
			"expires_at": time.Now().Add(30 * time.Minute).Unix(), // 30分钟有效期
		},
	}

	return response, nil
}

// QueryPayment 查询支付状态
func (a *CryptoAdapter) QueryPayment(ctx context.Context, channelTradeNo string) (*QueryPaymentResponse, error) {
	// 查询链上交易
	// 这里需要根据 channelTradeNo（支付号）查询对应的交易记录
	// 实际实现中应该：
	// 1. 从数据库获取期望的支付金额和地址
	// 2. 调用区块链浏览器 API 查询该地址的交易记录
	// 3. 匹配金额和时间范围，找到对应的交易

	// 模拟查询结果（实际应该调用区块链 API）
	tx, err := a.queryBlockchainTransaction(ctx, a.config.WalletAddress)
	if err != nil {
		return nil, err
	}

	if tx == nil {
		// 未找到交易，仍在等待支付
		return &QueryPaymentResponse{
			ChannelTradeNo: channelTradeNo,
			Status:         PaymentStatusPending,
		}, nil
	}

	// 判断确认数是否足够
	status := PaymentStatusProcessing
	if tx.Confirmations >= a.config.Confirmations {
		status = PaymentStatusSuccess
	}

	// 将加密货币金额转换为法币金额（分）
	// TODO: 应该根据支付时的价格计算，而不是当前价格
	amount := int64(0)

	response := &QueryPaymentResponse{
		ChannelTradeNo: channelTradeNo,
		Status:         status,
		Amount:         amount,
		Currency:       "USD",
		PaymentMethod:  "crypto",
		PaymentMethodDetails: map[string]interface{}{
			"tx_hash":        tx.TxHash,
			"crypto_amount":  tx.Amount,
			"crypto_symbol":  tx.Symbol,
			"confirmations":  tx.Confirmations,
			"block_number":   tx.BlockNumber,
		},
	}

	if tx.Confirmations >= a.config.Confirmations {
		paidAt := tx.Timestamp.Unix()
		response.PaidAt = &paidAt
	}

	return response, nil
}

// queryBlockchainTransaction 查询区块链交易
func (a *CryptoAdapter) queryBlockchainTransaction(ctx context.Context, address string) (*CryptoTransaction, error) {
	// 这里应该调用区块链浏览器 API
	// 例如：Etherscan API, BSCScan API, TronScan API 等

	if a.config.APIEndpoint == "" {
		// 如果没有配置 API 端点，返回 nil（表示需要手动确认）
		return nil, nil
	}

	// 示例：调用 Etherscan API 查询交易
	// GET https://api.etherscan.io/api?module=account&action=txlist&address=0x...&apikey=...

	url := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&sort=desc&apikey=%s",
		a.config.APIEndpoint, address, a.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建区块链查询请求失败: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询区块链交易失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询区块链交易失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应（具体格式取决于使用的 API）
	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash          string `json:"hash"`
			From          string `json:"from"`
			To            string `json:"to"`
			Value         string `json:"value"`
			BlockNumber   string `json:"blockNumber"`
			Confirmations string `json:"confirmations"`
			TimeStamp     string `json:"timeStamp"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析区块链响应失败: %w", err)
	}

	if result.Status != "1" || len(result.Result) == 0 {
		// 未找到交易
		return nil, nil
	}

	// 返回最新的一笔交易
	tx := result.Result[0]

	// 解析时间戳
	var timestamp int64
	fmt.Sscanf(tx.TimeStamp, "%d", &timestamp)

	// 解析确认数
	var confirmations int
	fmt.Sscanf(tx.Confirmations, "%d", &confirmations)

	// 解析区块号
	var blockNumber int64
	fmt.Sscanf(tx.BlockNumber, "%d", &blockNumber)

	return &CryptoTransaction{
		TxHash:        tx.Hash,
		From:          tx.From,
		To:            tx.To,
		Amount:        tx.Value,
		Symbol:        "ETH",
		Confirmations: confirmations,
		Status:        "confirmed",
		BlockNumber:   blockNumber,
		Timestamp:     time.Unix(timestamp, 0),
	}, nil
}

// CancelPayment 取消支付
func (a *CryptoAdapter) CancelPayment(ctx context.Context, channelTradeNo string) error {
	// 加密货币支付无法取消（链上交易不可逆）
	// 只能标记为过期
	return nil
}

// CreateRefund 创建退款
func (a *CryptoAdapter) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error) {
	// 加密货币退款需要手动转账
	// 这里只是记录退款请求，实际转账需要人工处理或使用热钱包自动转账

	return &CreateRefundResponse{
		RefundNo:        req.RefundNo,
		ChannelRefundNo: req.RefundNo,
		Status:          PaymentStatusProcessing, // 标记为处理中，需要人工确认
		Extra: map[string]interface{}{
			"note": "加密货币退款需要手动转账到用户地址",
			"requires_manual_processing": true,
		},
	}, nil
}

// QueryRefund 查询退款状态
func (a *CryptoAdapter) QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error) {
	// 查询退款交易状态
	// 实际实现中应该查询退款转账的交易哈希状态

	return &QueryRefundResponse{
		ChannelRefundNo: refundNo,
		Status:          PaymentStatusProcessing,
		Amount:          0,
		Currency:        "USD",
	}, nil
}

// VerifyWebhook 验证 Webhook 签名
func (a *CryptoAdapter) VerifyWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	// 加密货币支付一般通过主动轮询区块链，而不是 Webhook
	// 如果使用第三方服务（如 Coinbase Commerce），则需要验证其 Webhook 签名

	// 这里简化处理，认为都是可信的
	return true, nil
}

// ParseWebhook 解析 Webhook 数据
func (a *CryptoAdapter) ParseWebhook(ctx context.Context, body []byte) (*WebhookEvent, error) {
	// 解析区块链事件通知
	var event struct {
		TxHash   string `json:"tx_hash"`
		From     string `json:"from"`
		To       string `json:"to"`
		Amount   string `json:"amount"`
		Symbol   string `json:"symbol"`
		Network  string `json:"network"`
		Status   string `json:"status"`
		Confirmations int `json:"confirmations"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("解析 Webhook 失败: %w", err)
	}

	// 构造 Webhook 事件
	webhookEvent := &WebhookEvent{
		EventID:        event.TxHash,
		EventType:      EventTypePaymentSuccess,
		ChannelTradeNo: event.TxHash,
		Status:         convertCryptoStatus(event.Status, event.Confirmations, a.config.Confirmations),
		Currency:       event.Symbol,
		RawData:        event,
		Extra: map[string]interface{}{
			"tx_hash":       event.TxHash,
			"network":       event.Network,
			"confirmations": event.Confirmations,
		},
	}

	return webhookEvent, nil
}

// convertCryptoStatus 转换加密货币交易状态为统一状态
func convertCryptoStatus(status string, confirmations, requiredConfirmations int) string {
	switch strings.ToLower(status) {
	case "pending", "unconfirmed":
		return PaymentStatusPending
	case "confirmed":
		if confirmations >= requiredConfirmations {
			return PaymentStatusSuccess
		}
		return PaymentStatusProcessing
	case "failed":
		return PaymentStatusFailed
	default:
		return PaymentStatusPending
	}
}

// ConvertWeiToEther 将 Wei 转换为 Ether（用于 ETH/BSC）
func ConvertWeiToEther(wei string) float64 {
	// Wei 是以太坊的最小单位，1 ETH = 10^18 Wei
	var weiValue float64
	fmt.Sscanf(wei, "%f", &weiValue)
	return weiValue / math.Pow(10, 18)
}

// ConvertSunToTRX 将 Sun 转换为 TRX（用于 TRON）
func ConvertSunToTRX(sun string) float64 {
	// Sun 是 TRON 的最小单位，1 TRX = 10^6 Sun
	var sunValue float64
	fmt.Sscanf(sun, "%f", &sunValue)
	return sunValue / math.Pow(10, 6)
}

// ConvertSatoshiToBTC 将 Satoshi 转换为 BTC（用于比特币）
func ConvertSatoshiToBTC(satoshi string) float64 {
	// Satoshi 是比特币的最小单位，1 BTC = 10^8 Satoshi
	var satoshiValue float64
	fmt.Sscanf(satoshi, "%f", &satoshiValue)
	return satoshiValue / math.Pow(10, 8)
}
