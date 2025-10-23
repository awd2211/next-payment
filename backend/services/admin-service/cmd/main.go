package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/services/admin-service/internal/handler"
	"github.com/payment-platform/services/admin-service/internal/model"
	"github.com/payment-platform/services/admin-service/internal/repository"
	"github.com/payment-platform/services/admin-service/internal/service"
)

func main() {
	// 初始化日志
	env := config.GetEnv("ENV", "development")
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("正在启动 Admin Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_admin"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Fatal("数据库连接失败", logger.Log.With().Err(err).Logger())
	}
	logger.Info("数据库连接成功")

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.Admin{},
		&model.Role{},
		&model.Permission{},
		&model.AdminRole{},
		&model.RolePermission{},
		&model.AuditLog{},
		&model.SystemConfig{},
		&model.MerchantReview{},
		&model.ApprovalFlow{},
	); err != nil {
		logger.Fatal("数据库迁移失败", logger.Log.With().Err(err).Logger())
	}
	logger.Info("数据库迁移完成")

	// 初始化Redis
	redisConfig := db.RedisConfig{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnvInt("REDIS_PORT", 6379),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       config.GetEnvInt("REDIS_DB", 0),
	}

	redisClient, err := db.NewRedisClient(redisConfig)
	if err != nil {
		logger.Fatal("Redis连接失败", logger.Log.With().Err(err).Logger())
	}
	logger.Info("Redis连接成功")

	// 初始化JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// 初始化Repository
	adminRepo := repository.NewAdminRepository(database)
	roleRepo := repository.NewRoleRepository(database)

	// 初始化Service
	adminService := service.NewAdminService(adminRepo, roleRepo, jwtManager)

	// 初始化Handler
	adminHandler := handler.NewAdminHandler(adminService)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger.Log))

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "admin-service",
			"time":    time.Now().Unix(),
		})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		// 认证中间件
		authMiddleware := middleware.AuthMiddleware(jwtManager)

		// 注册管理员路由
		adminHandler.RegisterRoutes(api, authMiddleware)

		// TODO: 注册其他路由（角色、权限、系统配置等）
	}

	// 启动服务器
	port := config.GetEnvInt("PORT", 8001)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Admin Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败", logger.Log.With().Err(err).Logger())
	}
}
