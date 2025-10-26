package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/service"
)

// AdminServer gRPC服务实现
type AdminServer struct {
	pb.UnimplementedAdminServiceServer
	adminService        service.AdminService
	roleService         service.RoleService
	permissionService   service.PermissionService
	systemConfigService service.SystemConfigService
	auditLogService     service.AuditLogService
}

// NewAdminServer 创建gRPC服务实例
func NewAdminServer(
	adminService service.AdminService,
	roleService service.RoleService,
	permissionService service.PermissionService,
	systemConfigService service.SystemConfigService,
	auditLogService service.AuditLogService,
) *AdminServer {
	return &AdminServer{
		adminService:        adminService,
		roleService:         roleService,
		permissionService:   permissionService,
		systemConfigService: systemConfigService,
		auditLogService:     auditLogService,
	}
}

// ========== 管理员管理 ==========

// CreateAdmin 创建管理员
func (s *AdminServer) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.AdminResponse, error) {
	// 转换role_ids
	roleIDs := make([]uuid.UUID, 0, len(req.RoleIds))
	for _, idStr := range req.RoleIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的角色ID: %s", idStr)
		}
		roleIDs = append(roleIDs, id)
	}

	// 调用服务层
	admin, err := s.adminService.CreateAdmin(ctx, &service.CreateAdminRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
		RoleIDs:  roleIDs,
	})
	if err != nil {
		if err == service.ErrAdminExists {
			return nil, status.Errorf(codes.AlreadyExists, "管理员已存在")
		}
		return nil, status.Errorf(codes.Internal, "创建管理员失败: %v", err)
	}

	return &pb.AdminResponse{
		Admin: convertAdminToProto(admin),
	}, nil
}

// GetAdmin 获取管理员
func (s *AdminServer) GetAdmin(ctx context.Context, req *pb.GetAdminRequest) (*pb.AdminResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的管理员ID")
	}

	admin, err := s.adminService.GetAdmin(ctx, id)
	if err != nil {
		if err == service.ErrAdminNotFound {
			return nil, status.Errorf(codes.NotFound, "管理员不存在")
		}
		return nil, status.Errorf(codes.Internal, "获取管理员失败: %v", err)
	}

	return &pb.AdminResponse{
		Admin: convertAdminToProto(admin),
	}, nil
}

// ListAdmins 获取管理员列表
func (s *AdminServer) ListAdmins(ctx context.Context, req *pb.ListAdminsRequest) (*pb.ListAdminsResponse, error) {
	admins, total, err := s.adminService.ListAdmins(ctx, int(req.Page), int(req.PageSize), req.Status, req.Keyword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取管理员列表失败: %v", err)
	}

	pbAdmins := make([]*pb.Admin, 0, len(admins))
	for _, admin := range admins {
		pbAdmins = append(pbAdmins, convertAdminToProto(admin))
	}

	return &pb.ListAdminsResponse{
		Admins:   pbAdmins,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateAdmin 更新管理员
func (s *AdminServer) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.AdminResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的管理员ID")
	}

	// 转换role_ids
	roleIDs := make([]uuid.UUID, 0, len(req.RoleIds))
	for _, idStr := range req.RoleIds {
		roleID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的角色ID: %s", idStr)
		}
		roleIDs = append(roleIDs, roleID)
	}

	admin, err := s.adminService.UpdateAdmin(ctx, &service.UpdateAdminRequest{
		ID:       id,
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Status:   req.Status,
		RoleIDs:  roleIDs,
	})
	if err != nil {
		if err == service.ErrAdminNotFound {
			return nil, status.Errorf(codes.NotFound, "管理员不存在")
		}
		return nil, status.Errorf(codes.Internal, "更新管理员失败: %v", err)
	}

	return &pb.AdminResponse{
		Admin: convertAdminToProto(admin),
	}, nil
}

// DeleteAdmin 删除管理员
func (s *AdminServer) DeleteAdmin(ctx context.Context, req *pb.DeleteAdminRequest) (*pb.DeleteAdminResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的管理员ID")
	}

	if err := s.adminService.DeleteAdmin(ctx, id); err != nil {
		if err == service.ErrAdminNotFound {
			return nil, status.Errorf(codes.NotFound, "管理员不存在")
		}
		return nil, status.Errorf(codes.Internal, "删除管理员失败: %v", err)
	}

	return &pb.DeleteAdminResponse{
		Success: true,
	}, nil
}

