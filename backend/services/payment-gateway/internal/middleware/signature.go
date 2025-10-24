package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// APIKeyData API密钥完整数据
type APIKeyData struct {
	Secret       string
	MerchantID   uuid.UUID
	IsActive     bool
	ExpiresAt    *time.Time
	Environment  string
	IPWhitelist  string // IP白名单（可选）
	ShouldRotate bool   // 是否需要轮换密钥
}

// APIKeyUpdater API Key更新器接口
type APIKeyUpdater interface {
	UpdateLastUsedAt(ctx context.Context, apiKey string) error
}

// SignatureMiddleware API签名验证中间件
// 用于验证商户请求的合法性，防止重放攻击和暴力破解
type SignatureMiddleware struct {
	// 获取商户API Secret的函数（返回完整数据）
	getAPISecret    func(apiKey string) (*APIKeyData, error)
	redis           *redis.Client
	apiKeyUpdater   APIKeyUpdater // API Key更新器（用于更新last_used_at）
	maxBodySize     int64         // 最大请求体大小（默认10MB）
	timestampWindow time.Duration // 时间戳允许误差（默认2分钟）
}

// NewSignatureMiddleware 创建签名验证中间件
func NewSignatureMiddleware(
	getAPISecret func(apiKey string) (*APIKeyData, error),
	redisClient *redis.Client,
) *SignatureMiddleware {
	return &SignatureMiddleware{
		getAPISecret:    getAPISecret,
		redis:           redisClient,
		apiKeyUpdater:   nil, // 可选，通过SetAPIKeyUpdater设置
		maxBodySize:     10 * 1024 * 1024, // 10MB
		timestampWindow: 2 * time.Minute,   // 2分钟
	}
}

// SetAPIKeyUpdater 设置API Key更新器（可选）
func (m *SignatureMiddleware) SetAPIKeyUpdater(updater APIKeyUpdater) {
	m.apiKeyUpdater = updater
}

