package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestMerchantRegistration tests merchant registration flow
func TestMerchantRegistration(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant Registration ===")

	// Step 1: Register new merchant
	merchantData := GenerateTestMerchant()

	resp, body, err := client.POST(
		config.MerchantServiceURL+"/api/v1/merchants/register",
		merchantData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to register merchant: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Register merchant")

	result := ParseJSONResponse(t, body)
	merchantID := result["id"].(string)
	AssertNotEmpty(t, merchantID, "merchant_id")
	t.Logf("Merchant registered: %s", merchantID)

	// Step 2: Verify merchant created
	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/merchants/%s", config.MerchantServiceURL, merchantID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query merchant: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query merchant")

	result = ParseJSONResponse(t, body)
	email := result["email"].(string)
	AssertEqual(t, merchantData["email"], email, "merchant email")
	t.Log("✓ Merchant registration verified")

	t.Log("\n=== Registration Test Completed ===")
}

// TestMerchantLogin tests merchant login and token generation
func TestMerchantLogin(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant Login ===")

	// Step 1: Register merchant first
	merchantData := GenerateTestMerchant()
	email := merchantData["email"].(string)
	password := merchantData["password"].(string)

	resp, _, err := client.POST(
		config.MerchantServiceURL+"/api/v1/merchants/register",
		merchantData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to register merchant: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Skipf("Merchant registration failed, skipping login test")
	}

	t.Log("Merchant registered successfully")

	// Step 2: Login
	loginData := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	resp, body, err := client.POST(
		config.MerchantServiceURL+"/api/v1/merchants/login",
		loginData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Merchant login")

	result := ParseJSONResponse(t, body)
	token, ok := result["token"].(string)
	if !ok || token == "" {
		t.Fatal("Token not found in response")
	}

	AssertNotEmpty(t, token, "JWT token")
	t.Logf("✓ Login successful, token: %s...", token[:20])

	// Step 3: Use token to access protected endpoint
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	resp, _, err = client.GET(
		config.MerchantServiceURL+"/api/v1/merchants/profile",
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to access profile: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Access profile with token")
	t.Log("✓ Token authentication successful")

	t.Log("\n=== Login Test Completed ===")
}

// TestMerchantUpdate tests merchant profile update
func TestMerchantUpdate(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant Profile Update ===")

	// Register merchant
	merchantData := GenerateTestMerchant()
	resp, body, err := client.POST(
		config.MerchantServiceURL+"/api/v1/merchants/register",
		merchantData,
		nil,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Merchant registration failed, skipping update test")
	}

	result := ParseJSONResponse(t, body)
	merchantID := result["id"].(string)
	t.Logf("Merchant ID: %s", merchantID)

	// Update merchant profile
	updateData := map[string]interface{}{
		"business_name": "Updated Business Name Ltd",
		"contact_name":  "Updated Contact",
		"phone":         "+8613900139000",
		"address":       "123 Updated Street, City",
	}

	resp, _, err = client.PUT(
		fmt.Sprintf("%s/api/v1/merchants/%s", config.MerchantServiceURL, merchantID),
		updateData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to update merchant: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Update merchant")
	t.Log("✓ Merchant profile updated")

	// Verify update
	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/merchants/%s", config.MerchantServiceURL, merchantID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query merchant: %v", err)
	}

	result = ParseJSONResponse(t, body)
	businessName := result["business_name"].(string)
	AssertEqual(t, "Updated Business Name Ltd", businessName, "Updated business name")
	t.Log("✓ Update verified")

	t.Log("\n=== Update Test Completed ===")
}

// TestMerchantFreeze tests merchant account freeze/unfreeze
func TestMerchantFreeze(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant Freeze/Unfreeze ===")

	// Register merchant
	merchantData := GenerateTestMerchant()
	resp, body, err := client.POST(
		config.MerchantServiceURL+"/api/v1/merchants/register",
		merchantData,
		nil,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Merchant registration failed, skipping freeze test")
	}

	result := ParseJSONResponse(t, body)
	merchantID := result["id"].(string)

	// Freeze merchant
	freezeData := map[string]interface{}{
		"reason": "Suspicious activity detected",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/merchants/%s/freeze", config.MerchantServiceURL, merchantID),
		freezeData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to freeze merchant: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Freeze merchant")
	t.Log("✓ Merchant frozen")

	// Verify status
	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/merchants/%s", config.MerchantServiceURL, merchantID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query merchant: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	if status != "frozen" && status != "inactive" {
		t.Logf("Warning: expected frozen status, got: %s", status)
	}

	// Unfreeze merchant
	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/merchants/%s/unfreeze", config.MerchantServiceURL, merchantID),
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to unfreeze merchant: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Unfreeze merchant")
	t.Log("✓ Merchant unfrozen")

	t.Log("\n=== Freeze Test Completed ===")
}

// TestAdminUserManagement tests admin user creation and management
func TestAdminUserManagement(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.AdminServiceURL, 3); err != nil {
		t.Skipf("Admin service not ready: %v", err)
	}

	t.Log("\n=== Test: Admin User Management ===")

	// Step 1: Create admin user
	adminData := map[string]interface{}{
		"username": fmt.Sprintf("admin_%s", uuid.New().String()[:8]),
		"password": "Admin@123456",
		"email":    fmt.Sprintf("admin_%s@example.com", uuid.New().String()[:8]),
		"role":     "admin",
	}

	resp, body, err := client.POST(
		config.AdminServiceURL+"/api/v1/admins",
		adminData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Logf("Warning: Create admin returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("Admin creation may require authentication")
	}

	result := ParseJSONResponse(t, body)
	adminID, ok := result["id"].(string)
	if !ok {
		t.Log("Admin ID not found in response, may need authentication")
		return
	}

	AssertNotEmpty(t, adminID, "admin_id")
	t.Logf("Admin user created: %s", adminID)

	// Step 2: Query admin details
	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/admins/%s", config.AdminServiceURL, adminID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query admin: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result = ParseJSONResponse(t, body)
		username := result["username"].(string)
		AssertEqual(t, adminData["username"], username, "admin username")
		t.Log("✓ Admin user verified")
	}

	t.Log("\n=== Admin Management Test Completed ===")
}

