package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	pkggrpc "github.com/payment-platform/pkg/grpc"
	"github.com/payment-platform/pkg/health"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	pkgtls "github.com/payment-platform/pkg/tls"
	"github.com/payment-platform/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// ServiceConfig 服务配置（每个服务可以自定义）
type ServiceConfig struct {
	ServiceName    string   // 服务名称（必填）
	DBName         string   // 数据库名（必填）
	Port           int      // HTTP 端口（必填）
	GRPCPort       int      // gRPC 端口（可选）
	AutoMigrate    []any    // 需要自动迁移的模型（可选）

	// 可选配置
	EnableTracing     bool     // 是否启用追踪（默认：true）
	EnableMetrics     bool     // 是否启用指标（默认：true）
	EnableRedis       bool     // 是否启用 Redis（默认：false）
	EnableGRPC        bool     // 是否启用 gRPC（默认：false）
	EnableHealthCheck bool     // 是否启用增强健康检查（默认：true）
	EnableRateLimit   bool     // 是否启用速率限制（默认：false）
	EnableMTLS        bool     // 是否启用mTLS服务间认证（默认：false）

	// 速率限制配置（当 EnableRateLimit=true 时有效）
	RateLimitRequests int          // 请求数限制（默认：100）
	RateLimitWindow   time.Duration // 时间窗口（默认：1分钟）
}

// App 应用实例（每个服务独立拥有）
type App struct {
	// 基础设施（由框架提供）
	DB            *gorm.DB
	Redis         *redis.Client
	Router        *gin.Engine
	GRPCServer    *grpc.Server // gRPC 服务器（可选，当 EnableGRPC=true 时创建）
	Logger        *zap.Logger
	HealthChecker *health.HealthChecker

	// 配置信息
	Config      ServiceConfig
	Environment string

	// 内部状态
	shutdownFuncs []func(context.Context) error // 优雅关闭函数列表
}

