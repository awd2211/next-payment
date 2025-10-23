package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SignatureMiddleware API签名验证中间件
// 用于验证商户请求的合法性
type SignatureMiddleware struct {
	// 获取商户API Secret的函数
	getAPISecret func(apiKey string) (string, error)
}

// NewSignatureMiddleware 创建签名验证中间件
func NewSignatureMiddleware(getAPISecret func(apiKey string) (string, error)) *SignatureMiddleware {
	return &SignatureMiddleware{
		getAPISecret: getAPISecret,
	}
}

// Verify 验证签名
func (m *SignatureMiddleware) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取签名相关的头部信息
		apiKey := c.GetHeader("X-API-Key")
		signature := c.GetHeader("X-Signature")
		timestamp := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")

		// 验证必要参数
		if apiKey == "" || signature == "" || timestamp == "" || nonce == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少必要的签名参数",
			})
			c.Abort()
			return
		}

		// 验证时间戳（防止重放攻击）
		if err := validateTimestamp(timestamp); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("时间戳验证失败: %v", err),
			})
			c.Abort()
			return
		}

		// 获取API Secret
		apiSecret, err := m.getAPISecret(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的API Key",
			})
			c.Abort()
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "读取请求体失败",
			})
			c.Abort()
			return
		}

		// 重置请求体供后续处理使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 计算签名
		expectedSignature := calculateSignature(apiSecret, timestamp, nonce, string(body))

		// 验证签名
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "签名验证失败",
			})
			c.Abort()
			return
		}

		// 将API Key存入上下文，供后续使用
		c.Set("api_key", apiKey)

		c.Next()
	}
}

// calculateSignature 计算签名
// 签名算法：HMAC-SHA256(api_secret, timestamp + nonce + body)
func calculateSignature(apiSecret, timestamp, nonce, body string) string {
	// 构建待签名字符串
	signString := timestamp + nonce + body

	// 使用HMAC-SHA256计算签名
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(signString))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// validateTimestamp 验证时间戳（允许±5分钟的时间误差）
func validateTimestamp(timestampStr string) error {
	// 解析时间戳
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return fmt.Errorf("无效的时间戳格式")
	}

	// 计算时间差
	now := time.Now()
	diff := now.Sub(timestamp).Abs()

	// 允许±5分钟的误差
	if diff > 5*time.Minute {
		return fmt.Errorf("时间戳已过期")
	}

	return nil
}

// SignRequest 为请求签名（供SDK使用）
func SignRequest(apiKey, apiSecret, body string) map[string]string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	nonce := generateNonce()
	signature := calculateSignature(apiSecret, timestamp, nonce, body)

	return map[string]string{
		"X-API-Key":   apiKey,
		"X-Signature": signature,
		"X-Timestamp": timestamp,
		"X-Nonce":     nonce,
	}
}

// generateNonce 生成随机字符串
func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
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