// AdminLogin 管理员登录
func (s *AdminServer) AdminLogin(ctx context.Context, req *pb.AdminLoginRequest) (*pb.AdminLoginResponse, error) {
	loginResp, err := s.adminService.Login(ctx, req.Username, req.Password, req.Ip)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return nil, status.Errorf(codes.Unauthenticated, "用户名或密码错误")
		}
		if err == service.ErrInvalidStatus {
			return nil, status.Errorf(codes.PermissionDenied, "管理员账号已被禁用")
		}
		return nil, status.Errorf(codes.Internal, "登录失败: %v", err)
	}

	return &pb.AdminLoginResponse{
		Token:        loginResp.Token,
		RefreshToken: loginResp.RefreshToken,
		Admin:        convertAdminToProto(loginResp.Admin),
		ExpiresIn:    loginResp.ExpiresIn,
	}, nil
}

// ========== 角色管理 ==========

// CreateRole 创建角色
func (s *AdminServer) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.RoleResponse, error) {
	// 转换permission_ids
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIds))
	for _, idStr := range req.PermissionIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的权限ID: %s", idStr)
		}
		permissionIDs = append(permissionIDs, id)
	}

	role, err := s.roleService.CreateRole(ctx, &service.CreateRoleRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		PermissionIDs: permissionIDs,
	})
	if err != nil {
		if err == service.ErrRoleExists {
			return nil, status.Errorf(codes.AlreadyExists, "角色已存在")
		}
		return nil, status.Errorf(codes.Internal, "创建角色失败: %v", err)
	}

	return &pb.RoleResponse{
		Role: convertRoleToProto(role),
	}, nil
}

// ListRoles 获取角色列表
func (s *AdminServer) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	roles, total, err := s.roleService.ListRoles(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取角色列表失败: %v", err)
	}

	pbRoles := make([]*pb.Role, 0, len(roles))
	for _, role := range roles {
		pbRoles = append(pbRoles, convertRoleToProto(role))
	}

	return &pb.ListRolesResponse{
		Roles: pbRoles,
		Total: int32(total),
	}, nil
}

// UpdateRole 更新角色
func (s *AdminServer) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.RoleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的角色ID")
	}

	// 转换permission_ids
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIds))
	for _, idStr := range req.PermissionIds {
		permID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的权限ID: %s", idStr)
		}
		permissionIDs = append(permissionIDs, permID)
	}

	role, err := s.roleService.UpdateRole(ctx, &service.UpdateRoleRequest{
		ID:            id,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		PermissionIDs: permissionIDs,
	})
	if err != nil {
		if err == service.ErrRoleNotFound {
			return nil, status.Errorf(codes.NotFound, "角色不存在")
		}
		return nil, status.Errorf(codes.Internal, "更新角色失败: %v", err)
	}

	return &pb.RoleResponse{
		Role: convertRoleToProto(role),
	}, nil
}

// AssignRoleToAdmin 为管理员分配角色
func (s *AdminServer) AssignRoleToAdmin(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	adminID, err := uuid.Parse(req.AdminId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的管理员ID")
	}

	// 为管理员分配多个角色
	for _, roleIDStr := range req.RoleIds {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的角色ID: %s", roleIDStr)
		}

		if err := s.roleService.AssignRoleToAdmin(ctx, adminID, roleID); err != nil {
			return nil, status.Errorf(codes.Internal, "分配角色失败: %v", err)
		}
	}

	return &pb.AssignRoleResponse{
		Success: true,
	}, nil
}

// ========== 权限管理 ==========

// ListPermissions 获取权限列表
func (s *AdminServer) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	permissions, err := s.permissionService.ListPermissions(ctx, req.Resource)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取权限列表失败: %v", err)
	}

	pbPermissions := make([]*pb.Permission, 0, len(permissions))
	for _, perm := range permissions {
		pbPermissions = append(pbPermissions, convertPermissionToProto(perm))
	}

	return &pb.ListPermissionsResponse{
		Permissions: pbPermissions,
	}, nil
}

// AssignPermissionToRole 为角色分配权限
func (s *AdminServer) AssignPermissionToRole(ctx context.Context, req *pb.AssignPermissionRequest) (*pb.AssignPermissionResponse, error) {
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的角色ID")
	}

	// 转换permission_ids
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIds))
	for _, idStr := range req.PermissionIds {
		permID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的权限ID: %s", idStr)
		}
		permissionIDs = append(permissionIDs, permID)
	}

	if err := s.roleService.AssignPermissions(ctx, roleID, permissionIDs); err != nil {
		return nil, status.Errorf(codes.Internal, "分配权限失败: %v", err)
	}

	return &pb.AssignPermissionResponse{
		Success: true,
	}, nil
}

