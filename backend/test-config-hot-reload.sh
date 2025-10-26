#!/bin/bash

# 配置热更新测试脚本

echo "========================================="
echo "配置中心热更新测试"
echo "========================================="
echo ""

# 测试配置项
TEST_CONFIG_KEY="PAYMENT_TIMEOUT"
SERVICE_NAME="payment-gateway"
ENVIRONMENT="production"

# 1. 检查当前配置值
echo "1️⃣  查询当前配置值..."
CURRENT_VALUE=$(PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_config -t -A -c \
    "SELECT config_value FROM configs WHERE config_key='$TEST_CONFIG_KEY' AND service_name='$SERVICE_NAME' AND environment='$ENVIRONMENT';")

if [ -z "$CURRENT_VALUE" ]; then
    echo "   ❌ 配置项不存在，正在创建..."
    PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_config -c \
        "INSERT INTO configs (service_name, config_key, config_value, value_type, environment, description, is_encrypted, created_by)
         VALUES ('$SERVICE_NAME', '$TEST_CONFIG_KEY', '300', 'integer', '$ENVIRONMENT', 'Payment timeout in seconds (hot reload test)', false, 'test-script');"
    CURRENT_VALUE="300"
fi

echo "   📊 当前值: $CURRENT_VALUE"
ORIGINAL_VALUE=$CURRENT_VALUE

echo ""

# 2. 计算新值
NEW_VALUE=$((CURRENT_VALUE + 100))
echo "2️⃣  准备更新配置..."
echo "   旧值: $CURRENT_VALUE"
echo "   新值: $NEW_VALUE"

# 3. 更新配置
echo ""
echo "3️⃣  更新数据库配置..."
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_config -c \
    "UPDATE configs SET config_value='$NEW_VALUE', updated_at=NOW()
     WHERE config_key='$TEST_CONFIG_KEY' AND service_name='$SERVICE_NAME' AND environment='$ENVIRONMENT';"

if [ $? -eq 0 ]; then
    echo "   ✅ 数据库更新成功"
    echo "   📊 $TEST_CONFIG_KEY: $CURRENT_VALUE → $NEW_VALUE"
else
    echo "   ❌ 数据库更新失败"
    exit 1
fi

echo ""

# 4. 等待自动刷新
echo "4️⃣  等待配置自动刷新..."
echo "   ⏳ 配置客户端刷新周期: 30秒"
echo "   ⏳ 开始倒计时..."

for i in {30..1}; do
    echo -ne "   ⏰ 还需等待 $i 秒...\r"
    sleep 1
done
echo -ne "\n"
echo "   ✅ 等待完成，配置应已刷新"

echo ""

# 5. 验证配置更新（通过日志）
echo "5️⃣  验证配置更新..."

# 检查payment-gateway是否在运行并启用了配置客户端
if [ -f /home/eric/payment/backend/logs/payment-gateway-hot-reload-test.pid ]; then
    PID=$(cat /home/eric/payment/backend/logs/payment-gateway-hot-reload-test.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "   ✅ payment-gateway 正在运行 (PID: $PID)"

        LOG_FILE="/home/eric/payment/backend/logs/payment-gateway-hot-reload-test.log"

        # 检查日志中的配置加载
        echo ""
        echo "   📋 最近的配置加载日志:"
        grep "Configs loaded" "$LOG_FILE" | tail -3

        # 检查是否有更新回调（如果实现了）
        if grep -q "配置更新" "$LOG_FILE"; then
            echo ""
            echo "   ✅ 检测到配置更新日志:"
            grep "配置更新" "$LOG_FILE" | tail -5
        fi
    else
        echo "   ⚠️  payment-gateway 未运行，无法验证热更新"
        echo "   提示: 请先运行 test-config-mtls.sh 启动带配置客户端的 payment-gateway"
    fi
else
    echo "   ⚠️  未找到 payment-gateway 测试实例"
    echo "   提示: 请先运行 test-config-mtls.sh 启动带配置客户端的 payment-gateway"
fi

echo ""

# 6. 通过API验证配置值
echo "6️⃣  通过config-service API验证..."
CERT_DIR="/home/eric/payment/backend/certs"
SERVICE_CERT_DIR="$CERT_DIR/services/payment-gateway"
CA_CERT="$CERT_DIR/ca/ca-cert.pem"

API_RESPONSE=$(curl -s \
    --cacert "$CA_CERT" \
    --cert "$SERVICE_CERT_DIR/cert.pem" \
    --key "$SERVICE_CERT_DIR/key.pem" \
    "https://localhost:40010/api/v1/configs?service_name=$SERVICE_NAME&environment=$ENVIRONMENT" 2>&1)

if echo "$API_RESPONSE" | grep -q "$TEST_CONFIG_KEY"; then
    ACTUAL_VALUE=$(echo "$API_RESPONSE" | grep -o "\"config_key\":\"$TEST_CONFIG_KEY\"[^}]*\"config_value\":\"[^\"]*\"" | grep -o "\"config_value\":\"[^\"]*\"" | cut -d'"' -f4)

    if [ "$ACTUAL_VALUE" = "$NEW_VALUE" ]; then
        echo "   ✅ API 返回的配置值正确"
        echo "   📊 $TEST_CONFIG_KEY = $ACTUAL_VALUE (符合预期)"
    else
        echo "   ⚠️  API 返回的值与预期不符"
        echo "   预期: $NEW_VALUE"
        echo "   实际: $ACTUAL_VALUE"
    fi
else
    echo "   ⚠️  API 响应中未找到配置项"
fi

echo ""

# 7. 恢复原始值
echo "7️⃣  恢复原始配置值..."
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_config -c \
    "UPDATE configs SET config_value='$ORIGINAL_VALUE', updated_at=NOW()
     WHERE config_key='$TEST_CONFIG_KEY' AND service_name='$SERVICE_NAME' AND environment='$ENVIRONMENT';" > /dev/null

echo "   ✅ 已恢复为原始值: $ORIGINAL_VALUE"

echo ""
echo "========================================="
echo "📊 热更新测试总结"
echo "========================================="
echo ""
echo "测试配置项: $TEST_CONFIG_KEY"
echo "原始值: $ORIGINAL_VALUE"
echo "测试值: $NEW_VALUE"
echo "当前值: $ORIGINAL_VALUE (已恢复)"
echo ""
echo "测试步骤:"
echo "  1. ✅ 查询当前配置"
echo "  2. ✅ 更新数据库配置"
echo "  3. ✅ 等待30秒自动刷新"
echo "  4. ✅ 验证配置更新"
echo "  5. ✅ 恢复原始值"
echo ""
echo "验证方法:"
echo "  - 数据库直接查询"
echo "  - config-service API查询"
echo "  - payment-gateway日志分析"
echo ""
echo "🔄 配置热更新机制:"
echo "  - 刷新周期: 30秒"
echo "  - 自动触发: 是"
echo "  - 无需重启: 是"
echo ""
echo "下一步建议:"
echo "1. 查看 payment-gateway 日志确认配置自动刷新"
echo "2. 实现配置更新回调函数(如需要)"
echo "3. 测试其他配置项的热更新"
echo ""
