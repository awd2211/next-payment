package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/email"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

var (
	ErrTemplateNotFound = errors.New("邮件模板不存在")
	ErrTemplateExists   = errors.New("模板代码已存在")
)

// EmailTemplateService 邮件模板服务接口
type EmailTemplateService interface {
	// 模板管理
	CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*model.EmailTemplate, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*model.EmailTemplate, error)
	GetTemplateByCode(ctx context.Context, code string) (*model.EmailTemplate, error)
	ListTemplates(ctx context.Context, page, pageSize int, category string, isActive *bool) ([]*model.EmailTemplate, int64, error)
	UpdateTemplate(ctx context.Context, req *UpdateTemplateRequest) (*model.EmailTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	// 邮件发送
	SendEmail(ctx context.Context, req *SendEmailRequest) error
	SendTemplateEmail(ctx context.Context, req *SendTemplateEmailRequest) error
	TestTemplate(ctx context.Context, req *TestTemplateRequest) (string, error)

	// 邮件日志
	ListEmailLogs(ctx context.Context, page, pageSize int, status, to string) ([]*model.EmailLog, int64, error)

	// 初始化系统默认模板
	InitDefaultTemplates(ctx context.Context) error
}

type emailTemplateService struct {
	templateRepo repository.EmailTemplateRepository
	emailClient  *email.Client
}

// NewEmailTemplateService 创建邮件模板服务实例
func NewEmailTemplateService(
	templateRepo repository.EmailTemplateRepository,
	emailClient *email.Client,
) EmailTemplateService {
	return &emailTemplateService{
		templateRepo: templateRepo,
		emailClient:  emailClient,
	}
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Code        string
	Name        string
	Subject     string
	HTMLContent string
	TextContent string
	Description string
	Category    string
	Variables   []model.EmailTemplateVariable
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	ID          uuid.UUID
	Name        string
	Subject     string
	HTMLContent string
	TextContent string
	Description string
	IsActive    bool
	Variables   []model.EmailTemplateVariable
	UpdatedBy   uuid.UUID
}

// SendEmailRequest 发送邮件请求
type SendEmailRequest struct {
	To          []string
	Subject     string
	HTMLContent string
	TextContent string
}

// SendTemplateEmailRequest 使用模板发送邮件请求
type SendTemplateEmailRequest struct {
	To           []string
	TemplateCode string
	Data         map[string]interface{}
}

// TestTemplateRequest 测试模板请求
type TestTemplateRequest struct {
	TemplateID uuid.UUID
	Data       map[string]interface{}
}

