package middleware

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/idempotency"
)

// responseWriter 用于捕获响应
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware 幂等性中间件
// 使用 Idempotency-Key header 实现请求幂等性
func IdempotencyMiddleware(manager *idempotency.IdempotencyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对 POST、PUT、PATCH 方法启用幂等性检查
		if c.Request.Method != http.MethodPost &&
			c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// 获取幂等性Key
		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			// 未提供幂等性Key，正常处理
			c.Next()
			return
		}

		// 检查幂等性
		isProcessing, cachedResp, err := manager.Check(c.Request.Context(), idempotencyKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "幂等性检查失败",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// 如果正在处理中
		if isProcessing {
			c.JSON(http.StatusConflict, gin.H{
				"error": "请求正在处理中，请稍后重试",
				"idempotency_key": idempotencyKey,
			})
			c.Abort()
			return
		}

		// 如果有缓存的响应
		if cachedResp != nil {
			if cachedResp.Error != "" {
				c.JSON(cachedResp.StatusCode, gin.H{
					"error": cachedResp.Error,
					"cached": true,
					"created_at": cachedResp.CreatedAt,
				})
			} else {
				c.JSON(cachedResp.StatusCode, cachedResp.Body)
			}
			c.Abort()
			return
		}

		// 包装 ResponseWriter 以捕获响应
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = blw

		// 继续处理请求
		c.Next()

		// 请求处理完成后，缓存响应
		statusCode := c.Writer.Status()

		// 只缓存成功的响应（2xx）
		if statusCode >= 200 && statusCode < 300 {
			var responseBody interface{}

			// 尝试解析JSON响应
			bodyBytes := blw.body.Bytes()
			if len(bodyBytes) > 0 {
				// 简单起见，直接存储原始字节
				responseBody = string(bodyBytes)
			}

			// 存储响应到缓存
			errorMsg := ""
			if len(c.Errors) > 0 {
				errorMsg = c.Errors.String()
			}

			if err := manager.Store(c.Request.Context(), idempotencyKey, statusCode, responseBody, errorMsg); err != nil {
				// 缓存失败不影响正常响应，只记录日志
				// logger.Error("failed to store idempotency response", zap.Error(err))
			}
		}
	}
}
