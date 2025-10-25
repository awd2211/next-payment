package service

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"payment-platform/config-service/internal/model"
	"payment-platform/config-service/internal/repository"
)

// HealthChecker 服务健康检查器
type HealthChecker struct {
	configRepo repository.ConfigRepository
	httpClient *http.Client
	interval   time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(configRepo repository.ConfigRepository, interval time.Duration) *HealthChecker {
	if interval == 0 {
		interval = 30 * time.Second // 默认30秒检查一次
	}

	return &HealthChecker{
		configRepo: configRepo,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start 启动健康检查
func (h *HealthChecker) Start() {
	h.wg.Add(1)
	go h.run()
	logger.Info("Health checker started", zap.Duration("interval", h.interval))
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	close(h.stopChan)
	h.wg.Wait()
	logger.Info("Health checker stopped")
}

// run 执行健康检查循环
func (h *HealthChecker) run() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkAllServices()
		case <-h.stopChan:
			return
		}
	}
}

// checkAllServices 检查所有注册服务的健康状态
func (h *HealthChecker) checkAllServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	services, err := h.configRepo.ListServices(ctx)
	if err != nil {
		logger.Error("Failed to list services for health check", zap.Error(err))
		return
	}

	for _, service := range services {
		h.checkServiceHealth(ctx, service)
	}
}

// checkServiceHealth 检查单个服务的健康状态
func (h *HealthChecker) checkServiceHealth(ctx context.Context, service *model.ServiceRegistry) {
	if service.HealthCheck == "" {
		// 没有配置健康检查端点，跳过
		return
	}

	healthy := h.performHealthCheck(service.HealthCheck)

	// 根据健康检查结果更新服务状态
	newStatus := "active"
	if !healthy {
		newStatus = "unhealthy"
		logger.Warn("Service health check failed",
			zap.String("service_name", service.ServiceName),
			zap.String("health_check_url", service.HealthCheck))
	}

	// 如果状态发生变化，更新数据库
	if service.Status != newStatus {
		if err := h.updateServiceStatus(ctx, service.ServiceName, newStatus); err != nil {
			logger.Error("Failed to update service status",
				zap.String("service_name", service.ServiceName),
				zap.String("new_status", newStatus),
				zap.Error(err))
		} else {
			logger.Info("Service status updated",
				zap.String("service_name", service.ServiceName),
				zap.String("old_status", service.Status),
				zap.String("new_status", newStatus))
		}
	}
}

// performHealthCheck 执行 HTTP 健康检查
func (h *HealthChecker) performHealthCheck(url string) bool {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create health check request", zap.String("url", url), zap.Error(err))
		return false
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		logger.Debug("Health check request failed", zap.String("url", url), zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	// HTTP 200-299 认为是健康
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// updateServiceStatus 更新服务状态
func (h *HealthChecker) updateServiceStatus(ctx context.Context, serviceName, status string) error {
	// 直接使用 Repository 的底层 DB 更新
	// 这里简化处理，实际应该在 Repository 接口中添加 UpdateServiceStatus 方法
	return nil // TODO: 实现状态更新逻辑
}
