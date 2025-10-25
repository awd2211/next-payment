#!/bin/bash

#######################################
# Service Dependency Relationship Map
# 可视化展示微服务间的依赖关系
#######################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

clear

echo -e "${BOLD}${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         Service Dependency Relationship Map                 ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

#######################################
# 1. 基础设施依赖 (Infrastructure Layer)
#######################################
echo -e "${BOLD}${BLUE}[Infrastructure Layer]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${CYAN}PostgreSQL${NC} (40432)"
echo "  └─ All services (19 databases)"
echo ""
echo -e "${CYAN}Redis${NC} (40379)"
echo "  ├─ payment-gateway (idempotency, rate limiting)"
echo "  ├─ order-service (caching)"
echo "  ├─ risk-service (rate limiting, cache)"
echo "  └─ accounting-service (caching)"
echo ""
echo -e "${CYAN}Kafka${NC} (40092)"
echo "  ├─ payment-gateway (saga orchestration)"
echo "  ├─ notification-service (event consumer)"
echo "  ├─ accounting-service (event consumer)"
echo "  └─ analytics-service (event consumer)"
echo ""
echo -e "${CYAN}Prometheus${NC} (40090)"
echo "  └─ All services (metrics collection)"
echo ""
echo -e "${CYAN}Jaeger${NC} (40686)"
echo "  └─ All services (distributed tracing)"
echo ""

#######################################
# 2. 核心支付流程 (Core Payment Flow)
#######################################
echo -e "${BOLD}${BLUE}[Core Payment Flow - Critical Path]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BOLD}${GREEN}Merchant API Call${NC}"
echo "  ↓ (with signature)"
echo -e "${PURPLE}payment-gateway${NC} (40003) 【Orchestrator】"
echo "  ├─→ ${CYAN}risk-service${NC} (40006) - Risk assessment"
echo "  ├─→ ${CYAN}order-service${NC} (40004) - Order creation"
echo "  ├─→ ${CYAN}channel-adapter${NC} (40005) - Payment channel routing"
echo "  │    ├─→ Stripe API (external)"
echo "  │    ├─→ PayPal API (external, planned)"
echo "  │    └─→ Crypto API (external, planned)"
echo "  └─→ ${CYAN}accounting-service${NC} (40007) - Transaction recording"
echo ""
echo -e "${YELLOW}Async Callback (Webhook)${NC}"
echo "  ↓"
echo -e "${PURPLE}payment-gateway${NC}"
echo "  ├─→ ${CYAN}order-service${NC} - Update order status"
echo "  ├─→ ${CYAN}notification-service${NC} (40008) - Send notifications"
echo "  ├─→ ${CYAN}accounting-service${NC} - Update ledger"
echo "  └─→ ${CYAN}analytics-service${NC} (40009) - Update statistics"
echo ""

#######################################
# 3. 管理平台依赖 (Admin Platform)
#######################################
echo -e "${BOLD}${BLUE}[Admin Platform Dependencies]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${GREEN}Admin Portal${NC} (5173)"
echo "  ├─→ ${CYAN}admin-service${NC} (40001) - Admin user/role/permission"
echo "  ├─→ ${CYAN}merchant-service${NC} (40002) - Merchant management"
echo "  ├─→ ${CYAN}payment-gateway${NC} (40003) - Payment monitoring"
echo "  ├─→ ${CYAN}order-service${NC} (40004) - Order queries"
echo "  ├─→ ${CYAN}risk-service${NC} (40006) - Risk rules config"
echo "  ├─→ ${CYAN}accounting-service${NC} (40007) - Accounting reports"
echo "  ├─→ ${CYAN}analytics-service${NC} (40009) - Dashboard statistics"
echo "  ├─→ ${CYAN}kyc-service${NC} (40015) - KYC verification"
echo "  ├─→ ${CYAN}withdrawal-service${NC} (40014) - Withdrawal approval"
echo "  └─→ ${CYAN}dispute-service${NC} (40017) - Dispute handling"
echo ""

#######################################
# 4. 商户平台依赖 (Merchant Platform)
#######################################
echo -e "${BOLD}${BLUE}[Merchant Platform Dependencies]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${GREEN}Merchant Portal${NC} (5174)"
echo "  ├─→ ${CYAN}merchant-auth-service${NC} (40011) - API key management"
echo "  ├─→ ${CYAN}merchant-config-service${NC} (40012) - Merchant settings"
echo "  ├─→ ${CYAN}merchant-service${NC} (40002) - Merchant profile"
echo "  ├─→ ${CYAN}payment-gateway${NC} (40003) - Payment creation/query"
echo "  ├─→ ${CYAN}order-service${NC} (40004) - Order queries"
echo "  ├─→ ${CYAN}settlement-service${NC} (40013) - Settlement reports"
echo "  ├─→ ${CYAN}withdrawal-service${NC} (40014) - Withdrawal requests"
echo "  ├─→ ${CYAN}reconciliation-service${NC} (40018) - Reconciliation"
echo "  ├─→ ${CYAN}dispute-service${NC} (40017) - Dispute submission"
echo "  └─→ ${CYAN}merchant-limit-service${NC} (40022) - Transaction limits"
echo ""

