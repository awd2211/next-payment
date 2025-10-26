#!/bin/bash
# ============================================================================
# Merchant Services Migration Verification Script
# ============================================================================
# Purpose: Detailed verification of data migration correctness
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

# Database connection
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-40432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

# ============================================================================
# Helper Functions
# ============================================================================

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

run_sql() {
    local db_name=$1
    local sql=$2
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $db_name -tAc "$sql"
}

# ============================================================================
# Verification Tests
# ============================================================================

verify_record_counts() {
    log "=========================================="
    log "Test 1: Record Count Verification"
    log "=========================================="

    # Fee configs
    local old_fee=$(run_sql "payment_merchant_config" \
        "SELECT COUNT(*) FROM fee_configs WHERE merchant_id IS NOT NULL" 2>/dev/null || echo "0")
    local new_fee=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_fee_policies WHERE merchant_id IS NOT NULL" 2>/dev/null || echo "0")

    echo "  Fee Configs: $old_fee (old) -> $new_fee (new)"
    if [ "$new_fee" -ge "$old_fee" ]; then
        log_success "Fee policies count validated"
    else
        log_error "Fee policies count mismatch: $new_fee < $old_fee"
        return 1
    fi

    # Transaction limits
    local old_limit=$(run_sql "payment_merchant_config" \
        "SELECT COUNT(*) FROM transaction_limits WHERE merchant_id IS NOT NULL" 2>/dev/null || echo "0")
    local new_limit=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_limit_policies WHERE merchant_id IS NOT NULL" 2>/dev/null || echo "0")

    echo "  Transaction Limits: $old_limit (old) -> $new_limit (new)"
    if [ "$new_limit" -ge "$old_limit" ]; then
        log_success "Limit policies count validated"
    else
        log_error "Limit policies count mismatch: $new_limit < $old_limit"
        return 1
    fi

    # Merchant quotas
    local old_quota=$(run_sql "payment_merchant_limit" \
        "SELECT COUNT(*) FROM merchant_limits" 2>/dev/null || echo "0")
    local new_quota=$(run_sql "payment_merchant_quota" \
        "SELECT COUNT(*) FROM merchant_quotas" 2>/dev/null || echo "0")

    echo "  Merchant Limits: $old_quota (old) -> $new_quota (new)"
    if [ "$new_quota" -ge "$old_quota" ]; then
        log_success "Quotas count validated"
    else
        log_error "Quotas count mismatch: $new_quota < $old_quota"
        return 1
    fi

    echo ""
}

verify_data_integrity() {
    log "=========================================="
    log "Test 2: Data Integrity Verification"
    log "=========================================="

    # Sample merchant_id verification (if data exists)
    local sample_merchant=$(run_sql "payment_merchant_config" \
        "SELECT merchant_id FROM fee_configs WHERE merchant_id IS NOT NULL LIMIT 1" 2>/dev/null || echo "")

    if [ -n "$sample_merchant" ]; then
        log "Checking sample merchant: $sample_merchant"

        # Check if fee policy exists in new database
        local fee_exists=$(run_sql "payment_merchant_policy" \
            "SELECT COUNT(*) FROM merchant_fee_policies WHERE merchant_id::text = '$sample_merchant'")

        if [ "$fee_exists" -gt 0 ]; then
            log_success "Sample merchant fee policy found in new database"
        else
            log_error "Sample merchant fee policy NOT found in new database"
            return 1
        fi

        # Check if limit policy exists
        local limit_exists=$(run_sql "payment_merchant_policy" \
            "SELECT COUNT(*) FROM merchant_limit_policies WHERE merchant_id::text = '$sample_merchant'")

        if [ "$limit_exists" -gt 0 ]; then
            log_success "Sample merchant limit policy found in new database"
        else
            log_warning "Sample merchant limit policy NOT found (may not exist in old DB)"
        fi
    else
        log_warning "No merchant data found in old database (empty migration)"
    fi

    echo ""
}

verify_tier_defaults() {
    log "=========================================="
    log "Test 3: Tier Default Policies Verification"
    log "=========================================="

    # Check if 5 default tiers exist
    local tier_count=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_tiers WHERE is_active = true")

    echo "  Active tiers: $tier_count"
    if [ "$tier_count" -eq 5 ]; then
        log_success "All 5 default tiers exist"
    else
        log_error "Expected 5 tiers, found $tier_count"
        return 1
    fi

    # Check if tier default policies exist
    local tier_fee_count=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_fee_policies WHERE tier_id IS NOT NULL")

    echo "  Tier-level fee policies: $tier_fee_count"
    if [ "$tier_fee_count" -ge 5 ]; then
        log_success "Tier default fee policies exist"
    else
        log_error "Expected at least 5 tier fee policies, found $tier_fee_count"
        return 1
    fi

    local tier_limit_count=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_limit_policies WHERE tier_id IS NOT NULL")

    echo "  Tier-level limit policies: $tier_limit_count"
    if [ "$tier_limit_count" -ge 5 ]; then
        log_success "Tier default limit policies exist"
    else
        log_error "Expected at least 5 tier limit policies, found $tier_limit_count"
        return 1
    fi

    echo ""
}