// ========== 商户审核（Stub实现 - 需要实际业务逻辑） ==========

// ReviewMerchant 审核商户
func (s *AdminServer) ReviewMerchant(ctx context.Context, req *pb.ReviewMerchantRequest) (*pb.ReviewMerchantResponse, error) {
	// TODO: 实现商户审核逻辑
	// 这需要与merchant-service交互或实现MerchantReviewService
	return nil, status.Errorf(codes.Unimplemented, "商户审核功能待实现")
}

// ListMerchantReviews 获取商户审核列表
func (s *AdminServer) ListMerchantReviews(ctx context.Context, req *pb.ListMerchantReviewsRequest) (*pb.ListMerchantReviewsResponse, error) {
	// TODO: 实现商户审核列表查询
	return nil, status.Errorf(codes.Unimplemented, "商户审核列表功能待实现")
}

// ========== 系统配置 ==========

// GetSystemConfig 获取系统配置
func (s *AdminServer) GetSystemConfig(ctx context.Context, req *pb.GetSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	config, err := s.systemConfigService.GetConfigByKey(ctx, req.Key)
	if err != nil {
		if err == service.ErrConfigNotFound {
			return nil, status.Errorf(codes.NotFound, "配置不存在")
		}
		return nil, status.Errorf(codes.Internal, "获取配置失败: %v", err)
	}

	return &pb.SystemConfigResponse{
		Config: convertSystemConfigToProto(config),
	}, nil
}

// UpdateSystemConfig 更新系统配置
func (s *AdminServer) UpdateSystemConfig(ctx context.Context, req *pb.UpdateSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	// 先获取配置以得到ID
	config, err := s.systemConfigService.GetConfigByKey(ctx, req.Key)
	if err != nil {
		if err == service.ErrConfigNotFound {
			return nil, status.Errorf(codes.NotFound, "配置不存在")
		}
		return nil, status.Errorf(codes.Internal, "获取配置失败: %v", err)
	}

	// 更新配置
	updated, err := s.systemConfigService.UpdateConfig(ctx, &service.UpdateConfigRequest{
		ID:          config.ID,
		Key:         req.Key,
		Value:       req.Value,
		Description: req.Description,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新配置失败: %v", err)
	}

	return &pb.SystemConfigResponse{
		Config: convertSystemConfigToProto(updated),
	}, nil
}

// ListSystemConfigs 获取系统配置列表
func (s *AdminServer) ListSystemConfigs(ctx context.Context, req *pb.ListSystemConfigsRequest) (*pb.ListSystemConfigsResponse, error) {
	configs, total, err := s.systemConfigService.ListConfigs(ctx, req.Category, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取配置列表失败: %v", err)
	}

	pbConfigs := make([]*pb.SystemConfig, 0, len(configs))
	for _, config := range configs {
		pbConfigs = append(pbConfigs, convertSystemConfigToProto(config))
	}

	return &pb.ListSystemConfigsResponse{
		Configs: pbConfigs,
		Total:   int32(total),
	}, nil
}

// ========== 审批流程（Stub实现 - 需要实际业务逻辑） ==========

// CreateApproval 创建审批
func (s *AdminServer) CreateApproval(ctx context.Context, req *pb.CreateApprovalRequest) (*pb.ApprovalResponse, error) {
	// TODO: 实现审批流程创建逻辑
	return nil, status.Errorf(codes.Unimplemented, "审批流程创建功能待实现")
}

// ProcessApproval 处理审批
func (s *AdminServer) ProcessApproval(ctx context.Context, req *pb.ProcessApprovalRequest) (*pb.ApprovalResponse, error) {
	// TODO: 实现审批处理逻辑
	return nil, status.Errorf(codes.Unimplemented, "审批处理功能待实现")
}

// ListApprovals 获取审批列表
func (s *AdminServer) ListApprovals(ctx context.Context, req *pb.ListApprovalsRequest) (*pb.ListApprovalsResponse, error) {
	// TODO: 实现审批列表查询
	return nil, status.Errorf(codes.Unimplemented, "审批列表功能待实现")
}

// ========== 审计日志 ==========

// ListAuditLogs 获取审计日志列表
func (s *AdminServer) ListAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	// 转换时间戳
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		startTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		endTime = &t
	}

	// 转换admin_id
	var adminID *uuid.UUID
	if req.AdminId != "" {
		id, err := uuid.Parse(req.AdminId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的管理员ID")
		}
		adminID = &id
	}

	logs, total, err := s.auditLogService.ListLogs(ctx, &service.ListAuditLogsRequest{
		AdminID:   adminID,
		Resource:  req.Resource,
		Action:    req.Action,
		StartTime: startTime,
		EndTime:   endTime,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取审计日志失败: %v", err)
	}

	pbLogs := make([]*pb.AuditLog, 0, len(logs))
	for _, log := range logs {
		pbLogs = append(pbLogs, convertAuditLogToProto(log))
	}

	return &pb.ListAuditLogsResponse{
		Logs:  pbLogs,
		Total: int32(total),
	}, nil
}

