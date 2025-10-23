package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"payment-platform/admin-service/internal/model"
)

func main() {
	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 40432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_admin"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 创建第一个管理员
	passwordHash, err := auth.HashPassword("admin123456")
	if err != nil {
		log.Fatalf("密码加密失败: %v", err)
	}

	admin := &model.Admin{
		ID:           uuid.New(),
		Username:     "admin",
		Email:        "admin@payment-platform.com",
		FullName:     "System Administrator",
		PasswordHash: passwordHash,
		Status:       "active",
		IsSuper:      true,
	}

	if err := database.Create(admin).Error; err != nil {
		log.Fatalf("创建管理员失败: %v", err)
	}

	fmt.Println("✅ 初始管理员创建成功！")
	fmt.Println("   用户名: admin")
	fmt.Println("   密码: admin123456")
	fmt.Printf("   ID: %s\n", admin.ID.String())
}
