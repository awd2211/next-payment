package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestOrderCreation tests order creation
func TestOrderCreation(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.OrderServiceURL, 3); err != nil {
		t.Skipf("Order service not ready: %v", err)
	}

	t.Log("\n=== Test: Order Creation ===")

	merchantID := uuid.New()
	orderData := map[string]interface{}{
		"merchant_id":       merchantID.String(),
		"merchant_order_no": fmt.Sprintf("ORD%d", time.Now().Unix()),
		"amount":            10000, // 100.00 USD
		"currency":          "USD",
		"subject":           "Test Product",
		"description":       "Integration test order",
		"buyer_email":       "buyer@example.com",
	}

	resp, body, err := client.POST(
		config.OrderServiceURL+"/api/v1/orders",
		orderData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create order")

	result := ParseJSONResponse(t, body)
	orderID := result["id"].(string)
	orderNo := result["order_no"].(string)

	AssertNotEmpty(t, orderID, "order_id")
	AssertNotEmpty(t, orderNo, "order_no")
	t.Logf("Order created: %s (%s)", orderID, orderNo)

	t.Log("\n=== Order Creation Test Completed ===")
}

// TestOrderQuery tests order query by ID and order_no
func TestOrderQuery(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.OrderServiceURL, 3); err != nil {
		t.Skipf("Order service not ready: %v", err)
	}

	t.Log("\n=== Test: Order Query ===")

	// Create order first
	merchantID := uuid.New()
	orderData := map[string]interface{}{
		"merchant_id":       merchantID.String(),
		"merchant_order_no": fmt.Sprintf("ORD%d", time.Now().Unix()),
		"amount":            5000,
		"currency":          "USD",
		"subject":           "Query Test Product",
	}

	resp, body, err := client.POST(
		config.OrderServiceURL+"/api/v1/orders",
		orderData,
		nil,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Order creation failed, skipping query test")
	}

	result := ParseJSONResponse(t, body)
	orderID := result["id"].(string)
	orderNo := result["order_no"].(string)

	// Query by order ID
	t.Log("\n--- Query by Order ID ---")
	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/orders/%s", config.OrderServiceURL, orderID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query order by ID: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query order by ID")

	result = ParseJSONResponse(t, body)
	queriedOrderNo := result["order_no"].(string)
	AssertEqual(t, orderNo, queriedOrderNo, "order_no")
	t.Log("✓ Query by ID successful")

	// Query by order_no
	t.Log("\n--- Query by Order No ---")
	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/orders/by-no/%s", config.OrderServiceURL, orderNo),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query order by order_no: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result = ParseJSONResponse(t, body)
		queriedOrderID := result["id"].(string)
		AssertEqual(t, orderID, queriedOrderID, "order_id")
		t.Log("✓ Query by order_no successful")
	} else {
		t.Logf("Query by order_no returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Order Query Test Completed ===")
}

