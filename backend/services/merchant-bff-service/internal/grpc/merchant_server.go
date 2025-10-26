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
	apiKeyService   service.APIKeyService
	channelService  service.ChannelService
}

// NewMerchantServer 创建gRPC服务实例
func NewMerchantServer(merchantService service.MerchantService, apiKeyService service.APIKeyService, channelService service.ChannelService) *MerchantServer {
	return &MerchantServer{
		merchantService: merchantService,
		apiKeyService:   apiKeyService,
		channelService:  channelService,
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

// GenerateAPIKey 生成API密钥
func (s *MerchantServer) GenerateAPIKey(ctx context.Context, req *pb.GenerateAPIKeyRequest) (*pb.APIKeyResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	input := &service.CreateAPIKeyInput{
		Name:        req.Name,
		Environment: req.Environment,
	}

	// 处理过期时间
	if req.ExpiresInDays > 0 {
		expiresAt := timestamppb.Now().AsTime().AddDate(0, 0, int(req.ExpiresInDays))
		input.ExpiresAt = &expiresAt
	}

	apiKey, err := s.apiKeyService.Create(ctx, merchantID, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成API密钥失败: %v", err)
	}

	var lastUsedAt *timestamppb.Timestamp
	if apiKey.LastUsedAt != nil {
		lastUsedAt = timestamppb.New(*apiKey.LastUsedAt)
	}

	var expiresAt *timestamppb.Timestamp
	if apiKey.ExpiresAt != nil {
		expiresAt = timestamppb.New(*apiKey.ExpiresAt)
	}

	return &pb.APIKeyResponse{
		ApiKey: &pb.APIKey{
			Id:          apiKey.ID.String(),
			MerchantId:  apiKey.MerchantID.String(),
			ApiKey:      apiKey.APIKey,
			ApiSecret:   apiKey.APISecret, // 仅首次返回
			Name:        apiKey.Name,
			Environment: apiKey.Environment,
			IsActive:    apiKey.IsActive,
			LastUsedAt:  lastUsedAt,
			ExpiresAt:   expiresAt,
			CreatedAt:   timestamppb.New(apiKey.CreatedAt),
		},
	}, nil
}

// ListAPIKeys 列出API密钥
func (s *MerchantServer) ListAPIKeys(ctx context.Context, req *pb.ListAPIKeysRequest) (*pb.ListAPIKeysResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	apiKeys, err := s.apiKeyService.ListByMerchant(ctx, merchantID, req.Environment)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取API密钥列表失败: %v", err)
	}

	pbKeys := make([]*pb.APIKey, len(apiKeys))
	for i, key := range apiKeys {
		var lastUsedAt *timestamppb.Timestamp
		if key.LastUsedAt != nil {
			lastUsedAt = timestamppb.New(*key.LastUsedAt)
		}

		var expiresAt *timestamppb.Timestamp
		if key.ExpiresAt != nil {
			expiresAt = timestamppb.New(*key.ExpiresAt)
		}

		pbKeys[i] = &pb.APIKey{
			Id:          key.ID.String(),
			MerchantId:  key.MerchantID.String(),
			ApiKey:      key.APIKey,
			ApiSecret:   key.APISecret, // 列表中已被隐藏
			Name:        key.Name,
			Environment: key.Environment,
			IsActive:    key.IsActive,
			LastUsedAt:  lastUsedAt,
			ExpiresAt:   expiresAt,
			CreatedAt:   timestamppb.New(key.CreatedAt),
		}
	}

	return &pb.ListAPIKeysResponse{
		ApiKeys: pbKeys,
	}, nil
}

// RevokeAPIKey 撤销API密钥
func (s *MerchantServer) RevokeAPIKey(ctx context.Context, req *pb.RevokeAPIKeyRequest) (*pb.RevokeAPIKeyResponse, error) {
	keyID, err := uuid.Parse(req.ApiKeyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的API密钥ID")
	}

	if err := s.apiKeyService.Revoke(ctx, keyID); err != nil {
		return nil, status.Errorf(codes.Internal, "撤销API密钥失败: %v", err)
	}

	return &pb.RevokeAPIKeyResponse{
		Success: true,
	}, nil
}

// RotateAPISecret 轮换API密钥
func (s *MerchantServer) RotateAPISecret(ctx context.Context, req *pb.RotateAPISecretRequest) (*pb.APIKeyResponse, error) {
	keyID, err := uuid.Parse(req.ApiKeyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的API密钥ID")
	}

	apiKey, err := s.apiKeyService.Rotate(ctx, keyID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "轮换API密钥失败: %v", err)
	}

	var lastUsedAt *timestamppb.Timestamp
	if apiKey.LastUsedAt != nil {
		lastUsedAt = timestamppb.New(*apiKey.LastUsedAt)
	}

	var expiresAt *timestamppb.Timestamp
	if apiKey.ExpiresAt != nil {
		expiresAt = timestamppb.New(*apiKey.ExpiresAt)
	}

	return &pb.APIKeyResponse{
		ApiKey: &pb.APIKey{
			Id:          apiKey.ID.String(),
			MerchantId:  apiKey.MerchantID.String(),
			ApiKey:      apiKey.APIKey,
			ApiSecret:   apiKey.APISecret, // 返回新的secret
			Name:        apiKey.Name,
			Environment: apiKey.Environment,
			IsActive:    apiKey.IsActive,
			LastUsedAt:  lastUsedAt,
			ExpiresAt:   expiresAt,
			CreatedAt:   timestamppb.New(apiKey.CreatedAt),
		},
	}, nil
}

