package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
)

// ServiceClient 微服务客户端
type ServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewServiceClient 创建微服务客户端(支持mTLS)
func NewServiceClient(baseURL string) *ServiceClient {
	// 检查是否启用mTLS
	enableMTLS := os.Getenv("ENABLE_MTLS") == "true"

	var httpClient *http.Client

	if enableMTLS {
		// 配置mTLS客户端
		tlsConfig, err := createMTLSConfig()
		if err != nil {
			logger.Error("创建mTLS配置失败,将使用标准HTTP客户端", zap.Error(err))
			httpClient = &http.Client{
				Timeout: 30 * time.Second,
			}
		} else {
			// 创建带mTLS的HTTP客户端
			httpClient = &http.Client{
				Timeout: 30 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: tlsConfig,
				},
			}
			logger.Info("ServiceClient已启用mTLS", zap.String("baseURL", baseURL))
		}
	} else {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &ServiceClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// createMTLSConfig 创建mTLS TLS配置
func createMTLSConfig() (*tls.Config, error) {
	// 读取客户端证书
	certFile := os.Getenv("TLS_CLIENT_CERT")
	keyFile := os.Getenv("TLS_CLIENT_KEY")
	caFile := os.Getenv("TLS_CA_FILE")

	if certFile == "" || keyFile == "" || caFile == "" {
		return nil, fmt.Errorf("mTLS环境变量未配置: TLS_CLIENT_CERT=%s, TLS_CLIENT_KEY=%s, TLS_CA_FILE=%s", certFile, keyFile, caFile)
	}

	// 加载客户端证书
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("加载客户端证书失败: %w", err)
	}

	// 加载CA证书
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("读取CA证书失败: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("解析CA证书失败")
	}

	logger.Info("mTLS客户端配置成功",
		zap.String("cert", certFile),
		zap.String("key", keyFile),
		zap.String("ca", caFile),
	)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// Get 发送GET请求
func (c *ServiceClient) Get(ctx context.Context, path string, queryParams map[string]string) (map[string]interface{}, int, error) {
	// 构建完整URL
	fullURL := c.buildURL(path, queryParams)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// Post 发送POST请求
func (c *ServiceClient) Post(ctx context.Context, path string, body interface{}) (map[string]interface{}, int, error) {
	fullURL := c.buildURL(path, nil)

	// 序列化请求体
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// Put 发送PUT请求
func (c *ServiceClient) Put(ctx context.Context, path string, body interface{}) (map[string]interface{}, int, error) {
	fullURL := c.buildURL(path, nil)

	// 序列化请求体
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "PUT", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, resp.StatusCode, nil
}

// Delete 发送DELETE请求
func (c *ServiceClient) Delete(ctx context.Context, path string) (map[string]interface{}, int, error) {
	fullURL := c.buildURL(path, nil)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "DELETE", fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
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
