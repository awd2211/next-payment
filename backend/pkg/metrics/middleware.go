package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware 返回 Prometheus 指标收集中间件
func PrometheusMiddleware(metrics *HTTPMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取请求大小
		reqSize := 0
		if c.Request.ContentLength > 0 {
			reqSize = int(c.Request.ContentLength)
		}

		// 处理请求
		c.Next()

		// 计算处理时长
		duration := time.Since(start)

		// 获取路径（使用路由模式而不是实际路径，避免高基数）
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 获取响应大小
		respSize := c.Writer.Size()

		// 获取状态码
		status := strconv.Itoa(c.Writer.Status())

		// 记录指标
		metrics.RecordHTTPRequest(
			c.Request.Method,
			path,
			status,
			duration,
			reqSize,
			respSize,
		)
	}
}
