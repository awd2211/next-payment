package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/merchant_auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment-platform/merchant-auth-service/internal/service"
)

// MerchantAuthServer implements the MerchantAuthService gRPC service
type MerchantAuthServer struct {
	pb.UnimplementedMerchantAuthServiceServer
	securityService service.SecurityService
}

// NewMerchantAuthServer creates a new MerchantAuth gRPC server
func NewMerchantAuthServer(securityService service.SecurityService) *MerchantAuthServer {
	return &MerchantAuthServer{
		securityService: securityService,
	}
}

// Login implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

// RefreshToken implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshToken not implemented")
}

// Logout implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}

// EnableTwoFactor implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) EnableTwoFactor(ctx context.Context, req *pb.EnableTwoFactorRequest) (*pb.TwoFactorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnableTwoFactor not implemented")
}

// VerifyTwoFactor implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) VerifyTwoFactor(ctx context.Context, req *pb.VerifyTwoFactorRequest) (*pb.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyTwoFactor not implemented")
}

// DisableTwoFactor implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) DisableTwoFactor(ctx context.Context, req *pb.DisableTwoFactorRequest) (*pb.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisableTwoFactor not implemented")
}

// GetLoginActivities implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) GetLoginActivities(ctx context.Context, req *pb.GetLoginActivitiesRequest) (*pb.LoginActivitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLoginActivities not implemented")
}

// ListSessions implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.SessionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSessions not implemented")
}

// RevokeSession implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevokeSession not implemented")
}

// UpdateSecuritySettings implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) UpdateSecuritySettings(ctx context.Context, req *pb.UpdateSecuritySettingsRequest) (*pb.SecuritySettingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSecuritySettings not implemented")
}

// ChangePassword implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
