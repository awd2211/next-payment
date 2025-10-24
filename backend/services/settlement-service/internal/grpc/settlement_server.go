package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/settlement"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/service"
)

// SettlementServer implements the SettlementService gRPC service
type SettlementServer struct {
	pb.UnimplementedSettlementServiceServer
	settlementService service.SettlementService
}

// NewSettlementServer creates a new Settlement gRPC server
func NewSettlementServer(settlementService service.SettlementService) *SettlementServer {
	return &SettlementServer{
		settlementService: settlementService,
	}
}

// CreateSettlement implements settlement.SettlementService
func (s *SettlementServer) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.SettlementResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	// Convert items
	items := make([]service.TransactionItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.TransactionItem{
			TransactionID: uuid.New(), // Generate new ID for transaction
			OrderNo:       item.OrderNo,
			PaymentNo:     item.PaymentNo,
			Amount:        item.Amount,
			Fee:           item.Fee,
			TransactionAt: time.Now(),
		})
	}

	input := &service.CreateSettlementInput{
		MerchantID:   merchantID,
		Cycle:        model.SettlementCycleManual,
		StartDate:    req.StartDate.AsTime(),
		EndDate:      req.EndDate.AsTime(),
		Transactions: items,
	}

	settlement, err := s.settlementService.CreateSettlement(ctx, input)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.SettlementResponse{
		Code:    0,
		Message: "Success",
		Data:    convertSettlementToProto(settlement),
	}, nil
}

// GetSettlement implements settlement.SettlementService
func (s *SettlementServer) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.SettlementResponse, error) {
	settlementID, err := uuid.Parse(req.SettlementId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid settlement_id format",
		}, nil
	}

	detail, err := s.settlementService.GetSettlement(ctx, settlementID)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    404,
			Message: err.Error(),
		}, nil
	}

	return &pb.SettlementResponse{
		Code:    0,
		Message: "Success",
		Data:    convertSettlementToProto(detail.Settlement),
	}, nil
}

// ListSettlements implements settlement.SettlementService
func (s *SettlementServer) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	query := &service.ListSettlementQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return &pb.ListSettlementsResponse{
				Code:    400,
				Message: "Invalid merchant_id format",
			}, nil
		}
		query.MerchantID = &merchantID
	}

	if req.Status != "" {
		status := model.SettlementStatus(req.Status)
		query.Status = &status
	}

	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		query.StartDate = &t
	}

	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		query.EndDate = &t
	}

	result, err := s.settlementService.ListSettlements(ctx, query)
	if err != nil {
		return &pb.ListSettlementsResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	settlements := make([]*pb.SettlementData, 0, len(result.Settlements))
	for _, settlement := range result.Settlements {
		settlements = append(settlements, convertSettlementToProto(settlement))
	}

	return &pb.ListSettlementsResponse{
		Code:        0,
		Message:     "Success",
		Settlements: settlements,
		Total:       result.Total,
	}, nil
}

// ApproveSettlement implements settlement.SettlementService
func (s *SettlementServer) ApproveSettlement(ctx context.Context, req *pb.ApproveSettlementRequest) (*pb.SettlementResponse, error) {
	settlementID, err := uuid.Parse(req.SettlementId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid settlement_id format",
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid approver_id format",
		}, nil
	}

	err = s.settlementService.ApproveSettlement(ctx, settlementID, approverID, "System", req.Comments)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated settlement
	detail, err := s.settlementService.GetSettlement(ctx, settlementID)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.SettlementResponse{
		Code:    0,
		Message: "Success",
		Data:    convertSettlementToProto(detail.Settlement),
	}, nil
}

// RejectSettlement implements settlement.SettlementService
func (s *SettlementServer) RejectSettlement(ctx context.Context, req *pb.RejectSettlementRequest) (*pb.SettlementResponse, error) {
	settlementID, err := uuid.Parse(req.SettlementId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid settlement_id format",
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid approver_id format",
		}, nil
	}

	err = s.settlementService.RejectSettlement(ctx, settlementID, approverID, "System", req.RejectReason)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated settlement
	detail, err := s.settlementService.GetSettlement(ctx, settlementID)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.SettlementResponse{
		Code:    0,
		Message: "Success",
		Data:    convertSettlementToProto(detail.Settlement),
	}, nil
}

