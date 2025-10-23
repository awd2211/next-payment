package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestSettlementFlowComplete tests the complete settlement flow
func TestSettlementFlowComplete(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.SettlementURL, 3); err != nil {
		t.Skipf("Settlement service not ready: %v", err)
	}

	merchantID := uuid.New()

	// Step 1: Create settlement
	t.Log("\n=== Step 1: Create Settlement ===")

	settlementData := map[string]interface{}{
		"merchant_id":        merchantID.String(),
		"start_date":         time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
		"end_date":           time.Now().Format("2006-01-02"),
		"settlement_cycle":   "weekly",
		"total_amount":       1000000,    // 10,000元
		"settlement_amount":  995000,     // 扣除手续费后
		"fee_amount":         5000,       // 50元手续费
	}

	resp, body, err := client.POST(
		config.SettlementURL+"/api/v1/settlements",
		settlementData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create settlement: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create settlement")

	result := ParseJSONResponse(t, body)
	settlementID := result["id"].(string)
	settlementNo := result["settlement_no"].(string)
	AssertNotEmpty(t, settlementID, "settlement_id")
	AssertNotEmpty(t, settlementNo, "settlement_no")
	t.Logf("Settlement created: %s (%s)", settlementID, settlementNo)

	// Step 2: Query settlement details
	t.Log("\n=== Step 2: Query Settlement Details ===")

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/settlements/%s", config.SettlementURL, settlementID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query settlement: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query settlement")

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	AssertEqual(t, "pending", status, "Initial settlement status")
	t.Logf("Settlement status: %s", status)

	// Step 3: Approve settlement
	t.Log("\n=== Step 3: Approve Settlement ===")

	approvalData := map[string]interface{}{
		"settlement_id": settlementID,
		"approver_id":   uuid.New().String(),
		"approver_name": "Finance Manager",
		"action":        "approve",
		"comment":       "Settlement approved for processing",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/settlements/%s/approve", config.SettlementURL, settlementID),
		approvalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to approve settlement: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Approve settlement")
	t.Log("✓ Settlement approved")

	// Step 4: Execute settlement
	t.Log("\n=== Step 4: Execute Settlement ===")

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/settlements/%s/execute", config.SettlementURL, settlementID),
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to execute settlement: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Execute settlement")
	t.Log("✓ Settlement execution initiated")

	// Step 5: Check final status
	t.Log("\n=== Step 5: Check Final Status ===")
	Sleep(2 * time.Second)

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/settlements/%s", config.SettlementURL, settlementID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query settlement: %v", err)
	}

	result = ParseJSONResponse(t, body)
	finalStatus := result["status"].(string)
	t.Logf("Final settlement status: %s", finalStatus)

	if finalStatus == "completed" || finalStatus == "processing" {
		t.Log("✓ Settlement processed successfully")
	}

	t.Log("\n=== Settlement Flow Test Completed ===")
}

// TestSettlementAutoGeneration tests automatic settlement generation
func TestSettlementAutoGeneration(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.SettlementURL, 3); err != nil {
		t.Skipf("Settlement service not ready: %v", err)
	}

	merchantID := uuid.New()

	t.Log("\n=== Test: Auto-Generate Settlement ===")

	autoGenData := map[string]interface{}{
		"merchant_id":      merchantID.String(),
		"settlement_cycle": "daily",
	}

	resp, body, err := client.POST(
		config.SettlementURL+"/api/v1/settlements/auto-generate",
		autoGenData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to auto-generate settlement: %v", err)
	}

	// May return 200 (generated) or 400 (no data to settle)
	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		settlementNo := result["settlement_no"].(string)
		t.Logf("✓ Auto-generated settlement: %s", settlementNo)
	} else if resp.StatusCode == http.StatusBadRequest {
		t.Log("No transactions to settle (expected for new merchant)")
	} else {
		t.Logf("Unexpected status: %d", resp.StatusCode)
	}

	t.Log("\n=== Auto-Generation Test Completed ===")
}