// CreateTemplate 创建邮件模板
func (s *emailTemplateService) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*model.EmailTemplate, error) {
	// 检查代码是否已存在
	existing, err := s.templateRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTemplateExists
	}

	// 序列化变量列表
	variablesJSON, err := json.Marshal(req.Variables)
	if err != nil {
		return nil, err
	}

	template := &model.EmailTemplate{
		Code:        req.Code,
		Name:        req.Name,
		Subject:     req.Subject,
		HTMLContent: req.HTMLContent,
		TextContent: req.TextContent,
		Description: req.Description,
		Category:    req.Category,
		Variables:   string(variablesJSON),
		IsActive:    true,
		IsSystem:    false,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetTemplate 获取模板详情
func (s *emailTemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*model.EmailTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

// GetTemplateByCode 根据代码获取模板
func (s *emailTemplateService) GetTemplateByCode(ctx context.Context, code string) (*model.EmailTemplate, error) {
	template, err := s.templateRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

// ListTemplates 获取模板列表
func (s *emailTemplateService) ListTemplates(ctx context.Context, page, pageSize int, category string, isActive *bool) ([]*model.EmailTemplate, int64, error) {
	return s.templateRepo.List(ctx, page, pageSize, category, isActive)
}

// UpdateTemplate 更新模板
func (s *emailTemplateService) UpdateTemplate(ctx context.Context, req *UpdateTemplateRequest) (*model.EmailTemplate, error) {
	// 获取现有模板
	template, err := s.templateRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}

	// 系统模板不允许修改代码和分类
	if template.IsSystem {
		// 只允许修改内容和状态
		template.HTMLContent = req.HTMLContent
		template.TextContent = req.TextContent
		template.Subject = req.Subject
		template.IsActive = req.IsActive
	} else {
		// 更新所有字段
		template.Name = req.Name
		template.Subject = req.Subject
		template.HTMLContent = req.HTMLContent
		template.TextContent = req.TextContent
		template.Description = req.Description
		template.IsActive = req.IsActive

		// 序列化变量列表
		if len(req.Variables) > 0 {
			variablesJSON, err := json.Marshal(req.Variables)
			if err != nil {
				return nil, err
			}
			template.Variables = string(variablesJSON)
		}
	}

	template.UpdatedBy = req.UpdatedBy

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteTemplate 删除模板
func (s *emailTemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return s.templateRepo.Delete(ctx, id)
}

// SendEmail 直接发送邮件
func (s *emailTemplateService) SendEmail(ctx context.Context, req *SendEmailRequest) error {
	// 创建邮件日志
	log := &model.EmailLog{
		To:       req.To[0], // 记录第一个收件人
		Subject:  req.Subject,
		Status:   "pending",
		Provider: "smtp", // 从配置获取
	}

	if err := s.templateRepo.CreateLog(ctx, log); err != nil {
		return err
	}

	// 发送邮件
	msg := &email.EmailMessage{
		To:       req.To,
		Subject:  req.Subject,
		HTMLBody: req.HTMLContent,
		TextBody: req.TextContent,
	}

	err := s.emailClient.Send(msg)

	// 更新日志状态
	if err != nil {
		log.Status = "failed"
		log.ErrorMsg = err.Error()
	} else {
		log.Status = "sent"
		now := time.Now()
		log.SentAt = &now
	}

	s.templateRepo.UpdateLog(ctx, log)

	return err
}

// SendTemplateEmail 使用模板发送邮件
func (s *emailTemplateService) SendTemplateEmail(ctx context.Context, req *SendTemplateEmailRequest) error {
	// 获取模板
	tmpl, err := s.templateRepo.GetByCode(ctx, req.TemplateCode)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return ErrTemplateNotFound
	}

	// 渲染主题
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return err
	}
	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, req.Data); err != nil {
		return err
	}

	// 渲染HTML内容
	htmlTmpl, err := template.New("html").Parse(tmpl.HTMLContent)
	if err != nil {
		return err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, req.Data); err != nil {
		return err
	}

	// 渲染文本内容（可选）
	var textBuf bytes.Buffer
	if tmpl.TextContent != "" {
		textTmpl, err := template.New("text").Parse(tmpl.TextContent)
		if err != nil {
			return err
		}
		if err := textTmpl.Execute(&textBuf, req.Data); err != nil {
			return err
		}
	}

	// 创建邮件日志
	log := &model.EmailLog{
		TemplateID: tmpl.ID,
		To:         req.To[0],
		Subject:    subjectBuf.String(),
		Status:     "pending",
		Provider:   "smtp",
	}

	if err := s.templateRepo.CreateLog(ctx, log); err != nil {
		return err
	}

	// 发送邮件
	msg := &email.EmailMessage{
		To:       req.To,
		Subject:  subjectBuf.String(),
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	err = s.emailClient.Send(msg)

	// 更新日志状态
	if err != nil {
		log.Status = "failed"
		log.ErrorMsg = err.Error()
	} else {
		log.Status = "sent"
		now := time.Now()
		log.SentAt = &now
	}

	s.templateRepo.UpdateLog(ctx, log)

	return err
}

// TestTemplate 测试模板渲染
func (s *emailTemplateService) TestTemplate(ctx context.Context, req *TestTemplateRequest) (string, error) {
	// 获取模板
	tmpl, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return "", err
	}
	if tmpl == nil {
		return "", ErrTemplateNotFound
	}

	// 渲染HTML内容
	htmlTmpl, err := template.New("html").Parse(tmpl.HTMLContent)
	if err != nil {
		return "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, req.Data); err != nil {
		return "", err
	}

	return htmlBuf.String(), nil
}

