package health

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPChecker HTTP服务健康检查器
type HTTPChecker struct {
	name        string
	url         string
	method      string
	timeout     time.Duration
	client      *http.Client
	expectedCode int
}

// NewHTTPChecker 创建HTTP服务健康检查器
func NewHTTPChecker(name, url string) *HTTPChecker {
	return &HTTPChecker{
		name:         name,
		url:          url,
		method:       http.MethodGet,
		timeout:      5 * time.Second,
		expectedCode: http.StatusOK,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// WithMethod 设置HTTP方法
func (c *HTTPChecker) WithMethod(method string) *HTTPChecker {
	c.method = method
	return c
}

// WithTimeout 设置超时时间
func (c *HTTPChecker) WithTimeout(timeout time.Duration) *HTTPChecker {
	c.timeout = timeout
	c.client.Timeout = timeout
	return c
}

// WithExpectedCode 设置预期的HTTP状态码
func (c *HTTPChecker) WithExpectedCode(code int) *HTTPChecker {
	c.expectedCode = code
	return c
}

// Name 返回检查器名称
func (c *HTTPChecker) Name() string {
	return c.name
}

// Check 执行HTTP服务健康检查
func (c *HTTPChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  make(map[string]interface{}),
	}

	result.Metadata["url"] = c.url
	result.Metadata["method"] = c.method

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, c.method, c.url, nil)
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "创建HTTP请求失败"
		return result
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("HTTP请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Duration = time.Since(startTime)
	result.Metadata["status_code"] = resp.StatusCode
	result.Metadata["response_time_ms"] = result.Duration.Milliseconds()

	// 检查状态码
	if resp.StatusCode != c.expectedCode {
		// 如果返回5xx，认为不健康
		if resp.StatusCode >= 500 {
			result.Status = StatusUnhealthy
			result.Message = fmt.Sprintf("服务返回错误状态码: %d", resp.StatusCode)
		} else {
			// 其他非预期状态码认为是降级
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("状态码不符合预期: 期望 %d, 实际 %d", c.expectedCode, resp.StatusCode)
		}
		return result
	}

	// 检查响应时间
	if result.Duration > 3*time.Second {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("响应时间过长: %v", result.Duration)
		return result
	}

	result.Status = StatusHealthy
	result.Message = "服务正常"

	return result
}

// ServiceHealthChecker 微服务健康检查器（检查/health端点）
type ServiceHealthChecker struct {
	name        string
	serviceURL  string
	timeout     time.Duration
	client      *http.Client
}

// NewServiceHealthChecker 创建微服务健康检查器
func NewServiceHealthChecker(name, serviceURL string) *ServiceHealthChecker {
	return &ServiceHealthChecker{
		name:       name,
		serviceURL: serviceURL,
		timeout:    5 * time.Second,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// WithTimeout 设置超时时间
func (c *ServiceHealthChecker) WithTimeout(timeout time.Duration) *ServiceHealthChecker {
	c.timeout = timeout
	c.client.Timeout = timeout
	return c
}

// Name 返回检查器名称
func (c *ServiceHealthChecker) Name() string {
	return c.name
}

// Check 执行微服务健康检查
func (c *ServiceHealthChecker) Check(ctx context.Context) *CheckResult {
	startTime := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// 构建健康检查URL
	healthURL := c.serviceURL + "/health"
	result.Metadata["url"] = healthURL

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "创建健康检查请求失败"
		return result
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("服务 %s 不可达", c.name)
		return result
	}
	defer resp.Body.Close()

	result.Duration = time.Since(startTime)
	result.Metadata["status_code"] = resp.StatusCode
	result.Metadata["response_time_ms"] = result.Duration.Milliseconds()

	// 检查状态码
	if resp.StatusCode == http.StatusOK {
		result.Status = StatusHealthy
		result.Message = "服务健康"
	} else if resp.StatusCode == http.StatusServiceUnavailable {
		result.Status = StatusUnhealthy
		result.Message = "服务不可用"
	} else {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("服务状态异常: %d", resp.StatusCode)
	}

	return result
}
