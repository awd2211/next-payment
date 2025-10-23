package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/services/channel-adapter/internal/adapter"
	"github.com/payment-platform/services/channel-adapter/internal/handler"
	"github.com/payment-platform/services/channel-adapter/internal/model"
	"github.com/payment-platform/services/channel-adapter/internal/repository"
	"github.com/payment-platform/services/channel-adapter/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	config := loadConfig()

	// 连接数据库
	db, err := connectDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&model.ChannelConfig{},
		&model.Transaction{},
		&model.WebhookLog{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建适配器工厂
	adapterFactory := adapter.NewAdapterFactory()

	// 注册 Stripe 适配器（使用默认配置，实际应从数据库加载）
	// 这里只是示例，实际使用时应该为每个商户动态加载配置
	stripeConfig := &model.StripeConfig{
		APIKey:              os.Getenv("STRIPE_API_KEY"),
		WebhookSecret:       os.Getenv("STRIPE_WEBHOOK_SECRET"),
		PublishableKey:      os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		StatementDescriptor: "Payment Platform",
		CaptureMethod:       "automatic",
	}
	stripeAdapter := adapter.NewStripeAdapter(stripeConfig)
	adapterFactory.Register(model.ChannelStripe, stripeAdapter)

	// 创建仓储层
	channelRepo := repository.NewChannelRepository(db)

	// 创建服务层
	channelService := service.NewChannelService(channelRepo, adapterFactory)

	// 创建处理器层
	channelHandler := handler.NewChannelHandler(channelService)

	// 创建 HTTP 服务器
	router := gin.Default()

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册路由
	channelHandler.RegisterRoutes(router)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Channel Adapter 服务启动在 %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// Config 配置结构
type Config struct {
	DatabaseURL string
	ServerPort  string
}

// loadConfig 加载配置
func loadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payment_platform?sslmode=disable"),
		ServerPort:  getEnv("PORT", "8003"),
	}
}

// connectDB 连接数据库
func connectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// loadChannelConfigs 加载渠道配置（示例函数）
// 实际使用时应该从数据库为每个商户加载配置
func loadChannelConfigs(db *gorm.DB) map[string]*model.ChannelConfig {
	configs := make(map[string]*model.ChannelConfig)

	var configList []*model.ChannelConfig
	if err := db.Where("is_enabled = true").Find(&configList).Error; err != nil {
		log.Printf("加载渠道配置失败: %v", err)
		return configs
	}

	for _, config := range configList {
		key := fmt.Sprintf("%s_%s", config.MerchantID.String(), config.Channel)
		configs[key] = config
	}

	return configs
}

// parseStripeConfig 解析 Stripe 配置
func parseStripeConfig(configJSON string) (*model.StripeConfig, error) {
	var config model.StripeConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}
	return &config, nil
}
