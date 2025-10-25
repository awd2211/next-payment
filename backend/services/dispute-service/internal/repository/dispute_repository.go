package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payment-platform/dispute-service/internal/model"
)

// DisputeRepository 拒付数据仓库接口
type DisputeRepository interface {
	// Dispute operations
	CreateDispute(ctx context.Context, dispute *model.Dispute) error
	GetDisputeByID(ctx context.Context, id uuid.UUID) (*model.Dispute, error)
	GetDisputeByNo(ctx context.Context, disputeNo string) (*model.Dispute, error)
	GetDisputeByChannelID(ctx context.Context, channelDisputeID string) (*model.Dispute, error)
	UpdateDispute(ctx context.Context, dispute *model.Dispute) error
	ListDisputes(ctx context.Context, filters DisputeFilters, page, pageSize int) ([]*model.Dispute, int64, error)
	CountDisputesByStatus(ctx context.Context, merchantID *uuid.UUID, status string) (int, error)

	// Evidence operations
	CreateEvidence(ctx context.Context, evidence *model.DisputeEvidence) error
	GetEvidenceByID(ctx context.Context, id uuid.UUID) (*model.DisputeEvidence, error)
	ListEvidenceByDispute(ctx context.Context, disputeID uuid.UUID) ([]*model.DisputeEvidence, error)
	UpdateEvidence(ctx context.Context, evidence *model.DisputeEvidence) error
	DeleteEvidence(ctx context.Context, id uuid.UUID) error
	MarkEvidenceAsSubmitted(ctx context.Context, disputeID uuid.UUID) error

	// Timeline operations
	CreateTimelineEvent(ctx context.Context, event *model.DisputeTimeline) error
	ListTimelineByDispute(ctx context.Context, disputeID uuid.UUID) ([]*model.DisputeTimeline, error)

	// Statistics
	GetDisputeStatistics(ctx context.Context, merchantID *uuid.UUID, startDate, endDate *time.Time) (*DisputeStatistics, error)
}

// Filters and DTOs

type DisputeFilters struct {
	MerchantID         *uuid.UUID
	Channel            string
	Status             string
	Reason             string
	AssignedTo         *uuid.UUID
	EvidenceSubmitted  *bool
	StartDate          *time.Time
	EndDate            *time.Time
	PaymentNo          string
}

type DisputeStatistics struct {
	TotalCount       int     `json:"total_count"`
	WonCount         int     `json:"won_count"`
	LostCount        int     `json:"lost_count"`
	PendingCount     int     `json:"pending_count"`
	WinRate          float64 `json:"win_rate"`
	TotalAmount      int64   `json:"total_amount"`
	RefundedAmount   int64   `json:"refunded_amount"`
}

// disputeRepository 拒付仓库实现
type disputeRepository struct {
	db *gorm.DB
}

// NewDisputeRepository 创建拒付仓库实例
func NewDisputeRepository(db *gorm.DB) DisputeRepository {
	return &disputeRepository{db: db}
}

// CreateDispute 创建拒付记录
func (r *disputeRepository) CreateDispute(ctx context.Context, dispute *model.Dispute) error {
	if err := r.db.WithContext(ctx).Create(dispute).Error; err != nil {
		return fmt.Errorf("create dispute failed: %w", err)
	}
	return nil
}

// GetDisputeByID 根据ID获取拒付记录
func (r *disputeRepository) GetDisputeByID(ctx context.Context, id uuid.UUID) (*model.Dispute, error) {
	var dispute model.Dispute
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dispute).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get dispute by id failed: %w", err)
	}
	return &dispute, nil
}

// GetDisputeByNo 根据DisputeNo获取拒付记录
func (r *disputeRepository) GetDisputeByNo(ctx context.Context, disputeNo string) (*model.Dispute, error) {
	var dispute model.Dispute
	if err := r.db.WithContext(ctx).Where("dispute_no = ?", disputeNo).First(&dispute).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get dispute by no failed: %w", err)
	}
	return &dispute, nil
}

