package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor 日志拦截器（一元调用）
func LoggingInterceptor(logger interface{ Info(msg string, fields ...interface{}) }) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// 处理请求
		resp, err := handler(ctx, req)

		// 记录日志
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		logger.Info("gRPC请求",
			"method", info.FullMethod,
			"duration", duration.Milliseconds(),
			"code", code.String(),
		)

		return resp, err
	}
}

// RecoveryInterceptor 恢复拦截器（防止panic导致服务崩溃）
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Errorf(codes.Internal, "panic: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}

// AuthInterceptor 认证拦截器
func AuthInterceptor(authFunc func(ctx context.Context, token string) error) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从metadata中获取token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "缺少认证信息")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "缺少认证token")
		}

		// 验证token
		if err := authFunc(ctx, tokens[0]); err != nil {
			return nil, status.Error(codes.Unauthenticated, "认证失败")
		}

		return handler(ctx, req)
	}
}

// TimeoutInterceptor 超时拦截器
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}

// RateLimitInterceptor 限流拦截器
func RateLimitInterceptor(limiter interface{ Allow() bool }) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "请求过于频繁")
		}

		return handler(ctx, req)
	}
}

// ChainUnaryInterceptors 链式组合多个拦截器
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 构建拦截器链
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			currentInterceptor := interceptors[i]
			nextHandler := chain
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, info, nextHandler)
			}
		}

		return chain(ctx, req)
	}
}
