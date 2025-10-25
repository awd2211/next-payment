package adapter

import (
	"context"
	"fmt"
)

// DefaultPreAuthNotSupported 预授权默认实现（不支持）
// 其他 adapter 可以嵌入这个结构体来提供默认的"不支持"实现

type DefaultPreAuthNotSupported struct{}

func (d *DefaultPreAuthNotSupported) CreatePreAuth(ctx context.Context, req *CreatePreAuthRequest) (*CreatePreAuthResponse, error) {
	return nil, fmt.Errorf("当前支付渠道不支持预授权功能")
}

func (d *DefaultPreAuthNotSupported) CapturePreAuth(ctx context.Context, req *CapturePreAuthRequest) (*CapturePreAuthResponse, error) {
	return nil, fmt.Errorf("当前支付渠道不支持预授权功能")
}

func (d *DefaultPreAuthNotSupported) CancelPreAuth(ctx context.Context, req *CancelPreAuthRequest) (*CancelPreAuthResponse, error) {
	return nil, fmt.Errorf("当前支付渠道不支持预授权功能")
}

func (d *DefaultPreAuthNotSupported) QueryPreAuth(ctx context.Context, channelPreAuthNo string) (*QueryPreAuthResponse, error) {
	return nil, fmt.Errorf("当前支付渠道不支持预授权功能")
}
