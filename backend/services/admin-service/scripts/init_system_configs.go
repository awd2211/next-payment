package main

import (
	"fmt"
	"log"

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

	fmt.Println("🔧 开始创建默认系统配置...")

	// 获取admin用户ID作为创建者
	var admin model.Admin
	if err := database.Where("username = ?", "admin").First(&admin).Error; err != nil {
		log.Fatalf("未找到 admin 用户: %v", err)
	}

	// 定义默认配置
	configs := []model.SystemConfig{
		// 支付配置
		{
			Key:         "payment.default_currency",
			Value:       "USD",
			Type:        "string",
			Category:    "payment",
			Description: "默认货币类型",
			IsPublic:    true,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "payment.min_amount",
			Value:       "100",
			Type:        "number",
			Category:    "payment",
			Description: "最小支付金额（分）",
			IsPublic:    true,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "payment.max_amount",
			Value:       "1000000",
			Type:        "number",
			Category:    "payment",
			Description: "最大支付金额（分）",
			IsPublic:    true,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "payment.timeout",
			Value:       "1800",
			Type:        "number",
			Category:    "payment",
			Description: "支付超时时间（秒）",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},

		// 通知配置
		{
			Key:         "notification.email_enabled",
			Value:       "true",
			Type:        "boolean",
			Category:    "notification",
			Description: "是否启用邮件通知",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "notification.sms_enabled",
			Value:       "false",
			Type:        "boolean",
			Category:    "notification",
			Description: "是否启用短信通知",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "notification.webhook_retry_times",
			Value:       "3",
			Type:        "number",
			Category:    "notification",
			Description: "Webhook重试次数",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},

		// 风控配置
		{
			Key:         "risk.daily_limit_per_merchant",
			Value:       "10000000",
			Type:        "number",
			Category:    "risk",
			Description: "商户每日交易限额（分）",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "risk.single_transaction_limit",
			Value:       "500000",
			Type:        "number",
			Category:    "risk",
			Description: "单笔交易限额（分）",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "risk.ip_whitelist",
			Value:       `["127.0.0.1", "192.168.0.0/16"]`,
			Type:        "json",
			Category:    "risk",
			Description: "IP白名单",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "risk.fraud_detection_enabled",
			Value:       "true",
			Type:        "boolean",
			Category:    "risk",
			Description: "是否启用欺诈检测",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},

		// 系统配置
		{
			Key:         "system.maintenance_mode",
			Value:       "false",
			Type:        "boolean",
			Category:    "system",
			Description: "系统维护模式",
			IsPublic:    true,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "system.api_rate_limit",
			Value:       "100",
			Type:        "number",
			Category:    "system",
			Description: "API每分钟限流次数",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "system.session_timeout",
			Value:       "3600",
			Type:        "number",
			Category:    "system",
			Description: "会话超时时间（秒）",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},

		// 结算配置
		{
			Key:         "settlement.frequency",
			Value:       "daily",
			Type:        "string",
			Category:    "settlement",
			Description: "结算频率（daily/weekly/monthly）",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
		{
			Key:         "settlement.hold_days",
			Value:       "2",
			Type:        "number",
			Category:    "settlement",
			Description: "资金冻结天数",
			IsPublic:    false,
			UpdatedBy:   admin.ID,
		},
	}

	// 插入配置
	for i, cfg := range configs {
		// 检查配置是否已存在
		var existingConfig model.SystemConfig
		err := database.Where("key = ?", cfg.Key).First(&existingConfig).Error
		if err == nil {
			fmt.Printf("⚠️  配置 %d: %s 已存在，跳过\n", i+1, cfg.Key)
			continue
		}

		if err := database.Create(&cfg).Error; err != nil {
			log.Printf("创建配置 %d 失败: %v", i+1, err)
			continue
		}
		fmt.Printf("✅ 创建配置 %d: %s = %s (%s)\n", i+1, cfg.Key, cfg.Value, cfg.Description)
	}

	fmt.Printf("\n🎉 系统配置初始化完成！共创建 %d 个配置项\n", len(configs))

	// 按类别统计
	fmt.Println("\n📊 配置统计:")
	categories := make(map[string]int)
	for _, cfg := range configs {
		categories[cfg.Category]++
	}
	for category, count := range categories {
		fmt.Printf("  • %s: %d 个配置\n", category, count)
	}
}
