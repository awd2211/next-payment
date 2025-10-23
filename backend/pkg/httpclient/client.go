package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client HTTP客户端
type Client struct {
	httpClient *http.Client
	config     *Config
}

// Config HTTP客户端配置
type Config struct {
	Timeout        time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	EnableLogging  bool
	DefaultHeaders map[string]string
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		EnableLogging:  true,
		DefaultHeaders: make(map[string]string),
	}
}

// NewClient 创建HTTP客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// Request HTTP请求参数
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Ctx     context.Context
}

// Response HTTP响应
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	RawResp    *http.Response
}

// Do 发送HTTP请求
func (c *Client) Do(req *Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		resp, err := c.doRequest(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 如果是最后一次尝试，不再重试
		if attempt == c.config.MaxRetries {
			break
		}

		// 只有特定错误才重试
		if !isRetryableError(err, resp) {
			break
		}

		// 等待后重试
		if c.config.RetryDelay > 0 {
			time.Sleep(c.config.RetryDelay * time.Duration(attempt+1))
		}
	}

	return nil, lastErr
}

// doRequest 执行单次HTTP请求
func (c *Client) doRequest(req *Request) (*Response, error) {
	// 准备请求体
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置上下文
	if req.Ctx != nil {
		httpReq = httpReq.WithContext(req.Ctx)
	}

	// 设置默认头
	for k, v := range c.config.DefaultHeaders {
		httpReq.Header.Set(k, v)
	}

	// 设置请求头
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// 如果有Body，设置Content-Type
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		if c.config.EnableLogging {
			fmt.Printf("[HTTP] %s %s - Error: %v (Duration: %v)\n", req.Method, req.URL, err, duration)
		}
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       respBody,
		RawResp:    httpResp,
	}

	// 日志
	if c.config.EnableLogging {
		fmt.Printf("[HTTP] %s %s - %d (Duration: %v, Size: %d bytes)\n",
			req.Method, req.URL, httpResp.StatusCode, duration, len(respBody))
	}

	// 检查HTTP状态码
	if httpResp.StatusCode >= 400 {
		return resp, fmt.Errorf("HTTP错误: %d %s", httpResp.StatusCode, http.StatusText(httpResp.StatusCode))
	}

	return resp, nil
}

// Get 发送GET请求
func (c *Client) Get(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headers,
	})
}

// Post 发送POST请求
func (c *Client) Post(url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  http.MethodPost,
		URL:     url,
		Body:    body,
		Headers: headers,
	})
}

// Put 发送PUT请求
func (c *Client) Put(url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  http.MethodPut,
		URL:     url,
		Body:    body,
		Headers: headers,
	})
}

// Delete 发送DELETE请求
func (c *Client) Delete(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: headers,
	})
}

// ParseJSON 解析JSON响应
func (r *Response) ParseJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String 返回响应体字符串
func (r *Response) String() string {
	return string(r.Body)
}

// isRetryableError 判断错误是否可重试
func isRetryableError(err error, resp *Response) bool {
	if err != nil {
		// 网络错误通常可以重试
		return true
	}

	if resp != nil {
		// 5xx 服务器错误可以重试
		// 429 Too Many Requests 可以重试
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			return true
		}
	}

	return false
}
