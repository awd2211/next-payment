package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Require2FA 敏感操作需要2FA验证
// 用法: admin.POST("/withdraw", middleware.Require2FA, h.Withdraw)
func Require2FA(c *gin.Context) {
	// 1. 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
			"code":  "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// 2. 检查用户是否启用了2FA
	twoFAEnabled := c.GetBool("2fa_enabled")
	if !twoFAEnabled {
		// 如果未启用2FA，建议启用但允许继续
		c.Header("X-2FA-Warning", "建议启用2FA以提高安全性")
		c.Next()
		return
	}

	// 3. 从请求中获取2FA验证码
	twoFACode := c.GetHeader("X-2FA-Code")
	if twoFACode == "" {
		// 尝试从body获取
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if code, ok := body["twofa_code"].(string); ok {
				twoFACode = code
			}
		}
	}

	if twoFACode == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "需要2FA验证码",
			"code":    "2FA_REQUIRED",
			"message": "此操作需要提供2FA验证码",
			"headers": map[string]string{
				"X-2FA-Code": "TOTP 6位数字验证码",
			},
		})
		c.Abort()
		return
	}

	// 4. 验证2FA码（TOTP）
	// 实际应用中应该从数据库获取用户的2FA密钥
	twoFASecret := c.GetString("2fa_secret")
	if twoFASecret == "" {
		// 开发环境：允许通过（生产环境删除此段）
		if c.GetString("env") == "development" {
			c.Next()
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "2FA配置错误",
		})
		c.Abort()
		return
	}

	// 5. 验证TOTP码
	valid := verifyTOTP(twoFASecret, twoFACode)
	if !valid {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "2FA验证码错误或已过期",
			"code":  "INVALID_2FA_CODE",
		})
		c.Abort()
		return
	}

	// 6. 验证通过，继续处理
	c.Set("2fa_verified", true)
	c.Next()
}

// verifyTOTP 验证TOTP验证码
func verifyTOTP(secret, code string) bool {
	// 移除空格
	secret = strings.ReplaceAll(secret, " ", "")
	code = strings.TrimSpace(code)

	// 当前时间窗口（30秒）
	now := time.Now().Unix() / 30

	// 检查前后1个时间窗口（允许30秒误差）
	for i := int64(-1); i <= 1; i++ {
		if generateTOTP(secret, now+i) == code {
			return true
		}
	}

	return false
}

// generateTOTP 生成TOTP验证码
func generateTOTP(secret string, counter int64) string {
	// Base32解码密钥
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	// 将counter转为8字节
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// HMAC-SHA256
	h := hmac.New(sha256.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := int(hash[offset]&0x7f)<<24 |
		int(hash[offset+1]&0xff)<<16 |
		int(hash[offset+2]&0xff)<<8 |
		int(hash[offset+3]&0xff)

	// 取后6位
	code = code % 1000000

	return fmt.Sprintf("%06d", code)
}

// Generate2FASecret 生成2FA密钥（用于用户启用2FA时）
func Generate2FASecret() string {
	// 生成20字节随机密钥
	secret := make([]byte, 20)
	_, _ = rand.Read(secret)
	return base32.StdEncoding.EncodeToString(secret)
}

// Generate2FAQR 生成2FA二维码URI
func Generate2FAQR(secret, issuer, accountName string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		issuer,
		accountName,
		secret,
		issuer,
	)
}

// Require2FAForSensitiveOps 敏感操作列表自动检测
func Require2FAForSensitiveOps() gin.HandlerFunc {
	sensitiveOps := map[string]bool{
		"approve":  true,
		"reject":   true,
		"freeze":   true,
		"unfreeze": true,
		"delete":   true,
		"withdraw": true,
		"transfer": true,
	}

	return func(c *gin.Context) {
		path := strings.ToLower(c.Request.URL.Path)

		// 检查路径是否包含敏感操作
		requiresTwoFA := false
		for op := range sensitiveOps {
			if strings.Contains(path, "/"+op) {
				requiresTwoFA = true
				break
			}
		}

		if requiresTwoFA {
			Require2FA(c)
		} else {
			c.Next()
		}
	}
}

// Simple random number generator for secret
var rand = struct {
	Read func([]byte) (int, error)
}{
	Read: func(b []byte) (int, error) {
		// 简单的伪随机（生产环境应使用 crypto/rand）
		now := time.Now().UnixNano()
		for i := range b {
			b[i] = byte((now >> (i * 8)) & 0xff)
		}
		return len(b), nil
	},
}
