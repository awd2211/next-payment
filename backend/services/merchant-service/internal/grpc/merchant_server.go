package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/payment-platform/proto/merchant"
	"payment-platform/merchant-service/internal/service"
)

// MerchantServer gRPC服务实现
type MerchantServer struct {
	pb.UnimplementedMerchantServiceServer
	merchantService service.MerchantService
}

// NewMerchantServer 创建gRPC服务实例
func NewMerchantServer(merchantService service.MerchantService) *MerchantServer {
	return &MerchantServer{
		merchantService: merchantService,
	}
}

// RegisterMerchant 商户注册
func (s *MerchantServer) RegisterMerchant(ctx context.Context, req *pb.RegisterMerchantRequest) (*pb.MerchantResponse, error) {
	input := &service.RegisterMerchantInput{
		Name:         req.Name,
		Email:        req.Email,
		Password:     req.Password,
		CompanyName:  req.CompanyName,
		BusinessType: req.BusinessType,
		Country:      req.Country,
		Website:      req.Website,
	}

	merchant, err := s.merchantService.Register(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "注册失败: %v", err)
	}

	return &pb.MerchantResponse{
		Merchant: &pb.Merchant{
			Id:           merchant.ID.String(),
			Name:         merchant.Name,
			Email:        merchant.Email,
			Phone:        merchant.Phone,
			CompanyName:  merchant.CompanyName,
			BusinessType: merchant.BusinessType,
			Country:      merchant.Country,
			Website:      merchant.Website,
			Status:       merchant.Status,
			KycStatus:    merchant.KYCStatus,
			IsTestMode:   merchant.IsTestMode,
			CreatedAt:    timestamppb.New(merchant.CreatedAt),
			UpdatedAt:    timestamppb.New(merchant.UpdatedAt),
		},
	}, nil
}

// GetMerchant 获取商户信息
func (s *MerchantServer) GetMerchant(ctx context.Context, req *pb.GetMerchantRequest) (*pb.MerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	merchant, err := s.merchantService.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "商户不存在")
	}

	return &pb.MerchantResponse{
		Merchant: &pb.Merchant{
			Id:           merchant.ID.String(),
			Name:         merchant.Name,
			Email:        merchant.Email,
			Phone:        merchant.Phone,
			CompanyName:  merchant.CompanyName,
			BusinessType: merchant.BusinessType,
			Country:      merchant.Country,
			Website:      merchant.Website,
			Status:       merchant.Status,
			KycStatus:    merchant.KYCStatus,
			IsTestMode:   merchant.IsTestMode,
			CreatedAt:    timestamppb.New(merchant.CreatedAt),
			UpdatedAt:    timestamppb.New(merchant.UpdatedAt),
		},
	}, nil
}