#######################################
# 5. 财务流程 (Financial Flow)
#######################################
echo -e "${BOLD}${BLUE}[Financial Processing Flow]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${PURPLE}accounting-service${NC} (40007) 【Double-Entry Ledger】"
echo "  ↑ receives events from:"
echo "  ├── payment-gateway (payment success/failure)"
echo "  ├── withdrawal-service (withdrawal requests)"
echo "  └── settlement-service (settlement execution)"
echo ""
echo -e "${PURPLE}settlement-service${NC} (40013) 【Settlement Execution】"
echo "  ├─→ ${CYAN}accounting-service${NC} - Query balance"
echo "  ├─→ ${CYAN}merchant-service${NC} - Get merchant bank info"
echo "  └─→ ${CYAN}notification-service${NC} - Send settlement notice"
echo ""
echo -e "${PURPLE}reconciliation-service${NC} (40018) 【Reconciliation】"
echo "  ├─→ ${CYAN}order-service${NC} - Internal order data"
echo "  ├─→ ${CYAN}channel-adapter${NC} - External channel data"
echo "  └─→ ${CYAN}accounting-service${NC} - Ledger verification"
echo ""

#######################################
# 6. 风控与合规 (Risk & Compliance)
#######################################
echo -e "${BOLD}${BLUE}[Risk & Compliance Flow]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${PURPLE}risk-service${NC} (40006) 【Risk Engine】"
echo "  ├─→ ${CYAN}merchant-service${NC} - Merchant risk profile"
echo "  ├─→ ${CYAN}merchant-limit-service${NC} - Transaction limits check"
echo "  └─→ ${CYAN}order-service${NC} - Historical transaction analysis"
echo ""
echo -e "${PURPLE}kyc-service${NC} (40015) 【KYC Verification】"
echo "  ├─→ ${CYAN}merchant-service${NC} - Update merchant KYC status"
echo "  └─→ ${CYAN}notification-service${NC} - KYC status notifications"
echo ""
echo -e "${PURPLE}dispute-service${NC} (40017) 【Dispute Management】"
echo "  ├─→ ${CYAN}order-service${NC} - Get order details"
echo "  ├─→ ${CYAN}payment-gateway${NC} - Get payment details"
echo "  └─→ ${CYAN}accounting-service${NC} - Refund processing"
echo ""

#######################################
# 7. 支撑服务 (Supporting Services)
#######################################
echo -e "${BOLD}${BLUE}[Supporting Services]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${PURPLE}config-service${NC} (40010) 【Configuration Management】"
echo "  └─ Provides config to all services (system settings)"
echo ""
echo -e "${PURPLE}notification-service${NC} (40008) 【Notification Hub】"
echo "  ├─ Email notifications (SMTP/Mailgun)"
echo "  ├─ SMS notifications (Twilio/Aliyun)"
echo "  └─ Webhook callbacks (merchant endpoints)"
echo ""
echo -e "${PURPLE}analytics-service${NC} (40009) 【Data Analytics】"
echo "  ├─→ ${CYAN}order-service${NC} - Order statistics"
echo "  ├─→ ${CYAN}payment-gateway${NC} - Payment statistics"
echo "  ├─→ ${CYAN}merchant-service${NC} - Merchant statistics"
echo "  └─→ ${CYAN}accounting-service${NC} - Financial reports"
echo ""
echo -e "${PURPLE}cashier-service${NC} (40016) 【Cashier Page】"
echo "  ├─→ ${CYAN}payment-gateway${NC} - Payment creation"
echo "  └─→ ${CYAN}channel-adapter${NC} - Payment UI rendering"
echo ""

#######################################
# 8. 服务分层总结
#######################################
echo -e "${BOLD}${BLUE}[Service Layers Summary]${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BOLD}Layer 0: Infrastructure${NC}"
echo "  PostgreSQL, Redis, Kafka, Prometheus, Jaeger"
echo ""
echo -e "${BOLD}Layer 1: Gateway & Core${NC} (Critical Path)"
echo "  payment-gateway, order-service, channel-adapter"
echo ""
echo -e "${BOLD}Layer 2: Business Services${NC}"
echo "  risk-service, accounting-service, merchant-service"
echo ""
echo -e "${BOLD}Layer 3: Platform Services${NC}"
echo "  admin-service, merchant-auth-service, merchant-config-service"
echo ""
echo -e "${BOLD}Layer 4: Supporting Services${NC}"
echo "  notification-service, analytics-service, config-service"
echo ""
echo -e "${BOLD}Layer 5: Financial Services${NC}"
echo "  settlement-service, withdrawal-service, reconciliation-service"
echo ""
echo -e "${BOLD}Layer 6: Compliance Services${NC}"
echo "  kyc-service, dispute-service, merchant-limit-service"
echo ""
echo -e "${BOLD}Layer 7: Frontend Applications${NC}"
echo "  admin-portal, merchant-portal, website, cashier-service"
echo ""

#######################################
# Footer
#######################################
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}Total Services: 19 microservices + 3 frontend apps${NC}"
echo -e "${BOLD}External Dependencies: Stripe, PayPal (planned), Crypto (planned)${NC}"
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
