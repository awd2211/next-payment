package validator

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// StringValidator 字符串验证器
type StringValidator struct{}

// NewStringValidator 创建字符串验证器
func NewStringValidator() *StringValidator {
	return &StringValidator{}
}

var (
	// 邮箱正则表达式
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// 手机号正则表达式（支持国际格式）
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// 信用卡号正则表达式（Luhn算法验证）
	creditCardRegex = regexp.MustCompile(`^\d{13,19}$`)
)

// ValidateEmail 验证邮箱地址
func (v *StringValidator) ValidateEmail(email string) error {
	if email == "" {
		return errors.New("邮箱地址不能为空")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱地址格式无效")
	}
	return nil
}

// ValidatePhone 验证手机号
func (v *StringValidator) ValidatePhone(phone string) error {
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	// 移除空格和短横线
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
	if !phoneRegex.MatchString(cleanPhone) {
		return errors.New("手机号格式无效")
	}
	return nil
}

// ValidateURL 验证URL
func (v *StringValidator) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("URL不能为空")
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("URL格式无效")
	}
	if u.Scheme == "" || u.Host == "" {
		return errors.New("URL必须包含scheme和host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL scheme必须是http或https")
	}
	return nil
}

// ValidateCreditCard 验证信用卡号（Luhn算法）
func (v *StringValidator) ValidateCreditCard(cardNumber string) error {
	if cardNumber == "" {
		return errors.New("信用卡号不能为空")
	}

	// 移除空格
	cleanCard := strings.ReplaceAll(cardNumber, " ", "")

	if !creditCardRegex.MatchString(cleanCard) {
		return errors.New("信用卡号格式无效")
	}

	// Luhn算法验证
	if !luhnCheck(cleanCard) {
		return errors.New("信用卡号校验失败")
	}

	return nil
}

// ValidateLength 验证字符串长度
func (v *StringValidator) ValidateLength(str string, min, max int) error {
	length := len(str)
	if length < min {
		return errors.New("字符串长度不足")
	}
	if max > 0 && length > max {
		return errors.New("字符串长度超过限制")
	}
	return nil
}

// ValidateNotEmpty 验证字符串非空
func (v *StringValidator) ValidateNotEmpty(str string) error {
	if strings.TrimSpace(str) == "" {
		return errors.New("字符串不能为空")
	}
	return nil
}

// ValidateAlphanumeric 验证字符串只包含字母和数字
func (v *StringValidator) ValidateAlphanumeric(str string) error {
	if str == "" {
		return errors.New("字符串不能为空")
	}
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericRegex.MatchString(str) {
		return errors.New("字符串只能包含字母和数字")
	}
	return nil
}

// luhnCheck Luhn算法校验信用卡号
func luhnCheck(cardNumber string) bool {
	sum := 0
	alternate := false

	// 从右往左遍历
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
