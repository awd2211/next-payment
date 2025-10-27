package main

import (
	"context"
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/merchant-quota-service/internal/client"
	"payment-platform/merchant-quota-service/internal/handler"
	"payment-platform/merchant-quota-service/internal/model"
	"payment-platform/merchant-quota-service/internal/repository"
	"payment-platform/merchant-quota-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

//	@title						Merchant Quota Service API
//	@version					1.0
//	@description				商户配额服务 - 实时追踪配额消耗,提供配额检查、消耗、释放和预警功能
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40022
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 初始化配置客户端
	var configClient *configclient.Client
	if config.GetEnv("ENABLE_CONFIG_CLIENT", "false") == "true" {
		clientCfg := configclient.ClientConfig{
			ServiceName: "merchant-quota-service",
			Environment: config.GetEnv("ENV", "production"),
			ConfigURL:   config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"),
			RefreshRate: 30 * time.Second,
		}
		if config.GetEnvBool("CONFIG_CLIENT_MTLS", false) {
			clientCfg.EnableMTLS = true
			clientCfg.TLSCertFile = config.GetEnv("TLS_CERT_FILE", "")
			clientCfg.TLSKeyFile = config.GetEnv("TLS_KEY_FILE", "")
			clientCfg.TLSCAFile = config.GetEnv("TLS_CA_FILE", "")
		}
		client, err := configclient.NewClient(clientCfg)
		if err != nil {
			logger.Warn("配置客户端初始化失败", zap.Error(err))
		} else {
			configClient = client
			defer configClient.Stop()
			logger.Info("配置中心客户端初始化成功")
		}
	}

	getConfig := func(key, defaultValue string) string {
		if configClient != nil {
			if val := configClient.Get(key); val != "" {
				return val
			}
		}
		return config.GetEnv(key, defaultValue)
	}

	// 2. 使用 Bootstrap 框架初始化应用
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "merchant-quota-service",
		DBName:      config.GetEnv("DB_NAME", "payment_merchant_quota"),
		Port:        config.GetEnvInt("PORT", 40022),

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.MerchantQuota{},
			&model.QuotaUsageLog{},
			&model.QuotaAlert{},
		},

		// 启用企业级功能
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true, // Redis用于缓存配额数据
		EnableGRPC:        false, // 系统使用 HTTP/REST 通信
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false),

		// 速率限制配置 (配额服务高频调用)
		RateLimitRequests: 500,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap失败: %v", err)
	}

	logger.Info("正在启动 Merchant Quota Service...")

	// 3. 初始化服务客户端
	policyServiceURL := getConfig("POLICY_SERVICE_URL", "http://localhost:40012")
	policyClient := client.NewPolicyClient(policyServiceURL)
	logger.Info("PolicyClient初始化成功", zap.String("url", policyServiceURL))

	notificationServiceURL := getConfig("NOTIFICATION_SERVICE_URL", "http://localhost:40008")
	notificationClient := client.NewNotificationClient(notificationServiceURL)
	logger.Info("NotificationClient初始化成功", zap.String("url", notificationServiceURL))

	// 4. 初始化 Repository
	quotaRepo := repository.NewQuotaRepository(application.DB)
	usageLogRepo := repository.NewUsageLogRepository(application.DB)
	alertRepo := repository.NewAlertRepository(application.DB)

	// 5. 初始化 Service
	quotaService := service.NewQuotaService(quotaRepo, usageLogRepo, policyClient)
	alertService := service.NewAlertService(alertRepo, quotaRepo, policyClient, notificationClient)

	// 6. JWT 认证中间件（优先从配置中心获取）
	// ⚠️ 安全要求: JWT_SECRET必须在生产环境中设置，不能使用默认值
	jwtSecret := getConfig("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required and cannot be empty")
	}
	if len(jwtSecret) < 32 {
		logger.Fatal("JWT_SECRET must be at least 32 characters for security",
			zap.Int("current_length", len(jwtSecret)),
			zap.Int("minimum_length", 32))
	}
	logger.Info("JWT_SECRET validation passed", zap.Int("length", len(jwtSecret)))
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 7. 注册 Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 8. 注册 API 路由
	handler.RegisterRoutes(application.Router, authMiddleware, quotaService, alertService)

	logger.Info("Merchant Quota Service 路由注册完成")
	logger.Info("服务职责: 实时追踪配额消耗,提供配额检查、消耗、释放和预警功能")
	logger.Info("核心功能: 配额追踪(动态消耗管理)")
	logger.Info("策略配置功能由 merchant-policy-service 负责")

	// 10. 启动定时任务
	startScheduledTasks(quotaService, alertService)

	// 11. 启动HTTP服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal("服务启动失败: " + err.Error())
	}
}

// startScheduledTasks 启动定时任务
func startScheduledTasks(quotaService service.QuotaService, alertService service.AlertService) {
	// 定时任务1: 每日零点重置日配额
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			// 每天00:00执行
			if now.Hour() == 0 && now.Minute() < 5 {
				logger.Info("开始重置日配额...")
				ctx := context.Background()
				if err := quotaService.ResetDailyQuotas(ctx); err != nil {
					logger.Error("重置日配额失败", zap.Error(err))
				} else {
					logger.Info("日配额重置完成")
				}
			}
		}
	}()

	// 定时任务2: 每月1日零点重置月配额
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			// 每月1日00:00执行
			if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 5 {
				logger.Info("开始重置月配额...")
				ctx := context.Background()
				if err := quotaService.ResetMonthlyQuotas(ctx); err != nil {
					logger.Error("重置月配额失败", zap.Error(err))
				} else {
					logger.Info("月配额重置完成")
				}
			}
		}
	}()

	// 定时任务3: 每5分钟检查配额预警
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			logger.Info("开始检查配额预警...")
			ctx := context.Background()
			if err := alertService.CheckQuotaAlerts(ctx); err != nil {
				logger.Error("配额预警检查失败", zap.Error(err))
			} else {
				logger.Info("配额预警检查完成")
			}
		}
	}()

	logger.Info("定时任务已启动",
		zap.String("task1", "日配额重置 (每日00:00)"),
		zap.String("task2", "月配额重置 (每月1日00:00)"),
		zap.String("task3", "配额预警检查 (每5分钟)"),
	)
}

// ============================================================
// Merchant Quota Service - 架构说明
//
// 本服务由 merchant-limit-service 重构而来,职责更聚焦:
//
// 核心职责:
// ✅ 配额追踪 (MerchantQuota)
// ✅ 配额操作 (Consume, Release, Reset, Adjust)
// ✅ 使用日志 (QuotaUsageLog)
// ✅ 配额预警 (QuotaAlert)
// ✅ 定时任务 (重置配额, 预警检查)
//
// 与 merchant-policy-service 的区别:
// - policy-service: 策略配置(静态规则)
// - quota-service: 配额追踪(动态消耗)
//
// 数据模型:
// - MerchantQuota: 商户配额追踪 (实时使用量)
// - QuotaUsageLog: 使用日志 (审计)
// - QuotaAlert: 配额预警 (监控)
//
// 定时任务:
// - 每日00:00重置日配额
// - 每月1日00:00重置月配额
// - 每5分钟检查配额预警
//
// 端口: 40022
// 数据库: payment_merchant_quota
// ============================================================