// Bootstrap 启动应用（统一的初始化逻辑）
// 使用示例：
//   app, err := app.Bootstrap(app.ServiceConfig{
//       ServiceName: "payment-gateway",
//       DBName:      "payment_gateway",
//       Port:        40003,
//       AutoMigrate: []any{&model.Payment{}, &model.Refund{}},
//   })
func Bootstrap(cfg ServiceConfig) (*App, error) {
	// 1. 初始化环境变量
	env := config.GetEnv("ENV", "development")

	// 2. 初始化日志
	if err := logger.InitLogger(env); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	logger.Info(fmt.Sprintf("正在启动 %s...", cfg.ServiceName))

	// 3. 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", cfg.DBName),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
	}

	// 配置验证（生产环境安全检查）
	if env == "production" && dbConfig.Password == "postgres" {
		return nil, fmt.Errorf("生产环境禁止使用默认数据库密码")
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Fatal("数据库连接失败")
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	logger.Info("数据库连接成功")

	// 4. 自动迁移（如果提供了模型）
	if len(cfg.AutoMigrate) > 0 {
		if err := database.AutoMigrate(cfg.AutoMigrate...); err != nil {
			logger.Fatal("数据库迁移失败")
			return nil, fmt.Errorf("数据库迁移失败: %w", err)
		}
		logger.Info("数据库迁移成功")
	}

	app := &App{
		DB:          database,
		Config:      cfg,
		Environment: env,
	}

	// 5. 初始化 Redis（可选）
	if cfg.EnableRedis {
		redisConfig := db.RedisConfig{
			Host:     config.GetEnv("REDIS_HOST", "localhost"),
			Port:     config.GetEnvInt("REDIS_PORT", 6379),
			Password: config.GetEnv("REDIS_PASSWORD", ""),
			DB:       config.GetEnvInt("REDIS_DB", 0),
		}

		redisClient, err := db.NewRedisClient(redisConfig)
		if err != nil {
			logger.Fatal("Redis 连接失败")
			return nil, fmt.Errorf("Redis 连接失败: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Fatal("Redis Ping 失败")
			return nil, fmt.Errorf("Redis 连接测试失败: %w", err)
		}

		app.Redis = redisClient
		logger.Info("Redis 连接成功")
	}

	// 6. 初始化 Gin 路由
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// 7. 添加全局中间件（统一顺序）
	router.Use(gin.Recovery())                              // Panic 恢复
	router.Use(middleware.RequestID())                      // 请求 ID
	router.Use(middleware.CORS())                           // CORS

	// 8. 初始化追踪（可选）
	if cfg.EnableTracing {
		tracerShutdown, err := tracing.InitTracer(tracing.Config{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: "1.0.0",
			Environment:    env,
			JaegerEndpoint: config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			SamplingRate:   float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0,
		})
		if err != nil {
			logger.Error("初始化追踪失败", zap.Error(err))
		} else {
			router.Use(tracing.TracingMiddleware(cfg.ServiceName))
			logger.Info("追踪初始化成功")

			// 注册清理函数
			// 注意：实际应用中应该在 main 函数中调用
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				tracerShutdown(ctx)
			}()
		}
	}

	router.Use(middleware.Logger(logger.Log))               // 日志中间件

	// 9. 初始化指标（可选）
	if cfg.EnableMetrics {
		// Prometheus metric names must use underscores, not hyphens
		metricsNamespace := strings.ReplaceAll(cfg.ServiceName, "-", "_")
		httpMetrics := metrics.NewHTTPMetrics(metricsNamespace)
		router.Use(metrics.PrometheusMiddleware(httpMetrics))
		logger.Info("指标收集已启用")
	}

	// 10. 初始化速率限制（可选）
	if cfg.EnableRateLimit {
		if app.Redis == nil {
			logger.Warn("速率限制需要 Redis，但 Redis 未启用")
		} else {
			requests := cfg.RateLimitRequests
			if requests <= 0 {
				requests = 100 // 默认100请求
			}
			window := cfg.RateLimitWindow
			if window <= 0 {
				window = time.Minute // 默认1分钟
			}

			rateLimiter := middleware.NewRateLimiter(app.Redis, requests, window)
			router.Use(rateLimiter.RateLimit())
			logger.Info(fmt.Sprintf("速率限制已启用 (%d req/%v)", requests, window))
		}
	}

	app.Router = router

	// 11. 初始化健康检查器（可选）
	if cfg.EnableHealthCheck {
		healthChecker := health.NewHealthChecker()

		// 注册数据库健康检查
		healthChecker.Register(health.NewDBChecker("database", database))

		// 注册Redis健康检查（如果启用）
		if app.Redis != nil {
			healthChecker.Register(health.NewRedisChecker("redis", app.Redis))
		}

		app.HealthChecker = healthChecker

		// 注册健康检查端点
		healthHandler := health.NewGinHandler(healthChecker)
		router.GET("/health", healthHandler.Handle)                    // 完整健康检查
		router.GET("/health/live", healthHandler.HandleLiveness)       // 存活探针
		router.GET("/health/ready", healthHandler.HandleReadiness)     // 就绪探针

		logger.Info("健康检查已启用")
	} else {
		// 简单健康检查端点
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": cfg.ServiceName,
				"time":    time.Now().Unix(),
			})
		})
	}

	// 12. 初始化 gRPC 服务器（可选）
	if cfg.EnableGRPC {
		if cfg.GRPCPort <= 0 {
			logger.Warn("EnableGRPC=true 但未指定 GRPCPort，gRPC 服务器将不会启动")
		} else {
			// 创建 gRPC 服务器
			app.GRPCServer = pkggrpc.NewSimpleServer()
			logger.Info(fmt.Sprintf("gRPC 服务器已创建，将监听端口 %d", cfg.GRPCPort))
		}
	}

	// 13. 初始化 mTLS（可选）
	if cfg.EnableMTLS {
		tlsConfig := pkgtls.LoadFromEnv()
		if err := pkgtls.ValidateServerConfig(tlsConfig); err != nil {
			return nil, fmt.Errorf("mTLS 配置验证失败: %w", err)
		}
		logger.Info("mTLS 服务间认证已启用")

		// 添加 mTLS 中间件（记录客户端证书信息）
		router.Use(pkgtls.MTLSMiddleware())
	}

	// 14. Prometheus 指标端点
	if cfg.EnableMetrics {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	logger.Info(fmt.Sprintf("%s 初始化完成", cfg.ServiceName))

	return app, nil
}

// Run 启动 HTTP 服务器（简单模式，无优雅关闭）
func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.Config.Port)
	logger.Info(fmt.Sprintf("%s 正在监听 %s", a.Config.ServiceName, addr))

	if err := a.Router.Run(addr); err != nil {
		logger.Fatal("服务启动失败")
		return fmt.Errorf("服务启动失败: %w", err)
	}

	return nil
}

