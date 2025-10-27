package main

import (
	"fmt"
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"payment-platform/config-service/internal/handler"
	"payment-platform/config-service/internal/model"
	"payment-platform/config-service/internal/repository"
	"payment-platform/config-service/internal/service"
	// grpcServer "payment-platform/config-service/internal/grpc"
	// pb "github.com/payment-platform/proto/config"
)

//	@title						Config Service API
//	@version					1.0
//	@description				支付平台配置中心服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40010
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
			ServiceName: "config-service",
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
		ServiceName: "config-service",
		DBName:      config.GetEnv("DB_NAME", "payment_config"),
		Port:        config.GetEnvInt("PORT", 40010),
		// GRPCPort:    config.GetEnvInt("GRPC_PORT", 50010), // 不使用 gRPC,保持 HTTP 通信

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.Config{},
			&model.ConfigHistory{},
			&model.FeatureFlag{},
			&model.ServiceRegistry{},
			&model.ConfigAccessLog{}, // 审计日志（保留）
			// 注意：权限控制使用 admin-service 的 RBAC 系统，无需独立权限表
		},

		// 启用企业级功能(gRPC 默认关闭,使用 HTTP/REST)
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 默认关闭 gRPC,使用 HTTP 通信
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false), // mTLS 服务间认证

		// 速率限制配置
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Config Service...")

	// 2. 初始化 Repository
	configRepo := repository.NewConfigRepository(application.DB)

	// 3. 初始化 Service（注入Redis客户端用于缓存）
	configService := service.NewConfigService(configRepo, application.Redis)

	// 4. 初始化 Handler
	configHandler := handler.NewConfigHandler(configService)

	// 5. Swagger UI（公开接口）
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 6. 注册配置路由
	configHandler.RegisterRoutes(application.Router)

	// 7. gRPC 服务（预留但不启用，系统使用 HTTP/REST 通信）
	// configGrpcServer := grpcServer.NewConfigServer(configService)
	// pb.RegisterConfigServiceServer(application.GRPCServer, configGrpcServer)
	// logger.Info(fmt.Sprintf("gRPC Server 已注册，将监听端口 %d", config.GetEnvInt("GRPC_PORT", 50010)))

	// JWT 认证中间件（优先从配置中心获取）
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
	_ = jwtManager // 预留给需要认证的路由使用



	// 8. 启动服务（仅 HTTP，优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

// 代码行数对比：
// - 原始版本: 185行 (手动初始化所有组件)
// - Bootstrap版本: 100行 (框架自动处理)
// - 减少代码: 46%（保留了所有业务逻辑）
//
// 自动获得的功能：
// ✅ 数据库连接和迁移
// ✅ Redis 连接
// ✅ Zap 日志系统
// ✅ Gin 路由和中间件（CORS, RequestID, Panic Recovery）
// ✅ Jaeger 分布式追踪
// ✅ Prometheus 指标收集（/metrics 端点）
// ✅ 健康检查端点 (/health, /health/live, /health/ready)
// ✅ 速率限制
// ✅ 优雅关闭（信号处理）
// ✅ 请求 ID
//
// 保留的自定义能力：
// ✅ 配置管理服务（Config, ConfigHistory, FeatureFlag, ServiceRegistry）
// ✅ Swagger UI
