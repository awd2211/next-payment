package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"

	"payment-platform/dispute-service/internal/client"
	"payment-platform/dispute-service/internal/model"
	"payment-platform/dispute-service/internal/repository"
)

// DisputeService 拒付服务接口
type DisputeService interface {
	// Dispute management
	CreateDispute(ctx context.Context, input *CreateDisputeInput) (*model.Dispute, error)
	GetDisputeByID(ctx context.Context, id uuid.UUID) (*DisputeDetails, error)
	ListDisputes(ctx context.Context, filters *DisputeFilters, page, pageSize int) (*DisputeListResult, error)
	UpdateDisputeStatus(ctx context.Context, id uuid.UUID, status string) error
	AssignDispute(ctx context.Context, id, assignedTo uuid.UUID) error

	// Evidence management
	UploadEvidence(ctx context.Context, input *UploadEvidenceInput) (*model.DisputeEvidence, error)
	ListEvidence(ctx context.Context, disputeID uuid.UUID) ([]*model.DisputeEvidence, error)
	DeleteEvidence(ctx context.Context, evidenceID uuid.UUID) error

	// Stripe integration
	SubmitToStripe(ctx context.Context, disputeID uuid.UUID) error
	SyncFromStripe(ctx context.Context, channelDisputeID string) (*model.Dispute, error)

	// Statistics
	GetStatistics(ctx context.Context, merchantID *uuid.UUID, startDate, endDate *time.Time) (*repository.DisputeStatistics, error)
}

// Input/Output DTOs

type CreateDisputeInput struct {
	Channel          string    `json:"channel" binding:"required"`
	ChannelDisputeID string    `json:"channel_dispute_id"`
	PaymentNo        string    `json:"payment_no" binding:"required"`
	OrderNo          string    `json:"order_no"`
	MerchantID       uuid.UUID `json:"merchant_id" binding:"required"`
	ChannelTradeNo   string    `json:"channel_trade_no"`
	Amount           int64     `json:"amount" binding:"required"`
	Currency         string    `json:"currency" binding:"required"`
	Reason           string    `json:"reason"`
	ReasonCode       string    `json:"reason_code"`
	EvidenceDueBy    *time.Time `json:"evidence_due_by"`
}

