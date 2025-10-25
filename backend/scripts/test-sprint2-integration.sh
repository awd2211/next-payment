#!/bin/bash

# Integration Tests for Sprint 2 Services
# Tests the complete flow of Reconciliation, Dispute, and Merchant Limit services

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
RECON_URL="${RECON_URL:-http://localhost:40020}"
DISPUTE_URL="${DISPUTE_URL:-http://localhost:40021}"
LIMIT_URL="${LIMIT_URL:-http://localhost:40022}"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Test function
test_api() {
    local test_name=$1
    local method=$2
    local url=$3
    local data=$4
    local expected_code=$5

    ((TESTS_RUN++))
    log_test "$test_name"

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "$expected_code" ]; then
        log_pass "$test_name (HTTP $http_code)"
        echo "$body"
        return 0
    else
        log_fail "$test_name (Expected HTTP $expected_code, got $http_code)"
        echo "$body"
        return 1
    fi
}

echo "=========================================="
echo "Sprint 2 Integration Tests"
echo "=========================================="
echo ""

# Health checks
log_info "Checking service health..."
test_api "Reconciliation health check" GET "$RECON_URL/health" "" 200 || true
test_api "Dispute health check" GET "$DISPUTE_URL/health" "" 200 || true
test_api "Merchant Limit health check" GET "$LIMIT_URL/health" "" 200 || true
echo ""

# Test 1: Merchant Limit Service
echo "=========================================="
echo "Test Suite 1: Merchant Limit Service"
echo "=========================================="
echo ""

# 1.1: List tiers
log_test "1.1: List all merchant tiers"
TIER_RESPONSE=$(test_api "List merchant tiers" GET "$LIMIT_URL/api/v1/tiers" "" 200 || echo "FAILED")

if [ "$TIER_RESPONSE" != "FAILED" ]; then
    # Extract starter tier ID
    STARTER_TIER_ID=$(echo "$TIER_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    log_info "Starter Tier ID: $STARTER_TIER_ID"
fi

# 1.2: Initialize merchant limit
log_test "1.2: Initialize merchant limit"
MERCHANT_ID="e55feb66-16f9-41be-a68b-a8961df898b6"  # Test merchant ID

INIT_LIMIT_DATA='{
  "merchant_id": "'$MERCHANT_ID'",
  "tier_id": "'$STARTER_TIER_ID'"
}'

test_api "Initialize merchant limit" POST "$LIMIT_URL/api/v1/limits/initialize" "$INIT_LIMIT_DATA" 200 || true

# 1.3: Get merchant limit
log_test "1.3: Get merchant limit details"
test_api "Get merchant limit" GET "$LIMIT_URL/api/v1/limits/$MERCHANT_ID" "" 200 || true

# 1.4: Check limit (should pass)
log_test "1.4: Check limit for $50"
CHECK_LIMIT_DATA='{
  "merchant_id": "'$MERCHANT_ID'",
  "amount": 5000
}'

CHECK_RESULT=$(test_api "Check limit (pass)" POST "$LIMIT_URL/api/v1/limits/check" "$CHECK_LIMIT_DATA" 200 || echo "FAILED")

# 1.5: Consume limit
log_test "1.5: Consume $50 from limit"
CONSUME_DATA='{
  "merchant_id": "'$MERCHANT_ID'",
  "payment_no": "PAY-TEST-001",
  "order_no": "ORDER-TEST-001",
  "amount": 5000,
  "currency": "USD"
}'

test_api "Consume limit" POST "$LIMIT_URL/api/v1/limits/consume" "$CONSUME_DATA" 200 || true

# 1.6: Get usage statistics
log_test "1.6: Get usage statistics"
test_api "Get statistics" GET "$LIMIT_URL/api/v1/limits/$MERCHANT_ID/statistics" "" 200 || true

# 1.7: Release limit
log_test "1.7: Release $50 (refund)"
RELEASE_DATA='{
  "merchant_id": "'$MERCHANT_ID'",
  "payment_no": "PAY-TEST-001",
  "order_no": "ORDER-TEST-001",
  "amount": 5000,
  "currency": "USD",
  "reason": "refund"
}'

test_api "Release limit" POST "$LIMIT_URL/api/v1/limits/release" "$RELEASE_DATA" 200 || true

# 1.8: Get usage history
log_test "1.8: Get usage history"
test_api "Get usage history" GET "$LIMIT_URL/api/v1/limits/$MERCHANT_ID/usage-history?page=1&page_size=10" "" 200 || true

