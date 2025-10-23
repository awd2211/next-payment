package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"gorm.io/gorm"
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

	fmt.Println("🔧 开始初始化角色和权限...")

	// 1. 创建权限
	permissions := createPermissions(database)
	fmt.Printf("✅ 创建了 %d 个权限\n", len(permissions))

	// 2. 创建角色
	roles := createRoles(database)
	fmt.Printf("✅ 创建了 %d 个角色\n", len(roles))

	// 3. 分配权限给角色
	assignPermissionsToRoles(database, roles, permissions)
	fmt.Println("✅ 权限分配完成")

	// 4. 给 admin 用户分配超级管理员角色
	assignRoleToAdmin(database, roles["super_admin"])
	fmt.Println("✅ admin 用户已设置为超级管理员")

	fmt.Println("\n🎉 角色和权限初始化完成！")
	printSummary(roles, permissions)
}

func createPermissions(db *gorm.DB) map[string]*model.Permission {
	permissions := make(map[string]*model.Permission)

	// 定义所有权限
	permDefs := []struct {
		code     string
		name     string
		resource string
		action   string
		desc     string
	}{
		// 管理员管理
		{"admin.view", "查看管理员", "admin", "view", "查看管理员列表和详情"},
		{"admin.create", "创建管理员", "admin", "create", "创建新管理员账户"},
		{"admin.edit", "编辑管理员", "admin", "edit", "修改管理员信息"},
		{"admin.delete", "删除管理员", "admin", "delete", "删除管理员账户"},
		{"admin.reset_password", "重置密码", "admin", "reset_password", "重置管理员密码"},

		// 角色管理
		{"role.view", "查看角色", "role", "view", "查看角色列表和详情"},
		{"role.create", "创建角色", "role", "create", "创建新角色"},
		{"role.edit", "编辑角色", "role", "edit", "修改角色信息"},
		{"role.delete", "删除角色", "role", "delete", "删除角色"},
		{"role.assign", "分配角色", "role", "assign", "给管理员分配角色"},

		// 权限管理
		{"permission.view", "查看权限", "permission", "view", "查看权限列表"},
		{"permission.assign", "分配权限", "permission", "assign", "给角色分配权限"},

		// 商户管理
		{"merchant.view", "查看商户", "merchant", "view", "查看商户列表和详情"},
		{"merchant.create", "创建商户", "merchant", "create", "创建新商户"},
		{"merchant.edit", "编辑商户", "merchant", "edit", "修改商户信息"},
		{"merchant.delete", "删除商户", "merchant", "delete", "删除商户"},
		{"merchant.approve", "审核商户", "merchant", "approve", "审核商户资质"},
		{"merchant.freeze", "冻结商户", "merchant", "freeze", "冻结/解冻商户账户"},

		// 支付管理
		{"payment.view", "查看支付", "payment", "view", "查看支付记录"},
		{"payment.refund", "退款操作", "payment", "refund", "处理退款申请"},
		{"payment.cancel", "取消支付", "payment", "cancel", "取消支付订单"},

		// 订单管理
		{"order.view", "查看订单", "order", "view", "查看订单列表和详情"},
		{"order.edit", "编辑订单", "order", "edit", "修改订单信息"},
		{"order.cancel", "取消订单", "order", "cancel", "取消订单"},

		// 账务管理
		{"accounting.view", "查看账务", "accounting", "view", "查看账务记录"},
		{"accounting.settlement", "结算操作", "accounting", "settlement", "处理结算"},
		{"accounting.export", "导出账务", "accounting", "export", "导出账务报表"},

		// 风控管理
		{"risk.view", "查看风控", "risk", "view", "查看风控规则和记录"},
		{"risk.edit", "编辑风控", "risk", "edit", "修改风控规则"},
		{"risk.review", "风控审核", "risk", "review", "审核风险订单"},

		// 系统配置
		{"config.view", "查看配置", "config", "view", "查看系统配置"},
		{"config.edit", "编辑配置", "config", "edit", "修改系统配置"},

		// 审计日志
		{"audit.view", "查看日志", "audit", "view", "查看审计日志"},
		{"audit.export", "导出日志", "audit", "export", "导出审计日志"},

		// 邮件模板
		{"email.view", "查看模板", "email", "view", "查看邮件模板"},
		{"email.edit", "编辑模板", "email", "edit", "编辑邮件模板"},
		{"email.send", "发送邮件", "email", "send", "发送邮件"},
	}

	for _, pd := range permDefs {
		perm := &model.Permission{
			ID:          uuid.New(),
			Code:        pd.code,
			Name:        pd.name,
			Resource:    pd.resource,
			Action:      pd.action,
			Description: pd.desc,
		}

		if err := db.Create(perm).Error; err != nil {
			log.Printf("创建权限 %s 失败: %v", pd.code, err)
			continue
		}

		permissions[pd.code] = perm
	}

	return permissions
}