// Verify 验证签名
// 实现完整的安全验证流程：
// 1. 请求头验证
// 2. 失败次数限制（防止暴力破解）
// 3. 时间戳验证（2分钟窗口）
// 4. API Key验证（活跃状态、过期时间）
// 5. Nonce去重验证（防止重放攻击）
// 6. 请求体大小限制（防止DoS）
// 7. 签名验证（HMAC-SHA256）
// 8. 记录安全事件日志
func (m *SignatureMiddleware) Verify() gin.HandlerFunc {
	const authErrorMsg = "Authentication failed"

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. 提取签名相关的头部信息
		apiKey := c.GetHeader("X-API-Key")
		signature := c.GetHeader("X-Signature")
		timestamp := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		signatureVersion := c.GetHeader("X-Signature-Version") // 签名算法版本（可选，默认v1）
		clientIP := c.ClientIP()

		// 默认签名版本为v1（向后兼容）
		if signatureVersion == "" {
			signatureVersion = "v1"
		}

		// 2. 验证签名版本是否支持
		if !isSupportedSignatureVersion(signatureVersion) {
			logger.Debug("Unsupported signature version",
				zap.String("version", signatureVersion),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 3. 验证必要参数是否存在（不泄露具体缺失哪个参数）
		if apiKey == "" || signature == "" || timestamp == "" || nonce == "" {
			logger.Debug("Missing authentication headers",
				zap.String("client_ip", clientIP),
				zap.Bool("has_api_key", apiKey != ""),
				zap.Bool("has_signature", signature != ""),
				zap.Bool("has_timestamp", timestamp != ""),
				zap.Bool("has_nonce", nonce != ""))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 3. 检查失败次数限制（防止暴力破解）
		failedKey := fmt.Sprintf("sig_failed:%s", apiKey)
		failedCount, err := m.redis.Incr(ctx, failedKey).Result()
		if err != nil {
			logger.Error("Failed to check rate limit", zap.Error(err))
			// 即使Redis失败也继续验证，但记录错误
		} else if failedCount == 1 {
			// 首次失败，设置15分钟过期时间
			m.redis.Expire(ctx, failedKey, 15*time.Minute)
		} else if failedCount > 10 {
			// 超过10次失败，锁定账户
			logger.Warn("Account locked due to too many failed signature attempts",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.Int64("failed_count", failedCount),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many failed authentication attempts. Please try again later.",
			})
			c.Abort()
			return
		}

		// 4. 验证时间戳（2分钟窗口，防止重放攻击）
		if err := m.validateTimestamp(timestamp); err != nil {
			logger.Debug("Timestamp validation failed",
				zap.Error(err),
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 5. 获取API密钥数据（包含Secret、过期时间、活跃状态）
		keyData, err := m.getAPISecret(apiKey)
		if err != nil {
			logger.Warn("API key lookup failed",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("client_ip", clientIP),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 6. 验证API Key是否活跃
		if !keyData.IsActive {
			logger.Warn("Inactive API key attempted",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("merchant_id", keyData.MerchantID.String()),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 7. 验证API Key是否过期
		if keyData.ExpiresAt != nil && time.Now().After(*keyData.ExpiresAt) {
			logger.Warn("Expired API key attempted",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("merchant_id", keyData.MerchantID.String()),
				zap.Time("expired_at", *keyData.ExpiresAt),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 8. 验证IP白名单（可选功能）
		if keyData.IPWhitelist != "" {
			if !isIPInWhitelist(clientIP, keyData.IPWhitelist) {
				logger.Warn("IP not in whitelist",
					zap.String("api_key", maskAPIKey(apiKey)),
					zap.String("merchant_id", keyData.MerchantID.String()),
					zap.String("client_ip", clientIP),
					zap.String("whitelist", keyData.IPWhitelist))
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied from this IP"})
				c.Abort()
				return
			}
		}

		// 8. 验证Nonce唯一性（防止重放攻击）
		nonceKey := fmt.Sprintf("nonce:%s:%s", apiKey, nonce)
		exists, err := m.redis.Exists(ctx, nonceKey).Result()
		if err != nil {
			logger.Error("Redis nonce check failed",
				zap.Error(err),
				zap.String("api_key", maskAPIKey(apiKey)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			c.Abort()
			return
		}
		if exists > 0 {
			logger.Warn("Replay attack detected - nonce reused",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("nonce", nonce),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 9. 限制请求体大小（防止DoS攻击）
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, m.maxBodySize)

		// 10. 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			if err.Error() == "http: request body too large" {
				logger.Warn("Request body too large",
					zap.String("api_key", maskAPIKey(apiKey)),
					zap.String("client_ip", clientIP))
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": "Request body too large",
				})
			} else {
				logger.Error("Failed to read request body",
					zap.Error(err),
					zap.String("api_key", maskAPIKey(apiKey)))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			}
			c.Abort()
			return
		}

		// 重置请求体供后续处理使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 11. 计算并验证签名（使用常量时间比较）
		// 根据签名版本使用不同的签名算法
		expectedSignature := m.calculateSignatureByVersion(
			signatureVersion,
			keyData.Secret,
			c.Request.Method,
			c.Request.URL.Path,
			timestamp,
			nonce,
			string(body),
		)
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			logger.Warn("Signature verification failed",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("merchant_id", keyData.MerchantID.String()),
				zap.String("signature_version", signatureVersion),
				zap.Int64("failed_attempts", failedCount),
				zap.String("client_ip", clientIP))
			c.JSON(http.StatusUnauthorized, gin.H{"error": authErrorMsg})
			c.Abort()
			return
		}

		// 12. 签名验证成功 - 存储nonce（10分钟TTL，比时间戳窗口长）
		if err := m.redis.Set(ctx, nonceKey, "1", 10*time.Minute).Err(); err != nil {
			logger.Error("Failed to store nonce in Redis",
				zap.Error(err),
				zap.String("api_key", maskAPIKey(apiKey)))
			// 继续处理，不阻塞正常请求
		}

		// 13. 重置失败计数器
		m.redis.Del(ctx, failedKey)

		// 14. 异步更新最后使用时间（不阻塞请求）
		go m.updateLastUsed(apiKey, keyData.MerchantID)

		// 15. 记录成功的认证事件
		logger.Info("API authentication successful",
			zap.String("api_key", maskAPIKey(apiKey)),
			zap.String("merchant_id", keyData.MerchantID.String()),
			zap.String("client_ip", clientIP),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path))

		// 16. 检查是否需要轮换密钥（添加响应头提醒）
		if keyData.ShouldRotate {
			c.Header("X-API-Key-Rotation-Warning", "true")
			c.Header("X-API-Key-Rotation-Message", "API key should be rotated for security")
			logger.Warn("API key rotation recommended",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("merchant_id", keyData.MerchantID.String()))
		}

		// 17. 将关键信息存入上下文，供后续使用
		c.Set("api_key", apiKey)
		c.Set("merchant_id", keyData.MerchantID)
		c.Set("environment", keyData.Environment)

		c.Next()
	}
}

// isSupportedSignatureVersion 检查签名版本是否支持
func isSupportedSignatureVersion(version string) bool {
	supportedVersions := map[string]bool{
		"v1": true, // 原始版本: timestamp + nonce + body
		"v2": true, // 增强版本: method + path + timestamp + nonce + body
	}
	return supportedVersions[version]
}

// calculateSignatureByVersion 根据版本计算签名
func (m *SignatureMiddleware) calculateSignatureByVersion(
	version string,
	apiSecret string,
	method string,
	path string,
	timestamp string,
	nonce string,
	body string,
) string {
	var signString string

	switch version {
	case "v1":
		// v1: 向后兼容，仅使用 timestamp + nonce + body
		signString = timestamp + nonce + body
	case "v2":
		// v2: 增强安全性，包含 HTTP method + path
		signString = method + path + timestamp + nonce + body
	default:
		// 不支持的版本，返回空字符串（验证会失败）
		if logger.Log != nil {
			logger.Error("Unsupported signature version in calculation",
				zap.String("version", version))
		}
		return ""
	}

	// 使用HMAC-SHA256计算签名
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(signString))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// calculateSignature 计算签名（v1版本，向后兼容）
// 签名算法：HMAC-SHA256(api_secret, timestamp + nonce + body)
// 已弃用：建议使用 calculateSignatureByVersion
func calculateSignature(apiSecret, timestamp, nonce, body string) string {
	// 构建待签名字符串
	signString := timestamp + nonce + body

	// 使用HMAC-SHA256计算签名
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(signString))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// validateTimestamp 验证时间戳（允许±2分钟的时间误差）
func (m *SignatureMiddleware) validateTimestamp(timestampStr string) error {
	// 解析时间戳（RFC3339格式）
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}

	// 计算时间差（绝对值）
	now := time.Now()
	diff := now.Sub(timestamp).Abs()

	// 验证时间戳在允许的窗口内（默认2分钟）
	if diff > m.timestampWindow {
		return fmt.Errorf("timestamp expired (window: %v)", m.timestampWindow)
	}

	return nil
}

// maskAPIKey 遮蔽API Key敏感信息（只显示前8个字符）
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:8] + "..."
}

// updateLastUsed 异步更新API Key的最后使用时间
// 这个函数在goroutine中调用，不应阻塞主请求
func (m *SignatureMiddleware) updateLastUsed(apiKey string, merchantID uuid.UUID) {
	if m.apiKeyUpdater == nil {
		// 如果没有设置更新器，只记录日志
		logger.Debug("API key updater not configured, skipping last_used_at update",
			zap.String("api_key", maskAPIKey(apiKey)),
			zap.String("merchant_id", merchantID.String()))
		return
	}

	// 使用带超时的context，避免goroutine泄漏
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.apiKeyUpdater.UpdateLastUsedAt(ctx, apiKey); err != nil {
		logger.Error("Failed to update API key last_used_at",
			zap.Error(err),
			zap.String("api_key", maskAPIKey(apiKey)),
			zap.String("merchant_id", merchantID.String()))
	} else {
		logger.Debug("API key last_used_at updated successfully",
			zap.String("api_key", maskAPIKey(apiKey)),
			zap.String("merchant_id", merchantID.String()))
	}
}

// SignRequest 为请求签名（供SDK使用，v1版本）
// 已弃用：建议使用 SignRequestV2
func SignRequest(apiKey, apiSecret, body string) map[string]string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	nonce := generateNonce()
	signature := calculateSignature(apiSecret, timestamp, nonce, body)

	return map[string]string{
		"X-API-Key":           apiKey,
		"X-Signature":         signature,
		"X-Timestamp":         timestamp,
		"X-Nonce":             nonce,
		"X-Signature-Version": "v1",
	}
}

// SignRequestV2 为请求签名（v2版本，包含HTTP method和path）
func SignRequestV2(apiKey, apiSecret, method, path, body string) map[string]string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	nonce := generateNonce()

	// v2签名算法: method + path + timestamp + nonce + body
	signString := method + path + timestamp + nonce + body
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(signString))
	signature := hex.EncodeToString(h.Sum(nil))

	return map[string]string{
		"X-API-Key":           apiKey,
		"X-Signature":         signature,
		"X-Timestamp":         timestamp,
		"X-Nonce":             nonce,
		"X-Signature-Version": "v2",
	}
}

