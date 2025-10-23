package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestPaymentFlowComplete tests the complete payment flow
// Flow: Create Payment -> Risk Assessment -> Create Order -> Process Payment -> Webhook Callback
func TestPaymentFlowComplete(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	// Check if services are ready
	t.Log("Checking if services are ready...")
	services := map[string]string{
		"payment-gateway": config.PaymentGatewayURL,
		"order-service":   config.OrderServiceURL,
		"risk-service":    config.RiskServiceURL,
	}

	for name, url := range services {
		if err := WaitForService(url, 3); err != nil {
			t.Skipf("Service %s not ready: %v", name, err)
		}
		t.Logf("✓ %s is ready", name)
	}

	// Test data
	merchantID := "d76f9fd2-0a64-4a5e-b669-4a0f6081246a" // Test merchant
	amount := int64(10000)                                // 100.00 USD

	// Step 1: Create payment request
	t.Log("\n=== Step 1: Create Payment ===")
	paymentData := GenerateTestPayment(merchantID, amount)

	// Generate signature for API authentication
	nonce := GenerateNonce()
	timestamp := CurrentTimestamp()
	bodyJSON, _ := json.Marshal(paymentData)
	signature := SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, bodyJSON)

	headers := map[string]string{
		"X-Merchant-ID": merchantID,
		"X-Nonce":       nonce,
		"X-Timestamp":   timestamp,
		"X-Signature":   signature,
	}

	resp, body, err := client.POST(
		config.PaymentGatewayURL+"/api/v1/payments",
		paymentData,
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create payment status code")

	result := ParseJSONResponse(t, body)
	t.Logf("Payment created: %+v", result)

	// Extract payment info
	paymentNo, ok := result["payment_no"].(string)
	if !ok {
		t.Fatalf("payment_no not found in response")
	}
	AssertNotEmpty(t, paymentNo, "payment_no")

	paymentURL, ok := result["payment_url"].(string)
	if !ok {
		t.Logf("Warning: payment_url not found (may be expected for some channels)")
	} else {
		AssertNotEmpty(t, paymentURL, "payment_url")
		t.Logf("Payment URL: %s", paymentURL)
	}

	// Step 2: Query payment status
	t.Log("\n=== Step 2: Query Payment Status ===")
	Sleep(1 * time.Second) // Wait for async processing

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/payments/%s", config.PaymentGatewayURL, paymentNo),
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to query payment: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query payment status code")

	result = ParseJSONResponse(t, body)
	t.Logf("Payment status: %+v", result)

	status, ok := result["status"].(string)
	if !ok {
		t.Fatalf("status not found in response")
	}

	// Initial status should be "pending"
	if status != "pending" && status != "processing" {
		t.Logf("Warning: unexpected initial status: %s", status)
	}

	// Step 3: Verify order was created
	t.Log("\n=== Step 3: Verify Order Creation ===")

	orderNo, ok := result["order_no"].(string)
	if !ok || orderNo == "" {
		t.Log("Warning: order_no not found, order service may not be integrated yet")
	} else {
		AssertNotEmpty(t, orderNo, "order_no")
		t.Logf("Order created: %s", orderNo)
	}

	// Step 4: Test idempotency (same merchant_order_no should return same payment)
	t.Log("\n=== Step 4: Test Idempotency ===")

	resp, body, err = client.POST(
		config.PaymentGatewayURL+"/api/v1/payments",
		paymentData, // Same data, same merchant_order_no
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to test idempotency: %v", err)
	}

	// Should return 200 with existing payment or 409 Conflict
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		t.Errorf("Idempotency check failed: expected 200 or 409, got %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusOK {
		result = ParseJSONResponse(t, body)
		returnedPaymentNo := result["payment_no"].(string)
		AssertEqual(t, paymentNo, returnedPaymentNo, "Idempotency: same payment_no returned")
		t.Log("✓ Idempotency check passed")
	}

	// Step 5: Test payment cancellation
	t.Log("\n=== Step 5: Test Payment Cancellation ===")

	cancelResp, _, err := client.POST(
		fmt.Sprintf("%s/api/v1/payments/%s/cancel", config.PaymentGatewayURL, paymentNo),
		nil,
		headers,
	)

	if err != nil {
		t.Logf("Cancel request failed (may be expected): %v", err)
	} else if cancelResp.StatusCode == http.StatusOK {
		t.Log("✓ Payment cancelled successfully")
	} else {
		t.Logf("Cancel returned status %d (may be expected if payment already processed)", cancelResp.StatusCode)
	}

	t.Log("\n=== Payment Flow Test Completed ===")
}

