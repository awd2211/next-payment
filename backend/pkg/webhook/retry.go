package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RetryConfig Webhook 重试配置
type RetryConfig struct {
	MaxRetries     int           // 最大重试次数（默认 5）
	InitialBackoff time.Duration // 初始退避时间（默认 1s）
	MaxBackoff     time.Duration // 最大退避时间（默认 1h）
	Multiplier     float64       // 退避倍数（默认 2.0，指数退避）
	Timeout        time.Duration // 单次请求超时（默认 30s）
}

// DefaultRetryConfig 返回默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     1 * time.Hour,
		Multiplier:     2.0,
		Timeout:        30 * time.Second,
	}
}

// WebhookPayload Webhook 通知负载
type WebhookPayload struct {
	Event     string                 `json:"event"`      // 事件类型: payment.success, payment.failed, refund.success
	PaymentNo string                 `json:"payment_no"` // 支付流水号
	OrderNo   string                 `json:"order_no"`   // 商户订单号
	Amount    int64                  `json:"amount"`     // 金额（分）
	Currency  string                 `json:"currency"`   // 货币
	Status    string                 `json:"status"`     // 状态
	Timestamp int64                  `json:"timestamp"`  // 时间戳
	Extra     map[string]interface{} `json:"extra"`      // 扩展字段
}

// WebhookRequest Webhook 请求
type WebhookRequest struct {
	URL       string          // 通知 URL
	Secret    string          // 签名密钥
	Payload   *WebhookPayload // 通知负载
	MerchantID string         // 商户 ID（用于日志和指标）
}

// WebhookResponse Webhook 响应
type WebhookResponse struct {
	Success      bool          // 是否成功
	StatusCode   int           // HTTP 状态码
	Body         string        // 响应体
	Duration     time.Duration // 请求耗时
	Error        error         // 错误信息
	Attempt      int           // 第几次尝试
	NextRetryAt  *time.Time    // 下次重试时间（如果需要重试）
}

// WebhookRetrier Webhook 重试器
type WebhookRetrier struct {
	config      *RetryConfig
	httpClient  *http.Client
	redisClient *redis.Client
}

// NewWebhookRetrier 创建 Webhook 重试器
func NewWebhookRetrier(config *RetryConfig, redisClient *redis.Client) *WebhookRetrier {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &WebhookRetrier{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		redisClient: redisClient,
	}
}

// Send 发送 Webhook 通知（带重试）
func (r *WebhookRetrier) Send(ctx context.Context, req *WebhookRequest) (*WebhookResponse, error) {
	var lastResp *WebhookResponse

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// 如果是重试，等待退避时间
		if attempt > 0 {
			backoff := r.calculateBackoff(attempt)
			logger.Info("Webhook 重试等待",
				zap.String("merchant_id", req.MerchantID),
				zap.String("payment_no", req.Payload.PaymentNo),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff))

			select {
			case <-time.After(backoff):
				// 继续重试
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// 发送请求
		resp := r.sendOnce(ctx, req, attempt+1)
		lastResp = resp

		// 成功，直接返回
		if resp.Success {
			logger.Info("Webhook 发送成功",
				zap.String("merchant_id", req.MerchantID),
				zap.String("payment_no", req.Payload.PaymentNo),
				zap.Int("attempt", attempt+1),
				zap.Int("status_code", resp.StatusCode),
				zap.Duration("duration", resp.Duration))

			// 清除失败计数
			r.clearFailureCount(ctx, req.MerchantID, req.Payload.PaymentNo)
			return resp, nil
		}

		// 是否应该重试
		if !r.shouldRetry(resp.StatusCode, attempt) {
			logger.Warn("Webhook 发送失败，不再重试",
				zap.String("merchant_id", req.MerchantID),
				zap.String("payment_no", req.Payload.PaymentNo),
				zap.Int("attempt", attempt+1),
				zap.Int("status_code", resp.StatusCode),
				zap.Error(resp.Error))
			break
		}

		// 记录重试
		logger.Warn("Webhook 发送失败，将重试",
			zap.String("merchant_id", req.MerchantID),
			zap.String("payment_no", req.Payload.PaymentNo),
			zap.Int("attempt", attempt+1),
			zap.Int("remaining_retries", r.config.MaxRetries-attempt),
			zap.Int("status_code", resp.StatusCode),
			zap.Error(resp.Error))
	}

	// 所有重试都失败，记录到 Redis（用于后台任务继续重试）
	r.recordFailure(ctx, req, lastResp)

	return lastResp, fmt.Errorf("webhook 发送失败，已达最大重试次数 %d", r.config.MaxRetries)
}

