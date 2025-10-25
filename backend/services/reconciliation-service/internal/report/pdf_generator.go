package report

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"payment-platform/reconciliation-service/internal/model"
	"payment-platform/reconciliation-service/internal/repository"
)

// PDFGenerator PDF报告生成器
type PDFGenerator struct {
	repo     repository.ReconciliationRepository
	basePath string
}

// NewPDFGenerator 创建PDF生成器
func NewPDFGenerator(repo repository.ReconciliationRepository, basePath string) *PDFGenerator {
	return &PDFGenerator{
		repo:     repo,
		basePath: basePath,
	}
}

// Generate 生成对账报告
func (g *PDFGenerator) Generate(ctx context.Context, task *model.ReconciliationTask) (string, error) {
	// Create directory if not exists
	if err := os.MkdirAll(g.basePath, 0755); err != nil {
		return "", fmt.Errorf("create directory failed: %w", err)
	}

	// Generate file path
	fileName := fmt.Sprintf("report-%s-%s.txt", task.TaskNo, time.Now().Format("20060102150405"))
	filePath := filepath.Join(g.basePath, fileName)

	// Get diff records
	filters := repository.RecordFilters{TaskID: &task.ID}
	records, _, err := g.repo.ListRecords(ctx, filters, 1, 1000) // Get up to 1000 records
	if err != nil {
		return "", fmt.Errorf("get records failed: %w", err)
	}

	// Generate report content (simplified text format)
	content := g.generateTextReport(task, records)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return filePath, nil
}

// generateTextReport 生成文本报告内容
func (g *PDFGenerator) generateTextReport(task *model.ReconciliationTask, records []*model.ReconciliationRecord) string {
	content := fmt.Sprintf(`对账报告
========================================

任务信息:
  任务编号: %s
  对账日期: %s
  支付渠道: %s
  任务状态: %s
  创建时间: %s
  完成时间: %s

统计摘要:
  平台记录数: %d
  平台总金额: %.2f
  渠道记录数: %d
  渠道总金额: %.2f
  匹配记录数: %d
  匹配总金额: %.2f
  差异记录数: %d
  差异总金额: %.2f

`,
		task.TaskNo,
		task.TaskDate.Format("2006-01-02"),
		task.Channel,
		task.Status,
		task.CreatedAt.Format("2006-01-02 15:04:05"),
		formatTime(task.CompletedAt),
		task.PlatformCount,
		float64(task.PlatformAmount)/100.0,
		task.ChannelCount,
		float64(task.ChannelAmount)/100.0,
		task.MatchedCount,
		float64(task.MatchedAmount)/100.0,
		task.DiffCount,
		float64(task.DiffAmount)/100.0,
	)

	// Add diff records detail
	if len(records) > 0 {
		content += "差异明细:\n"
		content += "----------------------------------------\n"

		for i, record := range records {
			content += fmt.Sprintf(`
记录 %d:
  支付单号: %s
  渠道交易号: %s
  差异类型: %s
  平台金额: %.2f
  渠道金额: %.2f
  差异金额: %.2f
  差异原因: %s
  处理状态: %s
`,
				i+1,
				record.PaymentNo,
				record.ChannelTradeNo,
				translateDiffType(record.DiffType),
				float64(record.PlatformAmount)/100.0,
				float64(record.ChannelAmount)/100.0,
				float64(record.DiffAmount)/100.0,
				record.DiffReason,
				translateResolvedStatus(record.IsResolved),
			)
		}
	}

	content += "\n========================================\n"
	content += fmt.Sprintf("报告生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return content
}

// Helper functions

func formatTime(t *time.Time) string {
	if t == nil {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}

func translateDiffType(diffType string) string {
	switch diffType {
	case model.DiffTypeMatched:
		return "完全匹配"
	case model.DiffTypePlatformOnly:
		return "仅平台有记录"
	case model.DiffTypeChannelOnly:
		return "仅渠道有记录"
	case model.DiffTypeAmountDiff:
		return "金额不一致"
	case model.DiffTypeStatusDiff:
		return "状态不一致"
	default:
		return diffType
	}
}

func translateResolvedStatus(isResolved bool) string {
	if isResolved {
		return "已解决"
	}
	return "未解决"
}
