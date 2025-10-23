package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/services/payment-gateway/internal/handler"
	"github.com/payment-platform/services/payment-gateway/internal/model"
	"github.com/payment-platform/services/payment-gateway/internal/repository"
	"github.com/payment-platform/services/payment-gateway/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config := loadConfig()

	db, err := connectDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Payment{},
		&model.Refund{},
		&model.PaymentCallback{},
		&model.PaymentRoute{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	paymentRepo := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "payment-gateway"})
	})

	paymentHandler.RegisterRoutes(router)

	port := config.ServerPort
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Payment Gateway 服务启动在 %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

type Config struct {
	DatabaseURL string
	ServerPort  string
}

func loadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payment_platform?sslmode=disable"),
		ServerPort:  getEnv("PORT", "8002"),
	}
}

func connectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