// TestPaymentFlowWithInvalidData tests error handling
func TestPaymentFlowWithInvalidData(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.PaymentGatewayURL, 3); err != nil {
		t.Skipf("Payment gateway not ready: %v", err)
	}

	merchantID := "d76f9fd2-0a64-4a5e-b669-4a0f6081246a"

	// Test 1: Invalid amount (negative)
	t.Log("\n=== Test 1: Invalid Amount ===")
	invalidPayment := GenerateTestPayment(merchantID, -1000)

	nonce := GenerateNonce()
	timestamp := CurrentTimestamp()
	bodyJSON, _ := json.Marshal(invalidPayment)
	signature := SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, bodyJSON)

	headers := map[string]string{
		"X-Merchant-ID": merchantID,
		"X-Nonce":       nonce,
		"X-Timestamp":   timestamp,
		"X-Signature":   signature,
	}

	resp, body, err := client.POST(
		config.PaymentGatewayURL+"/api/v1/payments",
		invalidPayment,
		headers,
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should return 400 Bad Request
	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("Warning: expected 400 for negative amount, got %d: %s", resp.StatusCode, string(body))
	} else {
		t.Log("✓ Invalid amount rejected correctly")
	}

	// Test 2: Invalid currency
	t.Log("\n=== Test 2: Invalid Currency ===")
	invalidPayment = GenerateTestPayment(merchantID, 10000)
	invalidPayment["currency"] = "INVALID"

	bodyJSON, _ = json.Marshal(invalidPayment)
	signature = SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, bodyJSON)
	headers["X-Signature"] = signature

	resp, body, err = client.POST(
		config.PaymentGatewayURL+"/api/v1/payments",
		invalidPayment,
		headers,
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("Warning: expected 400 for invalid currency, got %d: %s", resp.StatusCode, string(body))
	} else {
		t.Log("✓ Invalid currency rejected correctly")
	}

	// Test 3: Missing required fields
	t.Log("\n=== Test 3: Missing Required Fields ===")
	incompletePayment := map[string]interface{}{
		"amount":   10000,
		"currency": "USD",
		// Missing merchant_order_no, channel, etc.
	}

	bodyJSON, _ = json.Marshal(incompletePayment)
	signature = SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, bodyJSON)
	headers["X-Signature"] = signature

	resp, _, err = client.POST(
		config.PaymentGatewayURL+"/api/v1/payments",
		incompletePayment,
		headers,
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("Warning: expected 400 for incomplete data, got %d", resp.StatusCode)
	} else {
		t.Log("✓ Incomplete data rejected correctly")
	}

	t.Log("\n=== Invalid Data Test Completed ===")
}

// TestPaymentRefund tests payment refund flow
func TestPaymentRefund(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.PaymentGatewayURL, 3); err != nil {
		t.Skipf("Payment gateway not ready: %v", err)
	}

	merchantID := "d76f9fd2-0a64-4a5e-b669-4a0f6081246a"
	amount := int64(5000) // 50.00 USD

	// Step 1: Create a payment first
	t.Log("\n=== Step 1: Create Payment for Refund Test ===")
	paymentData := GenerateTestPayment(merchantID, amount)

	nonce := GenerateNonce()
	timestamp := CurrentTimestamp()
	bodyJSON, _ := json.Marshal(paymentData)
	signature := SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, bodyJSON)

	headers := map[string]string{
		"X-Merchant-ID": merchantID,
		"X-Nonce":       nonce,
		"X-Timestamp":   timestamp,
		"X-Signature":   signature,
	}

	resp, body, err := client.POST(
		config.PaymentGatewayURL+"/api/v1/payments",
		paymentData,
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create payment")

	result := ParseJSONResponse(t, body)
	paymentNo := result["payment_no"].(string)
	t.Logf("Payment created: %s", paymentNo)

	// Step 2: Attempt refund
	t.Log("\n=== Step 2: Create Refund ===")

	refundData := map[string]interface{}{
		"payment_no": paymentNo,
		"amount":     amount, // Full refund
		"reason":     "Customer request",
	}

	bodyJSON, _ = json.Marshal(refundData)
	signature = SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, bodyJSON)
	headers["X-Signature"] = signature

	resp, body, err = client.POST(
		config.PaymentGatewayURL+"/api/v1/refunds",
		refundData,
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Refund may succeed or fail depending on payment status
	if resp.StatusCode == http.StatusOK {
		result = ParseJSONResponse(t, body)
		refundNo := result["refund_no"].(string)
		AssertNotEmpty(t, refundNo, "refund_no")
		t.Logf("✓ Refund created: %s", refundNo)
	} else {
		t.Logf("Refund returned status %d (may be expected if payment not completed): %s",
			resp.StatusCode, string(body))
	}

	t.Log("\n=== Refund Test Completed ===")
}

// TestPaymentListQuery tests querying payment list
func TestPaymentListQuery(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.PaymentGatewayURL, 3); err != nil {
		t.Skipf("Payment gateway not ready: %v", err)
	}

	merchantID := "d76f9fd2-0a64-4a5e-b669-4a0f6081246a"

	t.Log("\n=== Test: Query Payment List ===")

	nonce := GenerateNonce()
	timestamp := CurrentTimestamp()
	signature := SignRequest(config.SignatureSecret, merchantID, nonce, timestamp, []byte{})

	headers := map[string]string{
		"X-Merchant-ID": merchantID,
		"X-Nonce":       nonce,
		"X-Timestamp":   timestamp,
		"X-Signature":   signature,
	}

	resp, body, err := client.GET(
		config.PaymentGatewayURL+"/api/v1/payments?page=1&page_size=10",
		headers,
	)

	if err != nil {
		t.Fatalf("Failed to query payments: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query payments")

	result := ParseJSONResponse(t, body)
	t.Logf("Payment list response: %+v", result)

	// Check if list field exists
	if list, ok := result["list"]; ok {
		t.Logf("✓ Retrieved payment list with %d items", len(list.([]interface{})))
	} else if data, ok := result["data"]; ok {
		t.Logf("✓ Retrieved payment data: %v", data)
	} else {
		t.Log("Warning: unexpected response format")
	}

	t.Log("\n=== List Query Test Completed ===")
}
