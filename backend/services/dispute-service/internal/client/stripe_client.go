package client

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/dispute"

	"payment-platform/dispute-service/internal/model"
)

// StripeDisputeClient Stripe拒付客户端
type StripeDisputeClient struct {
	apiKey string
}

// NewStripeDisputeClient 创建Stripe拒付客户端
func NewStripeDisputeClient(apiKey string) *StripeDisputeClient {
	stripe.Key = apiKey
	return &StripeDisputeClient{
		apiKey: apiKey,
	}
}

// GetDispute 获取Stripe拒付详情
func (c *StripeDisputeClient) GetDispute(ctx context.Context, disputeID string) (*stripe.Dispute, error) {
	d, err := dispute.Get(disputeID, nil)
	if err != nil {
		return nil, fmt.Errorf("get stripe dispute failed: %w", err)
	}
	return d, nil
}

// SubmitEvidence 提交证据到Stripe
func (c *StripeDisputeClient) SubmitEvidence(ctx context.Context, disputeID string, evidenceList []*model.DisputeEvidence) error {
	// Build evidence params
	params := &stripe.DisputeParams{}
	evidence := &stripe.DisputeEvidenceParams{}

	// Map evidence to Stripe fields
	for _, e := range evidenceList {
		switch e.EvidenceType {
		case model.EvidenceTypeReceipt:
			if e.FileURL != "" {
				evidence.Receipt = stripe.String(e.FileURL)
			}
		case model.EvidenceTypeShippingProof:
			if e.FileURL != "" {
				evidence.ShippingDocumentation = stripe.String(e.FileURL)
			}
		case model.EvidenceTypeCommunication:
			if e.Description != "" {
				evidence.CustomerCommunication = stripe.String(e.Description)
			}
		case model.EvidenceTypeRefundPolicy:
			if e.FileURL != "" {
				evidence.RefundPolicy = stripe.String(e.FileURL)
			}
		case model.EvidenceTypeCancellationPolicy:
			if e.FileURL != "" {
				evidence.CancellationPolicy = stripe.String(e.FileURL)
			}
		case model.EvidenceTypeCustomerSignature:
			if e.FileURL != "" {
				evidence.CustomerSignature = stripe.String(e.FileURL)
			}
		case model.EvidenceTypeServiceDocumentation:
			if e.FileURL != "" {
				evidence.ServiceDocumentation = stripe.String(e.FileURL)
			}
		}
	}

	params.Evidence = evidence

	// Submit to Stripe
	_, err := dispute.Update(disputeID, params)
	if err != nil {
		return fmt.Errorf("update stripe dispute evidence failed: %w", err)
	}

	// Close dispute (submit for review)
	closeParams := &stripe.DisputeParams{
		Submit: stripe.Bool(true),
	}
	_, err = dispute.Update(disputeID, closeParams)
	if err != nil {
		return fmt.Errorf("submit stripe dispute failed: %w", err)
	}

	return nil
}

// ListDisputes 列出Stripe拒付
func (c *StripeDisputeClient) ListDisputes(ctx context.Context, params *stripe.DisputeListParams) ([]*stripe.Dispute, error) {
	var disputes []*stripe.Dispute

	i := dispute.List(params)
	for i.Next() {
		disputes = append(disputes, i.Dispute())
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("list stripe disputes failed: %w", err)
	}

	return disputes, nil
}