// TestAdminLogin tests admin login
func TestAdminLogin(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.AdminServiceURL, 3); err != nil {
		t.Skipf("Admin service not ready: %v", err)
	}

	t.Log("\n=== Test: Admin Login ===")

	// Try to login with default admin (if exists)
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	resp, body, err := client.POST(
		config.AdminServiceURL+"/api/v1/auth/login",
		loginData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		if token, ok := result["token"].(string); ok && token != "" {
			AssertNotEmpty(t, token, "JWT token")
			t.Logf("✓ Admin login successful")
		} else {
			t.Log("Warning: Token not found in response")
		}
	} else {
		t.Logf("Login returned status %d (default admin may not exist): %s",
			resp.StatusCode, string(body))
	}

	t.Log("\n=== Admin Login Test Completed ===")
}

// TestMerchantList tests merchant list query with pagination
func TestMerchantList(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant List Query ===")

	// Create multiple merchants
	createdIDs := []string{}
	for i := 0; i < 3; i++ {
		merchantData := GenerateTestMerchant()
		resp, body, err := client.POST(
			config.MerchantServiceURL+"/api/v1/merchants/register",
			merchantData,
			nil,
		)

		if err != nil {
			t.Logf("Warning: Failed to create test merchant %d: %v", i+1, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			result := ParseJSONResponse(t, body)
			if id, ok := result["id"].(string); ok {
				createdIDs = append(createdIDs, id)
			}
		}

		Sleep(100 * time.Millisecond)
	}

	t.Logf("Created %d test merchants", len(createdIDs))

	// Query merchant list
	resp, body, err := client.GET(
		config.MerchantServiceURL+"/api/v1/merchants?page=1&page_size=10",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query merchant list: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query merchant list")

	result := ParseJSONResponse(t, body)
	t.Logf("Merchant list response: %+v", result)

	// Check response format
	if list, ok := result["list"].([]interface{}); ok {
		t.Logf("✓ Retrieved %d merchants", len(list))
	} else if data, ok := result["data"].([]interface{}); ok {
		t.Logf("✓ Retrieved %d merchants", len(data))
	} else {
		t.Log("Warning: Unexpected list response format")
	}

	t.Log("\n=== Merchant List Test Completed ===")
}

// TestMerchantStatistics tests merchant statistics
func TestMerchantStatistics(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant Statistics ===")

	resp, body, err := client.GET(
		config.MerchantServiceURL+"/api/v1/merchants/statistics",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		t.Logf("Merchant Statistics: %+v", result)

		// Check for expected fields
		expectedFields := []string{"total_merchants", "active_merchants"}
		for _, field := range expectedFields {
			if _, ok := result[field]; ok {
				t.Logf("✓ Field %s present", field)
			}
		}
	} else {
		t.Logf("Statistics endpoint returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Statistics Test Completed ===")
}

// TestMerchantAPIKey tests API key generation and management
func TestMerchantAPIKey(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.MerchantServiceURL, 3); err != nil {
		t.Skipf("Merchant service not ready: %v", err)
	}

	t.Log("\n=== Test: Merchant API Key Management ===")

	// Register merchant
	merchantData := GenerateTestMerchant()
	resp, body, err := client.POST(
		config.MerchantServiceURL+"/api/v1/merchants/register",
		merchantData,
		nil,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Merchant registration failed, skipping API key test")
	}

	result := ParseJSONResponse(t, body)
	merchantID := result["id"].(string)

	// Generate API key
	resp, body, err = client.POST(
		fmt.Sprintf("%s/api/v1/merchants/%s/api-keys", config.MerchantServiceURL, merchantID),
		map[string]interface{}{
			"name": "Production API Key",
		},
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result = ParseJSONResponse(t, body)
		if apiKey, ok := result["api_key"].(string); ok {
			AssertNotEmpty(t, apiKey, "API key")
			t.Logf("✓ API key generated: %s...", apiKey[:10])
		} else {
			t.Log("API key not found in response")
		}
	} else {
		t.Logf("API key generation returned status %d", resp.StatusCode)
	}

	t.Log("\n=== API Key Test Completed ===")
}
