package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"gopkg.in/gomail.v2"
)

// EmailProvider 邮件提供商接口
type EmailProvider interface {
	Send(ctx context.Context, req *EmailRequest) (*EmailResponse, error)
	GetProviderName() string
}

// EmailRequest 邮件发送请求
type EmailRequest struct {
	From        string   `json:"from"`         // 发件人
	To          []string `json:"to"`           // 收件人列表
	CC          []string `json:"cc"`           // 抄送列表
	BCC         []string `json:"bcc"`          // 密送列表
	Subject     string   `json:"subject"`      // 主题
	TextBody    string   `json:"text_body"`    // 纯文本内容
	HTMLBody    string   `json:"html_body"`    // HTML内容
	Attachments []Attachment `json:"attachments"` // 附件列表
	Headers     map[string]string `json:"headers"` // 自定义头
}

// EmailResponse 邮件发送响应
type EmailResponse struct {
	MessageID string `json:"message_id"` // 消息ID
	Status    string `json:"status"`     // 状态
}

// Attachment 附件
type Attachment struct {
	Filename string `json:"filename"` // 文件名
	Content  []byte `json:"content"`  // 文件内容
	MimeType string `json:"mime_type"` // MIME类型
}

// SMTPProvider SMTP 邮件提供商
type SMTPProvider struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPProvider 创建 SMTP 提供商实例
func NewSMTPProvider(host string, port int, username, password, from string) *SMTPProvider {
	return &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// GetProviderName 获取提供商名称
func (p *SMTPProvider) GetProviderName() string {
	return "smtp"
}

// Send 发送邮件
func (p *SMTPProvider) Send(ctx context.Context, req *EmailRequest) (*EmailResponse, error) {
	m := gomail.NewMessage()

	// 设置发件人
	from := req.From
	if from == "" {
		from = p.from
	}
	m.SetHeader("From", from)

	// 设置收件人
	if len(req.To) == 0 {
		return nil, fmt.Errorf("收件人列表不能为空")
	}
	m.SetHeader("To", req.To...)

	// 设置抄送
	if len(req.CC) > 0 {
		m.SetHeader("Cc", req.CC...)
	}

	// 设置密送
	if len(req.BCC) > 0 {
		m.SetHeader("Bcc", req.BCC...)
	}

	// 设置主题
	m.SetHeader("Subject", req.Subject)

	// 设置内容
	if req.HTMLBody != "" {
		m.SetBody("text/html", req.HTMLBody)
		if req.TextBody != "" {
			m.AddAlternative("text/plain", req.TextBody)
		}
	} else {
		m.SetBody("text/plain", req.TextBody)
	}

	// 设置附件
	for _, att := range req.Attachments {
		m.Attach(att.Filename)
	}

	// 设置自定义头
	for key, value := range req.Headers {
		m.SetHeader(key, value)
	}

	// 创建邮件发送器
	d := gomail.NewDialer(p.host, p.port, p.username, p.password)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return nil, fmt.Errorf("发送邮件失败: %w", err)
	}

	// 生成消息ID
	messageID := fmt.Sprintf("%d@%s", time.Now().Unix(), p.host)

	return &EmailResponse{
		MessageID: messageID,
		Status:    "sent",
	}, nil
}

// MailgunProvider Mailgun 邮件提供商
type MailgunProvider struct {
	domain string
	apiKey string
	from   string
	mg     *mailgun.MailgunImpl
}

// NewMailgunProvider 创建 Mailgun 提供商实例
func NewMailgunProvider(domain, apiKey, from string) *MailgunProvider {
	mg := mailgun.NewMailgun(domain, apiKey)
	return &MailgunProvider{
		domain: domain,
		apiKey: apiKey,
		from:   from,
		mg:     mg,
	}
}

// GetProviderName 获取提供商名称
func (p *MailgunProvider) GetProviderName() string {
	return "mailgun"
}

// Send 发送邮件
func (p *MailgunProvider) Send(ctx context.Context, req *EmailRequest) (*EmailResponse, error) {
	// 设置发件人
	from := req.From
	if from == "" {
		from = p.from
	}

	// 创建消息
	message := p.mg.NewMessage(
		from,
		req.Subject,
		req.TextBody,
		req.To...,
	)

	// 设置 HTML 内容
	if req.HTMLBody != "" {
		message.SetHtml(req.HTMLBody)
	}

	// 设置抄送
	for _, cc := range req.CC {
		message.AddCC(cc)
	}

	// 设置密送
	for _, bcc := range req.BCC {
		message.AddBCC(bcc)
	}

	// 设置自定义头
	for key, value := range req.Headers {
		message.AddHeader(key, value)
	}

	// 添加附件
	for _, att := range req.Attachments {
		message.AddBufferAttachment(att.Filename, att.Content)
	}

	// 发送邮件
	_, id, err := p.mg.Send(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("Mailgun 发送邮件失败: %w", err)
	}

	return &EmailResponse{
		MessageID: id,
		Status:    "sent",
	}, nil
}

// EmailProviderFactory 邮件提供商工厂
type EmailProviderFactory struct {
	providers map[string]EmailProvider
}

// NewEmailProviderFactory 创建邮件提供商工厂
func NewEmailProviderFactory() *EmailProviderFactory {
	return &EmailProviderFactory{
		providers: make(map[string]EmailProvider),
	}
}

// Register 注册提供商
func (f *EmailProviderFactory) Register(name string, provider EmailProvider) {
	f.providers[name] = provider
}

// GetProvider 获取提供商
func (f *EmailProviderFactory) GetProvider(name string) (EmailProvider, bool) {
	provider, ok := f.providers[name]
	return provider, ok
}
