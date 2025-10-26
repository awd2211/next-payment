package configclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/payment-platform/pkg/httpclient"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
)

// Client 配置客户端
type Client struct {
	serviceName    string
	environment    string
	configURL      string
	httpClient     *httpclient.Client // 用于非mTLS场景
	rawHTTPClient  *http.Client        // 用于mTLS场景
	enableMTLS     bool
	cache          *ConfigCache
	refreshRate    time.Duration
	mu             sync.RWMutex
	stopCh         chan struct{}
	updateHooks    []func(key, value string)
}

// Config 配置项
type Config struct {
	Key         string `json:"config_key"`
	Value       string `json:"config_value"`
	ValueType   string `json:"value_type"`
	IsEncrypted bool   `json:"is_encrypted"`
}

// ConfigCache 配置缓存
type ConfigCache struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewConfigCache 创建配置缓存
func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		data: make(map[string]string),
	}
}

// Get 获取配置
func (c *ConfigCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// Set 设置配置
func (c *ConfigCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// GetAll 获取所有配置
func (c *ConfigCache) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// ClientConfig 客户端配置
type ClientConfig struct {
	ServiceName string        // 服务名称
	Environment string        // 环境 (development, production)
	ConfigURL   string        // config-service URL
	RefreshRate time.Duration // 刷新频率 (默认 30s)

	// mTLS 配置 (可选)
	EnableMTLS  bool   // 是否启用 mTLS
	TLSCertFile string // 客户端证书文件路径
	TLSKeyFile  string // 客户端私钥文件路径
	TLSCAFile   string // CA 证书文件路径
}

// NewClient 创建配置客户端
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if cfg.ConfigURL == "" {
		cfg.ConfigURL = "http://localhost:40010" // 默认 config-service 地址
	}
	if cfg.Environment == "" {
		cfg.Environment = "production"
	}
	if cfg.RefreshRate == 0 {
		cfg.RefreshRate = 30 * time.Second
	}

	// 创建 HTTP 客户端 (支持 mTLS)
	var httpClient *httpclient.Client
	var rawHTTPClient *http.Client
	var enableMTLS bool

	if cfg.EnableMTLS {
		// 加载 mTLS 配置
		tlsConfig, err := loadMTLSConfig(cfg.TLSCertFile, cfg.TLSKeyFile, cfg.TLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load mTLS config: %w", err)
		}

		// 创建原生 HTTP 客户端with mTLS
		rawHTTPClient = &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
		enableMTLS = true

		logger.Info("Config client mTLS enabled",
			zap.String("cert", cfg.TLSCertFile),
			zap.String("ca", cfg.TLSCAFile))
	} else {
		// 使用标准 httpclient
		httpClient = httpclient.NewClient(&httpclient.Config{
			Timeout:       10 * time.Second,
			MaxRetries:    3,
			RetryDelay:    1 * time.Second,
			EnableLogging: true,
		})
		enableMTLS = false
	}

	client := &Client{
		serviceName:   cfg.ServiceName,
		environment:   cfg.Environment,
		configURL:     cfg.ConfigURL,
		httpClient:    httpClient,
		rawHTTPClient: rawHTTPClient,
		enableMTLS:    enableMTLS,
		cache:         NewConfigCache(),
		refreshRate:   cfg.RefreshRate,
		stopCh:        make(chan struct{}),
		updateHooks:   make([]func(key, value string), 0),
	}

	// 初始加载配置
	if err := client.loadConfigs(context.Background()); err != nil {
		logger.Warn("Failed to load initial configs, will retry", zap.Error(err))
	}

	// 启动定时刷新
	go client.refreshLoop()

	logger.Info("Config client initialized",
		zap.String("service", cfg.ServiceName),
		zap.String("env", cfg.Environment),
		zap.String("url", cfg.ConfigURL))

	return client, nil
}

// Get 获取配置值
func (c *Client) Get(key string) string {
	if val, ok := c.cache.Get(key); ok {
		return val
	}
	logger.Warn("Config key not found in cache", zap.String("key", key))
	return ""
}

// GetWithDefault 获取配置值,如果不存在则返回默认值
func (c *Client) GetWithDefault(key, defaultValue string) string {
	if val := c.Get(key); val != "" {
		return val
	}
	return defaultValue
}

// GetInt 获取整数配置
func (c *Client) GetInt(key string, defaultValue int) int {
	val := c.Get(key)
	if val == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		logger.Warn("Failed to parse int config", zap.String("key", key), zap.Error(err))
		return defaultValue
	}
	return result
}

