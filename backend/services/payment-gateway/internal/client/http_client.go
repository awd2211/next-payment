package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/payment-platform/pkg/httpclient"
	pkgtls "github.com/payment-platform/pkg/tls"
)

// HTTPClient HTTP客户端封装
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient 创建HTTP客户端（支持mTLS）
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// 加载 TLS 配置
	tlsConfig := pkgtls.LoadFromEnv()
	var httpClient *http.Client

	if tlsConfig.EnableMTLS {
		// 验证客户端 TLS 配置
		if err := pkgtls.ValidateClientConfig(tlsConfig); err != nil {
			// 降级到普通 HTTP（记录错误）
			fmt.Printf("WARNING: mTLS 配置验证失败，降级到普通 HTTP: %v\n", err)
			httpClient = &http.Client{Timeout: timeout}
		} else {
			// 创建支持 mTLS 的客户端
			clientTLSConfig, err := pkgtls.NewClientTLSConfig(tlsConfig)
			if err != nil {
				fmt.Printf("WARNING: mTLS 配置失败，降级到普通 HTTP: %v\n", err)
				httpClient = &http.Client{Timeout: timeout}
			} else {
				httpClient = pkgtls.NewHTTPClient(clientTLSConfig, timeout)
			}
		}
	} else {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &HTTPClient{
		client:  httpClient,
		baseURL: baseURL,
	}
}

// Request HTTP请求结构
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
}

// Response HTTP响应结构
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do 执行HTTP请求
func (c *HTTPClient) Do(ctx context.Context, req *Request) (*Response, error) {
	// 构建完整URL
	url := c.baseURL + req.Path

	// 序列化请求体
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置默认头部
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// 设置自定义头部
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 执行请求
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("执行HTTP请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return &Response{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    httpResp.Header,
	}, nil
}

// Get 执行GET请求
func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodGet,
		Path:    path,
		Headers: headers,
	})
}

// Post 执行POST请求
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodPost,
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// Put 执行PUT请求
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodPut,
		Path:    path,
		Headers: headers,
		Body:    body,
	})
}

// Delete 执行DELETE请求
func (c *HTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	})
}

// ParseResponse 解析响应
func (r *Response) ParseResponse(v interface{}) error {
	if r.StatusCode >= 400 {
		return fmt.Errorf("HTTP错误: status=%d, body=%s", r.StatusCode, string(r.Body))
	}

	if v != nil && len(r.Body) > 0 {
		if err := json.Unmarshal(r.Body, v); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

// ServiceClient 微服务客户端基类
type ServiceClient struct {
	http    *HTTPClient
	breaker *httpclient.BreakerClient
	baseURL string // 保存baseURL用于构建完整URL
}

// NewServiceClient 创建微服务客户端（不带熔断器，用于向后兼容）
func NewServiceClient(baseURL string) *ServiceClient {
	return &ServiceClient{
		http:    NewHTTPClient(baseURL, 30*time.Second),
		baseURL: baseURL,
	}
}

// NewServiceClientWithBreaker 创建带熔断器的微服务客户端
func NewServiceClientWithBreaker(baseURL string, breakerName string) *ServiceClient {
	// 创建 pkg/httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig(breakerName)

	// 创建带熔断器的客户端
	breakerClient := httpclient.NewBreakerClient(config, breakerConfig)

	return &ServiceClient{
		http:    NewHTTPClient(baseURL, 30*time.Second), // 保留旧客户端用于向后兼容
		breaker: breakerClient,
		baseURL: baseURL,
	}
}

// Get 执行GET请求（自动使用熔断器）
func (sc *ServiceClient) Get(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	if sc.breaker != nil {
		return sc.doWithBreaker(ctx, http.MethodGet, path, nil, headers)
	}
	return sc.http.Get(ctx, path, headers)
}

// Post 执行POST请求（自动使用熔断器）
func (sc *ServiceClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	if sc.breaker != nil {
		return sc.doWithBreaker(ctx, http.MethodPost, path, body, headers)
	}
	return sc.http.Post(ctx, path, body, headers)
}

// Put 执行PUT请求（自动使用熔断器）
func (sc *ServiceClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	if sc.breaker != nil {
		return sc.doWithBreaker(ctx, http.MethodPut, path, body, headers)
	}
	return sc.http.Put(ctx, path, body, headers)
}

// Delete 执行DELETE请求（自动使用熔断器）
func (sc *ServiceClient) Delete(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	if sc.breaker != nil {
		return sc.doWithBreaker(ctx, http.MethodDelete, path, nil, headers)
	}
	return sc.http.Delete(ctx, path, headers)
}

// doWithBreaker 通过熔断器执行请求
func (sc *ServiceClient) doWithBreaker(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*Response, error) {
	// 构建完整URL
	fullURL := sc.baseURL + path

	// 创建 pkg/httpclient 请求
	req := &httpclient.Request{
		Method:  method,
		URL:     fullURL,
		Body:    body,
		Headers: headers,
		Ctx:     ctx,
	}

	// 通过熔断器执行请求
	resp, err := sc.breaker.Do(req)
	if err != nil {
		return nil, err
	}

	// 转换响应格式
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}, nil
}
