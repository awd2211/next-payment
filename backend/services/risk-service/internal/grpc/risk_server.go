package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/risk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/risk-service/internal/repository"
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

// ReportPaymentResult 报告支付结果 (暂未实现反馈机制)
func (s *RiskServer) ReportPaymentResult(ctx context.Context, req *pb.ReportPaymentResultRequest) (*pb.ReportPaymentResultResponse, error) {
	// TODO: 实现支付结果反馈机制，用于优化风控模型
	return &pb.ReportPaymentResultResponse{
		Success: true,
	}, nil
}

// ListRules 规则列表
func (s *RiskServer) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	query := &repository.RuleQuery{
		RuleType: req.RuleType,
		Status:   req.Status,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	rules, total, err := s.riskService.ListRules(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询规则列表失败: %v", err)
	}

	pbRules := make([]*pb.RiskRule, len(rules))
	for i, rule := range rules {
		pbRules[i] = &pb.RiskRule{
			Id:          rule.ID.String(),
			RuleName:    rule.RuleName,
			RuleType:    rule.RuleType,
			Priority:    int32(rule.Priority),
			Status:      rule.Status,
			Description: rule.Description,
			CreatedAt:   timestamppb.New(rule.CreatedAt),
			UpdatedAt:   timestamppb.New(rule.UpdatedAt),
		}
	}

	return &pb.ListRulesResponse{
		Rules: pbRules,
		Total: total,
	}, nil
}

// UpdateRule 更新规则
func (s *RiskServer) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.RuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的规则ID")
	}

	input := &service.UpdateRuleInput{
		RuleName:    req.RuleName,
		Priority:    int(req.Priority),
		Description: req.Description,
		// Conditions 和 Actions 在proto中未定义，需要proto扩展
	}

	rule, err := s.riskService.UpdateRule(ctx, id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新规则失败: %v", err)
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

// DeleteRule 删除规则
func (s *RiskServer) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的规则ID")
	}

	if err := s.riskService.DeleteRule(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "删除规则失败: %v", err)
	}

	return &pb.DeleteRuleResponse{
		Success: true,
	}, nil
}

// EnableRule 启用规则
func (s *RiskServer) EnableRule(ctx context.Context, req *pb.EnableRuleRequest) (*pb.RuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的规则ID")
	}

	if err := s.riskService.EnableRule(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "启用规则失败: %v", err)
	}

	rule, err := s.riskService.GetRule(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取规则失败: %v", err)
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

// DisableRule 禁用规则
func (s *RiskServer) DisableRule(ctx context.Context, req *pb.DisableRuleRequest) (*pb.RuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的规则ID")
	}

	if err := s.riskService.DisableRule(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "禁用规则失败: %v", err)
	}

	rule, err := s.riskService.GetRule(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取规则失败: %v", err)
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

// GetCheck 获取检查记录
func (s *RiskServer) GetCheck(ctx context.Context, req *pb.GetCheckRequest) (*pb.CheckResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的检查ID")
	}

	check, err := s.riskService.GetCheck(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "检查记录不存在: %v", err)
	}

	return &pb.CheckResponse{
		Check: &pb.RiskCheck{
			Id:          check.ID.String(),
			MerchantId:  check.MerchantID.String(),
			RelatedId:   check.RelatedID.String(),
			RelatedType: check.RelatedType,
			Amount:      0, // proto中有，但model中没有直接存储
			Currency:    "", // proto中有，但model中没有直接存储
			Decision:    check.Decision,
			Score:       int32(check.RiskScore),
			Reasons:     []string{check.Reason},
			Result:      check.Decision, // 使用decision作为result
			CreatedAt:   timestamppb.New(check.CreatedAt),
		},
	}, nil
}

// ListChecks 检查记录列表
func (s *RiskServer) ListChecks(ctx context.Context, req *pb.ListChecksRequest) (*pb.ListChecksResponse, error) {
	query := &repository.CheckQuery{
		Decision: req.Decision,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
		}
		query.MerchantID = &merchantID
	}

	if req.StartTime != nil {
		startTime := req.StartTime.AsTime()
		query.StartTime = &startTime
	}

	if req.EndTime != nil {
		endTime := req.EndTime.AsTime()
		query.EndTime = &endTime
	}

	checks, total, err := s.riskService.ListChecks(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询检查记录失败: %v", err)
	}

	pbChecks := make([]*pb.RiskCheck, len(checks))
	for i, check := range checks {
		pbChecks[i] = &pb.RiskCheck{
			Id:          check.ID.String(),
			MerchantId:  check.MerchantID.String(),
			RelatedId:   check.RelatedID.String(),
			RelatedType: check.RelatedType,
			Amount:      0,
			Currency:    "",
			Decision:    check.Decision,
			Score:       int32(check.RiskScore),
			Reasons:     []string{check.Reason},
			Result:      check.Decision,
			CreatedAt:   timestamppb.New(check.CreatedAt),
		}
	}

	return &pb.ListChecksResponse{
		Checks: pbChecks,
		Total:  total,
	}, nil
}

