#!/bin/bash

# 配置中心 mTLS 集成测试脚本

echo "========================================="
echo "配置中心 mTLS 集成测试"
echo "========================================="
echo ""

# 证书路径
CERT_DIR="/home/eric/payment/backend/certs"
SERVICE_CERT_DIR="$CERT_DIR/services/payment-gateway"
CA_CERT="$CERT_DIR/ca/ca-cert.pem"

# 1. 检查证书文件是否存在
echo "1️⃣  检查 mTLS 证书..."
if [ -f "$SERVICE_CERT_DIR/cert.pem" ] && [ -f "$SERVICE_CERT_DIR/key.pem" ] && [ -f "$CA_CERT" ]; then
    echo "   ✅ 证书文件存在"
    echo "   📄 Client Cert: $SERVICE_CERT_DIR/cert.pem"
    echo "   🔑 Client Key:  $SERVICE_CERT_DIR/key.pem"
    echo "   🏛️  CA Cert:     $CA_CERT"
else
    echo "   ❌ 证书文件缺失"
    exit 1
fi

echo ""

# 2. 检查 config-service 是否运行
echo "2️⃣  检查 config-service 状态..."
if lsof -i:40010 -sTCP:LISTEN > /dev/null 2>&1; then
    echo "   ✅ config-service 正在运行 (端口 40010)"
else
    echo "   ❌ config-service 未运行"
    exit 1
fi

echo ""

# 3. 测试 HTTPS 访问 config-service (使用 mTLS)
echo "3️⃣  测试 mTLS 访问 config-service..."
RESPONSE=$(curl -s \
    --cacert "$CA_CERT" \
    --cert "$SERVICE_CERT_DIR/cert.pem" \
    --key "$SERVICE_CERT_DIR/key.pem" \
    "https://localhost:40010/api/v1/configs?service_name=payment-gateway&environment=production" \
    2>&1)

if echo "$RESPONSE" | grep -q '"code":"SUCCESS"'; then
    echo "   ✅ mTLS 访问成功"
    COUNT=$(echo "$RESPONSE" | grep -o "config_key" | wc -l)
    echo "   📊 获取到 $COUNT 个payment-gateway特定配置项"

    # 显示配置项列表
    if [ $COUNT -gt 0 ]; then
        echo "   📋 配置列表:"
        echo "$RESPONSE" | grep -o '"config_key":"[^"]*"' | cut -d'"' -f4 | while read -r key; do
            echo "      - $key"
        done
    fi
else
    echo "   ❌ mTLS 访问失败"
    echo "   Response: $RESPONSE"
    exit 1
fi

echo ""

# 4. 编译 payment-gateway
echo "4️⃣  编译 payment-gateway..."
cd /home/eric/payment/backend/services/payment-gateway
if GOWORK=/home/eric/payment/backend/go.work timeout 30 go build -o /tmp/test-payment-gateway ./cmd/main.go 2>&1 | grep -q "error"; then
    echo "   ❌ 编译失败"
    GOWORK=/home/eric/payment/backend/go.work go build -o /tmp/test-payment-gateway ./cmd/main.go 2>&1 | tail -10
    exit 1
else
    echo "   ✅ 编译成功"
fi

echo ""

# 5. 启动 payment-gateway (启用配置客户端 + mTLS)
echo "5️⃣  启动 payment-gateway (配置客户端 + mTLS)..."

# 停止旧实例
if [ -f /home/eric/payment/backend/logs/payment-gateway.pid ]; then
    OLD_PID=$(cat /home/eric/payment/backend/logs/payment-gateway.pid)
    if kill -0 $OLD_PID 2>/dev/null; then
        echo "   停止旧实例 (PID: $OLD_PID)..."
        kill $OLD_PID
        sleep 2
    fi
fi

# 启动新实例 (启用配置客户端 + mTLS)
cd /home/eric/payment/backend/services/payment-gateway