echo ""

# Test 2: Dispute Service
echo "=========================================="
echo "Test Suite 2: Dispute Service"
echo "=========================================="
echo ""

# 2.1: Create dispute
log_test "2.1: Create a new dispute"
CREATE_DISPUTE_DATA='{
  "channel": "stripe",
  "channel_dispute_id": "dp_test_123",
  "payment_no": "PAY-TEST-001",
  "order_no": "ORDER-TEST-001",
  "merchant_id": "'$MERCHANT_ID'",
  "channel_trade_no": "ch_test_123",
  "amount": 10000,
  "currency": "USD",
  "reason": "fraudulent",
  "reason_code": "fraudulent"
}'

DISPUTE_RESPONSE=$(test_api "Create dispute" POST "$DISPUTE_URL/api/v1/disputes" "$CREATE_DISPUTE_DATA" 200 || echo "FAILED")

if [ "$DISPUTE_RESPONSE" != "FAILED" ]; then
    DISPUTE_ID=$(echo "$DISPUTE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    log_info "Created Dispute ID: $DISPUTE_ID"
fi

# 2.2: List disputes
log_test "2.2: List all disputes"
test_api "List disputes" GET "$DISPUTE_URL/api/v1/disputes?merchant_id=$MERCHANT_ID&page=1&page_size=10" "" 200 || true

# 2.3: Get dispute details
if [ -n "$DISPUTE_ID" ]; then
    log_test "2.3: Get dispute details"
    test_api "Get dispute details" GET "$DISPUTE_URL/api/v1/disputes/$DISPUTE_ID" "" 200 || true
fi

# 2.4: Upload evidence
if [ -n "$DISPUTE_ID" ]; then
    log_test "2.4: Upload evidence"
    UPLOAD_EVIDENCE_DATA='{
      "dispute_id": "'$DISPUTE_ID'",
      "evidence_type": "receipt",
      "title": "Payment Receipt",
      "description": "Customer paid in full on 2024-10-24",
      "uploaded_by": "'$MERCHANT_ID'"
    }'

    test_api "Upload evidence" POST "$DISPUTE_URL/api/v1/disputes/$DISPUTE_ID/evidence" "$UPLOAD_EVIDENCE_DATA" 200 || true
fi

# 2.5: List evidence
if [ -n "$DISPUTE_ID" ]; then
    log_test "2.5: List dispute evidence"
    test_api "List evidence" GET "$DISPUTE_URL/api/v1/disputes/$DISPUTE_ID/evidence" "" 200 || true
fi

# 2.6: Get statistics
log_test "2.6: Get dispute statistics"
test_api "Get dispute statistics" GET "$DISPUTE_URL/api/v1/disputes/statistics?merchant_id=$MERCHANT_ID" "" 200 || true

echo ""

# Test 3: Reconciliation Service
echo "=========================================="
echo "Test Suite 3: Reconciliation Service"
echo "=========================================="
echo ""

# 3.1: Create reconciliation task
log_test "3.1: Create reconciliation task"
CREATE_TASK_DATA='{
  "task_date": "2024-10-24",
  "channel": "stripe",
  "task_type": "manual"
}'

TASK_RESPONSE=$(test_api "Create reconciliation task" POST "$RECON_URL/api/v1/reconciliation/tasks" "$CREATE_TASK_DATA" 200 || echo "FAILED")

if [ "$TASK_RESPONSE" != "FAILED" ]; then
    TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    log_info "Created Task ID: $TASK_ID"
fi

# 3.2: List tasks
log_test "3.2: List reconciliation tasks"
test_api "List tasks" GET "$RECON_URL/api/v1/reconciliation/tasks?page=1&page_size=10" "" 200 || true

# 3.3: Get task details
if [ -n "$TASK_ID" ]; then
    log_test "3.3: Get task details"
    test_api "Get task details" GET "$RECON_URL/api/v1/reconciliation/tasks/$TASK_ID" "" 200 || true
fi

# 3.4: List records
log_test "3.4: List reconciliation records"
test_api "List records" GET "$RECON_URL/api/v1/reconciliation/records?page=1&page_size=10" "" 200 || true

# 3.5: List settlement files
log_test "3.5: List settlement files"
test_api "List settlement files" GET "$RECON_URL/api/v1/reconciliation/settlement-files?channel=stripe&page=1&page_size=10" "" 200 || true

echo ""

# Test Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "Total Tests:  $TESTS_RUN"
echo -e "Passed:       ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed:       ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
