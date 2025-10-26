package utils

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/service"
)

// AuditHelper 审计日志辅助工具
type AuditHelper struct {
	auditLogService service.AuditLogService
}

// NewAuditHelper 创建审计助手
func NewAuditHelper(auditLogService service.AuditLogService) *AuditHelper {
	return &AuditHelper{
		auditLogService: auditLogService,
	}
}

// LogCrossTenan访问跨租户访问
func (h *AuditHelper) LogCrossTenantAccess(c *gin.Context, action, resource, resourceID, targetMerchantID string, statusCode int) {
	go func() {
		adminID := c.GetString("user_id")
		adminUsername := c.GetString("username")
		reason := c.GetString("operation_reason")

		if reason == "" {
			reason = c.Query("reason")
		}

		adminUUID, _ := uuid.Parse(adminID)

		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       action,
			Resource:     resource,
			ResourceID:   resourceID,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
			RequestBody:  getRequestBody(c),
		}

		_ = h.auditLogService.CreateLog(context.Background(), logReq)
	}()
}

// LogSensitiveOperation 记录敏感操作
func (h *AuditHelper) LogSensitiveOperation(c *gin.Context, operation, target string, success bool) {
	go func() {
		adminID := c.GetString("user_id")
		adminUsername := c.GetString("username")
		reason := c.GetString("operation_reason")

		adminUUID, _ := uuid.Parse(adminID)

		status := "success"
		statusCode := 200
		if !success {
			status = "failed"
			statusCode = 500
		}

		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       operation,
			Resource:     "sensitive_operation",
			ResourceID:   target,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason + " | Status: " + status,
			ResponseCode: statusCode,
		}

		_ = h.auditLogService.CreateLog(context.Background(), logReq)
	}()
}

// getRequestBody 获取请求体（仅用于审计，不包含敏感数据）
func getRequestBody(c *gin.Context) string {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return ""
	}

	// 移除敏感字段
	delete(body, "password")
	delete(body, "secret")
	delete(body, "api_key")
	delete(body, "access_token")

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}
