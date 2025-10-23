package email

import (
	"bytes"
	"fmt"
	"html/template"
)

// EmailProvider 邮件提供商接口
type EmailProvider interface {
	Send(to []string, subject, htmlBody, textBody string, attachments []Attachment) error
	SendTemplate(to []string, subject, templateName string, data interface{}) error
}

// Attachment 邮件附件
type Attachment struct {
	Filename string
	Content  []byte
	MimeType string
}

// EmailMessage 邮件消息
type EmailMessage struct {
	To          []string            // 收件人列表
	Cc          []string            // 抄送列表
	Bcc         []string            // 密送列表
	Subject     string              // 主题
	HTMLBody    string              // HTML正文
	TextBody    string              // 纯文本正文
	Attachments []Attachment        // 附件
	Headers     map[string]string   // 自定义头部
	ReplyTo     string              // 回复地址
}

// Config 邮件配置
type Config struct {
	Provider string // smtp, mailgun

	// SMTP配置
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// Mailgun配置
	MailgunDomain     string
	MailgunAPIKey     string
	MailgunFrom       string
	MailgunFromName   string
	MailgunEURegion   bool // 是否使用欧盟区域

	// 模板配置
	TemplatePath string // 模板文件路径
}

// Client 邮件客户端
type Client struct {
	provider      EmailProvider
	templateCache map[string]*template.Template
	config        *Config
}

// NewClient 创建邮件客户端
func NewClient(cfg *Config) (*Client, error) {
	var provider EmailProvider
	var err error

	switch cfg.Provider {
	case "smtp":
		provider, err = NewSMTPProvider(SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
	case "mailgun":
		provider, err = NewMailgunProvider(MailgunConfig{
			Domain:   cfg.MailgunDomain,
			APIKey:   cfg.MailgunAPIKey,
			From:     cfg.MailgunFrom,
			FromName: cfg.MailgunFromName,
			EURegion: cfg.MailgunEURegion,
		})
	default:
		return nil, fmt.Errorf("不支持的邮件提供商: %s", cfg.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("初始化邮件提供商失败: %w", err)
	}

	client := &Client{
		provider:      provider,
		templateCache: make(map[string]*template.Template),
		config:        cfg,
	}

	// 加载模板
	if cfg.TemplatePath != "" {
		if err := client.loadTemplates(); err != nil {
			return nil, fmt.Errorf("加载邮件模板失败: %w", err)
		}
	}

	return client, nil
}

// Send 发送邮件
func (c *Client) Send(msg *EmailMessage) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("收件人列表不能为空")
	}

	return c.provider.Send(msg.To, msg.Subject, msg.HTMLBody, msg.TextBody, msg.Attachments)
}

// SendTemplate 使用模板发送邮件
func (c *Client) SendTemplate(to []string, subject, templateName string, data interface{}) error {
	if len(to) == 0 {
		return fmt.Errorf("收件人列表不能为空")
	}

	tmpl, ok := c.templateCache[templateName]
	if !ok {
		return fmt.Errorf("模板不存在: %s", templateName)
	}

	// 渲染HTML模板
	var htmlBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&htmlBuf, templateName+".html", data); err != nil {
		return fmt.Errorf("渲染HTML模板失败: %w", err)
	}

	// 渲染文本模板（可选）
	var textBuf bytes.Buffer
	if tmpl.Lookup(templateName+".txt") != nil {
		if err := tmpl.ExecuteTemplate(&textBuf, templateName+".txt", data); err != nil {
			return fmt.Errorf("渲染文本模板失败: %w", err)
		}
	}

	return c.provider.Send(to, subject, htmlBuf.String(), textBuf.String(), nil)
}

// loadTemplates 加载所有邮件模板
func (c *Client) loadTemplates() error {
	// 这里简化处理，实际可以遍历目录加载所有模板
	templates := []string{
		"welcome",              // 欢迎邮件
		"verify_email",         // 邮箱验证
		"reset_password",       // 重置密码
		"payment_success",      // 支付成功
		"payment_failed",       // 支付失败
		"refund_completed",     // 退款完成
		"merchant_approved",    // 商户审核通过
		"merchant_rejected",    // 商户审核拒绝
		"invoice",              // 账单
	}

	for _, name := range templates {
		htmlPath := fmt.Sprintf("%s/%s.html", c.config.TemplatePath, name)
		txtPath := fmt.Sprintf("%s/%s.txt", c.config.TemplatePath, name)

		tmpl := template.New(name)

		// 加载HTML模板
		if _, err := tmpl.New(name + ".html").ParseFiles(htmlPath); err != nil {
			// HTML模板不存在是错误
			return fmt.Errorf("加载HTML模板失败 %s: %w", name, err)
		}

		// 加载文本模板（可选）
		if _, err := tmpl.New(name + ".txt").ParseFiles(txtPath); err != nil {
			// 文本模板不存在不是错误，忽略
		}

		c.templateCache[name] = tmpl
	}

	return nil
}

// AddTemplate 动态添加模板
func (c *Client) AddTemplate(name, htmlContent, textContent string) error {
	tmpl := template.New(name)

	// 解析HTML模板
	if _, err := tmpl.New(name + ".html").Parse(htmlContent); err != nil {
		return fmt.Errorf("解析HTML模板失败: %w", err)
	}

	// 解析文本模板（可选）
	if textContent != "" {
		if _, err := tmpl.New(name + ".txt").Parse(textContent); err != nil {
			return fmt.Errorf("解析文本模板失败: %w", err)
		}
	}

	c.templateCache[name] = tmpl
	return nil
}
