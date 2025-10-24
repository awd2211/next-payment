package tls

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHTTPClient 创建支持mTLS的HTTP客户端
func NewHTTPClient(tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// WrapHTTPTransport 为现有Transport添加mTLS配置
func WrapHTTPTransport(transport *http.Transport, tlsConfig *tls.Config) *http.Transport {
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}
	return transport
}
