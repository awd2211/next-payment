package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	exportpkg "github.com/payment-platform/pkg/export"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/payment-gateway/internal/model"
)

// PaymentExportService 支付导出服务
type PaymentExportService struct {
	db            *gorm.DB
	exportService *exportpkg.ExportService
}

// NewPaymentExportService 创建支付导出服务
func NewPaymentExportService(db *gorm.DB, redisClient *redis.Client, storageDir string) *PaymentExportService {
	return &PaymentExportService{
		db:            db,
		exportService: exportpkg.NewExportService(db, redisClient, storageDir),
	}
}

// CreatePaymentExportTask 创建支付记录导出任务
func (s *PaymentExportService) CreatePaymentExportTask(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time, format string) (*exportpkg.ExportTask, error) {
	// 创建导出任务
	task, err := s.exportService.CreateExportTask(ctx, merchantID, "payment", format, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 异步执行导出
	go s.executePaymentExport(context.Background(), task)

	return task, nil
}

// CreateRefundExportTask 创建退款记录导出任务
func (s *PaymentExportService) CreateRefundExportTask(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time, format string) (*exportpkg.ExportTask, error) {
	// 创建导出任务
	task, err := s.exportService.CreateExportTask(ctx, merchantID, "refund", format, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 异步执行导出
	go s.executeRefundExport(context.Background(), task)

	return task, nil
}

// executePaymentExport 执行支付记录导出
func (s *PaymentExportService) executePaymentExport(ctx context.Context, task *exportpkg.ExportTask) {
	logger.Info("开始执行支付记录导出",
		zap.String("task_id", task.ID.String()),
		zap.String("merchant_id", task.MerchantID.String()))

	// 更新状态为处理中
	s.exportService.UpdateTaskStatus(ctx, task.ID, "processing", "")

	// 查询支付记录
	var payments []model.Payment
	err := s.db.WithContext(ctx).
		Where("merchant_id = ? AND created_at >= ? AND created_at <= ?",
			task.MerchantID, task.StartDate, task.EndDate).
		Order("created_at DESC").
		Find(&payments).Error

	if err != nil {
		logger.Error("查询支付记录失败",
			zap.String("task_id", task.ID.String()),
			zap.Error(err))
		s.exportService.UpdateTaskStatus(ctx, task.ID, "failed", err.Error())
		return
	}

	// 准备CSV数据
	headers := []string{
		"支付单号",
		"订单号",
		"商户ID",
		"金额(分)",
		"货币",
		"支付渠道",
		"状态",
		"渠道交易号",
		"支付者邮箱",
		"支付者IP",
		"创建时间",
		"支付完成时间",
	}

	data := make([][]string, 0, len(payments))
	for _, p := range payments {
		paidAtStr := ""
		if p.PaidAt != nil {
			paidAtStr = p.PaidAt.Format("2006-01-02 15:04:05")
		}

		row := []string{
			p.PaymentNo,
			p.OrderNo,
			p.MerchantID.String(),
			fmt.Sprintf("%d", p.Amount),
			p.Currency,
			p.Channel,
			p.Status,
			p.ChannelOrderNo,
			p.CustomerEmail,
			p.CustomerIP,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
			paidAtStr,
		}
		data = append(data, row)
	}

	// 导出到CSV
	if task.Format == "csv" {
		if err := s.exportService.ExportToCSV(ctx, task.ID, headers, data); err != nil {
			logger.Error("导出CSV失败",
				zap.String("task_id", task.ID.String()),
				zap.Error(err))
			s.exportService.UpdateTaskStatus(ctx, task.ID, "failed", err.Error())
			return
		}
	}

	// 更新状态为完成
	s.exportService.UpdateTaskStatus(ctx, task.ID, "completed", "")

	logger.Info("支付记录导出完成",
		zap.String("task_id", task.ID.String()),
		zap.Int("row_count", len(payments)))
}

// executeRefundExport 执行退款记录导出
func (s *PaymentExportService) executeRefundExport(ctx context.Context, task *exportpkg.ExportTask) {
	logger.Info("开始执行退款记录导出",
		zap.String("task_id", task.ID.String()),
		zap.String("merchant_id", task.MerchantID.String()))

	// 更新状态为处理中
	s.exportService.UpdateTaskStatus(ctx, task.ID, "processing", "")

	// 查询退款记录（预加载Payment信息以获取PaymentNo）
	var refunds []model.Refund
	err := s.db.WithContext(ctx).
		Preload("Payment").
		Where("merchant_id = ? AND created_at >= ? AND created_at <= ?",
			task.MerchantID, task.StartDate, task.EndDate).
		Order("created_at DESC").
		Find(&refunds).Error

	if err != nil {
		logger.Error("查询退款记录失败",
			zap.String("task_id", task.ID.String()),
			zap.Error(err))
		s.exportService.UpdateTaskStatus(ctx, task.ID, "failed", err.Error())
		return
	}

	// 准备CSV数据
	headers := []string{
		"退款单号",
		"支付单号",
		"商户ID",
		"退款金额(分)",
		"货币",
		"状态",
		"退款原因",
		"渠道退款号",
		"创建时间",
		"退款完成时间",
	}

	data := make([][]string, 0, len(refunds))
	for _, r := range refunds {
		refundedAtStr := ""
		if r.RefundedAt != nil {
			refundedAtStr = r.RefundedAt.Format("2006-01-02 15:04:05")
		}

		paymentNo := ""
		if r.Payment != nil {
			paymentNo = r.Payment.PaymentNo
		}

		row := []string{
			r.RefundNo,
			paymentNo,
			r.MerchantID.String(),
			fmt.Sprintf("%d", r.Amount),
			r.Currency,
			r.Status,
			r.Reason,
			r.ChannelRefundNo,
			r.CreatedAt.Format("2006-01-02 15:04:05"),
			refundedAtStr,
		}
		data = append(data, row)
	}

	// 导出到CSV
	if task.Format == "csv" {
		if err := s.exportService.ExportToCSV(ctx, task.ID, headers, data); err != nil {
			logger.Error("导出CSV失败",
				zap.String("task_id", task.ID.String()),
				zap.Error(err))
			s.exportService.UpdateTaskStatus(ctx, task.ID, "failed", err.Error())
			return
		}
	}

	// 更新状态为完成
	s.exportService.UpdateTaskStatus(ctx, task.ID, "completed", "")

	logger.Info("退款记录导出完成",
		zap.String("task_id", task.ID.String()),
		zap.Int("row_count", len(refunds)))
}

// GetExportTask 获取导出任务
func (s *PaymentExportService) GetExportTask(ctx context.Context, taskID, merchantID uuid.UUID) (*exportpkg.ExportTask, error) {
	return s.exportService.GetExportTask(ctx, taskID, merchantID)
}

// ListExportTasks 查询导出任务列表
func (s *PaymentExportService) ListExportTasks(ctx context.Context, merchantID uuid.UUID, page, pageSize int) ([]exportpkg.ExportTask, int64, error) {
	return s.exportService.ListExportTasks(ctx, merchantID, page, pageSize)
}
