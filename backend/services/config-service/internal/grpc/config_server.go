package grpc

import (
	"context"
	"encoding/json"
	"strconv"

	pb "github.com/payment-platform/proto/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/config-service/internal/repository"
	"payment-platform/config-service/internal/service"
)

// ConfigServer gRPC服务实现
type ConfigServer struct {
	pb.UnimplementedConfigServiceServer
	configService service.ConfigService
}

// NewConfigServer 创建gRPC服务实例
func NewConfigServer(configService service.ConfigService) *ConfigServer {
	return &ConfigServer{
		configService: configService,
	}
}

// GetSystemConfig 获取系统配置
func (s *ConfigServer) GetSystemConfig(ctx context.Context, req *pb.GetSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	// 系统配置使用固定的服务名 "system"
	config, err := s.configService.GetConfig(ctx, "system", req.Key, "production")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配置不存在: %v", err)
	}

	return &pb.SystemConfigResponse{
		Config: &pb.SystemConfig{
			Id:          config.ID.String(),
			Key:         config.ConfigKey,
			Value:       config.ConfigValue,
			Description: config.Description,
			Category:    config.ValueType,
			IsPublic:    !config.IsEncrypted,
			UpdatedAt:   timestamppb.New(config.UpdatedAt),
		},
	}, nil
}

// UpdateSystemConfig 更新系统配置
func (s *ConfigServer) UpdateSystemConfig(ctx context.Context, req *pb.UpdateSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	// 先获取现有配置
	config, err := s.configService.GetConfig(ctx, "system", req.Key, "production")
	if err != nil {
		// 如果不存在则创建
		createInput := &service.CreateConfigInput{
			ServiceName: "system",
			ConfigKey:   req.Key,
			ConfigValue: req.Value,
			Description: req.Description,
			Environment: "production",
			CreatedBy:   "grpc",
		}
		config, err = s.configService.CreateConfig(ctx, createInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "创建配置失败: %v", err)
		}
	} else {
		// 更新现有配置
		updateInput := &service.UpdateConfigInput{
			ConfigValue: req.Value,
			Description: req.Description,
			UpdatedBy:   "grpc",
		}
		config, err = s.configService.UpdateConfig(ctx, config.ID, updateInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "更新配置失败: %v", err)
		}
	}

	return &pb.SystemConfigResponse{
		Config: &pb.SystemConfig{
			Id:          config.ID.String(),
			Key:         config.ConfigKey,
			Value:       config.ConfigValue,
			Description: config.Description,
			Category:    config.ValueType,
			IsPublic:    !config.IsEncrypted,
			UpdatedAt:   timestamppb.New(config.UpdatedAt),
		},
	}, nil
}

// ListSystemConfigs 列出系统配置
func (s *ConfigServer) ListSystemConfigs(ctx context.Context, req *pb.ListSystemConfigsRequest) (*pb.ListSystemConfigsResponse, error) {
	query := &repository.ConfigQuery{
		ServiceName: "system",
		Page:        int(req.Page),
		PageSize:    int(req.PageSize),
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}

	configs, total, err := s.configService.ListConfigs(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbConfigs := make([]*pb.SystemConfig, len(configs))
	for i, c := range configs {
		pbConfigs[i] = &pb.SystemConfig{
			Id:          c.ID.String(),
			Key:         c.ConfigKey,
			Value:       c.ConfigValue,
			Description: c.Description,
			Category:    c.ValueType,
			IsPublic:    !c.IsEncrypted,
			UpdatedAt:   timestamppb.New(c.UpdatedAt),
		}
	}

	return &pb.ListSystemConfigsResponse{
		Configs: pbConfigs,
		Total:   total,
	}, nil
}

// GetMerchantConfig 获取商户配置
func (s *ConfigServer) GetMerchantConfig(ctx context.Context, req *pb.GetMerchantConfigRequest) (*pb.MerchantConfigResponse, error) {
	// 商户配置使用配置键 "merchant_config:{merchant_id}"
	configKey := "merchant_config:" + req.MerchantId
	config, err := s.configService.GetConfig(ctx, "merchant", configKey, "production")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "商户配置不存在: %v", err)
	}

	// 解析配置值为商户配置结构
	var merchantConfig pb.MerchantConfig
	if err := json.Unmarshal([]byte(config.ConfigValue), &merchantConfig); err != nil {
		return nil, status.Errorf(codes.Internal, "解析配置失败: %v", err)
	}
	merchantConfig.MerchantId = req.MerchantId
	merchantConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.MerchantConfigResponse{
		Config: &merchantConfig,
	}, nil
}

