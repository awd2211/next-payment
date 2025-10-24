package main

import (
	"fmt"
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/email"
	"github.com/payment-platform/pkg/idempotency"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"payment-platform/merchant-service/internal/client"
	"payment-platform/merchant-service/internal/handler"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
	"payment-platform/merchant-service/internal/service"
)

//	@title						Merchant Service API
//	@version					1.0
//	@description				支付平台商户管理服务API文档（Phase 10 清理后）
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40002
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 使用 Bootstrap 框架初始化应用
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "merchant-service",
		DBName:      config.GetEnv("DB_NAME", "payment_merchant"),
		Port:        config.GetEnvInt("PORT", 40002),

		// 自动迁移数据库模型（仅保留核心模型）
		// 已迁移模型：
		// - APIKey → merchant-auth-service
		// - ChannelConfig, MerchantFeeConfig, MerchantTransactionLimit → merchant-config-service
		// - SettlementAccount → settlement-service
		// - KYCDocument, BusinessQualification → kyc-service
		AutoMigrate: []any{
			&model.Merchant{},         // 核心：商户主表
			&model.MerchantUser{},     // 保留：商户子账户（Merchant聚合根）
			&model.MerchantContract{}, // 保留：商户合同（Merchant聚合根）
		},

		// 启用企业级功能
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 使用 HTTP 通信
		EnableHealthCheck: true,
		EnableRateLimit:   true,

		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Merchant Service (Phase 10 清理后)...")

	// 2. 初始化JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// 3. 初始化Repository（仅保留核心）
	merchantRepo := repository.NewMerchantRepository(application.DB)
	merchantUserRepo := repository.NewMerchantUserRepository(application.DB)

	// 4. 初始化Service
	// Note: apiKeyRepo removed - APIKey management now in merchant-auth-service
	merchantService := service.NewMerchantService(application.DB, merchantRepo, jwtManager)

	// 初始化邮件服务（可选）
	var emailProvider email.EmailProvider
	smtpHost := config.GetEnv("SMTP_HOST", "")
	if smtpHost != "" {
		smtpConfig := email.SMTPConfig{
			Host:     smtpHost,
			Port:     config.GetEnvInt("SMTP_PORT", 587),
			Username: config.GetEnv("SMTP_USERNAME", ""),
			Password: config.GetEnv("SMTP_PASSWORD", ""),
			From:     config.GetEnv("SMTP_FROM", "noreply@payment-platform.com"),
			FromName: config.GetEnv("SMTP_FROM_NAME", "支付平台"),
		}
		var err error
		emailProvider, err = email.NewSMTPProvider(smtpConfig)
		if err != nil {
			logger.Warn(fmt.Sprintf("邮件服务初始化失败: %v", err))
		}
	}

	// MerchantUser 服务（保留）
	// MerchantUser 服务（预留，待添加 Handler）
	_ = service.NewMerchantUserService(merchantUserRepo, merchantRepo, emailProvider)

	// 5. 初始化HTTP客户端（用于Dashboard聚合）
	analyticsClient := client.NewAnalyticsClient(config.GetEnv("ANALYTICS_SERVICE_URL", "http://localhost:40009"))
	accountingClient := client.NewAccountingClient(config.GetEnv("ACCOUNTING_SERVICE_URL", "http://localhost:40007"))
	riskClient := client.NewRiskClient(config.GetEnv("RISK_SERVICE_URL", "http://localhost:40006"))
	notificationClient := client.NewNotificationClient(config.GetEnv("NOTIFICATION_SERVICE_URL", "http://localhost:40008"))
	paymentClient := client.NewPaymentClient(config.GetEnv("PAYMENT_SERVICE_URL", "http://localhost:40003"))

	// Dashboard聚合服务
	dashboardService := service.NewDashboardService(
		analyticsClient,
		accountingClient,
		riskClient,
		notificationClient,
		paymentClient,
	)

	// 6. 初始化HTTP Handler
	merchantHandler := handler.NewMerchantHandler(merchantService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	
	// 注意：以下Handler已删除，功能迁移至新服务：
	// - apiKeyHandler → merchant-auth-service (port 40011)
	// - channelHandler → merchant-config-service (port 40012)
	// - businessHandler → settlement/kyc/config services

	// 7. 幂等性中间件
	idempotencyManager := idempotency.NewIdempotencyManager(application.Redis, "merchant-service", 24*time.Hour)
	application.Router.Use(middleware.IdempotencyMiddleware(idempotencyManager))

	// 8. Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 9. JWT 认证中间件
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 10. 注册路由（简化版）
	api := application.Router.Group("/api/v1")
	{
		// 商户核心路由
		merchantHandler.RegisterRoutes(api)

		// MerchantUser 路由
		userAPI := api.Group("/merchant-users")
		userAPI.Use(authMiddleware)
		{
			// TODO: 添加 MerchantUser 相关路由
		}

		// Dashboard聚合查询路由
		dashboardHandler.RegisterRoutes(api, authMiddleware)
	}

	logger.Info("Phase 10 清理完成：已迁移业务到新服务")
	logger.Info("- APIKey → merchant-auth-service:40011")
	logger.Info("- Config → merchant-config-service:40012")
	logger.Info("- Settlement → settlement-service:40013")
	logger.Info("- KYC → kyc-service:40015")

	// 11. 启动服务
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}