type UploadEvidenceInput struct {
	DisputeID    uuid.UUID `json:"dispute_id" binding:"required"`
	EvidenceType string    `json:"evidence_type" binding:"required"`
	Title        string    `json:"title" binding:"required"`
	Description  string    `json:"description"`
	FileURL      string    `json:"file_url"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	FileHash     string    `json:"file_hash"`
	UploadedBy   uuid.UUID `json:"uploaded_by" binding:"required"`
}

type DisputeFilters struct {
	MerchantID        *uuid.UUID
	Channel           string
	Status            string
	Reason            string
	AssignedTo        *uuid.UUID
	EvidenceSubmitted *bool
	StartDate         *time.Time
	EndDate           *time.Time
	PaymentNo         string
}

type DisputeDetails struct {
	Dispute  *model.Dispute            `json:"dispute"`
	Evidence []*model.DisputeEvidence  `json:"evidence"`
	Timeline []*model.DisputeTimeline  `json:"timeline"`
}

type DisputeListResult struct {
	Disputes   []*model.Dispute `json:"disputes"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// disputeService 拒付服务实现
type disputeService struct {
	repo          repository.DisputeRepository
	stripeClient  *client.StripeDisputeClient
	paymentClient client.PaymentClient
}

// NewDisputeService 创建拒付服务实例
func NewDisputeService(
	repo repository.DisputeRepository,
	stripeClient *client.StripeDisputeClient,
	paymentClient client.PaymentClient,
) DisputeService {
	return &disputeService{
		repo:          repo,
		stripeClient:  stripeClient,
		paymentClient: paymentClient,
	}
}

// CreateDispute 创建拒付记录
func (s *disputeService) CreateDispute(ctx context.Context, input *CreateDisputeInput) (*model.Dispute, error) {
	// Check if dispute already exists
	if input.ChannelDisputeID != "" {
		existing, err := s.repo.GetDisputeByChannelID(ctx, input.ChannelDisputeID)
		if err != nil {
			return nil, fmt.Errorf("check existing dispute failed: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("dispute already exists with channel_dispute_id: %s", input.ChannelDisputeID)
		}
	}

	// Generate dispute number
	disputeNo := generateDisputeNo(input.Channel)

	dispute := &model.Dispute{
		DisputeNo:        disputeNo,
		Channel:          input.Channel,
		ChannelDisputeID: input.ChannelDisputeID,
		PaymentNo:        input.PaymentNo,
		OrderNo:          input.OrderNo,
		MerchantID:       input.MerchantID,
		ChannelTradeNo:   input.ChannelTradeNo,
		Amount:           input.Amount,
		Currency:         input.Currency,
		Reason:           input.Reason,
		ReasonCode:       input.ReasonCode,
		Status:           model.DisputeStatusNeedsResponse,
		EvidenceDueBy:    input.EvidenceDueBy,
		EvidenceSubmitted: false,
	}

	if err := s.repo.CreateDispute(ctx, dispute); err != nil {
		return nil, fmt.Errorf("create dispute failed: %w", err)
	}

	// Create timeline event
	timeline := &model.DisputeTimeline{
		DisputeID:    dispute.ID,
		DisputeNo:    dispute.DisputeNo,
		EventType:    model.TimelineEventCreated,
		EventStatus:  dispute.Status,
		Description:  fmt.Sprintf("Dispute created for payment %s", input.PaymentNo),
		OperatorType: model.OperatorTypeSystem,
	}
	s.repo.CreateTimelineEvent(ctx, timeline)

	return dispute, nil
}

// GetDisputeByID 获取拒付详情
func (s *disputeService) GetDisputeByID(ctx context.Context, id uuid.UUID) (*DisputeDetails, error) {
	dispute, err := s.repo.GetDisputeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get dispute failed: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found")
	}

	// Get evidence
	evidence, err := s.repo.ListEvidenceByDispute(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get evidence failed: %w", err)
	}

	// Get timeline
	timeline, err := s.repo.ListTimelineByDispute(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get timeline failed: %w", err)
	}

	return &DisputeDetails{
		Dispute:  dispute,
		Evidence: evidence,
		Timeline: timeline,
	}, nil
}

// ListDisputes 查询拒付列表
func (s *disputeService) ListDisputes(ctx context.Context, filters *DisputeFilters, page, pageSize int) (*DisputeListResult, error) {
	repoFilters := repository.DisputeFilters{
		MerchantID:        filters.MerchantID,
		Channel:           filters.Channel,
		Status:            filters.Status,
		Reason:            filters.Reason,
		AssignedTo:        filters.AssignedTo,
		EvidenceSubmitted: filters.EvidenceSubmitted,
		StartDate:         filters.StartDate,
		EndDate:           filters.EndDate,
		PaymentNo:         filters.PaymentNo,
	}

	disputes, total, err := s.repo.ListDisputes(ctx, repoFilters, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list disputes failed: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &DisputeListResult{
		Disputes:   disputes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateDisputeStatus 更新拒付状态
func (s *disputeService) UpdateDisputeStatus(ctx context.Context, id uuid.UUID, status string) error {
	dispute, err := s.repo.GetDisputeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get dispute failed: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found")
	}

	oldStatus := dispute.Status
	dispute.Status = status

	// Update result based on status
	if status == model.DisputeStatusWon {
		dispute.Result = model.DisputeResultWon
		now := time.Now()
		dispute.ResolvedAt = &now
	} else if status == model.DisputeStatusLost {
		dispute.Result = model.DisputeResultLost
		now := time.Now()
		dispute.ResolvedAt = &now
	}

	if err := s.repo.UpdateDispute(ctx, dispute); err != nil {
		return fmt.Errorf("update dispute status failed: %w", err)
	}

	// Create timeline event
	timeline := &model.DisputeTimeline{
		DisputeID:    dispute.ID,
		DisputeNo:    dispute.DisputeNo,
		EventType:    model.TimelineEventUpdated,
		EventStatus:  status,
		Description:  fmt.Sprintf("Status changed from %s to %s", oldStatus, status),
		OperatorType: model.OperatorTypeAdmin,
	}
	s.repo.CreateTimelineEvent(ctx, timeline)

	return nil
}

// AssignDispute 分配拒付给处理人员
func (s *disputeService) AssignDispute(ctx context.Context, id, assignedTo uuid.UUID) error {
	dispute, err := s.repo.GetDisputeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get dispute failed: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found")
	}

	now := time.Now()
	dispute.AssignedTo = &assignedTo
	dispute.AssignedAt = &now

	if err := s.repo.UpdateDispute(ctx, dispute); err != nil {
		return fmt.Errorf("assign dispute failed: %w", err)
	}

	// Create timeline event
	timeline := &model.DisputeTimeline{
		DisputeID:    dispute.ID,
		DisputeNo:    dispute.DisputeNo,
		EventType:    model.TimelineEventAssigned,
		EventStatus:  dispute.Status,
		Description:  fmt.Sprintf("Dispute assigned to staff member"),
		OperatorID:   &assignedTo,
		OperatorType: model.OperatorTypeAdmin,
	}
	s.repo.CreateTimelineEvent(ctx, timeline)

	return nil
}

// UploadEvidence 上传证据
func (s *disputeService) UploadEvidence(ctx context.Context, input *UploadEvidenceInput) (*model.DisputeEvidence, error) {
	// Check if dispute exists
	dispute, err := s.repo.GetDisputeByID(ctx, input.DisputeID)
	if err != nil {
		return nil, fmt.Errorf("get dispute failed: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found")
	}

	// Check if evidence can still be submitted
	if dispute.EvidenceDueBy != nil && time.Now().After(*dispute.EvidenceDueBy) {
		return nil, fmt.Errorf("evidence submission deadline has passed")
	}

	evidence := &model.DisputeEvidence{
		DisputeID:    input.DisputeID,
		DisputeNo:    dispute.DisputeNo,
		EvidenceType: input.EvidenceType,
		Title:        input.Title,
		Description:  input.Description,
		FileURL:      input.FileURL,
		FileName:     input.FileName,
		FileSize:     input.FileSize,
		FileHash:     input.FileHash,
		UploadedBy:   input.UploadedBy,
		IsSubmitted:  false,
	}

	if err := s.repo.CreateEvidence(ctx, evidence); err != nil {
		return nil, fmt.Errorf("create evidence failed: %w", err)
	}

	// Create timeline event
	timeline := &model.DisputeTimeline{
		DisputeID:    dispute.ID,
		DisputeNo:    dispute.DisputeNo,
		EventType:    model.TimelineEventEvidenceUploaded,
		EventStatus:  dispute.Status,
		Description:  fmt.Sprintf("Evidence uploaded: %s (%s)", input.Title, input.EvidenceType),
		OperatorID:   &input.UploadedBy,
		OperatorType: model.OperatorTypeMerchant,
	}
	s.repo.CreateTimelineEvent(ctx, timeline)

	return evidence, nil
}

// ListEvidence 查询拒付的证据列表
func (s *disputeService) ListEvidence(ctx context.Context, disputeID uuid.UUID) ([]*model.DisputeEvidence, error) {
	evidence, err := s.repo.ListEvidenceByDispute(ctx, disputeID)
	if err != nil {
		return nil, fmt.Errorf("list evidence failed: %w", err)
	}
	return evidence, nil
}

// DeleteEvidence 删除证据
func (s *disputeService) DeleteEvidence(ctx context.Context, evidenceID uuid.UUID) error {
	evidence, err := s.repo.GetEvidenceByID(ctx, evidenceID)
	if err != nil {
		return fmt.Errorf("get evidence failed: %w", err)
	}
	if evidence == nil {
		return fmt.Errorf("evidence not found")
	}

	if evidence.IsSubmitted {
		return fmt.Errorf("cannot delete evidence that has been submitted")
	}

	if err := s.repo.DeleteEvidence(ctx, evidenceID); err != nil {
		return fmt.Errorf("delete evidence failed: %w", err)
	}

	return nil
}

// SubmitToStripe 提交证据到Stripe
func (s *disputeService) SubmitToStripe(ctx context.Context, disputeID uuid.UUID) error {
	dispute, err := s.repo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("get dispute failed: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found")
	}

	if dispute.Channel != "stripe" {
		return fmt.Errorf("only stripe disputes can be submitted via this method")
	}

	if dispute.EvidenceSubmitted {
		return fmt.Errorf("evidence already submitted")
	}

	// Get evidence
	evidenceList, err := s.repo.ListEvidenceByDispute(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("get evidence failed: %w", err)
	}

	if len(evidenceList) == 0 {
		return fmt.Errorf("no evidence to submit")
	}

	// Submit to Stripe
	if err := s.stripeClient.SubmitEvidence(ctx, dispute.ChannelDisputeID, evidenceList); err != nil {
		return fmt.Errorf("submit to stripe failed: %w", err)
	}

	// Update dispute
	now := time.Now()
	dispute.EvidenceSubmitted = true
	dispute.EvidenceSubmitTime = &now
	dispute.Status = model.DisputeStatusUnderReview

	if err := s.repo.UpdateDispute(ctx, dispute); err != nil {
		return fmt.Errorf("update dispute failed: %w", err)
	}

	// Mark all evidence as submitted
	if err := s.repo.MarkEvidenceAsSubmitted(ctx, disputeID); err != nil {
		return fmt.Errorf("mark evidence as submitted failed: %w", err)
	}

	// Create timeline event
	timeline := &model.DisputeTimeline{
		DisputeID:    dispute.ID,
		DisputeNo:    dispute.DisputeNo,
		EventType:    model.TimelineEventEvidenceSubmitted,
		EventStatus:  dispute.Status,
		Description:  fmt.Sprintf("Evidence submitted to Stripe (%d files)", len(evidenceList)),
		OperatorType: model.OperatorTypeSystem,
	}
	s.repo.CreateTimelineEvent(ctx, timeline)

	return nil
}

// SyncFromStripe 从Stripe同步拒付数据
func (s *disputeService) SyncFromStripe(ctx context.Context, channelDisputeID string) (*model.Dispute, error) {
	// Get dispute from Stripe
	stripeDispute, err := s.stripeClient.GetDispute(ctx, channelDisputeID)
	if err != nil {
		return nil, fmt.Errorf("get stripe dispute failed: %w", err)
	}

	// Check if dispute already exists
	existing, err := s.repo.GetDisputeByChannelID(ctx, channelDisputeID)
	if err != nil {
		return nil, fmt.Errorf("check existing dispute failed: %w", err)
	}

	if existing != nil {
		// Update existing dispute
		existing.Status = mapStripeStatus(stripeDispute.Status)
		existing.Reason = string(stripeDispute.Reason)
		if stripeDispute.EvidenceDetails != nil && stripeDispute.EvidenceDetails.DueBy > 0 {
			dueBy := time.Unix(stripeDispute.EvidenceDetails.DueBy, 0)
			existing.EvidenceDueBy = &dueBy
		}

		if err := s.repo.UpdateDispute(ctx, existing); err != nil {
			return nil, fmt.Errorf("update dispute failed: %w", err)
		}

		return existing, nil
	}

	// Create new dispute - Get payment info first
	var merchantID uuid.UUID
	var paymentNo string

	// Fetch payment information from payment-gateway using channel_trade_no (Stripe Charge ID)
	if s.paymentClient != nil && stripeDispute.Charge != nil {
		paymentInfo, err := s.paymentClient.GetPaymentByChannelTradeNo(ctx, stripeDispute.Charge.ID)
		if err != nil {
			// Log error but continue with partial data
			// In production, you might want to retry or queue this for later processing
			fmt.Printf("Warning: Failed to fetch payment info for charge %s: %v\n", stripeDispute.Charge.ID, err)
		} else {
			merchantID = paymentInfo.MerchantID
			paymentNo = paymentInfo.PaymentNo
		}
	}

	evidenceDueBy := (*time.Time)(nil)
	if stripeDispute.EvidenceDetails != nil && stripeDispute.EvidenceDetails.DueBy > 0 {
		dueBy := time.Unix(stripeDispute.EvidenceDetails.DueBy, 0)
		evidenceDueBy = &dueBy
	}

	input := &CreateDisputeInput{
		Channel:          "stripe",
		ChannelDisputeID: stripeDispute.ID,
		PaymentNo:        paymentNo,
		MerchantID:       merchantID,
		ChannelTradeNo:   stripeDispute.Charge.ID,
		Amount:           stripeDispute.Amount,
		Currency:         string(stripeDispute.Currency),
		Reason:           string(stripeDispute.Reason),
		EvidenceDueBy:    evidenceDueBy,
	}

	return s.CreateDispute(ctx, input)
}

// GetStatistics 获取拒付统计信息
func (s *disputeService) GetStatistics(ctx context.Context, merchantID *uuid.UUID, startDate, endDate *time.Time) (*repository.DisputeStatistics, error) {
	stats, err := s.repo.GetDisputeStatistics(ctx, merchantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get statistics failed: %w", err)
	}
	return stats, nil
}

// Helper functions

func generateDisputeNo(channel string) string {
	return fmt.Sprintf("DISPUTE-%s-%d", channel, time.Now().Unix())
}

func mapStripeStatus(status stripe.DisputeStatus) string {
	switch status {
	case stripe.DisputeStatusWarningNeedsResponse:
		return model.DisputeStatusWarningNeedsResponse
	case stripe.DisputeStatusWarningUnderReview:
		return model.DisputeStatusUnderReview
	case stripe.DisputeStatusNeedsResponse:
		return model.DisputeStatusNeedsResponse
	case stripe.DisputeStatusUnderReview:
		return model.DisputeStatusUnderReview
	case stripe.DisputeStatusWon:
		return model.DisputeStatusWon
	case stripe.DisputeStatusLost:
		return model.DisputeStatusLost
	default:
		// Handle other statuses including charge_refunded
		return string(status)
	}
}
