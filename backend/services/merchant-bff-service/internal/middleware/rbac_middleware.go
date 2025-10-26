package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 权限定义（可以从数据库加载）
var permissionMap = map[string][]string{
	// 超级管理员 - 所有权限
	"super_admin": {
		"*", // 通配符表示所有权限
	},

	// 运营管理员 - 商户和订单管理
	"operator": {
		"merchants.view",
		"merchants.approve",
		"merchants.freeze",
		"orders.view",
		"payments.view",
		"kyc.view",
		"kyc.approve",
		"analytics.view",
	},

	// 财务管理员 - 财务相关
	"finance": {
		"accounting.view",
		"settlements.view",
		"settlements.approve",
		"withdrawals.view",
		"withdrawals.approve",
		"reconciliation.view",
		"invoices.view",
	},

	// 风控管理员 - 风控和争议
	"risk_manager": {
		"risk.view",
		"risk.manage",
		"disputes.view",
		"disputes.manage",
		"orders.view",
		"payments.view",
	},

	// 客服 - 只读权限
	"support": {
		"merchants.view",
		"orders.view",
		"payments.view",
		"disputes.view",
	},

	// 审计员 - 审计日志查看
	"auditor": {
		"audit_logs.view",
		"analytics.view",
	},
}

// RequirePermission RBAC权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取用户角色（从JWT claims）
		roles := c.GetStringSlice("roles")
		if len(roles) == 0 {
			// 尝试从单个role字段获取
			if role := c.GetString("role"); role != "" {
				roles = []string{role}
			}
		}

		if len(roles) == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "未找到用户角色",
				"code":  "MISSING_ROLE",
			})
			c.Abort()
			return
		}

		// 2. 检查是否有所需权限
		hasPermission := false
		for _, role := range roles {
			permissions, exists := permissionMap[role]
			if !exists {
				continue
			}

			// 检查通配符
			for _, p := range permissions {
				if p == "*" {
					hasPermission = true
					break
				}
				if p == permission {
					hasPermission = true
					break
				}
				// 支持前缀匹配 (如 "merchants.*" 匹配 "merchants.view")
				if strings.HasSuffix(p, ".*") {
					prefix := strings.TrimSuffix(p, ".*")
					if strings.HasPrefix(permission, prefix+".") {
						hasPermission = true
						break
					}
				}
			}

			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "权限不足",
				"code":       "INSUFFICIENT_PERMISSION",
				"required":   permission,
				"user_roles": roles,
			})
			c.Abort()
			return
		}

		// 3. 权限验证通过，继续
		c.Next()
	}
}

// RequireAnyPermission 需要任一权限即可
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := c.GetStringSlice("roles")
		if len(roles) == 0 {
			if role := c.GetString("role"); role != "" {
				roles = []string{role}
			}
		}

		if len(roles) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "未找到用户角色"})
			c.Abort()
			return
		}

		hasPermission := false
		for _, role := range roles {
			perms, exists := permissionMap[role]
			if !exists {
				continue
			}

			for _, p := range perms {
				if p == "*" {
					hasPermission = true
					break
				}
				for _, required := range permissions {
					if p == required {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "权限不足",
				"required": permissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireReason 敏感操作必须提供原因
func RequireReason(c *gin.Context) {
	reason := c.Query("reason")
	if reason == "" {
		// 尝试从body获取
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if r, ok := body["reason"].(string); ok {
				reason = r
			}
		}
	}

	if reason == "" || len(reason) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "必须提供操作原因 (reason参数，至少5个字符)",
			"code":  "REASON_REQUIRED",
			"examples": []string{
				"客户投诉调查",
				"风险审核",
				"合规检查",
				"商户申诉处理",
			},
		})
		c.Abort()
		return
	}

	// 将reason存入context供后续使用
	c.Set("operation_reason", reason)
	c.Next()
}

// CheckIPWhitelist IP白名单检查（可选）
func CheckIPWhitelist(whitelist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(whitelist) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		allowed := false

		for _, ip := range whitelist {
			if ip == clientIP || ip == "*" {
				allowed = true
				break
			}
			// 支持CIDR格式（简化版，只支持前缀匹配）
			if strings.HasSuffix(ip, ".*") {
				prefix := strings.TrimSuffix(ip, ".*")
				if strings.HasPrefix(clientIP, prefix) {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "IP地址不在白名单中",
				"client_ip": clientIP,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HasPermission 检查用户是否有指定权限（辅助函数）
func HasPermission(roles []string, permission string) bool {
	for _, role := range roles {
		permissions, exists := permissionMap[role]
		if !exists {
			continue
		}

		for _, p := range permissions {
			if p == "*" || p == permission {
				return true
			}
			if strings.HasSuffix(p, ".*") {
				prefix := strings.TrimSuffix(p, ".*")
				if strings.HasPrefix(permission, prefix+".") {
					return true
				}
			}
		}
	}
	return false
}
