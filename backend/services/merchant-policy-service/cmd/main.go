package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/merchant-policy-service/internal/handler"
	"payment-platform/merchant-policy-service/internal/model"
	"payment-platform/merchant-policy-service/internal/repository"
	"payment-platform/merchant-policy-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

//	@title						Merchant Policy Service API
//	@version					1.0
//	@description				商户策略服务 - 统一管理商户等级、费率、限额和渠道策略
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40012
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
			ServiceName: "merchant-policy-service",
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
		ServiceName: "merchant-policy-service",
		DBName:      config.GetEnv("DB_NAME", "payment_merchant_policy"),
		Port:        config.GetEnvInt("PORT", 40012),

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.MerchantTier{},
			&model.MerchantPolicyBinding{},
			&model.MerchantFeePolicy{},
			&model.MerchantLimitPolicy{},
			&model.ChannelPolicy{},
		},

		// 启用企业级功能
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 系统使用 HTTP/REST 通信
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false),

		// 速率限制配置
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap失败: %v", err)
	}

	logger.Info("正在启动 Merchant Policy Service...")

	// 3. 初始化 Repository
	tierRepo := repository.NewTierRepository(application.DB)
	feePolicyRepo := repository.NewFeePolicyRepository(application.DB)
	limitPolicyRepo := repository.NewLimitPolicyRepository(application.DB)
	_ = repository.NewChannelPolicyRepository // TODO: 下阶段实现
	bindingRepo := repository.NewPolicyBindingRepository(application.DB)

	// 4. 初始化 Service
	tierService := service.NewTierService(tierRepo)
	policyEngineService := service.NewPolicyEngineService(feePolicyRepo, limitPolicyRepo, bindingRepo, tierRepo)
	policyBindingService := service.NewPolicyBindingService(bindingRepo, tierRepo)

	// 5. JWT 认证中间件（优先从配置中心获取）
	jwtSecret := getConfig("JWT_SECRET", "payment-platform-secret-key-2024")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 6. 注册 Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 7. 注册 API 路由
	handler.RegisterRoutes(application.Router, authMiddleware, tierService, policyEngineService, policyBindingService)

	logger.Info("Merchant Policy Service 路由注册完成")
	logger.Info("服务职责: 统一管理商户等级、费率策略、限额策略、渠道策略")
	logger.Info("核心功能: 策略配置管理(静态规则定义)")
	logger.Info("配额追踪功能由 merchant-quota-service 负责")

	// 9. 启动HTTP服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal("服务启动失败: " + err.Error())
	}
}

// ============================================================
// Merchant Policy Service - 架构说明
//
// 本服务由 merchant-config-service 重构而来,职责更聚焦:
//
// 核心职责:
// ✅ 商户等级管理 (MerchantTier)
// ✅ 费率策略管理 (MerchantFeePolicy)
// ✅ 限额策略管理 (MerchantLimitPolicy)
// ✅ 渠道策略管理 (ChannelPolicy)
// ✅ 策略生效引擎 (PolicyEngine)
//
// 与 merchant-quota-service 的区别:
// - policy-service: 策略配置(静态规则)
// - quota-service: 配额追踪(动态消耗)
//
// 数据模型:
// - MerchantTier: 商户等级定义
// - MerchantPolicyBinding: 商户策略绑定
// - MerchantFeePolicy: 费率策略
// - MerchantLimitPolicy: 限额策略
// - ChannelPolicy: 渠道策略
//
// 端口: 40012
// 数据库: payment_merchant_policy
// ============================================================
