#!/bin/bash
# ============================================================================
# Merchant Services Data Migration Execution Script
# ============================================================================
# Purpose: Execute data migration from old to new merchant services
# Author: Claude (Sonnet 4.5)
# Date: 2025-10-26
# Version: 1.0
# ============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database connection settings
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-40432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_SCRIPT="${SCRIPT_DIR}/migrate-merchant-data.sql"
LOG_FILE="${SCRIPT_DIR}/../logs/migration-$(date +%Y%m%d-%H%M%S).log"

# Ensure logs directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# ============================================================================
# Helper Functions
# ============================================================================

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

check_database_exists() {
    local db_name=$1
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tAc \
        "SELECT 1 FROM pg_database WHERE datname='$db_name'" | grep -q 1
}

count_table_records() {
    local db_name=$1
    local table_name=$2
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $db_name -tAc \
        "SELECT COUNT(*) FROM $table_name" 2>/dev/null || echo "0"
}

# ============================================================================
# Pre-flight Checks
# ============================================================================

preflight_checks() {
    log "Starting pre-flight checks..."

    # Check if psql is installed
    if ! command -v psql &> /dev/null; then
        log_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi

    # Check if we can connect to database
    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "SELECT 1" > /dev/null 2>&1; then
        log_error "Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT"
        exit 1
    fi

    # Check if old databases exist
    for db in payment_merchant_config payment_merchant_limit; do
        if check_database_exists $db; then
            log_success "Old database '$db' exists"
        else
            log_warning "Old database '$db' does not exist (may be empty migration)"
        fi
    done

    # Check if new databases exist
    for db in payment_merchant_policy payment_merchant_quota; do
        if check_database_exists $db; then
            log_success "New database '$db' exists"
        else
            log_error "New database '$db' does not exist. Please ensure new services have auto-migrated."
            exit 1
        fi
    done

    # Check if SQL script exists
    if [ ! -f "$SQL_SCRIPT" ]; then
        log_error "SQL script not found: $SQL_SCRIPT"
        exit 1
    fi

    log_success "Pre-flight checks completed"
}

# ============================================================================
# Migration Execution
# ============================================================================

execute_migration() {
    local dry_run=$1

    if [ "$dry_run" = "true" ]; then
        log_warning "DRY RUN MODE - No data will be modified"
        log "Would execute SQL script: $SQL_SCRIPT"
        log "Please review the SQL script before running actual migration"
        return 0
    fi

    log "Executing data migration..."
    log "SQL Script: $SQL_SCRIPT"
    log "Log File: $LOG_FILE"

    # Execute migration SQL
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER \
        -f "$SQL_SCRIPT" >> "$LOG_FILE" 2>&1; then
        log_success "Migration SQL executed successfully"
        return 0
    else
        log_error "Migration SQL execution failed. Check log: $LOG_FILE"
        return 1
    fi
}

# ============================================================================
# Post-Migration Validation
# ============================================================================

validate_migration() {
    log "Validating migration results..."

    # Count records in old databases
    local old_fee_count=$(count_table_records "payment_merchant_config" "fee_configs")
    local old_limit_count=$(count_table_records "payment_merchant_config" "transaction_limits")
    local old_quota_count=$(count_table_records "payment_merchant_limit" "merchant_limits")

    # Count records in new databases
    local new_fee_count=$(count_table_records "payment_merchant_policy" "merchant_fee_policies")
    local new_limit_count=$(count_table_records "payment_merchant_policy" "merchant_limit_policies")
    local new_quota_count=$(count_table_records "payment_merchant_quota" "merchant_quotas")

    log "Migration Record Counts:"
    log "  Old fee_configs: $old_fee_count -> New merchant_fee_policies: $new_fee_count"
    log "  Old transaction_limits: $old_limit_count -> New merchant_limit_policies: $new_limit_count"
    log "  Old merchant_limits: $old_quota_count -> New merchant_quotas: $new_quota_count"

    # Validate counts (new should be >= old due to tier default policies)
    local validation_failed=0

    if [ "$new_fee_count" -lt "$old_fee_count" ]; then
        log_error "Fee policies migration incomplete: $new_fee_count < $old_fee_count"
        validation_failed=1
    else
        log_success "Fee policies migration validated"
    fi

    if [ "$new_limit_count" -lt "$old_limit_count" ]; then
        log_error "Limit policies migration incomplete: $new_limit_count < $old_limit_count"
        validation_failed=1
    else
        log_success "Limit policies migration validated"
    fi

    if [ "$new_quota_count" -lt "$old_quota_count" ]; then
        log_error "Quotas migration incomplete: $new_quota_count < $old_quota_count"
        validation_failed=1
    else
        log_success "Quotas migration validated"
    fi

    return $validation_failed
}

# ============================================================================
# Rollback Function
# ============================================================================

rollback_migration() {
    log_warning "Rolling back migration..."

    # Restore from backup tables (created in SQL script)
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d payment_merchant_policy <<EOF
DELETE FROM merchant_fee_policies WHERE merchant_id IS NOT NULL;
DELETE FROM merchant_limit_policies WHERE merchant_id IS NOT NULL;
EOF

    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d payment_merchant_quota <<EOF
DELETE FROM merchant_quotas;
EOF

    log_warning "Rollback completed. Please review and re-run migration if needed."
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    local dry_run=false
    local skip_validation=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                dry_run=true
                shift
                ;;
            --skip-validation)
                skip_validation=true
                shift
                ;;
            --rollback)
                rollback_migration
                exit 0
                ;;
            --help)
                cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --dry-run           Run in dry-run mode (no changes)
  --skip-validation   Skip post-migration validation
  --rollback          Rollback migration (delete migrated data)
  --help              Show this help message

Environment Variables:
  DB_HOST             Database host (default: localhost)
  DB_PORT             Database port (default: 40432)
  DB_USER             Database user (default: postgres)
  DB_PASSWORD         Database password (default: postgres)

Examples:
  # Dry run (preview)
  $0 --dry-run

  # Execute migration
  $0

  # Rollback migration
  $0 --rollback
EOF
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Print header
    echo ""
    echo "============================================"
    echo "  Merchant Services Data Migration"
    echo "============================================"
    echo "  Date: $(date +'%Y-%m-%d %H:%M:%S')"
    echo "  Host: $DB_HOST:$DB_PORT"
    echo "  User: $DB_USER"
    echo "  Dry Run: $dry_run"
    echo "  Log File: $LOG_FILE"
    echo "============================================"
    echo ""

    # Execute migration workflow
    preflight_checks

    if ! execute_migration $dry_run; then
        log_error "Migration failed. Please review logs."
        exit 1
    fi

    if [ "$dry_run" = "true" ]; then
        log "Dry run completed. Review the SQL script and run without --dry-run to execute."
        exit 0
    fi

    if [ "$skip_validation" = "false" ]; then
        if ! validate_migration; then
            log_error "Migration validation failed!"
            log_warning "You can rollback with: $0 --rollback"
            exit 1
        fi
    fi

    # Success
    echo ""
    log_success "=========================================="
    log_success "Migration completed successfully!"
    log_success "=========================================="
    echo ""
    log "Next steps:"
    log "  1. Run: ./verify-migration.sh (detailed verification)"
    log "  2. Test new services API endpoints"
    log "  3. Review migration log: $LOG_FILE"
    log "  4. If all tests pass, proceed to Phase 2 (灰度切流)"
    echo ""
}

# Run main function
main "$@"
