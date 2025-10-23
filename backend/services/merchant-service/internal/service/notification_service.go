package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/email"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// NotificationService 通知服务接口
type NotificationService interface {
	// 发送登录通知
	SendLoginNotification(ctx context.Context, merchantID uuid.UUID, loginActivity *model.LoginActivity) error
	// 发送异常登录警告
	SendAbnormalLoginAlert(ctx context.Context, merchantID uuid.UUID, loginActivity *model.LoginActivity) error
	// 发送密码修改通知
	SendPasswordChangedNotification(ctx context.Context, merchantID uuid.UUID) error
	// 发送2FA启用通知
	Send2FAEnabledNotification(ctx context.Context, merchantID uuid.UUID) error
	// 发送2FA禁用通知
	Send2FADisabledNotification(ctx context.Context, merchantID uuid.UUID) error
}

type notificationService struct {
	merchantRepo  repository.MerchantRepository
	securityRepo  repository.SecurityRepository
	emailProvider email.EmailProvider
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(
	merchantRepo repository.MerchantRepository,
	securityRepo repository.SecurityRepository,
	emailProvider email.EmailProvider,
) NotificationService {
	return &notificationService{
		merchantRepo:  merchantRepo,
		securityRepo:  securityRepo,
		emailProvider: emailProvider,
	}
}

// SendLoginNotification 发送登录通知
func (s *notificationService) SendLoginNotification(ctx context.Context, merchantID uuid.UUID, loginActivity *model.LoginActivity) error {
	// 检查是否启用登录通知
	settings, err := s.securityRepo.GetSecuritySettings(ctx, merchantID)
	if err != nil || settings == nil || !settings.LoginNotification {
		return nil // 未启用，直接返回
	}

	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil || merchant == nil {
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 构建邮件内容
	subject := "新设备登录通知"
	body := fmt.Sprintf(`
		<h2>新设备登录通知</h2>
		<p>尊敬的 %s，</p>
		<p>您的账户在新设备上登录：</p>
		<ul>
			<li>登录时间: %s</li>
			<li>IP地址: %s</li>
			<li>设备类型: %s</li>
			<li>浏览器: %s</li>
			<li>操作系统: %s</li>
			<li>位置: %s</li>
		</ul>
		<p>如果这不是您的操作，请立即修改密码并联系我们。</p>
	`, merchant.Name,
	   loginActivity.LoginAt.Format("2006-01-02 15:04:05"),
	   loginActivity.IP,
	   loginActivity.DeviceType,
	   loginActivity.Browser,
	   loginActivity.OS,
	   loginActivity.Location)

	// 发送邮件
	return s.emailProvider.Send(
		[]string{merchant.Email},
		subject,
		body,
		"", // textBody
		nil, // attachments
	)
}

// SendAbnormalLoginAlert 发送异常登录警告
func (s *notificationService) SendAbnormalLoginAlert(ctx context.Context, merchantID uuid.UUID, loginActivity *model.LoginActivity) error {
	// 检查是否启用异常通知
	settings, err := s.securityRepo.GetSecuritySettings(ctx, merchantID)
	if err != nil || settings == nil || !settings.AbnormalNotification {
		return nil // 未启用，直接返回
	}

	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil || merchant == nil {
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 构建邮件内容
	subject := "⚠️ 检测到异常登录活动"
	body := fmt.Sprintf(`
		<h2 style="color: #d32f2f;">⚠️ 异常登录警告</h2>
		<p>尊敬的 %s，</p>
		<p>我们检测到您的账户存在异常登录活动：</p>
		<ul>
			<li>登录时间: %s</li>
			<li>IP地址: %s</li>
			<li>设备类型: %s</li>
			<li>浏览器: %s</li>
			<li>操作系统: %s</li>
			<li>位置: %s</li>
			<li><strong>异常原因: %s</strong></li>
		</ul>
		<p style="color: #d32f2f;"><strong>如果这不是您的操作，请立即：</strong></p>
		<ol>
			<li>修改密码</li>
			<li>启用双因素认证（2FA）</li>
			<li>联系我们的安全团队</li>
		</ol>
	`, merchant.Name,
	   loginActivity.LoginAt.Format("2006-01-02 15:04:05"),
	   loginActivity.IP,
	   loginActivity.DeviceType,
	   loginActivity.Browser,
	   loginActivity.OS,
	   loginActivity.Location,
	   loginActivity.AbnormalReason)

	// 发送邮件
	return s.emailProvider.Send(
		[]string{merchant.Email},
		subject,
		body,
		"", // textBody
		nil, // attachments
	)
}

// SendPasswordChangedNotification 发送密码修改通知
func (s *notificationService) SendPasswordChangedNotification(ctx context.Context, merchantID uuid.UUID) error {
	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil || merchant == nil {
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 构建邮件内容
	subject := "密码修改成功通知"
	body := fmt.Sprintf(`
		<h2>密码修改成功</h2>
		<p>尊敬的 %s，</p>
		<p>您的账户密码已成功修改。</p>
		<p>如果这不是您的操作，请立即联系我们的安全团队。</p>
	`, merchant.Name)

	// 发送邮件
	return s.emailProvider.Send(
		[]string{merchant.Email},
		subject,
		body,
		"", // textBody
		nil, // attachments
	)
}

// Send2FAEnabledNotification 发送2FA启用通知
func (s *notificationService) Send2FAEnabledNotification(ctx context.Context, merchantID uuid.UUID) error {
	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil || merchant == nil {
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 构建邮件内容
	subject := "双因素认证已启用"
	body := fmt.Sprintf(`
		<h2>双因素认证已启用</h2>
		<p>尊敬的 %s，</p>
		<p>您的账户已成功启用双因素认证（2FA）。</p>
		<p>今后登录时，除了密码外，还需要提供验证码。</p>
		<p>这将大大提高您的账户安全性。</p>
	`, merchant.Name)

	// 发送邮件
	return s.emailProvider.Send(
		[]string{merchant.Email},
		subject,
		body,
		"", // textBody
		nil, // attachments
	)
}

// Send2FADisabledNotification 发送2FA禁用通知
func (s *notificationService) Send2FADisabledNotification(ctx context.Context, merchantID uuid.UUID) error {
	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil || merchant == nil {
		return fmt.Errorf("获取商户信息失败: %w", err)
	}

	// 构建邮件内容
	subject := "双因素认证已禁用"
	body := fmt.Sprintf(`
		<h2>双因素认证已禁用</h2>
		<p>尊敬的 %s，</p>
		<p>您的账户已禁用双因素认证（2FA）。</p>
		<p>如果这不是您的操作，请立即联系我们的安全团队。</p>
		<p>建议您重新启用2FA以保护账户安全。</p>
	`, merchant.Name)

	// 发送邮件
	return s.emailProvider.Send(
		[]string{merchant.Email},
		subject,
		body,
		"", // textBody
		nil, // attachments
	)
}
