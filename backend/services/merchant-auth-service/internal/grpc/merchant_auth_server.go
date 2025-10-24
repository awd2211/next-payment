package grpc

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/merchant_auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
// Note: Login functionality should be implemented in a separate auth service with merchant client
// This is a placeholder that returns Unimplemented as the actual login logic
// requires password verification and JWT token generation which should be handled
// by a dedicated authentication handler, not the security service
func (s *MerchantAuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented - should use HTTP handler")
}

// RefreshToken implements merchant_auth.MerchantAuthService
// Note: Token refresh should be handled by JWT manager in HTTP handler
func (s *MerchantAuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshToken not implemented - should use HTTP handler")
}

// Logout implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.StatusResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Revoke session
	if req.SessionId != "" {
		if err := s.securityService.RevokeSession(ctx, req.SessionId); err != nil {
			return &pb.StatusResponse{
				Code:    int32(codes.Internal),
				Message: "Failed to revoke session: " + err.Error(),
			}, nil
		}
	} else {
		// Revoke all sessions
		if err := s.securityService.RevokeAllSessions(ctx, merchantID); err != nil {
			return &pb.StatusResponse{
				Code:    int32(codes.Internal),
				Message: "Failed to revoke all sessions: " + err.Error(),
			}, nil
		}
	}

	return &pb.StatusResponse{
		Code:    0,
		Message: "Logout successful",
	}, nil
}

// EnableTwoFactor implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) EnableTwoFactor(ctx context.Context, req *pb.EnableTwoFactorRequest) (*pb.TwoFactorResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.TwoFactorResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Enable 2FA
	result, err := s.securityService.Enable2FA(ctx, merchantID)
	if err != nil {
		return &pb.TwoFactorResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to enable 2FA: " + err.Error(),
		}, nil
	}

	return &pb.TwoFactorResponse{
		Code:    0,
		Message: "2FA enabled successfully",
		Data: &pb.TwoFactorData{
			Secret:      result.Secret,
			QrCode:      result.QRCode,
			BackupCodes: result.BackupCodes,
		},
	}, nil
}

// VerifyTwoFactor implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) VerifyTwoFactor(ctx context.Context, req *pb.VerifyTwoFactorRequest) (*pb.StatusResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Verify 2FA code
	result, err := s.securityService.Verify2FA(ctx, merchantID, req.Code)
	if err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to verify 2FA: " + err.Error(),
		}, nil
	}

	if !result.Success {
		return &pb.StatusResponse{
			Code:    int32(codes.PermissionDenied),
			Message: "Invalid 2FA code",
		}, nil
	}

	return &pb.StatusResponse{
		Code:    0,
		Message: "2FA verified successfully",
	}, nil
}

// DisableTwoFactor implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) DisableTwoFactor(ctx context.Context, req *pb.DisableTwoFactorRequest) (*pb.StatusResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Disable 2FA
	if err := s.securityService.Disable2FA(ctx, merchantID, req.Password); err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to disable 2FA: " + err.Error(),
		}, nil
	}

	return &pb.StatusResponse{
		Code:    0,
		Message: "2FA disabled successfully",
	}, nil
}

// GetLoginActivities implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) GetLoginActivities(ctx context.Context, req *pb.GetLoginActivitiesRequest) (*pb.LoginActivitiesResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.LoginActivitiesResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Set defaults
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get login activities
	activities, total, err := s.securityService.GetLoginActivities(ctx, merchantID, page, pageSize)
	if err != nil {
		return &pb.LoginActivitiesResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to get login activities: " + err.Error(),
		}, nil
	}

	// Convert to proto format
	pbActivities := make([]*pb.LoginActivityData, len(activities))
	for i, activity := range activities {
		pbActivities[i] = &pb.LoginActivityData{
			Id:            activity.ID.String(),
			MerchantId:    activity.MerchantID.String(),
			IpAddress:     activity.IP,
			UserAgent:     activity.UserAgent,
			Location:      activity.Location,
			Status:        activity.Status,
			FailureReason: activity.FailedReason,
			CreatedAt:     timestamppb.New(activity.LoginAt),
		}
	}

	return &pb.LoginActivitiesResponse{
		Code:       0,
		Message:    "Success",
		Activities: pbActivities,
		Total:      total,
	}, nil
}

