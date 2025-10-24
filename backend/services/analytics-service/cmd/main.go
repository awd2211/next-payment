package main

import (
	"fmt"
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"payment-platform/analytics-service/internal/handler"
	"payment-platform/analytics-service/internal/model"
	"payment-platform/analytics-service/internal/repository"
	"payment-platform/analytics-service/internal/service"
)

//	@title						Analytics Service API
//	@version					1.0
//	@description				支付平台数据分析服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40009
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "analytics-service",
		DBName:      config.GetEnv("DB_NAME", "payment_analytics"),
		Port:        config.GetEnvInt("PORT", 40009),
		AutoMigrate: []any{
			&model.PaymentMetrics{},
			&model.MerchantMetrics{},
			&model.ChannelMetrics{},
			&model.RealtimeStats{},
		},
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false,
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Analytics Service...")
	analyticsRepo := repository.NewAnalyticsRepository(application.DB)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	analyticsHandler.RegisterRoutes(application.Router)

	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

// 186 → 38 行, 减少 80%