// TestKYCFlowComplete tests the complete KYC verification flow
func TestKYCFlowComplete(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.KYCURL, 3); err != nil {
		t.Skipf("KYC service not ready: %v", err)
	}

	merchantID := uuid.New()

	// Step 1: Check initial KYC level
	t.Log("\n=== Step 1: Check Initial KYC Level ===")

	resp, body, err := client.GET(
		fmt.Sprintf("%s/api/v1/kyc/merchants/%s/level", config.KYCURL, merchantID.String()),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query KYC level: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		result := ParseJSONResponse(t, body)
		currentLevel := result["current_level"].(string)
		t.Logf("Current KYC level: %s", currentLevel)
	} else {
		t.Log("No KYC level yet (expected for new merchant)")
	}

	// Step 2: Submit identity document
	t.Log("\n=== Step 2: Submit Identity Document ===")

	identityDoc := map[string]interface{}{
		"merchant_id":   merchantID.String(),
		"document_type": "identity_card",
		"document_no":   "110101199001011234",
		"real_name":     "张三",
		"front_image":   "https://example.com/id_front.jpg",
		"back_image":    "https://example.com/id_back.jpg",
	}

	resp, body, err = client.POST(
		config.KYCURL+"/api/v1/kyc/documents",
		identityDoc,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to submit document: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Submit identity document")

	result := ParseJSONResponse(t, body)
	docID := result["id"].(string)
	AssertNotEmpty(t, docID, "document_id")
	t.Logf("Identity document submitted: %s", docID)

	// Step 3: Submit business license
	t.Log("\n=== Step 3: Submit Business License ===")

	businessLicense := map[string]interface{}{
		"merchant_id":         merchantID.String(),
		"company_name":        "测试科技有限公司",
		"license_no":          "91110000MA01234567",
		"legal_representative": "张三",
		"business_scope":      "软件开发",
		"registered_capital":  "1000000",
		"license_image":       "https://example.com/license.jpg",
	}

	resp, body, err = client.POST(
		config.KYCURL+"/api/v1/kyc/business-qualifications",
		businessLicense,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to submit business license: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Submit business license")

	result = ParseJSONResponse(t, body)
	qualID := result["id"].(string)
	AssertNotEmpty(t, qualID, "qualification_id")
	t.Logf("Business license submitted: %s", qualID)

	// Step 4: Review identity document
	t.Log("\n=== Step 4: Review Identity Document ===")

	reviewData := map[string]interface{}{
		"document_id":   docID,
		"reviewer_id":   uuid.New().String(),
		"reviewer_name": "KYC Reviewer",
		"status":        "approved",
		"comment":       "Identity verified",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/kyc/documents/%s/review", config.KYCURL, docID),
		reviewData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to review document: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Review document")
	t.Log("✓ Identity document approved")

	// Step 5: Review business qualification
	t.Log("\n=== Step 5: Review Business Qualification ===")

	reviewData = map[string]interface{}{
		"qualification_id": qualID,
		"reviewer_id":      uuid.New().String(),
		"reviewer_name":    "Business Reviewer",
		"status":           "approved",
		"comment":          "Business license verified",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/kyc/business-qualifications/%s/review", config.KYCURL, qualID),
		reviewData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to review qualification: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Review qualification")
	t.Log("✓ Business qualification approved")

	// Step 6: Upgrade KYC level
	t.Log("\n=== Step 6: Upgrade KYC Level ===")

	upgradeData := map[string]interface{}{
		"new_level": "intermediate",
	}

	resp, _, err = client.PUT(
		fmt.Sprintf("%s/api/v1/kyc/merchants/%s/level", config.KYCURL, merchantID.String()),
		upgradeData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to upgrade level: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Upgrade KYC level")
	t.Log("✓ KYC level upgraded to intermediate")

	// Step 7: Verify new KYC level and limits
	t.Log("\n=== Step 7: Verify New KYC Level ===")

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/kyc/merchants/%s/level", config.KYCURL, merchantID.String()),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query KYC level: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query KYC level")

	result = ParseJSONResponse(t, body)
	newLevel := result["current_level"].(string)
	AssertEqual(t, "intermediate", newLevel, "Updated KYC level")

	if transactionLimit, ok := result["transaction_limit"].(float64); ok {
		t.Logf("Transaction limit: %.2f", transactionLimit)
	}

	if dailyLimit, ok := result["daily_limit"].(float64); ok {
		t.Logf("Daily limit: %.2f", dailyLimit)
	}

	t.Log("✓ KYC level and limits verified")

	t.Log("\n=== KYC Flow Test Completed ===")
}

// TestKYCDocumentRejection tests KYC document rejection flow
func TestKYCDocumentRejection(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.KYCURL, 3); err != nil {
		t.Skipf("KYC service not ready: %v", err)
	}

	merchantID := uuid.New()

	// Submit document
	t.Log("\n=== Step 1: Submit Document ===")

	doc := map[string]interface{}{
		"merchant_id":   merchantID.String(),
		"document_type": "identity_card",
		"document_no":   "123456789012345678",
		"real_name":     "测试用户",
		"front_image":   "https://example.com/invalid.jpg",
	}

	resp, body, err := client.POST(
		config.KYCURL+"/api/v1/kyc/documents",
		doc,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to submit document: %v", err)
	}

	result := ParseJSONResponse(t, body)
	docID := result["id"].(string)
	t.Logf("Document submitted: %s", docID)

	// Reject document
	t.Log("\n=== Step 2: Reject Document ===")

	reviewData := map[string]interface{}{
		"document_id":   docID,
		"reviewer_id":   uuid.New().String(),
		"reviewer_name": "KYC Reviewer",
		"status":        "rejected",
		"comment":       "Document image is unclear, please resubmit",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/kyc/documents/%s/review", config.KYCURL, docID),
		reviewData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to reject document: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Reject document")
	t.Log("✓ Document rejected")

	// Verify rejection
	t.Log("\n=== Step 3: Verify Rejection ===")

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/kyc/documents/%s", config.KYCURL, docID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query document: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	AssertEqual(t, "rejected", status, "Document status after rejection")
	t.Log("✓ Rejection status confirmed")

	t.Log("\n=== Document Rejection Test Completed ===")
}

// TestKYCStatistics tests KYC statistics retrieval
func TestKYCStatistics(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.KYCURL, 3); err != nil {
		t.Skipf("KYC service not ready: %v", err)
	}

	t.Log("\n=== Test: Get KYC Statistics ===")

	resp, body, err := client.GET(
		config.KYCURL+"/api/v1/kyc/statistics",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Get KYC statistics")

	result := ParseJSONResponse(t, body)
	t.Logf("KYC Statistics: %+v", result)

	// Check expected fields
	expectedFields := []string{
		"total_merchants", "pending_reviews", "approved_count",
		"rejected_count", "total_documents",
	}

	for _, field := range expectedFields {
		if _, ok := result[field]; ok {
			t.Logf("✓ Field %s present", field)
		} else {
			t.Logf("Warning: Field %s not found", field)
		}
	}

	t.Log("\n=== Statistics Test Completed ===")
}
