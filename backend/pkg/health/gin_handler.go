package health

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// GinHandler Gin健康检查处理器
type GinHandler struct {
	checker *HealthChecker
	timeout time.Duration
}

// NewGinHandler 创建Gin健康检查处理器
func NewGinHandler(checker *HealthChecker) *GinHandler {
	return &GinHandler{
		checker: checker,
		timeout: 10 * time.Second, // 默认10秒超时
	}
}

// WithTimeout 设置超时时间
func (h *GinHandler) WithTimeout(timeout time.Duration) *GinHandler {
	h.timeout = timeout
	return h
}

// Handle 处理健康检查请求
func (h *GinHandler) Handle(c *gin.Context) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	// 执行健康检查
	report := h.checker.Check(ctx)

	// 返回结果
	c.JSON(report.GetStatusCode(), gin.H{
		"status":    report.Status,
		"timestamp": report.Timestamp.Format(time.RFC3339),
		"duration":  report.Duration.String(),
		"checks":    report.Checks,
	})
}

// HandleLiveness 处理存活探针（Kubernetes Liveness Probe）
// 只检查服务本身是否存活，不检查依赖
func (h *GinHandler) HandleLiveness(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HandleReadiness 处理就绪探针（Kubernetes Readiness Probe）
// 检查服务是否准备好接收流量（包括依赖检查）
func (h *GinHandler) HandleReadiness(c *gin.Context) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	// 执行健康检查
	report := h.checker.Check(ctx)

	// 只有完全健康才返回200
	if report.Status == StatusHealthy {
		c.JSON(200, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		c.JSON(503, gin.H{
			"status":    "not_ready",
			"reason":    report.Status,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// SimpleHealthHandler 简单健康检查处理器（向后兼容）
func SimpleHealthHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   serviceName,
			"timestamp": time.Now().Unix(),
		})
	}
}