// UpdateMerchantConfig 更新商户配置
func (s *ConfigServer) UpdateMerchantConfig(ctx context.Context, req *pb.UpdateMerchantConfigRequest) (*pb.MerchantConfigResponse, error) {
	configKey := "merchant_config:" + req.MerchantId

	// 构建配置值
	merchantConfig := &pb.MerchantConfig{
		MerchantId:          req.MerchantId,
		AutoSettlement:      req.AutoSettlement,
		SettlementCycle:     req.SettlementCycle,
		MinSettlementAmount: req.MinSettlementAmount,
		WebhookEnabled:      req.WebhookEnabled,
		WebhookUrl:          req.WebhookUrl,
		AllowedChannels:     req.AllowedChannels,
		Extra:               req.Extra,
	}

	configValue, err := json.Marshal(merchantConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "序列化配置失败: %v", err)
	}

	// 先尝试获取现有配置
	config, err := s.configService.GetConfig(ctx, "merchant", configKey, "production")
	if err != nil {
		// 创建新配置
		createInput := &service.CreateConfigInput{
			ServiceName: "merchant",
			ConfigKey:   configKey,
			ConfigValue: string(configValue),
			Environment: "production",
			Description: "商户配置",
			CreatedBy:   "grpc",
		}
		config, err = s.configService.CreateConfig(ctx, createInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "创建配置失败: %v", err)
		}
	} else {
		// 更新现有配置
		updateInput := &service.UpdateConfigInput{
			ConfigValue: string(configValue),
			UpdatedBy:   "grpc",
		}
		config, err = s.configService.UpdateConfig(ctx, config.ID, updateInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "更新配置失败: %v", err)
		}
	}

	merchantConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.MerchantConfigResponse{
		Config: merchantConfig,
	}, nil
}

// GetChannelConfig 获取渠道配置
func (s *ConfigServer) GetChannelConfig(ctx context.Context, req *pb.GetChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	configKey := "channel_config:" + req.MerchantId + ":" + req.Channel
	config, err := s.configService.GetConfig(ctx, "channel", configKey, "production")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "渠道配置不存在: %v", err)
	}

	var channelConfig pb.ChannelConfig
	if err := json.Unmarshal([]byte(config.ConfigValue), &channelConfig); err != nil {
		return nil, status.Errorf(codes.Internal, "解析配置失败: %v", err)
	}
	channelConfig.Id = config.ID.String()
	channelConfig.MerchantId = req.MerchantId
	channelConfig.Channel = req.Channel
	channelConfig.CreatedAt = timestamppb.New(config.CreatedAt)
	channelConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.ChannelConfigResponse{
		Config: &channelConfig,
	}, nil
}

// UpdateChannelConfig 更新渠道配置
func (s *ConfigServer) UpdateChannelConfig(ctx context.Context, req *pb.UpdateChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	configKey := "channel_config:" + req.MerchantId + ":" + req.Channel

	channelConfig := &pb.ChannelConfig{
		MerchantId:  req.MerchantId,
		Channel:     req.Channel,
		Credentials: req.Credentials,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
	}

	configValue, err := json.Marshal(channelConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "序列化配置失败: %v", err)
	}

	config, err := s.configService.GetConfig(ctx, "channel", configKey, "production")
	if err != nil {
		// 创建新配置
		createInput := &service.CreateConfigInput{
			ServiceName: "channel",
			ConfigKey:   configKey,
			ConfigValue: string(configValue),
			Environment: "production",
			Description: "渠道配置",
			CreatedBy:   "grpc",
		}
		config, err = s.configService.CreateConfig(ctx, createInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "创建配置失败: %v", err)
		}
	} else {
		// 更新现有配置
		updateInput := &service.UpdateConfigInput{
			ConfigValue: string(configValue),
			UpdatedBy:   "grpc",
		}
		config, err = s.configService.UpdateConfig(ctx, config.ID, updateInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "更新配置失败: %v", err)
		}
	}

	channelConfig.Id = config.ID.String()
	channelConfig.CreatedAt = timestamppb.New(config.CreatedAt)
	channelConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.ChannelConfigResponse{
		Config: channelConfig,
	}, nil
}

// ListChannelConfigs 列出渠道配置
func (s *ConfigServer) ListChannelConfigs(ctx context.Context, req *pb.ListChannelConfigsRequest) (*pb.ListChannelConfigsResponse, error) {
	query := &repository.ConfigQuery{
		ServiceName: "channel",
		ConfigKey:   "channel_config:" + req.MerchantId,
		Page:        1,
		PageSize:    100,
	}

	configs, _, err := s.configService.ListConfigs(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbConfigs := make([]*pb.ChannelConfig, 0)
	for _, c := range configs {
		var channelConfig pb.ChannelConfig
		if err := json.Unmarshal([]byte(c.ConfigValue), &channelConfig); err != nil {
			continue
		}

		// 如果只需要启用的配置，进行过滤
		if req.EnabledOnly && !channelConfig.Enabled {
			continue
		}

		channelConfig.Id = c.ID.String()
		channelConfig.CreatedAt = timestamppb.New(c.CreatedAt)
		channelConfig.UpdatedAt = timestamppb.New(c.UpdatedAt)
		pbConfigs = append(pbConfigs, &channelConfig)
	}

	return &pb.ListChannelConfigsResponse{
		Configs: pbConfigs,
	}, nil
}

// GetFeeConfig 获取费率配置
func (s *ConfigServer) GetFeeConfig(ctx context.Context, req *pb.GetFeeConfigRequest) (*pb.FeeConfigResponse, error) {
	configKey := "fee_config:" + req.MerchantId + ":" + req.Channel + ":" + req.PaymentMethod
	config, err := s.configService.GetConfig(ctx, "fee", configKey, "production")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "费率配置不存在: %v", err)
	}

	var feeConfig pb.FeeConfig
	if err := json.Unmarshal([]byte(config.ConfigValue), &feeConfig); err != nil {
		return nil, status.Errorf(codes.Internal, "解析配置失败: %v", err)
	}
	feeConfig.Id = config.ID.String()
	feeConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.FeeConfigResponse{
		Config: &feeConfig,
	}, nil
}