// ========== 辅助转换函数 ==========

// convertAdminToProto 将Admin模型转换为Proto消息
func convertAdminToProto(admin *model.Admin) *pb.Admin {
	if admin == nil {
		return nil
	}

	pbAdmin := &pb.Admin{
		Id:          admin.ID.String(),
		Username:    admin.Username,
		Email:       admin.Email,
		FullName:    admin.FullName,
		Phone:       admin.Phone,
		Avatar:      admin.Avatar,
		Status:      admin.Status,
		IsSuper:     admin.IsSuper,
		LastLoginIp: admin.LastLoginIP,
		CreatedAt:   timestamppb.New(admin.CreatedAt),
		UpdatedAt:   timestamppb.New(admin.UpdatedAt),
	}

	if admin.LastLoginAt != nil {
		pbAdmin.LastLoginAt = timestamppb.New(*admin.LastLoginAt)
	}

	// 转换角色
	if len(admin.Roles) > 0 {
		pbAdmin.Roles = make([]*pb.Role, 0, len(admin.Roles))
		for _, role := range admin.Roles {
			pbAdmin.Roles = append(pbAdmin.Roles, convertRoleToProto(&role))
		}
	}

	return pbAdmin
}

// convertRoleToProto 将Role模型转换为Proto消息
func convertRoleToProto(role *model.Role) *pb.Role {
	if role == nil {
		return nil
	}

	pbRole := &pb.Role{
		Id:          role.ID.String(),
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}

	// 转换权限
	if len(role.Permissions) > 0 {
		pbRole.Permissions = make([]*pb.Permission, 0, len(role.Permissions))
		for _, perm := range role.Permissions {
			pbRole.Permissions = append(pbRole.Permissions, convertPermissionToProto(&perm))
		}
	}

	return pbRole
}

// convertPermissionToProto 将Permission模型转换为Proto消息
func convertPermissionToProto(perm *model.Permission) *pb.Permission {
	if perm == nil {
		return nil
	}

	return &pb.Permission{
		Id:          perm.ID.String(),
		Code:        perm.Code,
		Name:        perm.Name,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
	}
}

// convertSystemConfigToProto 将SystemConfig模型转换为Proto消息
func convertSystemConfigToProto(config *model.SystemConfig) *pb.SystemConfig {
	if config == nil {
		return nil
	}

	return &pb.SystemConfig{
		Id:          config.ID.String(),
		Key:         config.Key,
		Value:       config.Value,
		Type:        config.Type,
		Category:    config.Category,
		Description: config.Description,
		IsPublic:    config.IsPublic,
		CreatedAt:   timestamppb.New(config.CreatedAt),
		UpdatedAt:   timestamppb.New(config.UpdatedAt),
	}
}

// convertAuditLogToProto 将AuditLog模型转换为Proto消息
func convertAuditLogToProto(log *model.AuditLog) *pb.AuditLog {
	if log == nil {
		return nil
	}

	return &pb.AuditLog{
		Id:           log.ID.String(),
		AdminId:      log.AdminID.String(),
		AdminName:    log.AdminName,
		Action:       log.Action,
		Resource:     log.Resource,
		ResourceId:   log.ResourceID,
		Method:       log.Method,
		Path:         log.Path,
		Ip:           log.IP,
		UserAgent:    log.UserAgent,
		ResponseCode: int32(log.ResponseCode),
		Description:  log.Description,
		CreatedAt:    timestamppb.New(log.CreatedAt),
	}
}
