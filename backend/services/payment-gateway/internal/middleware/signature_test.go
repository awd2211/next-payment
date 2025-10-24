package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSignatureVersionSupport 测试签名版本支持
func TestSignatureVersionSupport(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"v1 should be supported", "v1", true},
		{"v2 should be supported", "v2", true},
		{"v3 should not be supported", "v3", false},
		{"empty should be supported (defaults to v1)", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := tt.version
			if version == "" {
				version = "v1" // 默认版本
			}
			result := isSupportedSignatureVersion(version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCalculateSignatureByVersion 测试不同版本的签名计算
func TestCalculateSignatureByVersion(t *testing.T) {
	// 创建一个简单的中间件实例用于测试签名计算
	middleware := &SignatureMiddleware{}

	apiSecret := "test-secret"
	method := "POST"
	path := "/api/v1/payments"
	timestamp := "2025-10-24T12:00:00Z"
	nonce := "abc123"
	body := `{"amount":1000}`

	t.Run("v1 signature calculation", func(t *testing.T) {
		sig := middleware.calculateSignatureByVersion("v1", apiSecret, method, path, timestamp, nonce, body)
		assert.NotEmpty(t, sig)
		// v1不应包含method和path
		expectedV1 := calculateSignature(apiSecret, timestamp, nonce, body)
		assert.Equal(t, expectedV1, sig)
	})

	t.Run("v2 signature calculation", func(t *testing.T) {
		sig := middleware.calculateSignatureByVersion("v2", apiSecret, method, path, timestamp, nonce, body)
		assert.NotEmpty(t, sig)
		// v2应该与v1不同（因为包含method和path）
		v1Sig := middleware.calculateSignatureByVersion("v1", apiSecret, method, path, timestamp, nonce, body)
		assert.NotEqual(t, v1Sig, sig)
	})

	t.Run("unsupported version returns empty", func(t *testing.T) {
		sig := middleware.calculateSignatureByVersion("v99", apiSecret, method, path, timestamp, nonce, body)
		assert.Empty(t, sig)
	})
}

// TestIPWhitelist 测试IP白名单功能
func TestIPWhitelist(t *testing.T) {
	tests := []struct {
		name      string
		clientIP  string
		whitelist string
		expected  bool
	}{
		{
			name:      "empty whitelist allows all",
			clientIP:  "1.2.3.4",
			whitelist: "",
			expected:  true,
		},
		{
			name:      "exact match allows",
			clientIP:  "192.168.1.100",
			whitelist: "192.168.1.100",
			expected:  true,
		},
		{
			name:      "exact match denies different IP",
			clientIP:  "192.168.1.101",
			whitelist: "192.168.1.100",
			expected:  false,
		},
		{
			name:      "multiple IPs comma-separated",
			clientIP:  "192.168.1.101",
			whitelist: "192.168.1.100, 192.168.1.101, 192.168.1.102",
			expected:  true,
		},
		{
			name:      "CIDR range allows",
			clientIP:  "192.168.1.50",
			whitelist: "192.168.1.0/24",
			expected:  true,
		},
		{
			name:      "CIDR range denies different subnet",
			clientIP:  "192.168.2.50",
			whitelist: "192.168.1.0/24",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIPInWhitelist(tt.clientIP, tt.whitelist)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNonceGeneration 测试Nonce生成的随机性
func TestNonceGeneration(t *testing.T) {
	nonces := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		nonce := generateNonce()
		assert.NotEmpty(t, nonce)
		assert.Len(t, nonce, 32) // 16 bytes = 32 hex chars

		// 检查是否有重复
		if nonces[nonce] {
			t.Errorf("Duplicate nonce generated: %s", nonce)
		}
		nonces[nonce] = true
	}

	t.Logf("Generated %d unique nonces", len(nonces))
}

// TestSignRequestV2 测试v2签名请求生成
func TestSignRequestV2(t *testing.T) {
	apiKey := "test-api-key"
	apiSecret := "test-secret"
	method := "POST"
	path := "/api/v1/payments"
	body := `{"amount":1000}`

	headers := SignRequestV2(apiKey, apiSecret, method, path, body)

	assert.Equal(t, apiKey, headers["X-API-Key"])
	assert.NotEmpty(t, headers["X-Signature"])
	assert.NotEmpty(t, headers["X-Timestamp"])
	assert.NotEmpty(t, headers["X-Nonce"])
	assert.Equal(t, "v2", headers["X-Signature-Version"])

	// 验证时间戳格式
	_, err := time.Parse(time.RFC3339, headers["X-Timestamp"])
	assert.NoError(t, err)
}

// TestTimestampValidation 测试时间戳验证
func TestTimestampValidation(t *testing.T) {
	middleware := &SignatureMiddleware{
		timestampWindow: 2 * time.Minute,
	}

	tests := []struct {
		name      string
		timestamp string
		expectErr bool
	}{
		{
			name:      "current timestamp valid",
			timestamp: time.Now().UTC().Format(time.RFC3339),
			expectErr: false,
		},
		{
			name:      "1 minute ago valid",
			timestamp: time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339),
			expectErr: false,
		},
		{
			name:      "3 minutes ago invalid",
			timestamp: time.Now().Add(-3 * time.Minute).UTC().Format(time.RFC3339),
			expectErr: true,
		},
		{
			name:      "invalid format",
			timestamp: "not-a-timestamp",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := middleware.validateTimestamp(tt.timestamp)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSignatureMiddlewareHelpers 测试中间件辅助函数
func TestSignatureMiddlewareHelpers(t *testing.T) {
	t.Run("maskAPIKey", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"short", "***"},
			{"12345678ABC", "12345678..."},
			{"test-api-key-1234567890", "test-api..."},
		}

		for _, tt := range tests {
			result := maskAPIKey(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})
}

// BenchmarkSignatureCalculation 签名计算性能基准测试
func BenchmarkSignatureCalculation(b *testing.B) {
	apiSecret := "test-secret"
	timestamp := time.Now().UTC().Format(time.RFC3339)
	nonce := "test-nonce"
	body := `{"amount":1000}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSignature(apiSecret, timestamp, nonce, body)
	}
}

// BenchmarkNonceGeneration Nonce生成性能基准测试
func BenchmarkNonceGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateNonce()
	}
}
