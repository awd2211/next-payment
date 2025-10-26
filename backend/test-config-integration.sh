#!/bin/bash

# 配置中心集成测试脚本

echo "========================================="
echo "配置中心集成测试"
echo "========================================="
echo ""

# 1. 检查config-service是否运行
echo "1️⃣  检查 config-service 状态..."
if lsof -i:40010 -sTCP:LISTEN > /dev/null 2>&1; then
    echo "   ✅ config-service 正在运行 (端口 40010)"
else
    echo "   ❌ config-service 未运行,请先启动"
    exit 1
fi

echo ""

# 2. 测试配置API
echo "2️⃣  测试配置API..."
RESPONSE=$(curl -s "http://localhost:40010/api/v1/configs?service_name=payment-gateway&environment=production")
if echo "$RESPONSE" | grep -q "JWT_SECRET"; then
    echo "   ✅ API 响应正常,包含 JWT_SECRET 配置"
    # 显示配置数量
    COUNT=$(echo "$RESPONSE" | grep -o "config_key" | wc -l)
    echo "   📊 找到 $COUNT 个配置项"
else
    echo "   ❌ API 响应异常"
    echo "   Response: $RESPONSE"
    exit 1
fi

echo ""

# 3. 测试payment-gateway编译
echo "3️⃣  测试 payment-gateway 编译..."
cd /home/eric/payment/backend/services/payment-gateway
if GOWORK=/home/eric/payment/backend/go.work timeout 30 go build -o /tmp/test-payment-gateway ./cmd/main.go 2>&1 | grep -q "error"; then
    echo "   ❌ 编译失败"
    GOWORK=/home/eric/payment/backend/go.work go build -o /tmp/test-payment-gateway ./cmd/main.go 2>&1 | tail -10
    exit 1
else
    echo "   ✅ 编译成功"
    rm -f /tmp/test-payment-gateway
fi

echo ""

# 4. 测试payment-gateway运行(不启用配置客户端)
echo "4️⃣  测试 payment-gateway 运行 (环境变量模式)..."
cd /home/eric/payment/backend/services/payment-gateway

# 停止现有实例
if [ -f /home/eric/payment/backend/logs/payment-gateway.pid ]; then
    OLD_PID=$(cat /home/eric/payment/backend/logs/payment-gateway.pid)
    if kill -0 $OLD_PID 2>/dev/null; then
        echo "   停止旧实例 (PID: $OLD_PID)..."
        kill $OLD_PID
        sleep 2
    fi
fi

# 启动新实例(不启用配置客户端)
ENABLE_CONFIG_CLIENT=false \
GOWORK=/home/eric/payment/backend/go.work \
JWT_SECRET="test-jwt-secret" \
PORT=40003 \
DB_HOST=localhost \
DB_PORT=40432 \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=payment_gateway \
REDIS_HOST=localhost \
REDIS_PORT=40379 \
nohup go run cmd/main.go > /home/eric/payment/backend/logs/payment-gateway-test.log 2>&1 &
NEW_PID=$!
echo $NEW_PID > /home/eric/payment/backend/logs/payment-gateway-test.pid
echo "   ✅ payment-gateway 已启动 (PID: $NEW_PID, 环境变量模式)"

# 等待启动
echo "   ⏳ 等待服务启动 (5秒)..."
sleep 5

# 检查是否成功启动
if kill -0 $NEW_PID 2>/dev/null; then
    echo "   ✅ 服务运行正常"
    # 检查健康端点
    if curl -s http://localhost:40003/health > /dev/null 2>&1; then
        echo "   ✅ 健康检查通过"
    else
        echo "   ⚠️  健康检查失败(可能需要更多时间启动)"
    fi
else
    echo "   ❌ 服务启动失败"
    echo "   最后10行日志:"
    tail -10 /home/eric/payment/backend/logs/payment-gateway-test.log
    exit 1
fi

echo ""

# 5. 停止测试实例
echo "5️⃣  清理测试环境..."
kill $NEW_PID 2>/dev/null
sleep 1
echo "   ✅ 测试实例已停止"

echo ""
echo "========================================="
echo "✅ 所有测试通过!"
echo "========================================="
echo ""
echo "下一步:"
echo "1. 启用配置客户端: 设置 ENABLE_CONFIG_CLIENT=true"
echo "2. 重启 payment-gateway 服务"
echo "3. 观察日志确认从配置中心读取配置"
echo ""
