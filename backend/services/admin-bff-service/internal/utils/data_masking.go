package utils

import (
	"regexp"
	"strings"
)

// MaskSensitiveData 脱敏敏感数据
func MaskSensitiveData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return data
	}

	// 递归处理嵌套结构
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = maskString(key, v)
		case map[string]interface{}:
			data[key] = MaskSensitiveData(v)
		case []interface{}:
			data[key] = maskArray(v)
		}
	}

	return data
}

// maskString 根据字段名脱敏字符串
func maskString(fieldName, value string) string {
	lowerField := strings.ToLower(fieldName)

	// 手机号
	if strings.Contains(lowerField, "phone") || strings.Contains(lowerField, "mobile") {
		return maskPhone(value)
	}

	// 邮箱
	if strings.Contains(lowerField, "email") {
		return maskEmail(value)
	}

	// 身份证号
	if strings.Contains(lowerField, "id_card") || strings.Contains(lowerField, "identity") {
		return maskIDCard(value)
	}

	// 银行卡号
	if strings.Contains(lowerField, "bank") && strings.Contains(lowerField, "card") {
		return maskBankCard(value)
	}

	// 密码（完全隐藏）
	if strings.Contains(lowerField, "password") || strings.Contains(lowerField, "secret") {
		return "******"
	}

	// API Key（部分隐藏）
	if strings.Contains(lowerField, "api_key") || strings.Contains(lowerField, "access_key") {
		return maskAPIKey(value)
	}

	return value
}

// maskPhone 手机号脱敏 (138****5678)
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}

	// 中国手机号 (11位)
	if len(phone) == 11 {
		return phone[:3] + "****" + phone[7:]
	}

	// 其他格式
	if len(phone) > 7 {
		mid := len(phone) / 2
		return phone[:3] + "****" + phone[mid+2:]
	}

	return phone
}

// maskEmail 邮箱脱敏 (a****@example.com)
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return username + "****@" + domain
	}

	return username[:1] + "****" + username[len(username)-1:] + "@" + domain
}

// maskIDCard 身份证号脱敏 (310***********1234)
func maskIDCard(idCard string) string {
	if len(idCard) < 8 {
		return idCard
	}

	return idCard[:3] + "***********" + idCard[len(idCard)-4:]
}

// maskBankCard 银行卡号脱敏 (6222 **** **** 1234)
func maskBankCard(cardNo string) string {
	// 移除空格
	cardNo = strings.ReplaceAll(cardNo, " ", "")

	if len(cardNo) < 8 {
		return cardNo
	}

	masked := cardNo[:4] + " **** **** " + cardNo[len(cardNo)-4:]
	return masked
}

// maskAPIKey API密钥脱敏 (显示前8位和后4位)
func maskAPIKey(key string) string {
	if len(key) < 16 {
		return key[:4] + "************"
	}

	return key[:8] + "************" + key[len(key)-4:]
}

// maskArray 处理数组
func maskArray(arr []interface{}) []interface{} {
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]interface{}:
			arr[i] = MaskSensitiveData(v)
		case []interface{}:
			arr[i] = maskArray(v)
		}
	}
	return arr
}

// MaskCreditCard 信用卡号脱敏（用于payment相关）
func MaskCreditCard(cardNo string) string {
	// 移除所有非数字字符
	re := regexp.MustCompile(`\D`)
	cleaned := re.ReplaceAllString(cardNo, "")

	if len(cleaned) < 8 {
		return cardNo
	}

	// 保留前6位（BIN）和后4位
	return cleaned[:6] + "******" + cleaned[len(cleaned)-4:]
}

// MaskIPAddress IP地址部分脱敏 (192.168.***.*** )
func MaskIPAddress(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip
	}

	return parts[0] + "." + parts[1] + ".***." + "***"
}

// ShouldMaskField 判断字段是否应该脱敏
func ShouldMaskField(fieldName string) bool {
	sensitiveFields := []string{
		"phone", "mobile", "email", "password", "secret",
		"id_card", "identity", "bank_card", "api_key",
		"access_key", "credit_card", "cvv", "pin",
	}

	lowerField := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerField, sensitive) {
			return true
		}
	}

	return false
}