// generateNonce 生成加密安全的随机Nonce（128位）
// 使用crypto/rand确保不可预测性，符合OWASP和NIST标准
func generateNonce() string {
	nonce := make([]byte, 16) // 128-bit nonce
	if _, err := rand.Read(nonce); err != nil {
		// 如果crypto/rand失败，使用UUID作为fallback
		logger.Error("Failed to generate random nonce, using UUID fallback", zap.Error(err))
		return uuid.New().String()
	}
	return hex.EncodeToString(nonce)
}

// SignQueryString 为查询字符串签名
// 用于GET请求的签名验证
func SignQueryString(apiSecret string, params map[string]string) string {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signParts []string
	for _, k := range keys {
		if k != "sign" { // 排除sign字段本身
			signParts = append(signParts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	signString := strings.Join(signParts, "&")

	// 计算签名
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(signString))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyQuerySignature 验证查询字符串签名
func VerifyQuerySignature(apiSecret string, params map[string]string, providedSign string) bool {
	expectedSign := SignQueryString(apiSecret, params)
	return hmac.Equal([]byte(providedSign), []byte(expectedSign))
}

// isIPInWhitelist 检查IP是否在白名单中
func isIPInWhitelist(clientIP, whitelist string) bool {
	// 如果未配置白名单，允许所有IP
	if whitelist == "" {
		return true
	}

	// 解析白名单（逗号分隔）
	allowedIPs := strings.Split(whitelist, ",")
	for _, allowedIP := range allowedIPs {
		allowedIP = strings.TrimSpace(allowedIP)
		if allowedIP == "" {
			continue
		}

		// 支持CIDR格式（如 192.168.1.0/24）
		if strings.Contains(allowedIP, "/") {
			if isIPInCIDR(clientIP, allowedIP) {
				return true
			}
		} else {
			// 精确匹配
			if clientIP == allowedIP {
				return true
			}
		}
	}

	return false
}

// isIPInCIDR 检查IP是否在CIDR范围内（标准实现）
func isIPInCIDR(clientIP, cidr string) bool {
	// 如果不包含 / 则是单个IP直接比较
	if !strings.Contains(cidr, "/") {
		return strings.TrimSpace(clientIP) == strings.TrimSpace(cidr)
	}

	// 使用标准库解析CIDR
	ip := net.ParseIP(clientIP)
	if ip == nil {
		// 无效的IP地址
		return false
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// 无效的CIDR格式
		return false
	}

	// 标准CIDR包含检查
	return ipNet.Contains(ip)
}