// GetBool 获取布尔配置
func (c *Client) GetBool(key string, defaultValue bool) bool {
	val := c.Get(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1" || val == "yes"
}

// OnUpdate 注册配置更新回调
func (c *Client) OnUpdate(hook func(key, value string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updateHooks = append(c.updateHooks, hook)
}

// loadConfigs 加载配置
func (c *Client) loadConfigs(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/configs?service_name=%s&environment=%s",
		c.configURL, c.serviceName, c.environment)

	var bodyBytes []byte

	// 根据是否启用 mTLS 使用不同的客户端
	if c.enableMTLS {
		// 使用原生 http.Client (支持 mTLS)
		httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		httpResp, err := c.rawHTTPClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("failed to fetch configs: %w", err)
		}
		defer httpResp.Body.Close()

		bodyBytes, err = io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
	} else {
		// 使用标准 httpclient
		resp, err := c.httpClient.Get(url, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch configs: %w", err)
		}
		bodyBytes = resp.Body
	}

	var response struct {
		Code interface{} `json:"code"` // 支持 int 或 string 类型
		Data struct {
			Items []Config `json:"items"`
			List  []Config `json:"list"` // 兼容不同的字段名
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("failed to parse config response: %w", err)
	}

	// 验证响应code (支持string或int类型)
	codeStr := fmt.Sprintf("%v", response.Code)
	if codeStr != "0" && codeStr != "SUCCESS" {
		return fmt.Errorf("failed to fetch configs, code: %v", response.Code)
	}

	// 合并items和list (兼容不同的响应格式)
	items := response.Data.Items
	if len(items) == 0 && len(response.Data.List) > 0 {
		items = response.Data.List
	}

	// 更新缓存
	oldConfigs := c.cache.GetAll()
	for _, cfg := range items {
		oldValue, existed := oldConfigs[cfg.Key]
		c.cache.Set(cfg.Key, cfg.Value)

		// 触发更新回调
		if !existed || oldValue != cfg.Value {
			c.notifyUpdate(cfg.Key, cfg.Value)
		}
	}

	logger.Debug("Configs loaded",
		zap.String("service", c.serviceName),
		zap.Int("count", len(items)))

	return nil
}

// notifyUpdate 通知配置更新
func (c *Client) notifyUpdate(key, value string) {
	c.mu.RLock()
	hooks := c.updateHooks
	c.mu.RUnlock()

	for _, hook := range hooks {
		go func(h func(string, string)) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Config update hook panic", zap.Any("panic", r))
				}
			}()
			h(key, value)
		}(hook)
	}
}

// refreshLoop 定时刷新配置
func (c *Client) refreshLoop() {
	ticker := time.NewTicker(c.refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.loadConfigs(context.Background()); err != nil {
				logger.Error("Failed to refresh configs", zap.Error(err))
			}
		case <-c.stopCh:
			logger.Info("Config refresh loop stopped")
			return
		}
	}
}

// Stop 停止客户端
func (c *Client) Stop() {
	close(c.stopCh)
}

// GetAllConfigs 获取所有配置(用于调试)
func (c *Client) GetAllConfigs() map[string]string {
	return c.cache.GetAll()
}

// loadMTLSConfig 加载 mTLS 配置
func loadMTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// 加载客户端证书和私钥
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key: %w", err)
	}

	// 加载 CA 证书
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}