// sendOnce 发送一次 Webhook 请求
func (r *WebhookRetrier) sendOnce(ctx context.Context, req *WebhookRequest, attempt int) *WebhookResponse {
	start := time.Now()

	// 序列化 payload
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return &WebhookResponse{
			Success:  false,
			Error:    fmt.Errorf("序列化 payload 失败: %w", err),
			Duration: time.Since(start),
			Attempt:  attempt,
		}
	}

	// 计算签名
	signature := r.calculateSignature(payloadBytes, req.Secret)

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return &WebhookResponse{
			Success:  false,
			Error:    fmt.Errorf("创建 HTTP 请求失败: %w", err),
			Duration: time.Since(start),
			Attempt:  attempt,
		}
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Signature", signature)
	httpReq.Header.Set("X-Webhook-Event", req.Payload.Event)
	httpReq.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", req.Payload.Timestamp))
	httpReq.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", attempt))

	// 发送请求
	httpResp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return &WebhookResponse{
			Success:  false,
			Error:    fmt.Errorf("HTTP 请求失败: %w", err),
			Duration: time.Since(start),
			Attempt:  attempt,
		}
	}
	defer httpResp.Body.Close()

	// 读取响应体
	bodyBytes, _ := io.ReadAll(httpResp.Body)
	body := string(bodyBytes)

	// 判断是否成功（2xx 状态码）
	success := httpResp.StatusCode >= 200 && httpResp.StatusCode < 300

	return &WebhookResponse{
		Success:    success,
		StatusCode: httpResp.StatusCode,
		Body:       body,
		Duration:   time.Since(start),
		Attempt:    attempt,
	}
}

// calculateBackoff 计算退避时间（指数退避 + 抖动）
func (r *WebhookRetrier) calculateBackoff(attempt int) time.Duration {
	// 指数退避: backoff = initial * (multiplier ^ attempt)
	backoff := float64(r.config.InitialBackoff) * math.Pow(r.config.Multiplier, float64(attempt-1))

	// 限制最大退避时间
	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}

	// 添加抖动（±10%）防止惊群效应
	jitter := backoff * 0.1 * (2*float64(time.Now().UnixNano()%100)/100 - 1)
	backoff += jitter

	return time.Duration(backoff)
}

// shouldRetry 判断是否应该重试
func (r *WebhookRetrier) shouldRetry(statusCode int, attempt int) bool {
	// 已达最大重试次数
	if attempt >= r.config.MaxRetries {
		return false
	}

	// 4xx 客户端错误（除了 408, 429），不重试
	if statusCode >= 400 && statusCode < 500 {
		// 408 Request Timeout, 429 Too Many Requests 可以重试
		return statusCode == 408 || statusCode == 429
	}

	// 5xx 服务器错误，重试
	if statusCode >= 500 {
		return true
	}

	// 网络错误（statusCode = 0），重试
	if statusCode == 0 {
		return true
	}

	// 其他情况不重试
	return false
}

