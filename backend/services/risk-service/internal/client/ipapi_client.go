package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// IPAPIClient ipapi.co HTTP 客户端
type IPAPIClient struct {
	baseURL    string
	httpClient *http.Client
	redis      *redis.Client
	cacheTTL   time.Duration
}

// GeoIPInfo IP地理位置信息
type GeoIPInfo struct {
	IP            string  `json:"ip"`
	City          string  `json:"city"`
	Region        string  `json:"region"`
	RegionCode    string  `json:"region_code"`
	Country       string  `json:"country_name"`
	CountryCode   string  `json:"country_code"`
	ContinentCode string  `json:"continent_code"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timezone      string  `json:"timezone"`
	IsEU          bool    `json:"is_eu"`

	// 代理检测（ipapi.co 免费版不提供，需要付费）
	// IsProxy       bool    `json:"is_proxy"`
	// IsAnonymous   bool    `json:"is_anonymous"`
}

// NewIPAPIClient 创建 ipapi.co 客户端
func NewIPAPIClient(redis *redis.Client, cacheTTL time.Duration) *IPAPIClient {
	return &IPAPIClient{
		baseURL: "https://ipapi.co",
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // 2秒超时
		},
		redis:    redis,
		cacheTTL: cacheTTL,
	}
}

// LookupIP 查询IP地理位置（带缓存）
func (c *IPAPIClient) LookupIP(ctx context.Context, ip string) (*GeoIPInfo, error) {
	// 1. 尝试从缓存读取
	cacheKey := fmt.Sprintf("geoip:%s", ip)
	cached, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var info GeoIPInfo
		if err := json.Unmarshal([]byte(cached), &info); err == nil {
			logger.Debug("GeoIP 缓存命中", zap.String("ip", ip), zap.String("country", info.CountryCode))
			return &info, nil
		}
	}

	// 2. 调用 ipapi.co API
	url := fmt.Sprintf("%s/%s/json/", c.baseURL, ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 User-Agent（ipapi.co 要求）
	req.Header.Set("User-Agent", "payment-platform-risk-service/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API 调用失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		// 超出免费额度（1000次/天）
		logger.Warn("ipapi.co 超出免费额度", zap.String("ip", ip))
		return c.getFallbackGeoInfo(ip), nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 返回错误: status=%d", resp.StatusCode)
	}

	// 3. 解析响应
	var info GeoIPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 4. 写入缓存（24小时）
	if data, err := json.Marshal(info); err == nil {
		c.redis.Set(ctx, cacheKey, string(data), c.cacheTTL)
	}

	logger.Info("GeoIP 查询成功",
		zap.String("ip", ip),
		zap.String("country", info.CountryCode),
		zap.String("city", info.City))

	return &info, nil
}

// getFallbackGeoInfo 降级方案：返回默认信息
func (c *IPAPIClient) getFallbackGeoInfo(ip string) *GeoIPInfo {
	return &GeoIPInfo{
		IP:          ip,
		Country:     "Unknown",
		CountryCode: "XX",
		City:        "Unknown",
	}
}

// IsHighRiskCountry 判断是否为高风险国家
func IsHighRiskCountry(countryCode string) bool {
	// 高风险国家列表（示例）
	highRiskCountries := map[string]bool{
		"KP": true, // 朝鲜
		"IR": true, // 伊朗
		"SY": true, // 叙利亚
		"SD": true, // 苏丹
		"CU": true, // 古巴
		// 根据实际业务需求配置
	}

	return highRiskCountries[countryCode]
}

// IsHighRiskIP 判断IP是否属于高风险段
func IsHighRiskIP(ip string) bool {
	// 检查是否为已知的高风险IP段（Tor节点、代理等）
	// 这个列表应该从数据库或配置文件加载
	highRiskPrefixes := []string{
		"104.200.", // Tor 节点示例
		"185.220.", // Tor 节点示例
	}

	for _, prefix := range highRiskPrefixes {
		if len(ip) >= len(prefix) && ip[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}
