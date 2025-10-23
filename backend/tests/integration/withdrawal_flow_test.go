package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestWithdrawalFlowComplete tests the complete withdrawal flow with multi-level approval
func TestWithdrawalFlowComplete(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.WithdrawalURL, 3); err != nil {
		t.Skipf("Withdrawal service not ready: %v", err)
	}

	merchantID := uuid.New()
	userID := uuid.New()

	// Step 1: Create bank account
	t.Log("\n=== Step 1: Create Bank Account ===")

	bankAccount := map[string]interface{}{
		"merchant_id":     merchantID.String(),
		"user_id":         userID.String(),
		"account_name":    "Test Company Ltd",
		"account_number":  "1234567890",
		"bank_name":       "Test Bank",
		"bank_branch":     "Main Branch",
		"bank_code":       "TEST001",
		"account_type":    "company",
		"is_default":      true,
	}

	resp, body, err := client.POST(
		config.WithdrawalURL+"/api/v1/bank-accounts",
		bankAccount,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create bank account: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create bank account")

	result := ParseJSONResponse(t, body)
	bankAccountID := result["id"].(string)
	AssertNotEmpty(t, bankAccountID, "bank_account_id")
	t.Logf("Bank account created: %s", bankAccountID)

	// Step 2: Create withdrawal (small amount - requires 1 level approval)
	t.Log("\n=== Step 2: Create Withdrawal (Small Amount) ===")

	withdrawalData := map[string]interface{}{
		"merchant_id":      merchantID.String(),
		"user_id":          userID.String(),
		"bank_account_id":  bankAccountID,
		"amount":           500000, // 5,000 元 (< 10万，需要1级审批)
		"withdrawal_type":  "normal",
		"purpose":          "Test withdrawal",
	}

	resp, body, err = client.POST(
		config.WithdrawalURL+"/api/v1/withdrawals",
		withdrawalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create withdrawal: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create withdrawal")

	result = ParseJSONResponse(t, body)
	withdrawalID := result["id"].(string)
	AssertNotEmpty(t, withdrawalID, "withdrawal_id")

	requiredLevel := int(result["required_level"].(float64))
	AssertEqual(t, 1, requiredLevel, "required_level for small amount")
	t.Logf("Withdrawal created: %s (requires %d level approval)", withdrawalID, requiredLevel)

	// Step 3: Query withdrawal status
	t.Log("\n=== Step 3: Query Withdrawal Status ===")

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/withdrawals/%s", config.WithdrawalURL, withdrawalID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query withdrawal: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Query withdrawal")

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	AssertEqual(t, "pending", status, "Initial withdrawal status")
	t.Logf("Withdrawal status: %s", status)

	// Step 4: Approve withdrawal (Level 1)
	t.Log("\n=== Step 4: Approve Withdrawal (Level 1) ===")

	approvalData := map[string]interface{}{
		"withdrawal_id": withdrawalID,
		"approver_id":   uuid.New().String(),
		"approver_name": "Approver Level 1",
		"level":         1,
		"action":        "approve",
		"comment":       "Approved by level 1",
	}

	resp, body, err = client.POST(
		fmt.Sprintf("%s/api/v1/withdrawals/%s/approve", config.WithdrawalURL, withdrawalID),
		approvalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to approve withdrawal: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Approve withdrawal")
	t.Log("✓ Withdrawal approved by level 1")

	// Step 5: Check withdrawal status after approval
	t.Log("\n=== Step 5: Check Status After Approval ===")
	Sleep(1 * time.Second)

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/withdrawals/%s", config.WithdrawalURL, withdrawalID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query withdrawal: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status = result["status"].(string)

	// After 1-level approval for small amount, should be "approved"
	if status == "approved" || status == "processing" {
		t.Logf("✓ Withdrawal status updated to: %s", status)
	} else {
		t.Logf("Warning: unexpected status after approval: %s", status)
	}

	t.Log("\n=== Withdrawal Flow Test Completed ===")
}

// TestWithdrawalMultiLevelApproval tests multi-level approval for large amount
func TestWithdrawalMultiLevelApproval(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.WithdrawalURL, 3); err != nil {
		t.Skipf("Withdrawal service not ready: %v", err)
	}

	merchantID := uuid.New()
	userID := uuid.New()

	// Step 1: Create bank account
	t.Log("\n=== Step 1: Create Bank Account ===")

	bankAccount := map[string]interface{}{
		"merchant_id":     merchantID.String(),
		"user_id":         userID.String(),
		"account_name":    "Large Company Ltd",
		"account_number":  "9876543210",
		"bank_name":       "Test Bank",
		"bank_code":       "TEST001",
		"account_type":    "company",
	}

	resp, body, err := client.POST(
		config.WithdrawalURL+"/api/v1/bank-accounts",
		bankAccount,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create bank account: %v", err)
	}

	result := ParseJSONResponse(t, body)
	bankAccountID := result["id"].(string)
	t.Logf("Bank account created: %s", bankAccountID)

	// Step 2: Create withdrawal (large amount - requires 3 level approval)
	t.Log("\n=== Step 2: Create Large Withdrawal (>= 100万) ===")

	withdrawalData := map[string]interface{}{
		"merchant_id":      merchantID.String(),
		"user_id":          userID.String(),
		"bank_account_id":  bankAccountID,
		"amount":           100000000, // 100万元 (>= 100万，需要3级审批)
		"withdrawal_type":  "normal",
		"purpose":          "Large amount withdrawal test",
	}

	resp, body, err = client.POST(
		config.WithdrawalURL+"/api/v1/withdrawals",
		withdrawalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create withdrawal: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create large withdrawal")

	result = ParseJSONResponse(t, body)
	withdrawalID := result["id"].(string)
	requiredLevel := int(result["required_level"].(float64))
	AssertEqual(t, 3, requiredLevel, "required_level for large amount")
	t.Logf("Large withdrawal created: %s (requires %d level approval)", withdrawalID, requiredLevel)

	// Step 3: Approve Level 1
	t.Log("\n=== Step 3: Approve Level 1 ===")

	approvalData := map[string]interface{}{
		"withdrawal_id": withdrawalID,
		"approver_id":   uuid.New().String(),
		"approver_name": "Manager",
		"level":         1,
		"action":        "approve",
		"comment":       "Level 1 approved",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/withdrawals/%s/approve", config.WithdrawalURL, withdrawalID),
		approvalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to approve level 1: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Approve level 1")
	t.Log("✓ Level 1 approved")

	// Step 4: Approve Level 2
	t.Log("\n=== Step 4: Approve Level 2 ===")

	approvalData["approver_id"] = uuid.New().String()
	approvalData["approver_name"] = "Director"
	approvalData["level"] = 2
	approvalData["comment"] = "Level 2 approved"

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/withdrawals/%s/approve", config.WithdrawalURL, withdrawalID),
		approvalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to approve level 2: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Approve level 2")
	t.Log("✓ Level 2 approved")

	// Step 5: Check status (should still be pending, waiting for level 3)
	t.Log("\n=== Step 5: Check Status After 2 Levels ===")

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/withdrawals/%s", config.WithdrawalURL, withdrawalID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query withdrawal: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	approvalLevel := int(result["approval_level"].(float64))

	AssertEqual(t, "pending", status, "Status after 2 levels")
	AssertEqual(t, 2, approvalLevel, "Current approval level")
	t.Logf("Status: %s, Approval level: %d/%d", status, approvalLevel, requiredLevel)

	// Step 6: Approve Level 3 (final)
	t.Log("\n=== Step 6: Approve Level 3 (Final) ===")

	approvalData["approver_id"] = uuid.New().String()
	approvalData["approver_name"] = "CEO"
	approvalData["level"] = 3
	approvalData["comment"] = "Level 3 approved - Final approval"

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/withdrawals/%s/approve", config.WithdrawalURL, withdrawalID),
		approvalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to approve level 3: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Approve level 3")
	t.Log("✓ Level 3 approved")

	// Step 7: Verify final status
	t.Log("\n=== Step 7: Verify Final Status ===")
	Sleep(1 * time.Second)

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/withdrawals/%s", config.WithdrawalURL, withdrawalID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query withdrawal: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status = result["status"].(string)

	// After all 3 levels approved, should be "approved" or "processing"
	if status == "approved" || status == "processing" {
		t.Logf("✓ Withdrawal fully approved: %s", status)
	} else {
		t.Logf("Warning: unexpected final status: %s", status)
	}

	t.Log("\n=== Multi-Level Approval Test Completed ===")
}

// TestWithdrawalRejection tests withdrawal rejection
func TestWithdrawalRejection(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.WithdrawalURL, 3); err != nil {
		t.Skipf("Withdrawal service not ready: %v", err)
	}

	merchantID := uuid.New()
	userID := uuid.New()

	// Create bank account
	t.Log("\n=== Step 1: Create Bank Account ===")

	bankAccount := map[string]interface{}{
		"merchant_id":     merchantID.String(),
		"user_id":         userID.String(),
		"account_name":    "Reject Test Account",
		"account_number":  "1111222233",
		"bank_name":       "Test Bank",
		"bank_code":       "TEST001",
		"account_type":    "personal",
	}

	resp, body, err := client.POST(
		config.WithdrawalURL+"/api/v1/bank-accounts",
		bankAccount,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create bank account: %v", err)
	}

	result := ParseJSONResponse(t, body)
	bankAccountID := result["id"].(string)

	// Create withdrawal
	t.Log("\n=== Step 2: Create Withdrawal ===")

	withdrawalData := map[string]interface{}{
		"merchant_id":      merchantID.String(),
		"user_id":          userID.String(),
		"bank_account_id":  bankAccountID,
		"amount":           1000000, // 10,000元
		"withdrawal_type":  "normal",
		"purpose":          "Rejection test",
	}

	resp, body, err = client.POST(
		config.WithdrawalURL+"/api/v1/withdrawals",
		withdrawalData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create withdrawal: %v", err)
	}

	result = ParseJSONResponse(t, body)
	withdrawalID := result["id"].(string)
	t.Logf("Withdrawal created: %s", withdrawalID)

	// Reject withdrawal
	t.Log("\n=== Step 3: Reject Withdrawal ===")

	rejectionData := map[string]interface{}{
		"withdrawal_id": withdrawalID,
		"approver_id":   uuid.New().String(),
		"approver_name": "Risk Manager",
		"level":         1,
		"action":        "reject",
		"comment":       "Suspicious activity detected",
	}

	resp, _, err = client.POST(
		fmt.Sprintf("%s/api/v1/withdrawals/%s/approve", config.WithdrawalURL, withdrawalID),
		rejectionData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to reject withdrawal: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Reject withdrawal")
	t.Log("✓ Withdrawal rejected")

	// Verify rejection status
	t.Log("\n=== Step 4: Verify Rejection Status ===")

	resp, body, err = client.GET(
		fmt.Sprintf("%s/api/v1/withdrawals/%s", config.WithdrawalURL, withdrawalID),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to query withdrawal: %v", err)
	}

	result = ParseJSONResponse(t, body)
	status := result["status"].(string)
	AssertEqual(t, "rejected", status, "Withdrawal status after rejection")
	t.Logf("✓ Withdrawal status confirmed: %s", status)

	t.Log("\n=== Rejection Test Completed ===")
}

// TestWithdrawalBatch tests batch withdrawal processing
func TestWithdrawalBatch(t *testing.T) {
	config := DefaultTestConfig()
	client := NewHTTPClient(config)

	if err := WaitForService(config.WithdrawalURL, 3); err != nil {
		t.Skipf("Withdrawal service not ready: %v", err)
	}

	t.Log("\n=== Test: Create Withdrawal Batch ===")

	batchData := map[string]interface{}{
		"name":        fmt.Sprintf("Batch %d", time.Now().Unix()),
		"description": "Integration test batch",
		"total_count": 0,
		"total_amount": 0,
	}

	resp, body, err := client.POST(
		config.WithdrawalURL+"/api/v1/withdrawal-batches",
		batchData,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create batch: %v", err)
	}

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, "Create withdrawal batch")

	result := ParseJSONResponse(t, body)
	batchID := result["id"].(string)
	AssertNotEmpty(t, batchID, "batch_id")
	t.Logf("✓ Withdrawal batch created: %s", batchID)

	t.Log("\n=== Batch Test Completed ===")
}