// ListSessions implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.SessionsResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.SessionsResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Get active sessions
	sessions, err := s.securityService.GetActiveSessions(ctx, merchantID)
	if err != nil {
		return &pb.SessionsResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to get sessions: " + err.Error(),
		}, nil
	}

	// Convert to proto format
	pbSessions := make([]*pb.SessionData, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = &pb.SessionData{
			Id:             session.ID.String(),
			MerchantId:     session.MerchantID.String(),
			IpAddress:      session.IP,
			UserAgent:      session.UserAgent,
			DeviceType:     "", // Not stored in model
			IsCurrent:      false, // Would need current session ID to determine
			CreatedAt:      timestamppb.New(session.CreatedAt),
			ExpiresAt:      timestamppb.New(session.ExpiresAt),
			LastActivityAt: timestamppb.New(session.LastSeenAt),
		}
	}

	return &pb.SessionsResponse{
		Code:     0,
		Message:  "Success",
		Sessions: pbSessions,
	}, nil
}

// RevokeSession implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.StatusResponse, error) {
	// Revoke session
	if err := s.securityService.RevokeSession(ctx, req.SessionId); err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to revoke session: " + err.Error(),
		}, nil
	}

	return &pb.StatusResponse{
		Code:    0,
		Message: "Session revoked successfully",
	}, nil
}

// UpdateSecuritySettings implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) UpdateSecuritySettings(ctx context.Context, req *pb.UpdateSecuritySettingsRequest) (*pb.SecuritySettingsResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.SecuritySettingsResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Build update input
	input := &service.UpdateSecuritySettingsInput{}

	// Update session timeout if provided
	if req.SessionTimeoutMinutes > 0 {
		timeout := int(req.SessionTimeoutMinutes)
		input.SessionTimeoutMinutes = &timeout
	}

	// Update notification settings
	if req.LoginAlertsEnabled {
		loginNotif := req.LoginAlertsEnabled
		input.LoginNotification = &loginNotif
	}
	if req.SuspiciousActivityAlertsEnabled {
		abnormalNotif := req.SuspiciousActivityAlertsEnabled
		input.AbnormalNotification = &abnormalNotif
	}

	// Update security settings
	settings, err := s.securityService.UpdateSecuritySettings(ctx, merchantID, input)
	if err != nil {
		return &pb.SecuritySettingsResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to update security settings: " + err.Error(),
		}, nil
	}

	return &pb.SecuritySettingsResponse{
		Code:    0,
		Message: "Security settings updated successfully",
		Data: &pb.SecuritySettingsData{
			MerchantId:                       settings.MerchantID.String(),
			TwoFactorEnabled:                 false, // Would need to query TwoFactorAuth table
			TwoFactorMethod:                  "TOTP",
			LoginAlertsEnabled:               settings.LoginNotification,
			SuspiciousActivityAlertsEnabled:  settings.AbnormalNotification,
			SessionTimeoutMinutes:            int32(settings.SessionTimeoutMinutes),
			UpdatedAt:                        timestamppb.New(settings.UpdatedAt),
		},
	}, nil
}

// ChangePassword implements merchant_auth.MerchantAuthService
func (s *MerchantAuthServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.StatusResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.InvalidArgument),
			Message: "Invalid merchant ID format",
		}, nil
	}

	// Change password
	if err := s.securityService.ChangePassword(ctx, merchantID, req.OldPassword, req.NewPassword); err != nil {
		return &pb.StatusResponse{
			Code:    int32(codes.Internal),
			Message: "Failed to change password: " + err.Error(),
		}, nil
	}

	return &pb.StatusResponse{
		Code:    0,
		Message: "Password changed successfully",
	}, nil
}

// Helper function to parse JSON string arrays (for security settings)
func parseJSONStringArray(jsonStr string) []string {
	var arr []string
	if jsonStr == "" || jsonStr == "[]" {
		return arr
	}
	_ = json.Unmarshal([]byte(jsonStr), &arr)
	return arr
}
