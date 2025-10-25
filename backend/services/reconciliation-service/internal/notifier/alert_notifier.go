package notifier

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"go.uber.org/zap"

	"github.com/payment-platform/pkg/email"
	"payment-platform/reconciliation-service/internal/model"
)

// AlertNotifier 告警通知器
type AlertNotifier struct {
	emailClient     *email.Client
	logger          *zap.Logger
	alertRecipients []string // 告警接收人邮箱列表
}

// NewAlertNotifier 创建告警通知器
func NewAlertNotifier(emailClient *email.Client, logger *zap.Logger, recipients []string) *AlertNotifier {
	return &AlertNotifier{
		emailClient:     emailClient,
		logger:          logger,
		alertRecipients: recipients,
	}
}

// SendDifferenceAlert 发送差异告警
func (n *AlertNotifier) SendDifferenceAlert(ctx context.Context, task *model.ReconciliationTask, differences []*model.ReconciliationDifference) error {
	if len(differences) == 0 {
		return nil
	}

	// 分类差异
	critical := n.filterBySeverity(differences, "critical")
	high := n.filterBySeverity(differences, "high")
	medium := n.filterBySeverity(differences, "medium")

	// 生成告警内容
	subject := fmt.Sprintf("【对账告警】%s 发现 %d 笔差异", task.TaskDate.Format("2006-01-02"), len(differences))

	body, err := n.generateAlertEmail(task, differences, critical, high, medium)
	if err != nil {
		return fmt.Errorf("failed to generate alert email: %w", err)
	}

	// 发送邮件
	msg := &email.EmailMessage{
		To:       n.alertRecipients,
		Subject:  subject,
		HTMLBody: body,
	}

	if err := n.emailClient.Send(msg); err != nil {
		n.logger.Error("Failed to send alert email", zap.Error(err))
		return err
	}

	n.logger.Info("Alert email sent",
		zap.String("task_id", task.ID.String()),
		zap.Int("diff_count", len(differences)))

	return nil
}

// SendCriticalAlert 发送严重告警（紧急通知）
func (n *AlertNotifier) SendCriticalAlert(ctx context.Context, task *model.ReconciliationTask, criticalDiffs []*model.ReconciliationDifference) error {
	if len(criticalDiffs) == 0 {
		return nil
	}

	subject := fmt.Sprintf("【紧急告警】%s 发现 %d 笔严重差异", task.TaskDate.Format("2006-01-02"), len(criticalDiffs))

	body, err := n.generateCriticalAlertEmail(task, criticalDiffs)
	if err != nil {
		return fmt.Errorf("failed to generate critical alert email: %w", err)
	}

	// 发送邮件
	msg := &email.EmailMessage{
		To:       n.alertRecipients,
		Subject:  subject,
		HTMLBody: body,
	}

	if err := n.emailClient.Send(msg); err != nil {
		n.logger.Error("Failed to send critical alert", zap.Error(err))
		return err
	}

	n.logger.Warn("Critical alert sent",
		zap.String("task_id", task.ID.String()),
		zap.Int("critical_count", len(criticalDiffs)))

	return nil
}

// SendDailyReport 发送每日对账报告
func (n *AlertNotifier) SendDailyReport(ctx context.Context, report *model.ReconciliationReport) error {
	subject := fmt.Sprintf("【每日对账报告】%s", report.ReportDate.Format("2006-01-02"))

	body, err := n.generateDailyReportEmail(report)
	if err != nil {
		return fmt.Errorf("failed to generate daily report email: %w", err)
	}

	// 发送邮件
	msg := &email.EmailMessage{
		To:       n.alertRecipients,
		Subject:  subject,
		HTMLBody: body,
	}

	if err := n.emailClient.Send(msg); err != nil {
		n.logger.Error("Failed to send daily report", zap.Error(err))
		return err
	}

	n.logger.Info("Daily report sent",
		zap.Time("report_date", report.ReportDate),
		zap.Int("total_tasks", report.TotalTasks))

	return nil
}