// ConfirmSettlement implements settlement.SettlementService
func (s *SettlementServer) ConfirmSettlement(ctx context.Context, req *pb.ConfirmSettlementRequest) (*pb.SettlementResponse, error) {
	settlementID, err := uuid.Parse(req.SettlementId)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    400,
			Message: "Invalid settlement_id format",
		}, nil
	}

	err = s.settlementService.ExecuteSettlement(ctx, settlementID)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated settlement
	detail, err := s.settlementService.GetSettlement(ctx, settlementID)
	if err != nil {
		return &pb.SettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.SettlementResponse{
		Code:    0,
		Message: "Success",
		Data:    convertSettlementToProto(detail.Settlement),
	}, nil
}

// CalculateSettlement implements settlement.SettlementService
func (s *SettlementServer) CalculateSettlement(ctx context.Context, req *pb.CalculateSettlementRequest) (*pb.CalculateSettlementResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.CalculateSettlementResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	report, err := s.settlementService.GetSettlementReport(ctx, merchantID, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return &pb.CalculateSettlementResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.CalculateSettlementResponse{
		Code:    0,
		Message: "Success",
		Data: &pb.CalculateSettlementData{
			MerchantId:       req.MerchantId,
			TotalAmount:      report.TotalAmount,
			TotalFee:         report.TotalFee,
			NetAmount:        report.TotalSettlement,
			Currency:         "CNY",
			TransactionCount: int32(report.TotalCount),
			StartDate:        req.StartDate,
			EndDate:          req.EndDate,
		},
	}, nil
}

// GetSettlementStats implements settlement.SettlementService
func (s *SettlementServer) GetSettlementStats(ctx context.Context, req *pb.GetSettlementStatsRequest) (*pb.SettlementStatsResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.SettlementStatsResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	report, err := s.settlementService.GetSettlementReport(ctx, merchantID, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return &pb.SettlementStatsResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.SettlementStatsResponse{
		Code:    0,
		Message: "Success",
		Data: &pb.SettlementStatsData{
			MerchantId:              req.MerchantId,
			TotalSettledAmount:      report.TotalSettlement,
			TotalSettledFee:         report.TotalFee,
			PendingSettlementAmount: 0, // TODO: Calculate pending amount
			SettlementCount:         int32(report.CompletedCount),
			PendingCount:            int32(report.PendingCount),
			ByStatus: []*pb.SettlementByStatus{
				{Status: "completed", Count: int32(report.CompletedCount), Amount: report.TotalSettlement},
				{Status: "pending", Count: int32(report.PendingCount), Amount: 0},
				{Status: "rejected", Count: int32(report.RejectedCount), Amount: 0},
			},
		},
	}, nil
}

// Helper function to convert Settlement model to proto
func convertSettlementToProto(settlement *model.Settlement) *pb.SettlementData {
	data := &pb.SettlementData{
		Id:           settlement.ID.String(),
		SettlementNo: settlement.SettlementNo,
		MerchantId:   settlement.MerchantID.String(),
		TotalAmount:  settlement.TotalAmount,
		TotalFee:     settlement.FeeAmount,
		NetAmount:    settlement.SettlementAmount,
		Currency:     "CNY",
		Status:       string(settlement.Status),
		StartDate:    timestamppb.New(settlement.StartDate),
		EndDate:      timestamppb.New(settlement.EndDate),
		ItemCount:    int32(settlement.TotalCount),
		CreatedAt:    timestamppb.New(settlement.CreatedAt),
		UpdatedAt:    timestamppb.New(settlement.UpdatedAt),
	}

	if settlement.ApprovedAt != nil {
		data.SettlementDate = timestamppb.New(*settlement.ApprovedAt)
	}

	if settlement.WithdrawalNo != "" {
		data.BankAccount = settlement.WithdrawalNo
	}

	if settlement.ErrorMessage != "" {
		data.RejectReason = settlement.ErrorMessage
	}

	return data
}