// TestOrderStatusUpdate tests order status transitions
func TestOrderStatusUpdate(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.OrderServiceURL, 3); err != nil {
		t.Skipf("Order service not ready: %v", err)
	}

	t.Log("\n=== Test: Order Status Update ===")

	// Create order
	merchantID := uuid.New()
	orderData := map[string]interface{}{
		"merchant_id":       merchantID.String(),
		"merchant_order_no": fmt.Sprintf("ORD%d", time.Now().Unix()),
		"amount":            10000,
		"currency":          "USD",
		"subject":           "Status Test Product",
	}

	resp, body, err := client.POST(
		config.OrderServiceURL+"/api/v1/orders",
		orderData,
		nil,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Order creation failed, skipping status update test")
	}

	result := ParseJSONResponse(t, body)
	orderID := result["id"].(string)

	// Update to processing
	t.Log("\n--- Update to Processing ---")
	updateData := map[string]interface{}{
		"status": "processing",
	}

	resp, _, err = client.PUT(
		fmt.Sprintf("%s/api/v1/orders/%s/status", config.OrderServiceURL, orderID),
		updateData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to update order status: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Update status to processing")
	t.Log("✓ Status updated to processing")

	// Verify status change
	Sleep(500 * time.Millisecond)

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/orders/%s", config.OrderServiceURL, orderID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query order: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	AssertEqual(t, "processing", status, "Updated order status")

	// Update to completed
	t.Log("\n--- Update to Completed ---")
	updateData["status"] = "completed"

	resp, _, err = client.PUT(
		fmt.Sprintf("%s/api/v1/orders/%s/status", config.OrderServiceURL, orderID),
		updateData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to update to completed: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Update status to completed")
	t.Log("✓ Status updated to completed")

	t.Log("\n=== Order Status Test Completed ===")
}

// TestOrderList tests order list query
func TestOrderList(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.OrderServiceURL, 3); err != nil {
		t.Skipf("Order service not ready: %v", err)
	}

	t.Log("\n=== Test: Order List Query ===")

	resp, body, err := client.GET(
		config.OrderServiceURL+"/api/v1/orders?page=1&page_size=10",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query order list: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query order list")

	result := ParseJSONResponse(t, body)
	t.Logf("Order list response: %+v", result)

	if list, ok := result["list"].([]interface{}); ok {
		t.Logf("✓ Retrieved %d orders", len(list))
	} else if data, ok := result["data"].([]interface{}); ok {
		t.Logf("✓ Retrieved %d orders", len(data))
	}

	t.Log("\n=== Order List Test Completed ===")
}

// TestRiskAssessment tests risk assessment for transactions
func TestRiskAssessment(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.RiskServiceURL, 3); err != nil {
		t.Skipf("Risk service not ready: %v", err)
	}

	t.Log("\n=== Test: Risk Assessment ===")

	// Low risk transaction
	t.Log("\n--- Assessing Low Risk Transaction ---")
	riskData := map[string]interface{}{
		"merchant_id":  uuid.New().String(),
		"amount":       1000, // 10.00 USD
		"currency":     "USD",
		"user_id":      uuid.New().String(),
		"ip_address":   "192.168.1.100",
		"device_id":    "device123",
		"country":      "US",
	}

	resp, body, err := client.POST(
		config.RiskServiceURL+"/api/v1/risk/assess",
		riskData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to assess risk: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Risk assessment")

	result := ParseJSONResponse(t, body)
	riskLevel := result["risk_level"].(string)
	AssertNotEmpty(t, riskLevel, "risk_level")
	t.Logf("Risk level: %s", riskLevel)

	if score, ok := result["risk_score"].(float64); ok {
		t.Logf("Risk score: %.2f", score)
	}

	// High risk transaction (large amount)
	t.Log("\n--- Assessing High Risk Transaction ---")
	riskData["amount"] = 100000000 // 1,000,000 USD
	riskData["country"] = "XX"     // Unknown country

	resp, body, err = client.POST(
		config.RiskServiceURL+"/api/v1/risk/assess",
		riskData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to assess high risk: %v", err)
	}

	result = ParseJSONResponse(t, body)
	highRiskLevel := result["risk_level"].(string)
	t.Logf("High risk level: %s", highRiskLevel)

	if highRiskLevel == "high" || highRiskLevel == "block" {
		t.Log("✓ High risk transaction detected correctly")
	}

	t.Log("\n=== Risk Assessment Test Completed ===")
}

// TestRiskRuleManagement tests risk rule CRUD
func TestRiskRuleManagement(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.RiskServiceURL, 3); err != nil {
		t.Skipf("Risk service not ready: %v", err)
	}

	t.Log("\n=== Test: Risk Rule Management ===")

	// Create risk rule
	ruleData := map[string]interface{}{
		"name":        "High Amount Rule",
		"description": "Block transactions over 100,000",
		"rule_type":   "amount",
		"conditions": map[string]interface{}{
			"max_amount": 10000000, // 100,000 USD
		},
		"action":   "block",
		"priority": 1,
		"enabled":  true,
	}

	resp, body, err := client.POST(
		config.RiskServiceURL+"/api/v1/risk/rules",
		ruleData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create risk rule: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result := ParseJSONResponse(t, body)
		if ruleID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, ruleID, "rule_id")
			t.Logf("Risk rule created: %s", ruleID)

			// Query rule
			resp, body, err = client.GET(
				fmt.Sprintf("%s/api/v1/risk/rules/%s", config.RiskServiceURL, ruleID),
				nil,
			)

			if err != nil {
				t.Fatalf("Failed to query risk rule: %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				result = ParseJSONResponse(t, body)
				ruleName := result["name"].(string)
				AssertEqual(t, "High Amount Rule", ruleName, "rule name")
				t.Log("✓ Risk rule verified")
			}
		}
	} else {
		t.Logf("Create risk rule returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Risk Rule Test Completed ===")
}