// ListEmailLogs 获取邮件日志列表
func (s *emailTemplateService) ListEmailLogs(ctx context.Context, page, pageSize int, status, to string) ([]*model.EmailLog, int64, error) {
	return s.templateRepo.ListLogs(ctx, page, pageSize, status, to)
}

// InitDefaultTemplates 初始化系统默认模板
func (s *emailTemplateService) InitDefaultTemplates(ctx context.Context) error {
	// 定义默认模板
	defaultTemplates := []CreateTemplateRequest{
		{
			Code:        model.TemplateWelcome,
			Name:        "欢迎邮件",
			Subject:     "欢迎加入 Payment Platform",
			HTMLContent: getDefaultWelcomeHTML(),
			Category:    model.CategoryAccount,
			Variables: []model.EmailTemplateVariable{
				{Name: "Name", Placeholder: "{{.Name}}", Description: "用户名", Required: true},
				{Name: "Email", Placeholder: "{{.Email}}", Description: "邮箱地址", Required: true},
				{Name: "DashboardURL", Placeholder: "{{.DashboardURL}}", Description: "控制台链接", Required: true},
			},
		},
		{
			Code:        model.TemplateVerifyEmail,
			Name:        "邮箱验证",
			Subject:     "验证您的邮箱地址",
			HTMLContent: getDefaultVerifyEmailHTML(),
			Category:    model.CategorySecurity,
			Variables: []model.EmailTemplateVariable{
				{Name: "Name", Placeholder: "{{.Name}}", Description: "用户名", Required: true},
				{Name: "VerificationCode", Placeholder: "{{.VerificationCode}}", Description: "验证码", Required: true},
				{Name: "VerificationURL", Placeholder: "{{.VerificationURL}}", Description: "验证链接", Required: true},
				{Name: "ExpiresIn", Placeholder: "{{.ExpiresIn}}", Description: "过期时间（小时）", Required: true},
			},
		},
		{
			Code:        model.TemplatePaymentSuccess,
			Name:        "支付成功通知",
			Subject:     "支付成功 - {{.OrderNo}}",
			HTMLContent: getDefaultPaymentSuccessHTML(),
			Category:    model.CategoryPayment,
			Variables: []model.EmailTemplateVariable{
				{Name: "CustomerName", Placeholder: "{{.CustomerName}}", Description: "客户名称", Required: true},
				{Name: "Amount", Placeholder: "{{.Amount}}", Description: "支付金额", Required: true},
				{Name: "Currency", Placeholder: "{{.Currency}}", Description: "货币", Required: true},
				{Name: "OrderNo", Placeholder: "{{.OrderNo}}", Description: "订单号", Required: true},
				{Name: "PaymentID", Placeholder: "{{.PaymentID}}", Description: "支付ID", Required: true},
			},
		},
	}

	// 创建默认模板
	for _, tmpl := range defaultTemplates {
		// 检查是否已存在
		existing, _ := s.templateRepo.GetByCode(ctx, tmpl.Code)
		if existing == nil {
			if _, err := s.CreateTemplate(ctx, &tmpl); err != nil {
				return err
			}
		}
	}

	return nil
}

// 获取默认HTML模板（简化版）
func getDefaultWelcomeHTML() string {
	return `<!DOCTYPE html><html><body><h1>欢迎，{{.Name}}！</h1><p>感谢注册 Payment Platform。</p><a href="{{.DashboardURL}}">进入控制台</a></body></html>`
}

func getDefaultVerifyEmailHTML() string {
	return `<!DOCTYPE html><html><body><h1>验证您的邮箱</h1><p>验证码：<strong>{{.VerificationCode}}</strong></p><a href="{{.VerificationURL}}">点击验证</a></body></html>`
}

func getDefaultPaymentSuccessHTML() string {
	return `<!DOCTYPE html><html><body><h1>支付成功！</h1><p>金额：{{.Currency}} {{.Amount}}</p><p>订单号：{{.OrderNo}}</p></body></html>`
}