// GetDisputeByChannelID 根据ChannelDisputeID获取拒付记录
func (r *disputeRepository) GetDisputeByChannelID(ctx context.Context, channelDisputeID string) (*model.Dispute, error) {
	var dispute model.Dispute
	if err := r.db.WithContext(ctx).Where("channel_dispute_id = ?", channelDisputeID).First(&dispute).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get dispute by channel id failed: %w", err)
	}
	return &dispute, nil
}

// UpdateDispute 更新拒付记录
func (r *disputeRepository) UpdateDispute(ctx context.Context, dispute *model.Dispute) error {
	if err := r.db.WithContext(ctx).Save(dispute).Error; err != nil {
		return fmt.Errorf("update dispute failed: %w", err)
	}
	return nil
}

// ListDisputes 查询拒付列表
func (r *disputeRepository) ListDisputes(ctx context.Context, filters DisputeFilters, page, pageSize int) ([]*model.Dispute, int64, error) {
	var disputes []*model.Dispute
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Dispute{})

	// Apply filters
	if filters.MerchantID != nil {
		query = query.Where("merchant_id = ?", filters.MerchantID)
	}
	if filters.Channel != "" {
		query = query.Where("channel = ?", filters.Channel)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Reason != "" {
		query = query.Where("reason = ?", filters.Reason)
	}
	if filters.AssignedTo != nil {
		query = query.Where("assigned_to = ?", filters.AssignedTo)
	}
	if filters.EvidenceSubmitted != nil {
		query = query.Where("evidence_submitted = ?", *filters.EvidenceSubmitted)
	}
	if filters.PaymentNo != "" {
		query = query.Where("payment_no = ?", filters.PaymentNo)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", filters.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count disputes failed: %w", err)
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&disputes).Error; err != nil {
		return nil, 0, fmt.Errorf("list disputes failed: %w", err)
	}

	return disputes, total, nil
}

// CountDisputesByStatus 统计指定状态的拒付数量
func (r *disputeRepository) CountDisputesByStatus(ctx context.Context, merchantID *uuid.UUID, status string) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Dispute{}).Where("status = ?", status)

	if merchantID != nil {
		query = query.Where("merchant_id = ?", merchantID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count disputes by status failed: %w", err)
	}

	return int(count), nil
}

// CreateEvidence 创建证据记录
func (r *disputeRepository) CreateEvidence(ctx context.Context, evidence *model.DisputeEvidence) error {
	if err := r.db.WithContext(ctx).Create(evidence).Error; err != nil {
		return fmt.Errorf("create evidence failed: %w", err)
	}
	return nil
}

// GetEvidenceByID 根据ID获取证据记录
func (r *disputeRepository) GetEvidenceByID(ctx context.Context, id uuid.UUID) (*model.DisputeEvidence, error) {
	var evidence model.DisputeEvidence
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&evidence).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get evidence by id failed: %w", err)
	}
	return &evidence, nil
}

// ListEvidenceByDispute 查询拒付的所有证据
func (r *disputeRepository) ListEvidenceByDispute(ctx context.Context, disputeID uuid.UUID) ([]*model.DisputeEvidence, error) {
	var evidences []*model.DisputeEvidence
	if err := r.db.WithContext(ctx).
		Where("dispute_id = ?", disputeID).
		Order("created_at DESC").
		Find(&evidences).Error; err != nil {
		return nil, fmt.Errorf("list evidence by dispute failed: %w", err)
	}
	return evidences, nil
}

// UpdateEvidence 更新证据记录
func (r *disputeRepository) UpdateEvidence(ctx context.Context, evidence *model.DisputeEvidence) error {
	if err := r.db.WithContext(ctx).Save(evidence).Error; err != nil {
		return fmt.Errorf("update evidence failed: %w", err)
	}
	return nil
}

