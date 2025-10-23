package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConfigServer gRPC服务实现
type ConfigServer struct {
	pb.UnimplementedConfigServiceServer
}

// NewConfigServer 创建gRPC服务实例
func NewConfigServer() *ConfigServer {
	return &ConfigServer{}
}

// 所有方法暂时返回未实现
func (s *ConfigServer) GetSystemConfig(ctx context.Context, req *pb.GetSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) UpdateSystemConfig(ctx context.Context, req *pb.UpdateSystemConfigRequest) (*pb.SystemConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) ListSystemConfigs(ctx context.Context, req *pb.ListSystemConfigsRequest) (*pb.ListSystemConfigsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) GetMerchantConfig(ctx context.Context, req *pb.GetMerchantConfigRequest) (*pb.MerchantConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) UpdateMerchantConfig(ctx context.Context, req *pb.UpdateMerchantConfigRequest) (*pb.MerchantConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) GetChannelConfig(ctx context.Context, req *pb.GetChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) UpdateChannelConfig(ctx context.Context, req *pb.UpdateChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) ListChannelConfigs(ctx context.Context, req *pb.ListChannelConfigsRequest) (*pb.ListChannelConfigsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) GetFeeConfig(ctx context.Context, req *pb.GetFeeConfigRequest) (*pb.FeeConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) UpdateFeeConfig(ctx context.Context, req *pb.UpdateFeeConfigRequest) (*pb.FeeConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) GetLimitConfig(ctx context.Context, req *pb.GetLimitConfigRequest) (*pb.LimitConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *ConfigServer) UpdateLimitConfig(ctx context.Context, req *pb.UpdateLimitConfigRequest) (*pb.LimitConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}