func createRoles(db *gorm.DB) map[string]*model.Role {
	roles := make(map[string]*model.Role)

	roleDefs := []struct {
		name        string
		displayName string
		desc        string
		isSystem    bool
	}{
		{
			"super_admin",
			"超级管理员",
			"拥有系统所有权限，可以管理所有功能和数据",
			true,
		},
		{
			"admin",
			"普通管理员",
			"拥有大部分管理权限，适合日常运营管理",
			true,
		},
		{
			"operator",
			"运营人员",
			"只有查看权限，适合数据分析和客服人员",
			true,
		},
		{
			"finance",
			"财务人员",
			"拥有账务和结算相关权限",
			true,
		},
		{
			"risk_manager",
			"风控专员",
			"拥有风控管理和审核权限",
			true,
		},
	}

	for _, rd := range roleDefs {
		role := &model.Role{
			ID:          uuid.New(),
			Name:        rd.name,
			DisplayName: rd.displayName,
			Description: rd.desc,
			IsSystem:    rd.isSystem,
		}

		if err := db.Create(role).Error; err != nil {
			log.Printf("创建角色 %s 失败: %v", rd.name, err)
			continue
		}

		roles[rd.name] = role
	}

	return roles
}

func assignPermissionsToRoles(db *gorm.DB, roles map[string]*model.Role, permissions map[string]*model.Permission) {
	// 超级管理员 - 所有权限
	superAdmin := roles["super_admin"]
	for _, perm := range permissions {
		db.Create(&model.RolePermission{
			RoleID:       superAdmin.ID,
			PermissionID: perm.ID,
		})
	}

	// 普通管理员 - 大部分权限（不包括删除管理员、系统配置等敏感操作）
	admin := roles["admin"]
	adminPerms := []string{
		"admin.view", "admin.create", "admin.edit",
		"role.view",
		"permission.view",
		"merchant.view", "merchant.create", "merchant.edit", "merchant.approve",
		"payment.view", "payment.refund",
		"order.view", "order.edit", "order.cancel",
		"accounting.view", "accounting.export",
		"risk.view", "risk.review",
		"config.view",
		"audit.view",
		"email.view", "email.edit", "email.send",
	}
	for _, code := range adminPerms {
		if perm, ok := permissions[code]; ok {
			db.Create(&model.RolePermission{
				RoleID:       admin.ID,
				PermissionID: perm.ID,
			})
		}
	}

	// 运营人员 - 只有查看权限
	operator := roles["operator"]
	operatorPerms := []string{
		"admin.view",
		"role.view",
		"permission.view",
		"merchant.view",
		"payment.view",
		"order.view",
		"accounting.view",
		"risk.view",
		"config.view",
		"audit.view",
		"email.view",
	}
	for _, code := range operatorPerms {
		if perm, ok := permissions[code]; ok {
			db.Create(&model.RolePermission{
				RoleID:       operator.ID,
				PermissionID: perm.ID,
			})
		}
	}

	// 财务人员 - 账务相关权限
	finance := roles["finance"]
	financePerms := []string{
		"merchant.view",
		"payment.view", "payment.refund",
		"order.view",
		"accounting.view", "accounting.settlement", "accounting.export",
		"audit.view",
	}
	for _, code := range financePerms {
		if perm, ok := permissions[code]; ok {
			db.Create(&model.RolePermission{
				RoleID:       finance.ID,
				PermissionID: perm.ID,
			})
		}
	}

	// 风控专员 - 风控相关权限
	riskManager := roles["risk_manager"]
	riskPerms := []string{
		"merchant.view",
		"payment.view",
		"order.view",
		"risk.view", "risk.edit", "risk.review",
		"audit.view",
	}
	for _, code := range riskPerms {
		if perm, ok := permissions[code]; ok {
			db.Create(&model.RolePermission{
				RoleID:       riskManager.ID,
				PermissionID: perm.ID,
			})
		}
	}
}

func assignRoleToAdmin(db *gorm.DB, superAdminRole *model.Role) {
	// 查找 admin 用户
	var admin model.Admin
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		log.Printf("未找到 admin 用户: %v", err)
		return
	}

	// 检查是否已经分配了角色
	var count int64
	db.Model(&model.AdminRole{}).
		Where("admin_id = ? AND role_id = ?", admin.ID, superAdminRole.ID).
		Count(&count)

	if count > 0 {
		log.Println("admin 用户已经是超级管理员")
		return
	}

	// 分配超级管理员角色
	if err := db.Create(&model.AdminRole{
		AdminID: admin.ID,
		RoleID:  superAdminRole.ID,
	}).Error; err != nil {
		log.Printf("分配角色失败: %v", err)
	}
}

func printSummary(roles map[string]*model.Role, permissions map[string]*model.Permission) {
	fmt.Println("\n📋 初始化摘要:")
	fmt.Println("\n角色列表:")
	for name, role := range roles {
		fmt.Printf("  • %s (%s) - %s\n", role.DisplayName, name, role.Description)
	}

	fmt.Println("\n权限分类:")
	resourcePerms := make(map[string]int)
	for _, perm := range permissions {
		resourcePerms[perm.Resource]++
	}
	for resource, count := range resourcePerms {
		fmt.Printf("  • %s: %d 个权限\n", resource, count)
	}
}
