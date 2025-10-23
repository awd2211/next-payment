package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient HTTP客户端封装
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
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
	http *HTTPClient
}

// NewServiceClient 创建微服务客户端
func NewServiceClient(baseURL string) *ServiceClient {
	return &ServiceClient{
		http: NewHTTPClient(baseURL, 30*time.Second),
	}
}
