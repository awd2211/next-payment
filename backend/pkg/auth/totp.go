package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager TOTP双因素认证管理器
type TOTPManager struct {
	issuer string // 发行者名称（如："Payment Platform"）
}

// NewTOTPManager 创建TOTP管理器
func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		issuer: issuer,
	}
}

// GenerateSecret 生成TOTP密钥
// accountName: 用户账户名（如：邮箱或用户名）
// 返回：base32编码的密钥和otpauth URL
func (tm *TOTPManager) GenerateSecret(accountName string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tm.issuer,
		AccountName: accountName,
		SecretSize:  32, // 256位密钥
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", fmt.Errorf("生成TOTP密钥失败: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// VerifyCode 验证TOTP代码
// secret: base32编码的密钥
// code: 用户输入的6位数字代码
func (tm *TOTPManager) VerifyCode(secret, code string) bool {
	// 允许前后各30秒的时间窗口（共1分钟容错）
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1, // 允许1个时间窗口的偏差
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	return err == nil && valid
}

// GenerateCode 生成当前时间的TOTP代码（用于测试）
func (tm *TOTPManager) GenerateCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

// GenerateBackupCodes 生成备用恢复代码
// count: 生成的备用代码数量（建议8-10个）
func (tm *TOTPManager) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// 生成8字节随机数据
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("生成备用代码失败: %w", err)
		}

		// Base32编码并格式化为xxxx-xxxx格式
		encoded := base32.StdEncoding.EncodeToString(b)
		codes[i] = fmt.Sprintf("%s-%s", encoded[:4], encoded[4:8])
	}

	return codes, nil
}

// ValidateBackupCode 验证备用代码格式
func (tm *TOTPManager) ValidateBackupCode(code string) bool {
	// 备用代码应该是xxxx-xxxx格式，共9个字符
	if len(code) != 9 {
		return false
	}

	// 检查格式
	if code[4] != '-' {
		return false
	}

	// 检查是否为有效的base32字符
	allowed := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	for i, c := range code {
		if i == 4 {
			continue
		}
		found := false
		for _, a := range allowed {
			if c == a {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
