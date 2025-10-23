package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
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

	fmt.Println("🔧 开始创建测试审计日志...")

	// 获取admin用户ID
	var admin model.Admin
	if err := database.Where("username = ?", "admin").First(&admin).Error; err != nil {
		log.Fatalf("未找到 admin 用户: %v", err)
	}

	// 创建各种类型的审计日志
	logs := []model.AuditLog{
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "login",
			Resource:     "admin",
			ResourceID:   admin.ID.String(),
			Method:       "POST",
			Path:         "/api/v1/admin/login",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			ResponseCode: 200,
			Description:  "管理员登录",
			CreatedAt:    time.Now().Add(-48 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "create_role",
			Resource:     "role",
			ResourceID:   uuid.New().String(),
			Method:       "POST",
			Path:         "/api/v1/roles",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			RequestBody:  `{"name":"custom_role","display_name":"自定义角色"}`,
			ResponseCode: 201,
			Description:  "创建自定义角色",
			CreatedAt:    time.Now().Add(-36 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "update_role",
			Resource:     "role",
			ResourceID:   uuid.New().String(),
			Method:       "PUT",
			Path:         "/api/v1/roles/123",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			RequestBody:  `{"display_name":"更新后的角色"}`,
			ResponseCode: 200,
			Description:  "更新角色信息",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "assign_permissions",
			Resource:     "role",
			ResourceID:   uuid.New().String(),
			Method:       "POST",
			Path:         "/api/v1/roles/123/permissions",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			RequestBody:  `{"permission_ids":["perm1","perm2"]}`,
			ResponseCode: 200,
			Description:  "为角色分配权限",
			CreatedAt:    time.Now().Add(-12 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "create_admin",
			Resource:     "admin",
			ResourceID:   uuid.New().String(),
			Method:       "POST",
			Path:         "/api/v1/admin",
			IP:           "192.168.1.101",
			UserAgent:    "Mozilla/5.0",
			RequestBody:  `{"username":"newadmin","email":"new@example.com"}`,
			ResponseCode: 201,
			Description:  "创建新管理员",
			CreatedAt:    time.Now().Add(-6 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "view_permissions",
			Resource:     "permission",
			Method:       "GET",
			Path:         "/api/v1/permissions",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			ResponseCode: 200,
			Description:  "查看权限列表",
			CreatedAt:    time.Now().Add(-3 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "delete_role",
			Resource:     "role",
			ResourceID:   uuid.New().String(),
			Method:       "DELETE",
			Path:         "/api/v1/roles/456",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			ResponseCode: 403,
			Description:  "尝试删除系统角色（失败）",
			CreatedAt:    time.Now().Add(-1 * time.Hour),
		},
		{
			AdminID:      admin.ID,
			AdminName:    admin.FullName,
			Action:       "view_audit_logs",
			Resource:     "audit",
			Method:       "GET",
			Path:         "/api/v1/audit-logs",
			IP:           "192.168.1.100",
			UserAgent:    "Mozilla/5.0",
			ResponseCode: 200,
			Description:  "查看审计日志",
			CreatedAt:    time.Now().Add(-30 * time.Minute),
		},
	}

	// 插入日志
	for i, auditLog := range logs {
		if err := database.Create(&auditLog).Error; err != nil {
			log.Printf("创建日志 %d 失败: %v", i+1, err)
			continue
		}
		fmt.Printf("✅ 创建日志 %d: %s - %s\n", i+1, auditLog.Action, auditLog.Description)
	}

	fmt.Printf("\n🎉 成功创建 %d 条测试审计日志！\n", len(logs))
}