// UpdateWebhookConfig 更新Webhook配置 (暂未实现webhook功能)
func (s *MerchantServer) UpdateWebhookConfig(ctx context.Context, req *pb.UpdateWebhookConfigRequest) (*pb.WebhookConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Webhook功能暂未实现")
}

// GetWebhookConfig 获取Webhook配置 (暂未实现webhook功能)
func (s *MerchantServer) GetWebhookConfig(ctx context.Context, req *pb.GetWebhookConfigRequest) (*pb.WebhookConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Webhook功能暂未实现")
}

// TestWebhook 测试Webhook (暂未实现webhook功能)
func (s *MerchantServer) TestWebhook(ctx context.Context, req *pb.TestWebhookRequest) (*pb.TestWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Webhook功能暂未实现")
}

// ConfigureChannel 配置支付渠道
func (s *MerchantServer) ConfigureChannel(ctx context.Context, req *pb.ConfigureChannelRequest) (*pb.ChannelConfigResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	// 将 map[string]string 转换为 map[string]interface{}
	config := make(map[string]interface{})
	for k, v := range req.Config {
		config[k] = v
	}

	input := &service.CreateChannelInput{
		MerchantID: merchantID,
		Channel:    req.Channel,
		Config:     config,
		IsTestMode: req.IsTestMode,
	}

	channel, err := s.channelService.CreateChannel(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "配置渠道失败: %v", err)
	}

	return &pb.ChannelConfigResponse{
		Config: &pb.ChannelConfig{
			Id:         channel.ID.String(),
			MerchantId: channel.MerchantID.String(),
			Channel:    channel.Channel,
			Config:     req.Config, // 返回原始配置
			IsEnabled:  channel.IsEnabled,
			IsTestMode: channel.IsTestMode,
			CreatedAt:  timestamppb.New(channel.CreatedAt),
			UpdatedAt:  timestamppb.New(channel.UpdatedAt),
		},
	}, nil
}

// GetChannelConfig 获取渠道配置
func (s *MerchantServer) GetChannelConfig(ctx context.Context, req *pb.GetChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	channels, err := s.channelService.ListChannels(ctx, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取渠道配置失败: %v", err)
	}

	// 找到指定渠道
	var targetChannel *service.CreateChannelInput
	for _, ch := range channels {
		if ch.Channel == req.Channel {
			// 注意: 这里实际返回的是 model.ChannelConfig, 需要调整
			return &pb.ChannelConfigResponse{
				Config: &pb.ChannelConfig{
					Id:         ch.ID.String(),
					MerchantId: ch.MerchantID.String(),
					Channel:    ch.Channel,
					Config:     make(map[string]string), // Config在model中是JSON string
					IsEnabled:  ch.IsEnabled,
					IsTestMode: ch.IsTestMode,
					CreatedAt:  timestamppb.New(ch.CreatedAt),
					UpdatedAt:  timestamppb.New(ch.UpdatedAt),
				},
			}, nil
		}
	}

	if targetChannel == nil {
		return nil, status.Errorf(codes.NotFound, "渠道配置不存在")
	}

	return nil, status.Errorf(codes.Internal, "未找到渠道配置")
}

// ListChannelConfigs 列出所有渠道配置
func (s *MerchantServer) ListChannelConfigs(ctx context.Context, req *pb.ListChannelConfigsRequest) (*pb.ListChannelConfigsResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	channels, err := s.channelService.ListChannels(ctx, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取渠道配置列表失败: %v", err)
	}

	pbChannels := make([]*pb.ChannelConfig, len(channels))
	for i, ch := range channels {
		pbChannels[i] = &pb.ChannelConfig{
			Id:         ch.ID.String(),
			MerchantId: ch.MerchantID.String(),
			Channel:    ch.Channel,
			Config:     make(map[string]string), // Config在model中是JSON string，这里简化处理
			IsEnabled:  ch.IsEnabled,
			IsTestMode: ch.IsTestMode,
			CreatedAt:  timestamppb.New(ch.CreatedAt),
			UpdatedAt:  timestamppb.New(ch.UpdatedAt),
		}
	}

	return &pb.ListChannelConfigsResponse{
		Configs: pbChannels,
	}, nil
}

// DisableChannel 禁用渠道
func (s *MerchantServer) DisableChannel(ctx context.Context, req *pb.DisableChannelRequest) (*pb.DisableChannelResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	// 找到该渠道的配置ID
	channels, err := s.channelService.ListChannels(ctx, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取渠道配置失败: %v", err)
	}

	var channelID uuid.UUID
	found := false
	for _, ch := range channels {
		if ch.Channel == req.Channel {
			channelID = ch.ID
			found = true
			break
		}
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "渠道配置不存在")
	}

	if err := s.channelService.ToggleChannel(ctx, channelID, merchantID, false); err != nil {
		return nil, status.Errorf(codes.Internal, "禁用渠道失败: %v", err)
	}

	return &pb.DisableChannelResponse{
		Success: true,
	}, nil
}
