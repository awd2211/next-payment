package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/payment-platform/pkg/httpclient"
)

// ServiceClient 微服务客户端
type ServiceClient struct {
	baseURL    string
	httpClient *httpclient.Client
}

// NewServiceClient 创建微服务客户端
func NewServiceClient(baseURL string) *ServiceClient {
	return &ServiceClient{
		baseURL:    baseURL,
		httpClient: httpclient.NewClient(httpclient.DefaultConfig()),
	}
}

// Get 发送GET请求
func (c *ServiceClient) Get(ctx context.Context, path string, queryParams map[string]string) (map[string]interface{}, int, error) {
	// 构建完整URL
	fullURL := c.buildURL(path, queryParams)

	// 发送请求
	resp, err := c.httpClient.Do(&httpclient.Request{
		Method: "GET",
		URL:    fullURL,
		Ctx:    ctx,
	})
	if err != nil {
		return nil, 0, err
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// Post 发送POST请求
func (c *ServiceClient) Post(ctx context.Context, path string, body interface{}) (map[string]interface{}, int, error) {
	fullURL := c.buildURL(path, nil)

	resp, err := c.httpClient.Do(&httpclient.Request{
		Method: "POST",
		URL:    fullURL,
		Body:   body,
		Ctx:    ctx,
	})
	if err != nil {
		return nil, 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// Put 发送PUT请求
func (c *ServiceClient) Put(ctx context.Context, path string, body interface{}) (map[string]interface{}, int, error) {
	fullURL := c.buildURL(path, nil)

	resp, err := c.httpClient.Do(&httpclient.Request{
		Method: "PUT",
		URL:    fullURL,
		Body:   body,
		Ctx:    ctx,
	})
	if err != nil {
		return nil, 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// Delete 发送DELETE请求
func (c *ServiceClient) Delete(ctx context.Context, path string) (map[string]interface{}, int, error) {
	fullURL := c.buildURL(path, nil)

	resp, err := c.httpClient.Do(&httpclient.Request{
		Method: "DELETE",
		URL:    fullURL,
		Ctx:    ctx,
	})
	if err != nil {
		return nil, 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// buildURL 构建完整URL
func (c *ServiceClient) buildURL(path string, queryParams map[string]string) string {
	fullURL := c.baseURL + path

	if len(queryParams) > 0 {
		values := url.Values{}
		for k, v := range queryParams {
			values.Add(k, v)
		}
		fullURL += "?" + values.Encode()
	}

	return fullURL
}
