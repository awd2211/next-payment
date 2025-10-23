#!/bin/bash

# ============================================
# Payment Platform - Run Integration Tests
# ============================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_DIR="$BACKEND_DIR/tests/integration"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if services are running
check_services() {
    echo -e "${BLUE}Checking if required services are running...${NC}"

    local required_services=(
        "payment-gateway:8003"
        "order-service:8004"
        "risk-service:8006"
        "settlement-service:8012"
        "withdrawal-service:8013"
        "kyc-service:8014"
    )

    local missing_services=()

    for service_port in "${required_services[@]}"; do
        IFS=':' read -r service port <<< "$service_port"
        if ! curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            missing_services+=("$service")
        else
            echo -e "  ${GREEN}✓${NC} $service"
        fi
    done

    if [ ${#missing_services[@]} -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}Warning: The following services are not running:${NC}"
        for service in "${missing_services[@]}"; do
            echo -e "  ${YELLOW}⚠${NC} $service"
        done
        echo ""
        echo -e "${CYAN}Some tests may be skipped. Start services with:${NC}"
        echo -e "  ${CYAN}./scripts/start-all.sh${NC}"
        echo ""
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    echo ""
}

# Run tests
run_tests() {
    local test_pattern=$1
    local verbose=$2

    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}Running Integration Tests${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    cd "$TEST_DIR"

    # Initialize go module if needed
    if [ ! -f "go.sum" ]; then
        echo -e "${BLUE}Initializing Go module...${NC}"
        go mod tidy
        echo ""
    fi

    # Build test flags
    local test_flags="-v"
    if [ "$verbose" = true ]; then
        test_flags="-v -count=1"
    fi

    # Run tests
    if [ -n "$test_pattern" ]; then
        echo -e "${CYAN}Running tests matching: $test_pattern${NC}"
        echo ""
        go test $test_flags -run "$test_pattern"
    else
        echo -e "${CYAN}Running all integration tests${NC}"
        echo ""
        go test $test_flags
    fi

    local exit_code=$?

    echo ""
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed${NC}"
    else
        echo -e "${RED}✗ Some tests failed${NC}"
    fi

    return $exit_code
}

# Generate test report
generate_report() {
    echo -e "${BLUE}Generating test report...${NC}"

    cd "$TEST_DIR"

    go test -v -json > test-results.json 2>&1 || true

    # Parse results
    local total=$(grep -c '"Action":"pass"\|"Action":"fail"\|"Action":"skip"' test-results.json 2>/dev/null || echo "0")
    local passed=$(grep -c '"Action":"pass"' test-results.json 2>/dev/null || echo "0")
    local failed=$(grep -c '"Action":"fail"' test-results.json 2>/dev/null || echo "0")
    local skipped=$(grep -c '"Action":"skip"' test-results.json 2>/dev/null || echo "0")

    echo ""
    echo -e "${CYAN}Test Summary:${NC}"
    echo -e "  Total:   $total"
    echo -e "  ${GREEN}Passed:  $passed${NC}"
    if [ $failed -gt 0 ]; then
        echo -e "  ${RED}Failed:  $failed${NC}"
    else
        echo -e "  Failed:  $failed"
    fi
    if [ $skipped -gt 0 ]; then
        echo -e "  ${YELLOW}Skipped: $skipped${NC}"
    else
        echo -e "  Skipped: $skipped"
    fi
    echo ""

    echo -e "${BLUE}Test report saved to: $TEST_DIR/test-results.json${NC}"
}

# Run specific test suite
run_suite() {
    local suite=$1

    case $suite in
        payment)
            echo -e "${CYAN}Running Payment Flow Tests${NC}"
            run_tests "TestPayment" false
            ;;
        withdrawal)
            echo -e "${CYAN}Running Withdrawal Flow Tests${NC}"
            run_tests "TestWithdrawal" false
            ;;
        settlement)
            echo -e "${CYAN}Running Settlement Flow Tests${NC}"
            run_tests "TestSettlement" false
            ;;
        kyc)
            echo -e "${CYAN}Running KYC Flow Tests${NC}"
            run_tests "TestKYC" false
            ;;
        merchant)
            echo -e "${CYAN}Running Merchant Management Tests${NC}"
            run_tests "TestMerchant" false
            ;;
        admin)
            echo -e "${CYAN}Running Admin Service Tests${NC}"
            run_tests "TestAdmin" false
            ;;
        order)
            echo -e "${CYAN}Running Order Service Tests${NC}"
            run_tests "TestOrder" false
            ;;
        risk)
            echo -e "${CYAN}Running Risk Assessment Tests${NC}"
            run_tests "TestRisk|TestBlacklist" false
            ;;
        notification)
            echo -e "${CYAN}Running Notification Tests${NC}"
            run_tests "TestNotification" false
            ;;
        config)
            echo -e "${CYAN}Running Configuration Tests${NC}"
            run_tests "TestConfig" false
            ;;
        accounting)
            echo -e "${CYAN}Running Accounting Tests${NC}"
            run_tests "TestAccounting" false
            ;;
        analytics)
            echo -e "${CYAN}Running Analytics Tests${NC}"
            run_tests "TestAnalytics" false
            ;;
        *)
            echo -e "${RED}Unknown test suite: $suite${NC}"
            echo "Available suites:"
            echo "  payment, withdrawal, settlement, kyc"
            echo "  merchant, admin, order, risk"
            echo "  notification, config, accounting, analytics"
            exit 1
            ;;
    esac
}