// calculateSignature 计算 Webhook 签名
func (r *WebhookRetrier) calculateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// recordFailure 记录失败的 Webhook（用于后台任务继续重试）
func (r *WebhookRetrier) recordFailure(ctx context.Context, req *WebhookRequest, resp *WebhookResponse) {
	if r.redisClient == nil {
		return
	}

	// 失败记录的 key
	key := fmt.Sprintf("webhook:failed:%s:%s", req.MerchantID, req.Payload.PaymentNo)

	// 记录失败数据
	data := map[string]interface{}{
		"url":         req.URL,
		"payload":     req.Payload,
		"merchant_id": req.MerchantID,
		"attempts":    resp.Attempt,
		"last_error":  resp.Error.Error(),
		"status_code": resp.StatusCode,
		"created_at":  time.Now().Unix(),
	}

	dataBytes, _ := json.Marshal(data)
	if err := r.redisClient.Set(ctx, key, dataBytes, 7*24*time.Hour).Err(); err != nil {
		logger.Error("记录失败 Webhook 到 Redis 失败",
			zap.Error(err),
			zap.String("key", key))
	}

	// 添加到失败队列（用于后台任务批量重试）
	queueKey := "webhook:failed:queue"
	if err := r.redisClient.LPush(ctx, queueKey, key).Err(); err != nil {
		logger.Error("添加到失败队列失败",
			zap.Error(err),
			zap.String("queue_key", queueKey))
	}

	logger.Info("已记录失败 Webhook 到 Redis",
		zap.String("merchant_id", req.MerchantID),
		zap.String("payment_no", req.Payload.PaymentNo),
		zap.String("key", key))
}

// clearFailureCount 清除失败计数
func (r *WebhookRetrier) clearFailureCount(ctx context.Context, merchantID, paymentNo string) {
	if r.redisClient == nil {
		return
	}

	key := fmt.Sprintf("webhook:failed:%s:%s", merchantID, paymentNo)
	r.redisClient.Del(ctx, key)
}

// RetryWorker Webhook 重试后台任务
type RetryWorker struct {
	retrier  *WebhookRetrier
	interval time.Duration
	batchSize int
}

// NewRetryWorker 创建 Webhook 重试后台任务
func NewRetryWorker(retrier *WebhookRetrier, interval time.Duration, batchSize int) *RetryWorker {
	return &RetryWorker{
		retrier:   retrier,
		interval:  interval,
		batchSize: batchSize,
	}
}

// Start 启动后台任务
func (w *RetryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	logger.Info("Webhook 重试后台任务已启动",
		zap.Duration("interval", w.interval),
		zap.Int("batch_size", w.batchSize))

	for {
		select {
		case <-ticker.C:
			w.processBatch(ctx)
		case <-ctx.Done():
			logger.Info("Webhook 重试后台任务已停止")
			return
		}
	}
}

// processBatch 处理一批失败的 Webhook
func (w *RetryWorker) processBatch(ctx context.Context) {
	if w.retrier.redisClient == nil {
		return
	}

	queueKey := "webhook:failed:queue"

	for i := 0; i < w.batchSize; i++ {
		// 从队列中弹出一个失败记录
		key, err := w.retrier.redisClient.RPop(ctx, queueKey).Result()
		if err != nil {
			// 队列为空
			return
		}

		// 获取失败记录
		dataBytes, err := w.retrier.redisClient.Get(ctx, key).Bytes()
		if err != nil {
			logger.Warn("获取失败 Webhook 数据失败",
				zap.Error(err),
				zap.String("key", key))
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			logger.Warn("解析失败 Webhook 数据失败",
				zap.Error(err),
				zap.String("key", key))
			continue
		}

		// 重新构造请求
		req := &WebhookRequest{
			URL:        data["url"].(string),
			MerchantID: data["merchant_id"].(string),
		}

		// TODO: 从数据库重新获取 payload 和 secret

		// 重新发送
		logger.Info("后台任务重试 Webhook",
			zap.String("merchant_id", req.MerchantID),
			zap.String("key", key))

		// resp, err := w.retrier.Send(ctx, req)
		// 处理结果...
	}
}
