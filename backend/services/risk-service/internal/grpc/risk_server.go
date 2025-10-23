package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/risk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/risk-service/internal/service"
)

// RiskServer gRPC服务实现
type RiskServer struct {
	pb.UnimplementedRiskServiceServer
	riskService service.RiskService
}

// NewRiskServer 创建gRPC服务实例
func NewRiskServer(riskService service.RiskService) *RiskServer {
	return &RiskServer{
		riskService: riskService,
	}
}

// CheckPayment 风控检查
func (s *RiskServer) CheckPayment(ctx context.Context, req *pb.CheckPaymentRequest) (*pb.CheckPaymentResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	relatedID, _ := uuid.Parse(req.RelatedId)

	input := &service.PaymentCheckInput{
		MerchantID:    merchantID,
		RelatedID:     relatedID,
		RelatedType:   req.RelatedType,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PayerIP:       req.PayerIp,
		PayerEmail:    req.PayerEmail,
		PayerPhone:    req.PayerPhone,
		PaymentMethod: req.PaymentMethod,
	}

	result, err := s.riskService.CheckPayment(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "风控检查失败: %v", err)
	}

	return &pb.CheckPaymentResponse{
		Decision:     result.Decision,
		Score:        int32(result.RiskScore),
		Reasons:      []string{result.Reason}, // 将 Reason 转为数组
		MatchedRules: []string{},               // TODO: 添加匹配的规则列表
		CheckId:      result.ID.String(),
	}, nil
}

// CreateRule 创建规则
func (s *RiskServer) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.RuleResponse, error) {
	rule, err := s.riskService.CreateRule(ctx, &service.CreateRuleInput{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		Priority:    int(req.Priority),
		Description: req.Description,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建规则失败: %v", err)
	}

	return &pb.RuleResponse{
		Rule: &pb.RiskRule{
			Id:          rule.ID.String(),
			RuleName:    rule.RuleName,
			RuleType:    rule.RuleType,
			Priority:    int32(rule.Priority),
			Status:      rule.Status,
			Description: rule.Description,
			CreatedAt:   timestamppb.New(rule.CreatedAt),
			UpdatedAt:   timestamppb.New(rule.UpdatedAt),
		},
	}, nil
}

// GetRule 获取规则
func (s *RiskServer) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.RuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的规则ID")
	}

	rule, err := s.riskService.GetRule(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "规则不存在")
	}

	return &pb.RuleResponse{
		Rule: &pb.RiskRule{
			Id:          rule.ID.String(),
			RuleName:    rule.RuleName,
			RuleType:    rule.RuleType,
			Priority:    int32(rule.Priority),
			Status:      rule.Status,
			Description: rule.Description,
			CreatedAt:   timestamppb.New(rule.CreatedAt),
			UpdatedAt:   timestamppb.New(rule.UpdatedAt),
		},
	}, nil
}

// 其他方法暂时返回未实现
func (s *RiskServer) ReportPaymentResult(ctx context.Context, req *pb.ReportPaymentResultRequest) (*pb.ReportPaymentResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.RuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) EnableRule(ctx context.Context, req *pb.EnableRuleRequest) (*pb.RuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) DisableRule(ctx context.Context, req *pb.DisableRuleRequest) (*pb.RuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) GetCheck(ctx context.Context, req *pb.GetCheckRequest) (*pb.CheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) ListChecks(ctx context.Context, req *pb.ListChecksRequest) (*pb.ListChecksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) AddBlacklist(ctx context.Context, req *pb.AddBlacklistRequest) (*pb.BlacklistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) RemoveBlacklist(ctx context.Context, req *pb.RemoveBlacklistRequest) (*pb.RemoveBlacklistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) CheckBlacklist(ctx context.Context, req *pb.CheckBlacklistRequest) (*pb.CheckBlacklistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *RiskServer) ListBlacklist(ctx context.Context, req *pb.ListBlacklistRequest) (*pb.ListBlacklistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}