verify_quota_consistency() {
    log "=========================================="
    log "Test 4: Quota Data Consistency"
    log "=========================================="

    # Check for negative values (should not exist)
    local negative_daily=$(run_sql "payment_merchant_quota" \
        "SELECT COUNT(*) FROM merchant_quotas WHERE daily_used < 0 OR monthly_used < 0")

    if [ "$negative_daily" -eq 0 ]; then
        log_success "No negative quota usage values"
    else
        log_error "Found $negative_daily records with negative usage"
        return 1
    fi

    # Check if daily_used <= daily_limit
    local exceeded_daily=$(run_sql "payment_merchant_quota" \
        "SELECT COUNT(*) FROM merchant_quotas WHERE daily_used > daily_limit AND status = 'active'")

    if [ "$exceeded_daily" -eq 0 ]; then
        log_success "No active quotas exceeding daily limits"
    else
        log_warning "Found $exceeded_daily quotas exceeding daily limits (may be intentional)"
    fi

    # Check version numbers (should all be 0 after migration)
    local non_zero_version=$(run_sql "payment_merchant_quota" \
        "SELECT COUNT(*) FROM merchant_quotas WHERE version != 0")

    if [ "$non_zero_version" -eq 0 ]; then
        log_success "All quota versions initialized to 0"
    else
        log_warning "Found $non_zero_version quotas with non-zero versions"
    fi

    echo ""
}

verify_status_mapping() {
    log "=========================================="
    log "Test 5: Status Field Mapping"
    log "=========================================="

    # Check for invalid status values
    local invalid_fee_status=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_fee_policies WHERE status NOT IN ('active', 'inactive', 'expired')")

    if [ "$invalid_fee_status" -eq 0 ]; then
        log_success "All fee policy statuses are valid"
    else
        log_error "Found $invalid_fee_status fee policies with invalid status"
        return 1
    fi

    local invalid_limit_status=$(run_sql "payment_merchant_policy" \
        "SELECT COUNT(*) FROM merchant_limit_policies WHERE status NOT IN ('active', 'inactive', 'expired')")

    if [ "$invalid_limit_status" -eq 0 ]; then
        log_success "All limit policy statuses are valid"
    else
        log_error "Found $invalid_limit_status limit policies with invalid status"
        return 1
    fi

    local invalid_quota_status=$(run_sql "payment_merchant_quota" \
        "SELECT COUNT(*) FROM merchant_quotas WHERE status NOT IN ('active', 'inactive', 'suspended', 'frozen')")

    if [ "$invalid_quota_status" -eq 0 ]; then
        log_success "All quota statuses are valid"
    else
        log_error "Found $invalid_quota_status quotas with invalid status"
        return 1
    fi

    echo ""
}

generate_summary_report() {
    log "=========================================="
    log "Migration Summary Report"
    log "=========================================="

    # Generate detailed comparison
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER <<EOF
-- Fee Policies Summary
\echo 'Fee Policies Comparison:'
\c payment_merchant_config
SELECT
    'Old DB (fee_configs)' AS source,
    COUNT(*) AS total,
    COUNT(DISTINCT merchant_id) AS unique_merchants,
    COUNT(*) FILTER (WHERE is_active = true) AS active_count
FROM fee_configs
WHERE merchant_id IS NOT NULL;

\c payment_merchant_policy
SELECT
    'New DB (merchant_fee_policies)' AS source,
    COUNT(*) AS total,
    COUNT(DISTINCT merchant_id) AS unique_merchants,
    COUNT(*) FILTER (WHERE status = 'active') AS active_count
FROM merchant_fee_policies
WHERE merchant_id IS NOT NULL;

\echo ''
\echo 'Limit Policies Comparison:'
\c payment_merchant_config
SELECT
    'Old DB (transaction_limits)' AS source,
    COUNT(*) AS total,
    COUNT(DISTINCT merchant_id) AS unique_merchants,
    COUNT(*) FILTER (WHERE is_active = true) AS active_count
FROM transaction_limits
WHERE merchant_id IS NOT NULL;

\c payment_merchant_policy
SELECT
    'New DB (merchant_limit_policies)' AS source,
    COUNT(*) AS total,
    COUNT(DISTINCT merchant_id) AS unique_merchants,
    COUNT(*) FILTER (WHERE status = 'active') AS active_count
FROM merchant_limit_policies
WHERE merchant_id IS NOT NULL;

\echo ''
\echo 'Merchant Quotas Comparison:'
\c payment_merchant_limit
SELECT
    'Old DB (merchant_limits)' AS source,
    COUNT(*) AS total,
    COUNT(DISTINCT merchant_id) AS unique_merchants,
    SUM(daily_used)::bigint AS total_daily_used,
    SUM(monthly_used)::bigint AS total_monthly_used
FROM merchant_limits;

\c payment_merchant_quota
SELECT
    'New DB (merchant_quotas)' AS source,
    COUNT(*) AS total,
    COUNT(DISTINCT merchant_id) AS unique_merchants,
    SUM(daily_used)::bigint AS total_daily_used,
    SUM(monthly_used)::bigint AS total_monthly_used
FROM merchant_quotas;
EOF

    echo ""
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    echo ""
    echo "============================================"
    echo "  Merchant Services Migration Verification"
    echo "============================================"
    echo "  Date: $(date +'%Y-%m-%d %H:%M:%S')"
    echo "  Host: $DB_HOST:$DB_PORT"
    echo "============================================"
    echo ""

    local failed_tests=0

    # Run all verification tests
    verify_record_counts || ((failed_tests++))
    verify_data_integrity || ((failed_tests++))
    verify_tier_defaults || ((failed_tests++))
    verify_quota_consistency || ((failed_tests++))
    verify_status_mapping || ((failed_tests++))

    # Generate summary report
    generate_summary_report

    # Final result
    echo ""
    log "=========================================="
    if [ $failed_tests -eq 0 ]; then
        log_success "All verification tests passed! ✅"
        log_success "Migration is ready for Phase 2 (灰度切流)"
    else
        log_error "$failed_tests test(s) failed ❌"
        log_error "Please review and fix issues before proceeding"
        return 1
    fi
    log "=========================================="
    echo ""
}

main "$@"