// ListMerchants 商户列表
func (s *MerchantServer) ListMerchants(ctx context.Context, req *pb.ListMerchantsRequest) (*pb.ListMerchantsResponse, error) {
	merchants, total, err := s.merchantService.List(
		ctx,
		int(req.Page),
		int(req.PageSize),
		req.Status,
		req.KycStatus,
		req.Keyword,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbMerchants := make([]*pb.Merchant, len(merchants))
	for i, m := range merchants {
		pbMerchants[i] = &pb.Merchant{
			Id:           m.ID.String(),
			Name:         m.Name,
			Email:        m.Email,
			Phone:        m.Phone,
			CompanyName:  m.CompanyName,
			BusinessType: m.BusinessType,
			Country:      m.Country,
			Website:      m.Website,
			Status:       m.Status,
			KycStatus:    m.KYCStatus,
			IsTestMode:   m.IsTestMode,
			CreatedAt:    timestamppb.New(m.CreatedAt),
			UpdatedAt:    timestamppb.New(m.UpdatedAt),
		}
	}

	return &pb.ListMerchantsResponse{
		Merchants: pbMerchants,
		Total:     int32(total),
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// UpdateMerchant 更新商户信息
func (s *MerchantServer) UpdateMerchant(ctx context.Context, req *pb.UpdateMerchantRequest) (*pb.MerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	input := &service.UpdateMerchantInput{
		Name:        &req.Name,
		Phone:       &req.Phone,
		CompanyName: &req.CompanyName,
		Website:     &req.Website,
	}

	merchant, err := s.merchantService.Update(ctx, id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新失败: %v", err)
	}

	return &pb.MerchantResponse{
		Merchant: &pb.Merchant{
			Id:           merchant.ID.String(),
			Name:         merchant.Name,
			Email:        merchant.Email,
			Phone:        merchant.Phone,
			CompanyName:  merchant.CompanyName,
			BusinessType: merchant.BusinessType,
			Country:      merchant.Country,
			Website:      merchant.Website,
			Status:       merchant.Status,
			KycStatus:    merchant.KYCStatus,
			IsTestMode:   merchant.IsTestMode,
			CreatedAt:    timestamppb.New(merchant.CreatedAt),
			UpdatedAt:    timestamppb.New(merchant.UpdatedAt),
		},
	}, nil
}

// UpdateMerchantStatus 更新商户状态
func (s *MerchantServer) UpdateMerchantStatus(ctx context.Context, req *pb.UpdateMerchantStatusRequest) (*pb.MerchantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	if err := s.merchantService.UpdateStatus(ctx, id, req.Status); err != nil {
		return nil, status.Errorf(codes.Internal, "更新状态失败: %v", err)
	}

	merchant, err := s.merchantService.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取商户失败: %v", err)
	}

	return &pb.MerchantResponse{
		Merchant: &pb.Merchant{
			Id:           merchant.ID.String(),
			Name:         merchant.Name,
			Email:        merchant.Email,
			Phone:        merchant.Phone,
			CompanyName:  merchant.CompanyName,
			BusinessType: merchant.BusinessType,
			Country:      merchant.Country,
			Website:      merchant.Website,
			Status:       merchant.Status,
			KycStatus:    merchant.KYCStatus,
			IsTestMode:   merchant.IsTestMode,
			CreatedAt:    timestamppb.New(merchant.CreatedAt),
			UpdatedAt:    timestamppb.New(merchant.UpdatedAt),
		},
	}, nil
}

// MerchantLogin 商户登录
func (s *MerchantServer) MerchantLogin(ctx context.Context, req *pb.MerchantLoginRequest) (*pb.MerchantLoginResponse, error) {
	loginResp, err := s.merchantService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "登录失败: %v", err)
	}

	return &pb.MerchantLoginResponse{
		Token: loginResp.Token,
		Merchant: &pb.Merchant{
			Id:           loginResp.Merchant.ID.String(),
			Name:         loginResp.Merchant.Name,
			Email:        loginResp.Merchant.Email,
			Phone:        loginResp.Merchant.Phone,
			CompanyName:  loginResp.Merchant.CompanyName,
			BusinessType: loginResp.Merchant.BusinessType,
			Country:      loginResp.Merchant.Country,
			Website:      loginResp.Merchant.Website,
			Status:       loginResp.Merchant.Status,
			KycStatus:    loginResp.Merchant.KYCStatus,
			IsTestMode:   loginResp.Merchant.IsTestMode,
			CreatedAt:    timestamppb.New(loginResp.Merchant.CreatedAt),
			UpdatedAt:    timestamppb.New(loginResp.Merchant.UpdatedAt),
		},
		ExpiresIn: 86400, // 24小时
	}, nil
}

// 以下方法暂不实现，返回未实现错误
func (s *MerchantServer) GenerateAPIKey(ctx context.Context, req *pb.GenerateAPIKeyRequest) (*pb.APIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) ListAPIKeys(ctx context.Context, req *pb.ListAPIKeysRequest) (*pb.ListAPIKeysResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) RevokeAPIKey(ctx context.Context, req *pb.RevokeAPIKeyRequest) (*pb.RevokeAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) RotateAPISecret(ctx context.Context, req *pb.RotateAPISecretRequest) (*pb.APIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) UpdateWebhookConfig(ctx context.Context, req *pb.UpdateWebhookConfigRequest) (*pb.WebhookConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) GetWebhookConfig(ctx context.Context, req *pb.GetWebhookConfigRequest) (*pb.WebhookConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) TestWebhook(ctx context.Context, req *pb.TestWebhookRequest) (*pb.TestWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) ConfigureChannel(ctx context.Context, req *pb.ConfigureChannelRequest) (*pb.ChannelConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) GetChannelConfig(ctx context.Context, req *pb.GetChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) ListChannelConfigs(ctx context.Context, req *pb.ListChannelConfigsRequest) (*pb.ListChannelConfigsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *MerchantServer) DisableChannel(ctx context.Context, req *pb.DisableChannelRequest) (*pb.DisableChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}
