package repository

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKey API密钥模型（与merchant-service中的模型保持一致）
type APIKey struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	APIKey       string     `gorm:"type:varchar(64);unique;not null;index"`
	APISecret    string     `gorm:"type:varchar(128);not null"`
	Name         string     `gorm:"type:varchar(100)"`
	Environment  string     `gorm:"type:varchar(20);not null;index"`
	IsActive     bool       `gorm:"default:true"`
	IPWhitelist  string     `gorm:"type:text"` // IP白名单，逗号分隔（可选）
	LastUsedAt   *time.Time `gorm:"type:timestamptz"`
	ExpiresAt    *time.Time `gorm:"type:timestamptz"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;default:now()"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;default:now()"`
	RotationDays int        `gorm:"default:90"` // 密钥轮换提醒天数（0表示不提醒）
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// APIKeyRepository API密钥仓储接口
type APIKeyRepository interface {
	// GetByAPIKey 根据API Key查询密钥信息
	GetByAPIKey(ctx context.Context, apiKey string) (*APIKey, error)
	// GetByMerchantID 根据商户ID获取活跃的API密钥（用于通知签名）
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*APIKey, error)
	// UpdateLastUsedAt 更新最后使用时间
	UpdateLastUsedAt(ctx context.Context, apiKey string) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository 创建API密钥仓储实例
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// GetByAPIKey 根据API Key查询密钥信息
func (r *apiKeyRepository) GetByAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
	var key APIKey
	err := r.db.WithContext(ctx).
		Where("api_key = ?", apiKey).
		First(&key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}
	return &key, nil
}

// GetByMerchantID 根据商户ID获取活跃的API密钥
func (r *apiKeyRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*APIKey, error) {
	var key APIKey
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND is_active = ? AND (expires_at IS NULL OR expires_at > ?)",
			merchantID, true, time.Now()).
		Order("created_at DESC"). // 获取最新的密钥
		First(&key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active API key found for merchant %s", merchantID)
		}
		return nil, fmt.Errorf("failed to query merchant API key: %w", err)
	}
	return &key, nil
}

// UpdateLastUsedAt 更新最后使用时间
func (r *apiKeyRepository) UpdateLastUsedAt(ctx context.Context, apiKey string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&APIKey{}).
		Where("api_key = ?", apiKey).
		Update("last_used_at", now).Error
	if err != nil {
		return fmt.Errorf("failed to update last_used_at: %w", err)
	}
	return nil
}

// IsIPAllowed 检查IP是否在白名单中
func (k *APIKey) IsIPAllowed(clientIP string) bool {
	// 如果未配置白名单，允许所有IP
	if k.IPWhitelist == "" {
		return true
	}

	// 解析白名单（逗号分隔）
	allowedIPs := strings.Split(k.IPWhitelist, ",")
	for _, allowedIP := range allowedIPs {
		allowedIP = strings.TrimSpace(allowedIP)
		if allowedIP == "" {
			continue
		}

		// 支持CIDR格式（如 192.168.1.0/24）
		if strings.Contains(allowedIP, "/") {
			if isIPInCIDR(clientIP, allowedIP) {
				return true
			}
		} else {
			// 精确匹配
			if clientIP == allowedIP {
				return true
			}
		}
	}

	return false
}

// ShouldRotate 检查是否需要轮换密钥
func (k *APIKey) ShouldRotate() bool {
	// 如果未设置轮换天数，不提醒
	if k.RotationDays <= 0 {
		return false
	}

	// 计算创建后的天数
	daysSinceCreation := int(time.Since(k.CreatedAt).Hours() / 24)
	return daysSinceCreation >= k.RotationDays
}

// isIPInCIDR 检查IP是否在CIDR范围内
func isIPInCIDR(clientIP, cidr string) bool {
	// 解析客户端IP地址
	ip := net.ParseIP(clientIP)
	if ip == nil {
		// 无效的IP地址
		return false
	}

	// 解析CIDR表示法
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// 无效的CIDR格式，降级到前缀匹配
		// 例如：192.168.1.0/24 解析失败时，尝试前缀匹配
		prefix := strings.Split(cidr, "/")[0]
		if len(prefix) > 2 {
			return strings.HasPrefix(clientIP, prefix[:len(prefix)-2])
		}
		return false
	}

	// 使用标准库的 Contains 方法检查IP是否在CIDR范围内
	return ipNet.Contains(ip)
}