// filterBySeverity 按严重程度过滤差异
func (n *AlertNotifier) filterBySeverity(differences []*model.ReconciliationDifference, severity string) []*model.ReconciliationDifference {
	filtered := make([]*model.ReconciliationDifference, 0)
	for _, diff := range differences {
		if diff.Severity == severity {
			filtered = append(filtered, diff)
		}
	}
	return filtered
}

// generateAlertEmail 生成告警邮件内容
func (n *AlertNotifier) generateAlertEmail(
	task *model.ReconciliationTask,
	allDiffs, critical, high, medium []*model.ReconciliationDifference,
) (string, error) {

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { background: #f44336; color: white; padding: 20px; border-radius: 5px; }
        .summary { background: #fff3cd; padding: 15px; margin: 20px 0; border-left: 4px solid #ffc107; }
        .stats { display: flex; justify-content: space-around; margin: 20px 0; }
        .stat { text-align: center; padding: 15px; background: #f5f5f5; border-radius: 5px; flex: 1; margin: 0 10px; }
        .stat h3 { margin: 0; color: #666; font-size: 14px; }
        .stat p { margin: 10px 0 0 0; font-size: 24px; font-weight: bold; }
        .critical { color: #d32f2f; }
        .high { color: #f57c00; }
        .medium { color: #fbc02d; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f5f5f5; font-weight: bold; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 2px solid #eee; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>对账差异告警</h1>
            <p>对账日期: {{.TaskDate}}</p>
        </div>

        <div class="summary">
            <h2>概览</h2>
            <p>在 {{.TaskDate}} 的对账中发现 <strong>{{.TotalDifferences}}</strong> 笔差异</p>
        </div>

        <div class="stats">
            <div class="stat">
                <h3>严重差异</h3>
                <p class="critical">{{.CriticalCount}}</p>
            </div>
            <div class="stat">
                <h3>高级差异</h3>
                <p class="high">{{.HighCount}}</p>
            </div>
            <div class="stat">
                <h3>中级差异</h3>
                <p class="medium">{{.MediumCount}}</p>
            </div>
        </div>

        {{if .CriticalDiffs}}
        <h2>严重差异详情</h2>
        <table>
            <thead>
                <tr>
                    <th>订单号</th>
                    <th>类型</th>
                    <th>描述</th>
                    <th>金额差</th>
                </tr>
            </thead>
            <tbody>
                {{range .CriticalDiffs}}
                <tr>
                    <td>{{.OrderNo}}</td>
                    <td>{{.DifferenceType}}</td>
                    <td>{{.Description}}</td>
                    <td class="critical">{{.AmountDiff}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{end}}

        <div class="footer">
            <p>此邮件由对账系统自动发送，请勿回复。</p>
            <p>发送时间: {{.SentAt}}</p>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("alert").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"TaskDate":         task.TaskDate.Format("2006-01-02"),
		"TotalDifferences": len(allDiffs),
		"CriticalCount":    len(critical),
		"HighCount":        len(high),
		"MediumCount":      len(medium),
		"CriticalDiffs":    critical,
		"SentAt":           time.Now().Format("2006-01-02 15:04:05"),
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateCriticalAlertEmail 生成严重告警邮件
func (n *AlertNotifier) generateCriticalAlertEmail(task *model.ReconciliationTask, criticalDiffs []*model.ReconciliationDifference) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; color: #333; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { background: #c62828; color: white; padding: 25px; border-radius: 5px; }
        .urgent { background: #ffebee; padding: 20px; margin: 20px 0; border-left: 5px solid #c62828; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #ffcdd2; font-weight: bold; }
        .amount { color: #c62828; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚠️ 严重对账差异告警</h1>
            <p>对账日期: {{.TaskDate}}</p>
        </div>

        <div class="urgent">
            <h2>紧急处理</h2>
            <p><strong>发现 {{.CriticalCount}} 笔严重差异，请立即处理！</strong></p>
        </div>

        <table>
            <thead>
                <tr>
                    <th>订单号</th>
                    <th>差异类型</th>
                    <th>内部金额</th>
                    <th>渠道金额</th>
                    <th>差额</th>
                    <th>描述</th>
                </tr>
            </thead>
            <tbody>
                {{range .CriticalDiffs}}
                <tr>
                    <td>{{.OrderNo}}</td>
                    <td>{{.DifferenceType}}</td>
                    <td>{{.InternalAmount}}</td>
                    <td>{{.ChannelAmount}}</td>
                    <td class="amount">{{.AmountDiff}}</td>
                    <td>{{.Description}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>

        <p style="margin-top: 30px; color: #666; font-size: 12px;">
            发送时间: {{.SentAt}}<br>
            此邮件由对账系统自动发送
        </p>
    </div>
</body>
</html>
`

	t, err := template.New("critical").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"TaskDate":      task.TaskDate.Format("2006-01-02"),
		"CriticalCount": len(criticalDiffs),
		"CriticalDiffs": criticalDiffs,
		"SentAt":        time.Now().Format("2006-01-02 15:04:05"),
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateDailyReportEmail 生成每日报告邮件
func (n *AlertNotifier) generateDailyReportEmail(report *model.ReconciliationReport) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; color: #333; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { background: #1976d2; color: white; padding: 20px; border-radius: 5px; }
        .stats { display: flex; justify-content: space-around; margin: 20px 0; }
        .stat { text-align: center; padding: 15px; background: #e3f2fd; border-radius: 5px; flex: 1; margin: 0 10px; }
        .stat h3 { margin: 0; color: #1976d2; font-size: 14px; }
        .stat p { margin: 10px 0 0 0; font-size: 24px; font-weight: bold; color: #0d47a1; }
        .section { margin: 20px 0; padding: 15px; background: #f5f5f5; border-radius: 5px; }
        .success { color: #388e3c; }
        .warning { color: #f57c00; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>每日对账报告</h1>
            <p>报告日期: {{.ReportDate}}</p>
        </div>

        <div class="stats">
            <div class="stat">
                <h3>对账任务数</h3>
                <p>{{.TotalTasks}}</p>
            </div>
            <div class="stat">
                <h3>成功匹配</h3>
                <p class="success">{{.TotalMatched}}</p>
            </div>
            <div class="stat">
                <h3>发现差异</h3>
                <p class="warning">{{.TotalDifferences}}</p>
            </div>
        </div>

        <div class="section">
            <h2>对账结果</h2>
            <p>内部交易总数: {{.TotalInternal}}</p>
            <p>渠道交易总数: {{.TotalChannel}}</p>
            <p>匹配率: {{.MatchRate}}%</p>
        </div>

        {{if gt .TotalDifferences 0}}
        <div class="section">
            <h2>差异概况</h2>
            <p>严重差异: {{.CriticalDiffs}}</p>
            <p>高级差异: {{.HighDiffs}}</p>
            <p>中级差异: {{.MediumDiffs}}</p>
            <p>低级差异: {{.LowDiffs}}</p>
        </div>
        {{end}}

        <p style="margin-top: 30px; color: #666; font-size: 12px;">
            报告生成时间: {{.GeneratedAt}}<br>
            此邮件由对账系统自动发送
        </p>
    </div>
</body>
</html>
`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return "", err
	}

	matchRate := 0.0
	if report.TotalInternal > 0 {
		matchRate = float64(report.TotalMatched) / float64(report.TotalInternal) * 100
	}

	data := map[string]interface{}{
		"ReportDate":       report.ReportDate.Format("2006-01-02"),
		"TotalTasks":       report.TotalTasks,
		"TotalMatched":     report.TotalMatched,
		"TotalDifferences": report.TotalDifferences,
		"TotalInternal":    report.TotalInternal,
		"TotalChannel":     report.TotalChannel,
		"MatchRate":        fmt.Sprintf("%.2f", matchRate),
		"CriticalDiffs":    report.CriticalDiffs,
		"HighDiffs":        report.HighDiffs,
		"MediumDiffs":      report.MediumDiffs,
		"LowDiffs":         report.LowDiffs,
		"GeneratedAt":      time.Now().Format("2006-01-02 15:04:05"),
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