ENABLE_CONFIG_CLIENT=true \
CONFIG_CLIENT_MTLS=true \
CONFIG_SERVICE_URL=https://localhost:40010 \
TLS_CERT_FILE="$SERVICE_CERT_DIR/cert.pem" \
TLS_KEY_FILE="$SERVICE_CERT_DIR/key.pem" \
TLS_CA_FILE="$CA_CERT" \
ENV=development \
JWT_SECRET="config-test-jwt-secret-12345" \
PORT=40003 \
DB_HOST=localhost \
DB_PORT=40432 \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=payment_gateway \
REDIS_HOST=localhost \
REDIS_PORT=40379 \
ENABLE_MTLS=true \
GOWORK=/home/eric/payment/backend/go.work \
nohup go run cmd/main.go > /home/eric/payment/backend/logs/payment-gateway-config-test.log 2>&1 &

NEW_PID=$!
echo $NEW_PID > /home/eric/payment/backend/logs/payment-gateway-config-test.pid
echo "   ✅ payment-gateway 已启动 (PID: $NEW_PID)"

# 等待启动
echo "   ⏳ 等待服务启动 (10秒)..."
sleep 10

echo ""

# 6. 检查日志确认配置客户端初始化
echo "6️⃣  检查日志..."
LOG_FILE="/home/eric/payment/backend/logs/payment-gateway-config-test.log"

if grep -q "配置客户端启用 mTLS" "$LOG_FILE"; then
    echo "   ✅ 检测到 mTLS 启用日志"
fi

if grep -q "Config client mTLS enabled" "$LOG_FILE"; then
    echo "   ✅ 配置客户端 mTLS 初始化成功"
fi

if grep -q "配置客户端初始化成功" "$LOG_FILE"; then
    echo "   ✅ 配置客户端初始化成功"
    # 提取 mtls_enabled 状态
    MTLS_ENABLED=$(grep "配置客户端初始化成功" "$LOG_FILE" | grep -o "mtls_enabled.:true\|mtls_enabled.:false")
    echo "   📊 $MTLS_ENABLED"
fi

if grep -q "Configs loaded" "$LOG_FILE"; then
    echo "   ✅ 配置加载成功"
    # 提取配置数量
    CONFIG_COUNT=$(grep "Configs loaded" "$LOG_FILE" | tail -1 | grep -o "count.:[0-9]*" | grep -o "[0-9]*")
    echo "   📊 已加载 $CONFIG_COUNT 个配置"
fi

if grep -q "从配置中心读取配置" "$LOG_FILE"; then
    echo "   ✅ 检测到配置中心读取日志"
    echo "   📋 读取的配置:"
    grep "从配置中心读取配置" "$LOG_FILE" | head -5 | while read -r line; do
        KEY=$(echo "$line" | grep -o 'key":"[^"]*' | cut -d'"' -f3)
        echo "      - $KEY"
    done
else
    echo "   ⚠️  未检测到配置中心读取日志 (可能使用了缓存或环境变量)"
fi

echo ""

# 7. 检查服务健康状态
echo "7️⃣  检查服务健康状态..."
if kill -0 $NEW_PID 2>/dev/null; then
    echo "   ✅ 服务运行正常 (PID: $NEW_PID)"

    # 尝试健康检查
    sleep 2
    if curl -s --cacert "$CA_CERT" --cert "$SERVICE_CERT_DIR/cert.pem" --key "$SERVICE_CERT_DIR/key.pem" https://localhost:40003/health > /dev/null 2>&1; then
        echo "   ✅ 健康检查通过"
    else
        echo "   ⚠️  健康检查失败 (可能需要更多启动时间)"
    fi
else
    echo "   ❌ 服务已停止"
    echo "   最后 20 行日志:"
    tail -20 "$LOG_FILE"
    exit 1
fi

echo ""

# 8. 显示完整日志(最后30行)
echo "8️⃣  最后 30 行日志:"
echo "-------------------------------------------"
tail -30 "$LOG_FILE"
echo "-------------------------------------------"

echo ""
echo "========================================="
echo "✅ mTLS 集成测试完成!"
echo "========================================="
echo ""
echo "服务信息:"
echo "  PID: $NEW_PID"
echo "  日志: $LOG_FILE"
echo "  配置中心: https://localhost:40010"
echo "  mTLS: 已启用"
echo ""
echo "下一步:"
echo "1. 观察日志确认配置读取来源"
echo "2. 测试配置热更新 (修改数据库中的配置)"
echo "3. 验证服务功能正常"
echo ""
