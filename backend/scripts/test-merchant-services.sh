#!/bin/bash

# Merchant Services Testing Script
# Starts both services and tests basic API endpoints

set -e

echo "=================================================="
echo "Merchant Services Testing"
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if services are compiled
if [ ! -f "/tmp/merchant-policy-service" ]; then
    echo -e "${YELLOW}Compiling merchant-policy-service...${NC}"
    cd /home/eric/payment/backend/services/merchant-policy-service
    go build -o /tmp/merchant-policy-service ./cmd/main.go
    echo -e "${GREEN}✅ Compiled merchant-policy-service${NC}"
fi

if [ ! -f "/tmp/merchant-quota-service" ]; then
    echo -e "${YELLOW}Compiling merchant-quota-service...${NC}"
    cd /home/eric/payment/backend/services/merchant-quota-service
    go build -o /tmp/merchant-quota-service ./cmd/main.go
    echo -e "${GREEN}✅ Compiled merchant-quota-service${NC}"
fi

echo ""
echo "Starting services in background..."
echo "-----------------------------------"

# Start merchant-policy-service
cd /home/eric/payment/backend/services/merchant-policy-service
PORT=40012 DB_NAME=payment_merchant_policy /tmp/merchant-policy-service > /tmp/policy-service.log 2>&1 &
POLICY_PID=$!
echo -e "${GREEN}✅ Started merchant-policy-service (PID: $POLICY_PID, Port: 40012)${NC}"

# Start merchant-quota-service
cd /home/eric/payment/backend/services/merchant-quota-service
PORT=40022 DB_NAME=payment_merchant_quota /tmp/merchant-quota-service > /tmp/quota-service.log 2>&1 &
QUOTA_PID=$!
echo -e "${GREEN}✅ Started merchant-quota-service (PID: $QUOTA_PID, Port: 40022)${NC}"

echo ""
echo "Waiting for services to start (10 seconds)..."
sleep 10

# Test health endpoints
echo ""
echo "Testing Health Endpoints..."
echo "-----------------------------------"

if curl -s http://localhost:40012/health > /dev/null; then
    echo -e "${GREEN}✅ merchant-policy-service health check passed${NC}"
else
    echo -e "${RED}❌ merchant-policy-service health check failed${NC}"
fi

if curl -s http://localhost:40022/health > /dev/null; then
    echo -e "${GREEN}✅ merchant-quota-service health check passed${NC}"
else
    echo -e "${RED}❌ merchant-quota-service health check failed${NC}"
fi

echo ""
echo "=================================================="
echo "Services Started Successfully!"
echo "=================================================="
echo ""
echo "Service URLs:"
echo "  Policy Service:  http://localhost:40012"
echo "  Quota Service:   http://localhost:40022"
echo ""
echo "Swagger UI:"
echo "  Policy Service:  http://localhost:40012/swagger/index.html"
echo "  Quota Service:   http://localhost:40022/swagger/index.html"
echo ""
echo "Logs:"
echo "  Policy Service:  tail -f /tmp/policy-service.log"
echo "  Quota Service:   tail -f /tmp/quota-service.log"
echo ""
echo "To stop services:"
echo "  kill $POLICY_PID $QUOTA_PID"
echo ""
echo "PIDs saved to /tmp/merchant-services.pid"
echo "$POLICY_PID" > /tmp/merchant-services.pid
echo "$QUOTA_PID" >> /tmp/merchant-services.pid