# Show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -a, --all           Run all integration tests (default)"
    echo "  -s, --suite NAME    Run specific test suite"
    echo "  -t, --test PATTERN  Run tests matching pattern"
    echo "  -v, --verbose       Verbose output"
    echo "  -r, --report        Generate test report"
    echo "  -c, --check         Only check if services are running"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Available Test Suites:"
    echo "  Core Flows:"
    echo "    payment       - Payment flow tests (4 tests)"
    echo "    withdrawal    - Withdrawal flow tests (4 tests)"
    echo "    settlement    - Settlement flow tests (2 tests)"
    echo "    kyc           - KYC verification tests (4 tests)"
    echo ""
    echo "  Service Management:"
    echo "    merchant      - Merchant management tests (8 tests)"
    echo "    admin         - Admin service tests (2 tests)"
    echo "    order         - Order service tests (5 tests)"
    echo "    risk          - Risk assessment tests (5 tests)"
    echo ""
    echo "  System Services:"
    echo "    notification  - Notification tests (5 tests)"
    echo "    config        - Configuration tests (3 tests)"
    echo "    accounting    - Accounting tests (2 tests)"
    echo "    analytics     - Analytics tests (1 test)"
    echo ""
    echo "Examples:"
    echo "  $0                        # Run all 42 tests"
    echo "  $0 -s payment             # Run payment tests only"
    echo "  $0 -s merchant            # Run merchant tests"
    echo "  $0 -t TestPaymentFlow     # Run specific test"
    echo "  $0 -v -r                  # Verbose output with report"
    exit 0
}

# Main execution
main() {
    local run_all=true
    local verbose=false
    local report=false
    local check_only=false
    local suite=""
    local pattern=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -a|--all)
                run_all=true
                shift
                ;;
            -s|--suite)
                suite=$2
                run_all=false
                shift 2
                ;;
            -t|--test)
                pattern=$2
                run_all=false
                shift 2
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            -r|--report)
                report=true
                shift
                ;;
            -c|--check)
                check_only=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                usage
                ;;
        esac
    done

    # Check services
    check_services

    if [ "$check_only" = true ]; then
        exit 0
    fi

    # Run tests
    if [ -n "$suite" ]; then
        run_suite "$suite"
        exit_code=$?
    elif [ -n "$pattern" ]; then
        run_tests "$pattern" "$verbose"
        exit_code=$?
    else
        run_tests "" "$verbose"
        exit_code=$?
    fi

    # Generate report if requested
    if [ "$report" = true ]; then
        generate_report
    fi

    exit $exit_code
}

main "$@"