// RunWithGracefulShutdown 启动 HTTP 服务器并支持优雅关闭
// 监听 SIGINT 和 SIGTERM 信号，收到信号后优雅关闭服务
func (a *App) RunWithGracefulShutdown() error {
	addr := fmt.Sprintf(":%d", a.Config.Port)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	// 如果启用了 mTLS，配置 TLS
	var tlsConfig *tls.Config
	var certFile, keyFile string
	if a.Config.EnableMTLS {
		cfg := pkgtls.LoadFromEnv()
		var err error
		tlsConfig, err = pkgtls.NewServerTLSConfig(cfg)
		if err != nil {
			return fmt.Errorf("创建 TLS 配置失败: %w", err)
		}
		srv.TLSConfig = tlsConfig
		certFile = cfg.CertFile
		keyFile = cfg.KeyFile
		logger.Info("HTTP 服务器已启用 mTLS")
	}

	// 在 goroutine 中启动服务器
	go func() {
		var err error
		if a.Config.EnableMTLS {
			logger.Info(fmt.Sprintf("%s HTTPS服务器(mTLS)正在监听 %s", a.Config.ServiceName, addr))
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			logger.Info(fmt.Sprintf("%s HTTP服务器正在监听 %s", a.Config.ServiceName, addr))
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(fmt.Sprintf("收到关闭信号，正在优雅关闭 %s...", a.Config.ServiceName))

	// 创建关闭上下文（30秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP 服务器关闭失败", zap.Error(err))
		return fmt.Errorf("HTTP 服务器关闭失败: %w", err)
	}

	// 执行应用级关闭
	if err := a.Shutdown(ctx); err != nil {
		logger.Error("应用关闭失败", zap.Error(err))
		return fmt.Errorf("应用关闭失败: %w", err)
	}

	logger.Info(fmt.Sprintf("%s 已完全关闭", a.Config.ServiceName))
	return nil
}

// RunDualProtocol 同时启动 HTTP 和 gRPC 服务器并支持优雅关闭
// 仅在 EnableGRPC=true 且 GRPCPort>0 时启动 gRPC 服务器
func (a *App) RunDualProtocol() error {
	// 1. 创建 HTTP 服务器
	httpAddr := fmt.Sprintf(":%d", a.Config.Port)
	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: a.Router,
	}

	// 2. 启动 HTTP 服务器
	go func() {
		logger.Info(fmt.Sprintf("%s HTTP服务器正在监听 %s", a.Config.ServiceName, httpAddr))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务启动失败", zap.Error(err))
		}
	}()

	// 3. 启动 gRPC 服务器（如果启用）
	var grpcListener net.Listener
	if a.GRPCServer != nil && a.Config.GRPCPort > 0 {
		grpcAddr := fmt.Sprintf(":%d", a.Config.GRPCPort)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("gRPC 监听端口失败", zap.Error(err))
			return fmt.Errorf("gRPC 监听端口失败: %w", err)
		}
		grpcListener = lis

		go func() {
			logger.Info(fmt.Sprintf("%s gRPC服务器正在监听 %s", a.Config.ServiceName, grpcAddr))
			if err := a.GRPCServer.Serve(lis); err != nil {
				logger.Fatal("gRPC服务启动失败", zap.Error(err))
			}
		}()
	}

	// 4. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(fmt.Sprintf("收到关闭信号，正在优雅关闭 %s...", a.Config.ServiceName))

	// 5. 创建关闭上下文（30秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 6. 关闭 HTTP 服务器
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error("HTTP 服务器关闭失败", zap.Error(err))
	}

	// 7. 关闭 gRPC 服务器
	if a.GRPCServer != nil {
		a.GRPCServer.GracefulStop()
		logger.Info("gRPC 服务器已关闭")
	}
	if grpcListener != nil {
		grpcListener.Close()
	}

	// 8. 执行应用级关闭
	if err := a.Shutdown(ctx); err != nil {
		logger.Error("应用关闭失败", zap.Error(err))
		return fmt.Errorf("应用关闭失败: %w", err)
	}

	logger.Info(fmt.Sprintf("%s 已完全关闭", a.Config.ServiceName))
	return nil
}

// Shutdown 优雅关闭（关闭数据库连接等）
func (a *App) Shutdown(ctx context.Context) error {
	logger.Info(fmt.Sprintf("正在关闭 %s...", a.Config.ServiceName))

	// 关闭数据库连接
	if sqlDB, err := a.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error("关闭数据库连接失败", zap.Error(err))
		}
	}

	// 关闭 Redis 连接
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			logger.Error("关闭 Redis 连接失败", zap.Error(err))
		}
	}

	logger.Sync()
	logger.Info(fmt.Sprintf("%s 已关闭", a.Config.ServiceName))

	return nil
}
