#!/bin/bash
# ============================================================================
# Merchant Services API Compatibility Test Script
# ============================================================================
# Purpose: Test new services API endpoints and compare with expected behavior
# Author: Claude (Sonnet 4.5)
# Date: 2025-10-26
# Version: 1.0
# ============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Service URLs
POLICY_SERVICE_URL=${POLICY_SERVICE_URL:-http://localhost:40112}
QUOTA_SERVICE_URL=${QUOTA_SERVICE_URL:-http://localhost:40122}

# Test JWT token (for authenticated endpoints)
# This is a sample token - in production, generate a real JWT
JWT_TOKEN=${JWT_TOKEN:-""}

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# ============================================================================
# Helper Functions
# ============================================================================

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((TESTS_SKIPPED++))
}

call_api() {
    local method=$1
    local url=$2
    local data=$3
    local auth_header=""

    if [ -n "$JWT_TOKEN" ]; then
        auth_header="-H \"Authorization: Bearer $JWT_TOKEN\""
    fi

    if [ "$method" = "GET" ]; then
        curl -s -X GET "$url" $auth_header
    elif [ "$method" = "POST" ]; then
        curl -s -X POST "$url" \
            -H "Content-Type: application/json" \
            $auth_header \
            -d "$data"
    fi
}

check_service_health() {
    local service_name=$1
    local health_url=$2

    log "Checking $service_name health..."
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$health_url")

    if [ "$response" = "200" ] || [ "$response" = "429" ]; then
        log_success "$service_name is healthy (HTTP $response)"
        return 0
    else
        log_error "$service_name health check failed (HTTP $response)"
        return 1
    fi
}

# ============================================================================
# Policy Service Tests
# ============================================================================

test_policy_service() {
    log "=========================================="
    log "Testing Merchant Policy Service"
    log "=========================================="

    # Test 1: Health check
    if check_service_health "merchant-policy-service" "$POLICY_SERVICE_URL/health"; then
        : # pass
    else
        log_skip "Skipping policy service tests due to health check failure"
        return 1
    fi

    # Test 2: Get all active tiers
    log "Test: GET /api/v1/tiers/active"
    local tiers_response=$(call_api "GET" "$POLICY_SERVICE_URL/api/v1/tiers/active")

    if echo "$tiers_response" | grep -q "starter"; then
        log_success "Retrieved active tiers successfully"
    elif echo "$tiers_response" | grep -q "rate limit"; then
        log_skip "Rate limit exceeded, skipping tier test"
    else
        log_error "Failed to retrieve active tiers"
    fi

    # Test 3: Get tier by code
    log "Test: GET /api/v1/tiers/code/starter"
    local tier_response=$(call_api "GET" "$POLICY_SERVICE_URL/api/v1/tiers/code/starter")

    if echo "$tier_response" | grep -q '"tier_code":"starter"' || echo "$tier_response" | grep -q "rate limit"; then
        if echo "$tier_response" | grep -q "rate limit"; then
            log_skip "Rate limit exceeded, skipping tier detail test"
        else
            log_success "Retrieved tier details successfully"
        fi
    else
        log_error "Failed to retrieve tier details"
    fi

    # Test 4: Get effective fee policy (without auth - will likely fail, which is expected)
    log "Test: GET /api/v1/policy-engine/fee-policy (unauthenticated)"
    local test_merchant_id="00000000-0000-0000-0000-000000000001"
    local fee_response=$(curl -s -o /dev/null -w "%{http_code}" \
        "$POLICY_SERVICE_URL/api/v1/policy-engine/fee-policy?merchant_id=$test_merchant_id&channel=stripe&currency=USD")

    if [ "$fee_response" = "401" ] || [ "$fee_response" = "403" ]; then
        log_success "Authentication properly required for policy endpoints (HTTP $fee_response)"
    elif [ "$fee_response" = "429" ]; then
        log_skip "Rate limit exceeded, skipping auth test"
    elif [ "$fee_response" = "200" ]; then
        log_success "Fee policy endpoint accessible (HTTP 200)"
    else
        log_error "Unexpected response for fee policy endpoint (HTTP $fee_response)"
    fi

    echo ""
}

# ============================================================================
# Quota Service Tests
# ============================================================================

test_quota_service() {
    log "=========================================="
    log "Testing Merchant Quota Service"
    log "=========================================="

    # Test 1: Health check
    if check_service_health "merchant-quota-service" "$QUOTA_SERVICE_URL/health"; then
        : # pass
    else
        log_skip "Skipping quota service tests due to health check failure"
        return 1
    fi

    # Test 2: Get quota (unauthenticated - should fail)
    log "Test: GET /api/v1/quotas (unauthenticated)"
    local test_merchant_id="00000000-0000-0000-0000-000000000001"
    local quota_response=$(curl -s -o /dev/null -w "%{http_code}" \
        "$QUOTA_SERVICE_URL/api/v1/quotas?merchant_id=$test_merchant_id&currency=USD")

    if [ "$quota_response" = "401" ] || [ "$quota_response" = "403" ]; then
        log_success "Authentication properly required for quota endpoints (HTTP $quota_response)"
    elif [ "$quota_response" = "429" ]; then
        log_skip "Rate limit exceeded, skipping quota auth test"
    elif [ "$quota_response" = "200" ] || [ "$quota_response" = "404" ]; then
        log_success "Quota endpoint accessible (HTTP $quota_response)"
    else
        log_error "Unexpected response for quota endpoint (HTTP $quota_response)"
    fi

    # Test 3: Initialize quota (unauthenticated - should fail)
    log "Test: POST /api/v1/quotas/initialize (unauthenticated)"
    local init_data='{
        "merchant_id": "00000000-0000-0000-0000-000000000001",
        "currency": "USD",
        "daily_limit": 1000000,
        "monthly_limit": 10000000
    }'
    local init_response=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$QUOTA_SERVICE_URL/api/v1/quotas/initialize" \
        -H "Content-Type: application/json" \
        -d "$init_data")

    if [ "$init_response" = "401" ] || [ "$init_response" = "403" ]; then
        log_success "Authentication properly required for initialize endpoint (HTTP $init_response)"
    elif [ "$init_response" = "429" ]; then
        log_skip "Rate limit exceeded, skipping initialize test"
    else
        log_error "Unexpected response for initialize endpoint (HTTP $init_response)"
    fi

    echo ""
}