// UpdateFeeConfig 更新费率配置
func (s *ConfigServer) UpdateFeeConfig(ctx context.Context, req *pb.UpdateFeeConfigRequest) (*pb.FeeConfigResponse, error) {
	configKey := "fee_config:" + req.MerchantId + ":" + req.Channel + ":" + req.PaymentMethod

	feeConfig := &pb.FeeConfig{
		MerchantId:    req.MerchantId,
		Channel:       req.Channel,
		PaymentMethod: req.PaymentMethod,
		FeeRate:       req.FeeRate,
		FixedFee:      req.FixedFee,
		MinFee:        req.MinFee,
		MaxFee:        req.MaxFee,
	}

	configValue, err := json.Marshal(feeConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "序列化配置失败: %v", err)
	}

	config, err := s.configService.GetConfig(ctx, "fee", configKey, "production")
	if err != nil {
		// 创建新配置
		createInput := &service.CreateConfigInput{
			ServiceName: "fee",
			ConfigKey:   configKey,
			ConfigValue: string(configValue),
			Environment: "production",
			Description: "费率配置",
			CreatedBy:   "grpc",
		}
		config, err = s.configService.CreateConfig(ctx, createInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "创建配置失败: %v", err)
		}
	} else {
		// 更新现有配置
		updateInput := &service.UpdateConfigInput{
			ConfigValue: string(configValue),
			UpdatedBy:   "grpc",
		}
		config, err = s.configService.UpdateConfig(ctx, config.ID, updateInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "更新配置失败: %v", err)
		}
	}

	feeConfig.Id = config.ID.String()
	feeConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.FeeConfigResponse{
		Config: feeConfig,
	}, nil
}

// GetLimitConfig 获取限额配置
func (s *ConfigServer) GetLimitConfig(ctx context.Context, req *pb.GetLimitConfigRequest) (*pb.LimitConfigResponse, error) {
	configKey := "limit_config:" + req.MerchantId + ":" + req.LimitType
	config, err := s.configService.GetConfig(ctx, "limit", configKey, "production")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "限额配置不存在: %v", err)
	}

	var limitConfig pb.LimitConfig
	if err := json.Unmarshal([]byte(config.ConfigValue), &limitConfig); err != nil {
		return nil, status.Errorf(codes.Internal, "解析配置失败: %v", err)
	}
	limitConfig.Id = config.ID.String()
	limitConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.LimitConfigResponse{
		Config: &limitConfig,
	}, nil
}

// UpdateLimitConfig 更新限额配置
func (s *ConfigServer) UpdateLimitConfig(ctx context.Context, req *pb.UpdateLimitConfigRequest) (*pb.LimitConfigResponse, error) {
	configKey := "limit_config:" + req.MerchantId + ":" + req.LimitType

	limitConfig := &pb.LimitConfig{
		MerchantId: req.MerchantId,
		LimitType:  req.LimitType,
		SingleMin:  req.SingleMin,
		SingleMax:  req.SingleMax,
		DailyMax:   req.DailyMax,
		MonthlyMax: req.MonthlyMax,
	}

	configValue, err := json.Marshal(limitConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "序列化配置失败: %v", err)
	}

	config, err := s.configService.GetConfig(ctx, "limit", configKey, "production")
	if err != nil {
		// 创建新配置
		createInput := &service.CreateConfigInput{
			ServiceName: "limit",
			ConfigKey:   configKey,
			ConfigValue: string(configValue),
			Environment: "production",
			Description: "限额配置",
			CreatedBy:   "grpc",
		}
		config, err = s.configService.CreateConfig(ctx, createInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "创建配置失败: %v", err)
		}
	} else {
		// 更新现有配置
		updateInput := &service.UpdateConfigInput{
			ConfigValue: string(configValue),
			UpdatedBy:   "grpc",
		}
		config, err = s.configService.UpdateConfig(ctx, config.ID, updateInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "更新配置失败: %v", err)
		}
	}

	limitConfig.Id = config.ID.String()
	limitConfig.UpdatedAt = timestamppb.New(config.UpdatedAt)

	return &pb.LimitConfigResponse{
		Config: limitConfig,
	}, nil
}

// Helper function to convert map to structpb.Struct
func mapToStruct(m map[string]interface{}) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}
	return structpb.NewStruct(m)
}

// Helper function to convert structpb.Struct to map
func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

// Helper function to parse int64 from string
func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// Helper function to parseFloat64 from string
func parseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