// DeleteEvidence 删除证据记录（软删除）
func (r *disputeRepository) DeleteEvidence(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.DisputeEvidence{}, id).Error; err != nil {
		return fmt.Errorf("delete evidence failed: %w", err)
	}
	return nil
}

// MarkEvidenceAsSubmitted 标记拒付的所有证据为已提交
func (r *disputeRepository) MarkEvidenceAsSubmitted(ctx context.Context, disputeID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&model.DisputeEvidence{}).
		Where("dispute_id = ? AND is_submitted = false", disputeID).
		Updates(map[string]interface{}{
			"is_submitted": true,
			"submitted_at": now,
		}).Error; err != nil {
		return fmt.Errorf("mark evidence as submitted failed: %w", err)
	}
	return nil
}

// CreateTimelineEvent 创建时间线事件
func (r *disputeRepository) CreateTimelineEvent(ctx context.Context, event *model.DisputeTimeline) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("create timeline event failed: %w", err)
	}
	return nil
}

// ListTimelineByDispute 查询拒付的时间线
func (r *disputeRepository) ListTimelineByDispute(ctx context.Context, disputeID uuid.UUID) ([]*model.DisputeTimeline, error) {
	var timeline []*model.DisputeTimeline
	if err := r.db.WithContext(ctx).
		Where("dispute_id = ?", disputeID).
		Order("created_at ASC").
		Find(&timeline).Error; err != nil {
		return nil, fmt.Errorf("list timeline by dispute failed: %w", err)
	}
	return timeline, nil
}

// GetDisputeStatistics 获取拒付统计信息
func (r *disputeRepository) GetDisputeStatistics(ctx context.Context, merchantID *uuid.UUID, startDate, endDate *time.Time) (*DisputeStatistics, error) {
	stats := &DisputeStatistics{}

	query := r.db.WithContext(ctx).Model(&model.Dispute{})

	if merchantID != nil {
		query = query.Where("merchant_id = ?", merchantID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	// Total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count total disputes failed: %w", err)
	}
	stats.TotalCount = int(total)

	// Won count
	var wonCount int64
	if err := query.Where("result = ?", model.DisputeResultWon).Count(&wonCount).Error; err != nil {
		return nil, fmt.Errorf("count won disputes failed: %w", err)
	}
	stats.WonCount = int(wonCount)

	// Lost count
	var lostCount int64
	if err := query.Where("result = ?", model.DisputeResultLost).Count(&lostCount).Error; err != nil {
		return nil, fmt.Errorf("count lost disputes failed: %w", err)
	}
	stats.LostCount = int(lostCount)

	// Pending count (no result yet)
	var pendingCount int64
	if err := query.Where("result IS NULL OR result = ''").Count(&pendingCount).Error; err != nil {
		return nil, fmt.Errorf("count pending disputes failed: %w", err)
	}
	stats.PendingCount = int(pendingCount)

	// Win rate
	if stats.WonCount+stats.LostCount > 0 {
		stats.WinRate = float64(stats.WonCount) / float64(stats.WonCount+stats.LostCount) * 100
	}

	// Total amount
	var totalAmount int64
	if err := r.db.WithContext(ctx).Model(&model.Dispute{}).
		Select("COALESCE(SUM(amount), 0)").
		Where(query).
		Scan(&totalAmount).Error; err != nil {
		return nil, fmt.Errorf("sum total amount failed: %w", err)
	}
	stats.TotalAmount = totalAmount

	// Refunded amount
	var refundedAmount int64
	if err := r.db.WithContext(ctx).Model(&model.Dispute{}).
		Select("COALESCE(SUM(refund_amount), 0)").
		Where(query).
		Where("is_refunded = true").
		Scan(&refundedAmount).Error; err != nil {
		return nil, fmt.Errorf("sum refunded amount failed: %w", err)
	}
	stats.RefundedAmount = refundedAmount

	return stats, nil
}
