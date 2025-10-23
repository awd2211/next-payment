package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// ServerConfig gRPC服务端配置
type ServerConfig struct {
	Port            int           // 监听端口
	MaxRecvMsgSize  int           // 最大接收消息大小
	MaxSendMsgSize  int           // 最大发送消息大小
	EnableKeepalive bool          // 启用Keepalive
	MaxConnAge      time.Duration // 连接最大存活时间
}

// NewServer 创建gRPC服务器
func NewServer(config ServerConfig) *grpc.Server {
	var opts []grpc.ServerOption

	// 设置最大消息大小
	if config.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(config.MaxRecvMsgSize))
	}
	if config.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.MaxSendMsgSize(config.MaxSendMsgSize))
	}

	// 启用Keepalive
	if config.EnableKeepalive {
		kaParams := keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute, // 连接空闲时间
			MaxConnectionAge:      config.MaxConnAge,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}
		kaPolicy := keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}
		opts = append(opts,
			grpc.KeepaliveParams(kaParams),
			grpc.KeepaliveEnforcementPolicy(kaPolicy),
		)
	}

	return grpc.NewServer(opts...)
}

// NewSimpleServer 创建简单的gRPC服务器
func NewSimpleServer() *grpc.Server {
	return NewServer(ServerConfig{
		MaxRecvMsgSize:  4 << 20, // 4MB
		MaxSendMsgSize:  4 << 20, // 4MB
		EnableKeepalive: true,
		MaxConnAge:      30 * time.Minute,
	})
}

// StartServer 启动gRPC服务器
func StartServer(server *grpc.Server, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("监听端口失败 %d: %w", port, err)
	}

	fmt.Printf("gRPC服务器启动在端口 %d\n", port)
	return server.Serve(lis)
}
