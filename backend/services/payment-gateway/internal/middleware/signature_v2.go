package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"payment-platform/payment-gateway/internal/client"
)

// SignatureMiddlewareV2 新版签名验证中间件
// 委托给 merchant-auth-service 进行签名验证
type SignatureMiddlewareV2 struct {
	authClient client.MerchantAuthClient
}

// NewSignatureMiddlewareV2 创建新版签名验证中间件
func NewSignatureMiddlewareV2(authClient client.MerchantAuthClient) *SignatureMiddlewareV2 {
	return &SignatureMiddlewareV2{
		authClient: authClient,
	}
}

// Verify 验证签名
func (m *SignatureMiddlewareV2) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. 提取请求头
		apiKey := c.GetHeader("X-API-Key")
		signature := c.GetHeader("X-Signature")

		if apiKey == "" || signature == "" {
			logger.Warn("Missing authentication headers",
				zap.String("client_ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		// 2. 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 3. 调用 merchant-auth-service 验证签名
		result, err := m.authClient.ValidateSignature(ctx, apiKey, signature, string(bodyBytes))
		if err != nil {
			logger.Warn("Signature validation failed",
				zap.Error(err),
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("client_ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		if !result.Valid {
			logger.Warn("Invalid signature",
				zap.String("api_key", maskAPIKey(apiKey)),
				zap.String("client_ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		// 4. 签名验证成功，设置上下文
		logger.Info("API authentication successful",
			zap.String("api_key", maskAPIKey(apiKey)),
			zap.String("merchant_id", result.MerchantID.String()),
			zap.String("client_ip", c.ClientIP()))

		c.Set("api_key", apiKey)
		c.Set("merchant_id", result.MerchantID)
		c.Set("environment", result.Environment)

		c.Next()
	}
}
