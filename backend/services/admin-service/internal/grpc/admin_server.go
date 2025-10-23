package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/admin-service/internal/service"
)

// AdminServer gRPC服务实现
type AdminServer struct {
	pb.UnimplementedAdminServiceServer
	adminService service.AdminService
}

// NewAdminServer 创建gRPC服务实例
func NewAdminServer(adminService service.AdminService) *AdminServer {
	return &AdminServer{
		adminService: adminService,
	}
}

// CreateAdmin 创建管理员
func (s *AdminServer) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.AdminResponse, error) {
	input := &service.CreateAdminInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
	}

	admin, err := s.adminService.CreateAdmin(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建管理员失败: %v", err)
	}

	return &pb.AdminResponse{
		Admin: &pb.Admin{
			Id:        admin.ID.String(),
			Username:  admin.Username,
			Email:     admin.Email,
			Phone:     admin.Phone,
			RealName:  admin.RealName,
			Status:    admin.Status,
			CreatedAt: timestamppb.New(admin.CreatedAt),
			UpdatedAt: timestamppb.New(admin.UpdatedAt),
		},
	}, nil
}

// GetAdmin 获取管理员
func (s *AdminServer) GetAdmin(ctx context.Context, req *pb.GetAdminRequest) (*pb.AdminResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的管理员ID")
	}

	admin, err := s.adminService.GetAdminByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "管理员不存在")
	}

	return &pb.AdminResponse{
		Admin: &pb.Admin{
			Id:        admin.ID.String(),
			Username:  admin.Username,
			Email:     admin.Email,
			Phone:     admin.Phone,
			RealName:  admin.RealName,
			Status:    admin.Status,
			CreatedAt: timestamppb.New(admin.CreatedAt),
			UpdatedAt: timestamppb.New(admin.UpdatedAt),
		},
	}, nil
}

// AdminLogin 管理员登录
func (s *AdminServer) AdminLogin(ctx context.Context, req *pb.AdminLoginRequest) (*pb.AdminLoginResponse, error) {
	loginResp, err := s.adminService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "登录失败: %v", err)
	}

	return &pb.AdminLoginResponse{
		Token: loginResp.Token,
		Admin: &pb.Admin{
			Id:        loginResp.Admin.ID.String(),
			Username:  loginResp.Admin.Username,
			Email:     loginResp.Admin.Email,
			Phone:     loginResp.Admin.Phone,
			RealName:  loginResp.Admin.RealName,
			Status:    loginResp.Admin.Status,
			CreatedAt: timestamppb.New(loginResp.Admin.CreatedAt),
			UpdatedAt: timestamppb.New(loginResp.Admin.UpdatedAt),
		},
		ExpiresIn: 86400,
	}, nil
}

// 其他方法暂时返回未实现
func (s *AdminServer) ListAdmins(ctx context.Context, req *pb.ListAdminsRequest) (*pb.ListAdminsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.AdminResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) UpdateAdminStatus(ctx context.Context, req *pb.UpdateAdminStatusRequest) (*pb.AdminResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) DeleteAdmin(ctx context.Context, req *pb.DeleteAdminRequest) (*pb.DeleteAdminResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.RoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.RoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.RoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) AssignPermissions(ctx context.Context, req *pb.AssignPermissionsRequest) (*pb.AssignPermissionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) GetSystemConfig(ctx context.Context, req *pb.GetSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) UpdateSystemConfig(ctx context.Context, req *pb.UpdateSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) ListSystemConfigs(ctx context.Context, req *pb.ListSystemConfigsRequest) (*pb.ListSystemConfigsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) GetAuditLog(ctx context.Context, req *pb.GetAuditLogRequest) (*pb.AuditLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AdminServer) ListAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}
