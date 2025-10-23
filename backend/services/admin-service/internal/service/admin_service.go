package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

var (
	ErrAdminNotFound      = errors.New("管理员不存在")
	ErrAdminExists        = errors.New("管理员已存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrInvalidStatus      = errors.New("管理员状态异常")
)

// AdminService 管理员服务接口
type AdminService interface {
	// 管理员管理
	CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*model.Admin, error)
	GetAdmin(ctx context.Context, id uuid.UUID) (*model.Admin, error)
	ListAdmins(ctx context.Context, page, pageSize int, status, keyword string) ([]*model.Admin, int64, error)
	UpdateAdmin(ctx context.Context, req *UpdateAdminRequest) (*model.Admin, error)
	DeleteAdmin(ctx context.Context, id uuid.UUID) error
	Login(ctx context.Context, username, password, ip string) (*LoginResponse, error)
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, adminID uuid.UUID, newPassword string) error
}

type adminService struct {
	adminRepo repository.AdminRepository
	roleRepo  repository.RoleRepository
	jwtManager *auth.JWTManager
}

// NewAdminService 创建管理员服务实例
func NewAdminService(
	adminRepo repository.AdminRepository,
	roleRepo repository.RoleRepository,
	jwtManager *auth.JWTManager,
) AdminService {
	return &adminService{
		adminRepo:  adminRepo,
		roleRepo:   roleRepo,
		jwtManager: jwtManager,
	}
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string
	Email    string
	Password string
	FullName string
	Phone    string
	RoleIDs  []uuid.UUID
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	ID       uuid.UUID
	Email    string
	FullName string
	Phone    string
	Status   string
	RoleIDs  []uuid.UUID
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string
	RefreshToken string
	Admin        *model.Admin
	ExpiresIn    int64
}

// CreateAdmin 创建管理员
func (s *adminService) CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*model.Admin, error) {
	// 检查用户名是否已存在
	existingAdmin, err := s.adminRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingAdmin != nil {
		return nil, ErrAdminExists
	}

	// 检查邮箱是否已存在
	existingAdmin, err = s.adminRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingAdmin != nil {
		return nil, ErrAdminExists
	}

	// 哈希密码
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建管理员
	admin := &model.Admin{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Status:       "active",
		IsSuper:      false,
	}

	// 加载角色
	if len(req.RoleIDs) > 0 {
		roles := make([]model.Role, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			role, err := s.roleRepo.GetByID(ctx, roleID)
			if err != nil {
				return nil, err
			}
			if role != nil {
				roles = append(roles, *role)
			}
		}
		admin.Roles = roles
	}

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// GetAdmin 获取管理员详情
func (s *adminService) GetAdmin(ctx context.Context, id uuid.UUID) (*model.Admin, error) {
	admin, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrAdminNotFound
	}
	return admin, nil
}

// ListAdmins 获取管理员列表
func (s *adminService) ListAdmins(ctx context.Context, page, pageSize int, status, keyword string) ([]*model.Admin, int64, error) {
	return s.adminRepo.List(ctx, page, pageSize, status, keyword)
}

// UpdateAdmin 更新管理员
func (s *adminService) UpdateAdmin(ctx context.Context, req *UpdateAdminRequest) (*model.Admin, error) {
	// 获取现有管理员
	admin, err := s.adminRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrAdminNotFound
	}

	// 更新字段
	if req.Email != "" {
		admin.Email = req.Email
	}
	if req.FullName != "" {
		admin.FullName = req.FullName
	}
	if req.Phone != "" {
		admin.Phone = req.Phone
	}
	if req.Status != "" {
		admin.Status = req.Status
	}

	// 更新角色
	if len(req.RoleIDs) > 0 {
		roles := make([]model.Role, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			role, err := s.roleRepo.GetByID(ctx, roleID)
			if err != nil {
				return nil, err
			}
			if role != nil {
				roles = append(roles, *role)
			}
		}
		admin.Roles = roles
	}

	if err := s.adminRepo.Update(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// DeleteAdmin 删除管理员
func (s *adminService) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	admin, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrAdminNotFound
	}

	// 超级管理员不可删除
	if admin.IsSuper {
		return errors.New("超级管理员不可删除")
	}

	return s.adminRepo.Delete(ctx, id)
}

// Login 管理员登录
func (s *adminService) Login(ctx context.Context, username, password, ip string) (*LoginResponse, error) {
	// 获取管理员
	admin, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := auth.VerifyPassword(password, admin.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 检查状态
	if admin.Status != "active" {
		return nil, ErrInvalidStatus
	}

	// 提取角色和权限
	roles := make([]string, 0, len(admin.Roles))
	permissionsMap := make(map[string]bool)
	for _, role := range admin.Roles {
		roles = append(roles, role.Name)
		for _, perm := range role.Permissions {
			permissionsMap[perm.Code] = true
		}
	}

	permissions := make([]string, 0, len(permissionsMap))
	for code := range permissionsMap {
		permissions = append(permissions, code)
	}

	// 生成JWT Token
	token, err := s.jwtManager.GenerateToken(
		admin.ID,
		admin.Username,
		"admin",
		nil, // 管理员不需要tenant_id
		roles,
		permissions,
	)
	if err != nil {
		return nil, err
	}

	// 生成 Refresh Token（有效期更长）
	refreshToken, err := s.jwtManager.GenerateToken(
		admin.ID,
		admin.Username,
		"admin",
		nil,
		roles,
		permissions,
	)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	if err := s.adminRepo.UpdateLastLogin(ctx, admin.ID, ip); err != nil {
		// 不影响登录流程，记录日志即可
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Admin:        admin,
		ExpiresIn:    int64(time.Hour * 24 / time.Second), // 24小时
	}, nil
}

// ChangePassword 修改密码
func (s *adminService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	admin, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrAdminNotFound
	}

	// 验证旧密码
	if err := auth.VerifyPassword(oldPassword, admin.PasswordHash); err != nil {
		return errors.New("旧密码错误")
	}

	// 哈希新密码
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	admin.PasswordHash = hashedPassword
	return s.adminRepo.Update(ctx, admin)
}

// ResetPassword 重置密码（管理员为其他用户重置密码）
func (s *adminService) ResetPassword(ctx context.Context, adminID uuid.UUID, newPassword string) error {
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrAdminNotFound
	}

	// 超级管理员密码不能被重置
	if admin.IsSuper {
		return errors.New("超级管理员密码不能被重置")
	}

	// 哈希新密码
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	admin.PasswordHash = hashedPassword
	return s.adminRepo.Update(ctx, admin)
}
