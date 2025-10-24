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
	"payment-platform/accounting-service/internal/client"
	"payment-platform/accounting-service/internal/handler"
	"payment-platform/accounting-service/internal/model"
	"payment-platform/accounting-service/internal/repository"
	"payment-platform/accounting-service/internal/service"
	// grpcServer "payment-platform/accounting-service/internal/grpc"
	// pb "github.com/payment-platform/proto/accounting"
)

//	@title						Accounting Service API
//	@version					1.0
//	@description				支付平台财务核算服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40007
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 使用 Bootstrap 框架初始化应用
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "accounting-service",
		DBName:      config.GetEnv("DB_NAME", "payment_accounting"),
		Port:        config.GetEnvInt("PORT", 40007),

		// 自动迁移数据库模型（仅核心账户模型）
		AutoMigrate: []any{
			&model.Account{},
			&model.AccountTransaction{},
			&model.DoubleEntry{},
		},

		// 启用企业级功能
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

	logger.Info("正在启动 Accounting Service...")

	// 2. 初始化 HTTP 客户端
	channelAdapterURL := config.GetEnv("CHANNEL_SERVICE_URL", "http://localhost:40005")
	channelAdapterClient := client.NewChannelAdapterClient(channelAdapterURL)

	// 3. 初始化Repository
	accountRepo := repository.NewAccountRepository(application.DB)

	// 4. 初始化Service（传入 application.DB 用于事务支持）
	accountService := service.NewAccountService(application.DB, accountRepo, channelAdapterClient)

	// 5. 初始化Handler
	accountHandler := handler.NewAccountHandler(accountService)

	// 6. Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 7. 注册账户路由
	accountHandler.RegisterRoutes(application.Router)

	// 8. 启动服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

// 代码行数: 192 → 80 行, 减少 58%
