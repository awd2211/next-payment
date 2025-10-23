package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/services/notification-service/internal/handler"
	"github.com/payment-platform/services/notification-service/internal/model"
	"github.com/payment-platform/services/notification-service/internal/provider"
	"github.com/payment-platform/services/notification-service/internal/repository"
	"github.com/payment-platform/services/notification-service/internal/service"
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
		&model.Notification{},
		&model.NotificationTemplate{},
		&model.WebhookEndpoint{},
		&model.WebhookDelivery{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建邮件提供商工厂
	emailFactory := provider.NewEmailProviderFactory()

	// 注册 SMTP 提供商
	if config.SMTPHost != "" {
		smtpProvider := provider.NewSMTPProvider(
			config.SMTPHost,
			config.SMTPPort,
			config.SMTPUsername,
			config.SMTPPassword,
			config.SMTPFrom,
		)
		emailFactory.Register("smtp", smtpProvider)
	}

	// 注册 Mailgun 提供商
	if config.MailgunDomain != "" {
		mailgunProvider := provider.NewMailgunProvider(
			config.MailgunDomain,
			config.MailgunAPIKey,
			config.MailgunFrom,
		)
		emailFactory.Register("mailgun", mailgunProvider)
	}

	// 创建短信提供商工厂
	smsFactory := provider.NewSMSProviderFactory()

	// 注册 Twilio 提供商
	if config.TwilioAccountSID != "" {
		twilioProvider := provider.NewTwilioProvider(
			config.TwilioAccountSID,
			config.TwilioAuthToken,
			config.TwilioFrom,
		)
		smsFactory.Register("twilio", twilioProvider)
	}

	// 注册模拟短信提供商（用于测试）
	mockSMSProvider := provider.NewMockSMSProvider()
	smsFactory.Register("mock", mockSMSProvider)

	// 创建 Webhook 提供商
	webhookProvider := provider.NewWebhookProvider()

	// 创建仓储层
	notificationRepo := repository.NewNotificationRepository(db)

	// 创建服务层
	notificationService := service.NewNotificationService(
		notificationRepo,
		emailFactory,
		smsFactory,
		webhookProvider,
	)

	// 创建处理器层
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// 创建 HTTP 服务器
	router := gin.Default()

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册路由
	notificationHandler.RegisterRoutes(router)

	// 启动后台任务
	go startBackgroundWorkers(notificationService)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8007"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Notification Service 启动在 %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// Config 配置结构
type Config struct {
	DatabaseURL string

	// SMTP 配置
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Mailgun 配置
	MailgunDomain string
	MailgunAPIKey string
	MailgunFrom   string

	// Twilio 配置
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFrom       string
}

// loadConfig 加载配置
func loadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payment_platform?sslmode=disable"),

		// SMTP
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),

		// Mailgun
		MailgunDomain: getEnv("MAILGUN_DOMAIN", ""),
		MailgunAPIKey: getEnv("MAILGUN_API_KEY", ""),
		MailgunFrom:   getEnv("MAILGUN_FROM", ""),

		// Twilio
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFrom:       getEnv("TWILIO_FROM", ""),
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

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var intValue int
	fmt.Sscanf(value, "%d", &intValue)
	return intValue
}

// startBackgroundWorkers 启动后台任务
func startBackgroundWorkers(notificationService service.NotificationService) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 处理待发送的通知
		if err := notificationService.ProcessPendingNotifications(nil); err != nil {
			log.Printf("处理待发送通知失败: %v", err)
		}

		// 处理待投递的 Webhook
		if err := notificationService.ProcessPendingWebhookDeliveries(nil); err != nil {
			log.Printf("处理待投递 Webhook 失败: %v", err)
		}
	}
}
