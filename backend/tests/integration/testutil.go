package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	AdminServiceURL       string
	MerchantServiceURL    string
	PaymentGatewayURL     string
	OrderServiceURL       string
	ChannelAdapterURL     string
	RiskServiceURL        string
	NotificationURL       string
	AccountingURL         string
	AnalyticsURL          string
	ConfigURL             string
	SettlementURL         string
	WithdrawalURL         string
	KYCURL                string
	FeeURL                string
	JWTSecret             string
	SignatureSecret       string
	AdminToken            string
	MerchantToken         string
}

// DefaultTestConfig returns default configuration for tests
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		AdminServiceURL:    "http://localhost:8001",
		MerchantServiceURL: "http://localhost:8002",
		PaymentGatewayURL:  "http://localhost:8003",
		OrderServiceURL:    "http://localhost:8004",
		ChannelAdapterURL:  "http://localhost:8005",
		RiskServiceURL:     "http://localhost:8006",
		NotificationURL:    "http://localhost:8007",
		AccountingURL:      "http://localhost:8008",
		AnalyticsURL:       "http://localhost:8009",
		ConfigURL:          "http://localhost:8010",
		SettlementURL:      "http://localhost:8012",
		WithdrawalURL:      "http://localhost:8013",
		KYCURL:             "http://localhost:8014",
		FeeURL:             "http://localhost:8015",
		JWTSecret:          "default-jwt-secret-change-in-production",
		SignatureSecret:    "default-signature-secret",
	}
}

// HTTPClient wraps http.Client with helper methods
type HTTPClient struct {
	client *http.Client
	config *TestConfig
}

// NewHTTPClient creates a new test HTTP client
func NewHTTPClient(config *TestConfig) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// DoRequest performs an HTTP request and returns response
func (c *HTTPClient) DoRequest(method, url string, body interface{}, headers map[string]string) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("read response: %w", err)
	}

	return resp, respBody, nil
}

// POST performs a POST request
func (c *HTTPClient) POST(url string, body interface{}, headers map[string]string) (*http.Response, []byte, error) {
	return c.DoRequest("POST", url, body, headers)
}

// GET performs a GET request
func (c *HTTPClient) GET(url string, headers map[string]string) (*http.Response, []byte, error) {
	return c.DoRequest("GET", url, nil, headers)
}

// PUT performs a PUT request
func (c *HTTPClient) PUT(url string, body interface{}, headers map[string]string) (*http.Response, []byte, error) {
	return c.DoRequest("PUT", url, body, headers)
}

// DELETE performs a DELETE request
func (c *HTTPClient) DELETE(url string, headers map[string]string) (*http.Response, []byte, error) {
	return c.DoRequest("DELETE", url, nil, headers)
}

// SignRequest generates HMAC-SHA256 signature for API requests
func SignRequest(secret, merchantID, nonce, timestamp string, body []byte) string {
	message := fmt.Sprintf("%s%s%s%s", merchantID, nonce, timestamp, string(body))
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateTestMerchant creates test merchant data
func GenerateTestMerchant() map[string]interface{} {
	return map[string]interface{}{
		"email":         fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		"password":      "Test123456!",
		"business_name": "Test Business Inc",
		"contact_name":  "Test User",
		"phone":         "+8613800138000",
		"status":        "active",
	}
}

// GenerateTestPayment creates test payment data
func GenerateTestPayment(merchantID string, amount int64) map[string]interface{} {
	return map[string]interface{}{
		"merchant_id":     merchantID,
		"merchant_order_no": fmt.Sprintf("TEST%d", time.Now().Unix()),
		"amount":          amount,
		"currency":        "USD",
		"channel":         "stripe",
		"subject":         "Test Payment",
		"description":     "Integration test payment",
		"notify_url":      "https://example.com/notify",
		"return_url":      "https://example.com/return",
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEmpty checks if value is not empty
func AssertNotEmpty(t *testing.T, value string, message string) {
	if value == "" {
		t.Errorf("%s: value is empty", message)
	}
}

// AssertStatusCode checks HTTP status code
func AssertStatusCode(t *testing.T, expected, actual int, message string) {
	if expected != actual {
		t.Errorf("%s: expected status %d, got %d", message, expected, actual)
	}
}

// AssertJSONField checks if JSON response contains expected field value
func AssertJSONField(t *testing.T, data []byte, field string, expected interface{}) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	actual, ok := result[field]
	if !ok {
		t.Errorf("field %s not found in response", field)
		return
	}

	if actual != expected {
		t.Errorf("field %s: expected %v, got %v", field, expected, actual)
	}
}

// ParseJSONResponse parses JSON response into map
func ParseJSONResponse(t *testing.T, data []byte) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	return result
}

// WaitForService waits for a service to be ready
func WaitForService(url string, maxRetries int) error {
	client := &http.Client{Timeout: 2 * time.Second}
	healthURL := url + "/health"

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("service %s not ready after %d retries", url, maxRetries)
}

// CleanupTestData provides cleanup function for test data
type CleanupFunc func()

// TestDataManager manages test data lifecycle
type TestDataManager struct {
	cleanups []CleanupFunc
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager() *TestDataManager {
	return &TestDataManager{
		cleanups: make([]CleanupFunc, 0),
	}
}

// AddCleanup adds a cleanup function
func (m *TestDataManager) AddCleanup(fn CleanupFunc) {
	m.cleanups = append(m.cleanups, fn)
}

// Cleanup runs all cleanup functions
func (m *TestDataManager) Cleanup() {
	for i := len(m.cleanups) - 1; i >= 0; i-- {
		m.cleanups[i]()
	}
}

// GenerateNonce generates a random nonce
func GenerateNonce() string {
	return uuid.New().String()
}

// CurrentTimestamp returns current Unix timestamp as string
func CurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// Sleep waits for specified duration (for async operations)
func Sleep(duration time.Duration) {
	time.Sleep(duration)
}
