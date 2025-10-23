package email

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// MailgunConfig Mailgun配置
type MailgunConfig struct {
	Domain   string // Mailgun域名
	APIKey   string // API密钥
	From     string // 发件人地址
	FromName string // 发件人名称
	EURegion bool   // 是否使用欧盟区域
}

// MailgunProvider Mailgun邮件提供商
type MailgunProvider struct {
	config *MailgunConfig
	client *mailgun.MailgunImpl
}

// NewMailgunProvider 创建Mailgun提供商
func NewMailgunProvider(cfg MailgunConfig) (*MailgunProvider, error) {
	if cfg.Domain == "" {
		return nil, fmt.Errorf("Mailgun域名不能为空")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Mailgun API密钥不能为空")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("发件人地址不能为空")
	}

	// 创建Mailgun客户端
	mg := mailgun.NewMailgun(cfg.Domain, cfg.APIKey)

	// 设置区域
	if cfg.EURegion {
		mg.SetAPIBase(mailgun.APIBaseEU)
	}

	provider := &MailgunProvider{
		config: &cfg,
		client: mg,
	}

	return provider, nil
}

// Send 发送邮件
func (p *MailgunProvider) Send(to []string, subject, htmlBody, textBody string, attachments []Attachment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建邮件消息
	message := p.client.NewMessage(
		p.getFrom(),
		subject,
		textBody, // 纯文本版本
		to...,
	)

	// 设置HTML正文
	if htmlBody != "" {
		message.SetHtml(htmlBody)
	}

	// 添加附件
	for _, attachment := range attachments {
		message.AddBufferAttachment(attachment.Filename, attachment.Content)
	}

	// 发送邮件
	_, _, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun发送邮件失败: %w", err)
	}

	return nil
}

// SendTemplate 使用模板发送邮件（由Client统一处理）
func (p *MailgunProvider) SendTemplate(to []string, subject, templateName string, data interface{}) error {
	return fmt.Errorf("请使用 Client.SendTemplate 方法")
}

// SendWithTemplate 使用Mailgun原生模板发送
func (p *MailgunProvider) SendWithTemplate(to []string, templateName string, variables map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := p.client.NewMessage(
		p.getFrom(),
		"", // 模板中定义主题
		"", // 模板中定义正文
		to...,
	)

	// 使用Mailgun模板
	message.SetTemplate(templateName)

	// 设置模板变量
	for key, value := range variables {
		message.AddVariable(key, value)
	}

	// 发送邮件
	_, _, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun发送模板邮件失败: %w", err)
	}

	return nil
}

// SendBatch 批量发送邮件（使用Mailgun批量发送功能）
func (p *MailgunProvider) SendBatch(messages []*EmailMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, msg := range messages {
		message := p.client.NewMessage(
			p.getFrom(),
			msg.Subject,
			msg.TextBody,
			msg.To...,
		)

		if msg.HTMLBody != "" {
			message.SetHtml(msg.HTMLBody)
		}

		// 添加抄送
		if len(msg.Cc) > 0 {
			for _, cc := range msg.Cc {
				message.AddCC(cc)
			}
		}

		// 添加密送
		if len(msg.Bcc) > 0 {
			for _, bcc := range msg.Bcc {
				message.AddBCC(bcc)
			}
		}

		// 添加附件
		for _, attachment := range msg.Attachments {
			message.AddBufferAttachment(attachment.Filename, attachment.Content)
		}

		// 设置自定义头部
		for key, value := range msg.Headers {
			message.AddHeader(key, value)
		}

		// 设置回复地址
		if msg.ReplyTo != "" {
			message.SetReplyTo(msg.ReplyTo)
		}

		// 发送
		_, _, err := p.client.Send(ctx, message)
		if err != nil {
			return fmt.Errorf("批量发送邮件失败: %w", err)
		}
	}

	return nil
}

// ScheduleSend 定时发送邮件
func (p *MailgunProvider) ScheduleSend(to []string, subject, htmlBody, textBody string, deliveryTime time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := p.client.NewMessage(
		p.getFrom(),
		subject,
		textBody,
		to...,
	)

	if htmlBody != "" {
		message.SetHtml(htmlBody)
	}

	// 设置定时发送时间
	message.SetDeliveryTime(deliveryTime)

	// 发送邮件
	_, _, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun定时发送邮件失败: %w", err)
	}

	return nil
}

// TestConnection 测试Mailgun连接
func (p *MailgunProvider) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取域名信息来验证连接
	_, err := p.client.GetDomain(ctx, p.config.Domain)
	if err != nil {
		return fmt.Errorf("Mailgun连接测试失败: %w", err)
	}

	return nil
}

// GetStats 获取邮件统计信息
func (p *MailgunProvider) GetStats(event string, startDate, endDate time.Time) ([]mailgun.Stats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	it := p.client.ListStats(&mailgun.StatsOptions{
		Event:      []string{event},
		Start:      &startDate,
		End:        &endDate,
		Resolution: "day",
	})

	var stats []mailgun.Stats
	var page []mailgun.Stats

	for it.Next(ctx, &page) {
		stats = append(stats, page...)
	}

	if it.Err() != nil {
		return nil, fmt.Errorf("获取统计信息失败: %w", it.Err())
	}

	return stats, nil
}

// ValidateEmail 验证邮箱地址（使用Mailgun验证API）
func (p *MailgunProvider) ValidateEmail(email string) (*mailgun.EmailVerification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	validator := p.client
	verification, err := validator.ValidateEmail(ctx, email, false)
	if err != nil {
		return nil, fmt.Errorf("邮箱验证失败: %w", err)
	}

	return &verification, nil
}

// getFrom 获取完整的发件人地址
func (p *MailgunProvider) getFrom() string {
	if p.config.FromName != "" {
		return fmt.Sprintf("%s <%s>", p.config.FromName, p.config.From)
	}
	return p.config.From
}

// AddTag 为邮件添加标签（用于追踪）
func (p *MailgunProvider) SendWithTags(to []string, subject, htmlBody, textBody string, tags []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := p.client.NewMessage(
		p.getFrom(),
		subject,
		textBody,
		to...,
	)

	if htmlBody != "" {
		message.SetHtml(htmlBody)
	}

	// 添加标签
	for _, tag := range tags {
		message.AddTag(tag)
	}

	// 发送邮件
	_, _, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("发送带标签的邮件失败: %w", err)
	}

	return nil
}

// GetBounces 获取退信列表
func (p *MailgunProvider) GetBounces() ([]mailgun.Bounce, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	it := p.client.ListBounces(nil)

	var bounces []mailgun.Bounce
	var page []mailgun.Bounce

	for it.Next(ctx, &page) {
		bounces = append(bounces, page...)
	}

	if it.Err() != nil {
		return nil, fmt.Errorf("获取退信列表失败: %w", it.Err())
	}

	return bounces, nil
}

// AddToSuppressionList 添加邮箱到抑制列表
func (p *MailgunProvider) AddToSuppressionList(email, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := p.client.CreateBounce(ctx, &mailgun.Bounce{
		Address: email,
		Error:   reason,
	})

	if err != nil {
		return fmt.Errorf("添加到抑制列表失败: %w", err)
	}

	return nil
}
