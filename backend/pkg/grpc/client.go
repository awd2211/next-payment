package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientConfig gRPC客户端配置
type ClientConfig struct {
	Address         string        // 服务地址
	Timeout         time.Duration // 连接超时时间
	MaxRecvMsgSize  int           // 最大接收消息大小
	MaxSendMsgSize  int           // 最大发送消息大小
	EnableKeepalive bool          // 启用Keepalive
}

// NewClient 创建gRPC客户端连接
func NewClient(config ClientConfig) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// 暂时使用不安全连接（生产环境应使用TLS）
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 设置最大消息大小
	if config.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(config.MaxRecvMsgSize)))
	}
	if config.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(config.MaxSendMsgSize)))
	}

	// 启用Keepalive
	if config.EnableKeepalive {
		kaParams := keepalive.ClientParameters{
			Time:                10 * time.Second, // 发送ping的间隔
			Timeout:             3 * time.Second,  // ping超时时间
			PermitWithoutStream: true,             // 没有活动流时也发送ping
		}
		opts = append(opts, grpc.WithKeepaliveParams(kaParams))
	}

	// 设置连接超时
	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// 建立连接
	conn, err := grpc.DialContext(ctx, config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接gRPC服务失败 %s: %w", config.Address, err)
	}

	return conn, nil
}

// NewSimpleClient 创建简单的gRPC客户端连接
func NewSimpleClient(address string) (*grpc.ClientConn, error) {
	return NewClient(ClientConfig{
		Address:         address,
		Timeout:         5 * time.Second,
		MaxRecvMsgSize:  4 << 20, // 4MB
		MaxSendMsgSize:  4 << 20, // 4MB
		EnableKeepalive: true,
	})
}
