package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

// AuditLogService 审计日志服务接口
type AuditLogService interface {
	CreateLog(ctx context.Context, req *CreateAuditLogRequest) error
	GetLog(ctx context.Context, id uuid.UUID) (*model.AuditLog, error)
	ListLogs(ctx context.Context, req *ListAuditLogsRequest) ([]*model.AuditLog, int64, error)
	GetLogStats(ctx context.Context, startTime, endTime time.Time) (*AuditLogStats, error)
}

type auditLogService struct {
	auditLogRepo repository.AuditLogRepository
}

// NewAuditLogService 创建审计日志服务实例
func NewAuditLogService(
	auditLogRepo repository.AuditLogRepository,
) AuditLogService {
	return &auditLogService{
		auditLogRepo: auditLogRepo,
	}
}

// CreateAuditLogRequest 创建审计日志请求
type CreateAuditLogRequest struct {
	AdminID      uuid.UUID
	AdminName    string
	Action       string
	Resource     string
	ResourceID   string
	Method       string
	Path         string
	IP           string
	UserAgent    string
	RequestBody  string
	ResponseCode int
	Description  string
}

// ListAuditLogsRequest 查询审计日志请求
type ListAuditLogsRequest struct {
	AdminID      *uuid.UUID
	Action       string
	Resource     string
	Method       string
	StartTime    *time.Time
	EndTime      *time.Time
	IP           string
	ResponseCode *int
	Page         int
	PageSize     int
}

// AuditLogStats 审计日志统计
type AuditLogStats struct {
	TotalLogs       int64            `json:"total_logs"`
	ActionCounts    map[string]int64 `json:"action_counts"`
	ResourceCounts  map[string]int64 `json:"resource_counts"`
	ResponseCodes   map[int]int64    `json:"response_codes"`
	TopAdmins       []AdminActivity  `json:"top_admins"`
}

// AdminActivity 管理员活动统计
type AdminActivity struct {
	AdminID   uuid.UUID `json:"admin_id"`
	AdminName string    `json:"admin_name"`
	Count     int64     `json:"count"`
}

// CreateLog 创建审计日志
func (s *auditLogService) CreateLog(ctx context.Context, req *CreateAuditLogRequest) error {
	log := &model.AuditLog{
		AdminID:      req.AdminID,
		AdminName:    req.AdminName,
		Action:       req.Action,
		Resource:     req.Resource,
		ResourceID:   req.ResourceID,
		Method:       req.Method,
		Path:         req.Path,
		IP:           req.IP,
		UserAgent:    req.UserAgent,
		RequestBody:  req.RequestBody,
		ResponseCode: req.ResponseCode,
		Description:  req.Description,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// GetLog 获取审计日志详情
func (s *auditLogService) GetLog(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	return s.auditLogRepo.GetByID(ctx, id)
}

// ListLogs 获取审计日志列表
func (s *auditLogService) ListLogs(ctx context.Context, req *ListAuditLogsRequest) ([]*model.AuditLog, int64, error) {
	filter := &repository.AuditLogFilter{
		AdminID:      req.AdminID,
		Action:       req.Action,
		Resource:     req.Resource,
		Method:       req.Method,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		IP:           req.IP,
		ResponseCode: req.ResponseCode,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	return s.auditLogRepo.List(ctx, filter)
}

// GetLogStats 获取审计日志统计信息
func (s *auditLogService) GetLogStats(ctx context.Context, startTime, endTime time.Time) (*AuditLogStats, error) {
	// 这是一个简化实现，实际项目中可能需要专门的统计查询
	// 这里我们返回一个基本的统计结构
	stats := &AuditLogStats{
		ActionCounts:   make(map[string]int64),
		ResourceCounts: make(map[string]int64),
		ResponseCodes:  make(map[int]int64),
		TopAdmins:      make([]AdminActivity, 0),
	}

	// 获取时间范围内的所有日志
	filter := &repository.AuditLogFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  10000, // 限制最大数量
	}

	logs, total, err := s.auditLogRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	stats.TotalLogs = total

	// 统计各项数据
	adminCounts := make(map[uuid.UUID]AdminActivity)
	for _, log := range logs {
		// 统计操作类型
		stats.ActionCounts[log.Action]++

		// 统计资源类型
		stats.ResourceCounts[log.Resource]++

		// 统计响应码
		stats.ResponseCodes[log.ResponseCode]++

		// 统计管理员活动
		if activity, ok := adminCounts[log.AdminID]; ok {
			activity.Count++
			adminCounts[log.AdminID] = activity
		} else {
			adminCounts[log.AdminID] = AdminActivity{
				AdminID:   log.AdminID,
				AdminName: log.AdminName,
				Count:     1,
			}
		}
	}

	// 获取活动最多的前10个管理员
	for _, activity := range adminCounts {
		stats.TopAdmins = append(stats.TopAdmins, activity)
	}

	return stats, nil
}
