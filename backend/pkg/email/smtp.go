package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"strings"

	"gopkg.in/gomail.v2"
)

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// SMTPProvider SMTP邮件提供商
type SMTPProvider struct {
	config *SMTPConfig
	dialer *gomail.Dialer
}

// NewSMTPProvider 创建SMTP提供商
func NewSMTPProvider(cfg SMTPConfig) (*SMTPProvider, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("SMTP主机不能为空")
	}
	if cfg.Port == 0 {
		cfg.Port = 587 // 默认端口
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("发件人地址不能为空")
	}

	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	// 配置TLS
	dialer.TLSConfig = &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: false,
	}

	provider := &SMTPProvider{
		config: &cfg,
		dialer: dialer,
	}

	return provider, nil
}

// Send 发送邮件
func (p *SMTPProvider) Send(to []string, subject, htmlBody, textBody string, attachments []Attachment) error {
	m := gomail.NewMessage()

	// 设置发件人
	if p.config.FromName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", p.config.FromName, p.config.From))
	} else {
		m.SetHeader("From", p.config.From)
	}

	// 设置收件人
	m.SetHeader("To", to...)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置正文
	if htmlBody != "" {
		if textBody != "" {
			// 同时提供HTML和文本版本
			m.SetBody("text/plain", textBody)
			m.AddAlternative("text/html", htmlBody)
		} else {
			// 仅HTML版本
			m.SetBody("text/html", htmlBody)
		}
	} else if textBody != "" {
		// 仅文本版本
		m.SetBody("text/plain", textBody)
	} else {
		return fmt.Errorf("邮件正文不能为空")
	}

	// 添加附件
	for _, attachment := range attachments {
		m.Attach(attachment.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(attachment.Content)
			return err
		}))
	}

	// 发送邮件
	if err := p.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// SendTemplate 使用模板发送邮件（由Client统一处理）
func (p *SMTPProvider) SendTemplate(to []string, subject, templateName string, data interface{}) error {
	return fmt.Errorf("请使用 Client.SendTemplate 方法")
}

// TestConnection 测试SMTP连接
func (p *SMTPProvider) TestConnection() error {
	// 尝试连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	// 建立连接
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}
	defer client.Close()

	// STARTTLS
	if err := client.StartTLS(&tls.Config{
		ServerName:         p.config.Host,
		InsecureSkipVerify: false,
	}); err != nil {
		return fmt.Errorf("启动TLS失败: %w", err)
	}

	// 认证
	if p.config.Username != "" && p.config.Password != "" {
		auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}
	}

	return nil
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// BatchSend 批量发送邮件（使用连接池）
func (p *SMTPProvider) BatchSend(messages []*EmailMessage) error {
	// 打开一个持久连接
	sender, err := p.dialer.Dial()
	if err != nil {
		return fmt.Errorf("打开SMTP连接失败: %w", err)
	}
	defer sender.Close()

	// 批量发送
	for _, msg := range messages {
		m := gomail.NewMessage()

		// 设置发件人
		if p.config.FromName != "" {
			m.SetHeader("From", fmt.Sprintf("%s <%s>", p.config.FromName, p.config.From))
		} else {
			m.SetHeader("From", p.config.From)
		}

		// 设置收件人
		m.SetHeader("To", msg.To...)

		// 设置主题
		m.SetHeader("Subject", msg.Subject)

		// 设置正文
		if msg.HTMLBody != "" {
			if msg.TextBody != "" {
				m.SetBody("text/plain", msg.TextBody)
				m.AddAlternative("text/html", msg.HTMLBody)
			} else {
				m.SetBody("text/html", msg.HTMLBody)
			}
		} else if msg.TextBody != "" {
			m.SetBody("text/plain", msg.TextBody)
		}

		// 添加附件
		for _, attachment := range msg.Attachments {
			m.Attach(attachment.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Content)
				return err
			}))
		}

		// 发送
		if err := gomail.Send(sender, m); err != nil {
			return fmt.Errorf("批量发送邮件失败: %w", err)
		}
	}

	return nil
}