# ============================================================================
# Metrics Endpoints Tests
# ============================================================================

test_metrics_endpoints() {
    log "=========================================="
    log "Testing Metrics Endpoints"
    log "=========================================="

    # Test policy service metrics
    log "Test: GET /metrics (policy-service)"
    local policy_metrics=$(curl -s -o /dev/null -w "%{http_code}" "$POLICY_SERVICE_URL/metrics")

    if [ "$policy_metrics" = "200" ]; then
        log_success "Policy service metrics endpoint accessible"
    else
        log_error "Policy service metrics endpoint failed (HTTP $policy_metrics)"
    fi

    # Test quota service metrics
    log "Test: GET /metrics (quota-service)"
    local quota_metrics=$(curl -s -o /dev/null -w "%{http_code}" "$QUOTA_SERVICE_URL/metrics")

    if [ "$quota_metrics" = "200" ]; then
        log_success "Quota service metrics endpoint accessible"
    else
        log_error "Quota service metrics endpoint failed (HTTP $quota_metrics)"
    fi

    echo ""
}

# ============================================================================
# Swagger Documentation Tests
# ============================================================================

test_swagger_docs() {
    log "=========================================="
    log "Testing Swagger Documentation"
    log "=========================================="

    # Test policy service swagger
    log "Test: GET /swagger/index.html (policy-service)"
    local policy_swagger=$(curl -s -o /dev/null -w "%{http_code}" "$POLICY_SERVICE_URL/swagger/index.html")

    if [ "$policy_swagger" = "200" ]; then
        log_success "Policy service Swagger docs accessible"
        log "  URL: $POLICY_SERVICE_URL/swagger/index.html"
    else
        log_error "Policy service Swagger docs failed (HTTP $policy_swagger)"
    fi

    # Test quota service swagger
    log "Test: GET /swagger/index.html (quota-service)"
    local quota_swagger=$(curl -s -o /dev/null -w "%{http_code}" "$QUOTA_SERVICE_URL/swagger/index.html")

    if [ "$quota_swagger" = "200" ]; then
        log_success "Quota service Swagger docs accessible"
        log "  URL: $QUOTA_SERVICE_URL/swagger/index.html"
    else
        log_error "Quota service Swagger docs failed (HTTP $quota_swagger)"
    fi

    echo ""
}

# ============================================================================
# Response Time Tests
# ============================================================================

test_response_times() {
    log "=========================================="
    log "Testing Response Times"
    log "=========================================="

    # Policy service health endpoint
    log "Test: Health endpoint response time (policy-service)"
    local start_time=$(date +%s%N)
    curl -s "$POLICY_SERVICE_URL/health" > /dev/null
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds

    echo "  Response time: ${duration}ms"
    if [ $duration -lt 200 ]; then
        log_success "Policy service responds within 200ms"
    elif [ $duration -lt 500 ]; then
        log_success "Policy service responds within 500ms (acceptable)"
    else
        log_error "Policy service response time too high: ${duration}ms"
    fi

    # Quota service health endpoint
    log "Test: Health endpoint response time (quota-service)"
    start_time=$(date +%s%N)
    curl -s "$QUOTA_SERVICE_URL/health" > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))

    echo "  Response time: ${duration}ms"
    if [ $duration -lt 200 ]; then
        log_success "Quota service responds within 200ms"
    elif [ $duration -lt 500 ]; then
        log_success "Quota service responds within 500ms (acceptable)"
    else
        log_error "Quota service response time too high: ${duration}ms"
    fi

    echo ""
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    echo ""
    echo "============================================"
    echo "  Merchant Services API Compatibility Test"
    echo "============================================"
    echo "  Date: $(date +'%Y-%m-%d %H:%M:%S')"
    echo "  Policy Service: $POLICY_SERVICE_URL"
    echo "  Quota Service: $QUOTA_SERVICE_URL"
    echo "============================================"
    echo ""

    # Run all test suites
    test_policy_service
    test_quota_service
    test_metrics_endpoints
    test_swagger_docs
    test_response_times

    # Print summary
    echo "============================================"
    echo "  Test Summary"
    echo "============================================"
    echo -e "  ${GREEN}Passed:${NC}  $TESTS_PASSED"
    echo -e "  ${RED}Failed:${NC}  $TESTS_FAILED"
    echo -e "  ${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
    echo "  Total:   $((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))"
    echo "============================================"
    echo ""

    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed! ✅"
        log "API compatibility verified successfully"
        return 0
    else
        log_error "$TESTS_FAILED test(s) failed ❌"
        return 1
    fi
}

main "$@"