// TestBlacklistManagement tests blacklist add/remove
func TestBlacklistManagement(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.RiskServiceURL, 3); err != nil {
		t.Skipf("Risk service not ready: %v", err)
	}

	t.Log("\n=== Test: Blacklist Management ===")

	// Add to blacklist
	blacklistData := map[string]interface{}{
		"type":   "ip",
		"value":  "192.168.1.666",
		"reason": "Fraudulent activity",
	}

	resp, body, err := client.POST(
		config.RiskServiceURL+"/api/v1/risk/blacklist",
		blacklistData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to add to blacklist: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result := ParseJSONResponse(t, body)
		if blacklistID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, blacklistID, "blacklist_id")
			t.Logf("✓ Added to blacklist: %s", blacklistID)

			// Remove from blacklist
			resp, _, err = client.DELETE(
				fmt.Sprintf("%s/api/v1/risk/blacklist/%s", config.RiskServiceURL, blacklistID),
				nil,
			)

			if err != nil {
				t.Fatalf("Failed to remove from blacklist: %v", err)
			}

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				t.Log("✓ Removed from blacklist")
			}
		}
	} else {
		t.Logf("Add to blacklist returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Blacklist Test Completed ===")
}

// TestRiskStatistics tests risk statistics
func TestRiskStatistics(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.RiskServiceURL, 3); err != nil {
		t.Skipf("Risk service not ready: %v", err)
	}

	t.Log("\n=== Test: Risk Statistics ===")

	resp, body, err := client.GET(
		config.RiskServiceURL+"/api/v1/risk/statistics",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to get risk statistics: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		t.Logf("Risk Statistics: %+v", result)

		expectedFields := []string{"total_assessments", "blocked_count", "high_risk_count"}
		for _, field := range expectedFields {
			if _, ok := result[field]; ok {
				t.Logf("✓ Field %s present", field)
			}
		}
	} else {
		t.Logf("Risk statistics returned status %d", resp.StatusCode)
	}

	t.Log("\n=== Risk Statistics Test Completed ===")
}

// TestOrderRefund tests order refund creation
func TestOrderRefund(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.OrderServiceURL, 3); err != nil {
		t.Skipf("Order service not ready: %v", err)
	}

	t.Log("\n=== Test: Order Refund ===")

	// Create order first
	merchantID := uuid.New()
	orderData := map[string]interface{}{
		"merchant_id":       merchantID.String(),
		"merchant_order_no": fmt.Sprintf("ORD%d", time.Now().Unix()),
		"amount":            5000,
		"currency":          "USD",
		"subject":           "Refund Test Product",
	}

	resp, body, err := client.POST(
		config.OrderServiceURL+"/api/v1/orders",
		orderData,
		nil,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Order creation failed, skipping refund test")
	}

	result := ParseJSONResponse(t, body)
	orderID := result["id"].(string)

	// Create refund
	refundData := map[string]interface{}{
		"order_id": orderID,
		"amount":   5000, // Full refund
		"reason":   "Customer request",
	}

	resp, body, err = client.POST(
		config.OrderServiceURL+"/api/v1/refunds",
		refundData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result = ParseJSONResponse(t, body)
		if refundID, ok := result["id"].(string); ok {
			AssertNotEmpty(t, refundID, "refund_id")
			t.Logf("✓ Refund created: %s", refundID)
		}
	} else {
		t.Logf("Create refund returned status %d: %s", resp.StatusCode, string(body))
	}

	t.Log("\n=== Refund Test Completed ===")
}