// AddBlacklist 添加黑名单
func (s *RiskServer) AddBlacklist(ctx context.Context, req *pb.AddBlacklistRequest) (*pb.BlacklistResponse, error) {
	input := &service.AddBlacklistInput{
		EntityType:  req.EntityType,
		EntityValue: req.EntityValue,
		Reason:      req.Reason,
		AddedBy:     req.AddedBy,
	}

	if req.ExpireAt != nil {
		expireAt := req.ExpireAt.AsTime()
		input.ExpireAt = &expireAt
	}

	blacklist, err := s.riskService.AddBlacklist(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "添加黑名单失败: %v", err)
	}

	var expireAt *timestamppb.Timestamp
	if blacklist.ExpireAt != nil {
		expireAt = timestamppb.New(*blacklist.ExpireAt)
	}

	return &pb.BlacklistResponse{
		Blacklist: &pb.Blacklist{
			Id:          blacklist.ID.String(),
			EntityType:  blacklist.EntityType,
			EntityValue: blacklist.EntityValue,
			Reason:      blacklist.Reason,
			AddedBy:     blacklist.AddedBy,
			ExpireAt:    expireAt,
			CreatedAt:   timestamppb.New(blacklist.CreatedAt),
		},
	}, nil
}

// RemoveBlacklist 移除黑名单
func (s *RiskServer) RemoveBlacklist(ctx context.Context, req *pb.RemoveBlacklistRequest) (*pb.RemoveBlacklistResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的黑名单ID")
	}

	if err := s.riskService.RemoveBlacklist(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "移除黑名单失败: %v", err)
	}

	return &pb.RemoveBlacklistResponse{
		Success: true,
	}, nil
}

// CheckBlacklist 检查黑名单
func (s *RiskServer) CheckBlacklist(ctx context.Context, req *pb.CheckBlacklistRequest) (*pb.CheckBlacklistResponse, error) {
	isBlacklisted, blacklist, err := s.riskService.CheckBlacklist(ctx, req.EntityType, req.EntityValue)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "检查黑名单失败: %v", err)
	}

	response := &pb.CheckBlacklistResponse{
		IsBlacklisted: isBlacklisted,
	}

	if isBlacklisted && blacklist != nil {
		var expireAt *timestamppb.Timestamp
		if blacklist.ExpireAt != nil {
			expireAt = timestamppb.New(*blacklist.ExpireAt)
		}

		response.Blacklist = &pb.Blacklist{
			Id:          blacklist.ID.String(),
			EntityType:  blacklist.EntityType,
			EntityValue: blacklist.EntityValue,
			Reason:      blacklist.Reason,
			AddedBy:     blacklist.AddedBy,
			ExpireAt:    expireAt,
			CreatedAt:   timestamppb.New(blacklist.CreatedAt),
		}
	}

	return response, nil
}

// ListBlacklist 黑名单列表
func (s *RiskServer) ListBlacklist(ctx context.Context, req *pb.ListBlacklistRequest) (*pb.ListBlacklistResponse, error) {
	query := &repository.BlacklistQuery{
		EntityType: req.EntityType,
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
	}

	blacklists, total, err := s.riskService.ListBlacklist(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询黑名单失败: %v", err)
	}

	pbBlacklists := make([]*pb.Blacklist, len(blacklists))
	for i, bl := range blacklists {
		var expireAt *timestamppb.Timestamp
		if bl.ExpireAt != nil {
			expireAt = timestamppb.New(*bl.ExpireAt)
		}

		pbBlacklists[i] = &pb.Blacklist{
			Id:          bl.ID.String(),
			EntityType:  bl.EntityType,
			EntityValue: bl.EntityValue,
			Reason:      bl.Reason,
			AddedBy:     bl.AddedBy,
			ExpireAt:    expireAt,
			CreatedAt:   timestamppb.New(bl.CreatedAt),
		}
	}

	return &pb.ListBlacklistResponse{
		Blacklists: pbBlacklists,
		Total:      total,
	}, nil
}
