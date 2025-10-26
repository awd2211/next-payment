#!/bin/bash

# Admin Service BFF 启动脚本
# 用途: 启动 Admin Service 作为 BFF,聚合6个后端微服务

echo "=================================================="
echo "  启动 Admin Service (BFF 模式)"
echo "=================================================="

# 切换到 Admin Service 目录
cd /home/eric/payment/backend/services/admin-service

# 设置环境变量
export GOWORK=/home/eric/payment/backend/go.work
export JWT_SECRET="payment-platform-secret-key-2024"
export DB_HOST="localhost"
export DB_PORT="40432"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="payment_admin"
export PORT="40001"
export REDIS_HOST="localhost"
export REDIS_PORT="40379"
export ENABLE_MTLS="true"
export TLS_CERT_FILE="/home/eric/payment/backend/certs/services/admin-service/cert.pem"
export TLS_KEY_FILE="/home/eric/payment/backend/certs/services/admin-service/key.pem"
export TLS_CA_FILE="/home/eric/payment/backend/certs/ca/ca-cert.pem"

# BFF 后端服务地址
export CONFIG_SERVICE_URL="http://localhost:40010"
export RISK_SERVICE_URL="http://localhost:40006"
export KYC_SERVICE_URL="http://localhost:40015"
export MERCHANT_SERVICE_URL="http://localhost:40002"
export ANALYTICS_SERVICE_URL="http://localhost:40009"
export LIMIT_SERVICE_URL="http://localhost:40022"

echo ""
echo "环境变量配置:"
echo "  JWT_SECRET: $JWT_SECRET"
echo "  DB_NAME: $DB_NAME"
echo "  PORT: $PORT"
echo "  ENABLE_MTLS: $ENABLE_MTLS"
echo ""
echo "BFF 后端服务:"
echo "  Config Service:    $CONFIG_SERVICE_URL"
echo "  Risk Service:      $RISK_SERVICE_URL"
echo "  KYC Service:       $KYC_SERVICE_URL"
echo "  Merchant Service:  $MERCHANT_SERVICE_URL"
echo "  Analytics Service: $ANALYTICS_SERVICE_URL"
echo "  Limit Service:     $LIMIT_SERVICE_URL"
echo ""

# 检查日志目录
LOG_DIR="/home/eric/payment/backend/logs"
if [ ! -d "$LOG_DIR" ]; then
    echo "创建日志目录: $LOG_DIR"
    mkdir -p "$LOG_DIR"
fi

# 启动服务
echo "正在启动 Admin Service (BFF)..."
nohup go run cmd/main.go > "$LOG_DIR/admin-service.log" 2>&1 &
PID=$!
echo $PID > "$LOG_DIR/admin-service.pid"

echo ""
echo "✅ Admin Service (BFF) 已启动"
echo "   PID: $PID"
echo "   Port: $PORT"
echo "   Log: $LOG_DIR/admin-service.log"
echo ""

# 等待服务启动
echo "等待服务启动..."
sleep 3

# 检查服务状态
if lsof -i :$PORT -sTCP:LISTEN > /dev/null 2>&1; then
    echo "✅ Admin Service 正在监听端口 $PORT"
    echo ""
    echo "Swagger UI: http://localhost:$PORT/swagger/index.html"
    echo ""
    echo "BFF 路由测试:"
    echo "  curl -X GET 'http://localhost:$PORT/api/v1/admin/configs' -H 'Authorization: Bearer \$TOKEN'"
    echo "  curl -X GET 'http://localhost:$PORT/api/v1/admin/risk/rules' -H 'Authorization: Bearer \$TOKEN'"
    echo "  curl -X GET 'http://localhost:$PORT/api/v1/admin/kyc/documents/pending' -H 'Authorization: Bearer \$TOKEN'"
    echo ""
    echo "查看日志: tail -f $LOG_DIR/admin-service.log"
else
    echo "❌ Admin Service 启动失败"
    echo "查看日志: cat $LOG_DIR/admin-service.log"
    exit 1
fi

echo "=================================================="
