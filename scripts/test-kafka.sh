#!/bin/bash

# Kafka 测试脚本
# 测试 Kafka 生产者和消费者功能

set -e

KAFKA_HOST="${KAFKA_HOST:-localhost:40093}"
TEST_TOPIC="${TEST_TOPIC:-test.payment.event}"

echo "========================================="
echo "Kafka 功能测试"
echo "========================================="
echo "Kafka 地址: $KAFKA_HOST"
echo "测试 Topic: $TEST_TOPIC"
echo ""

# 检查 Kafka 连接
echo "1️⃣  检查 Kafka 连接..."
if docker exec payment-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 >/dev/null 2>&1; then
    echo "✅ Kafka 连接成功"
else
    echo "❌ Kafka 连接失败"
    exit 1
fi

# 创建测试 topic
echo ""
echo "2️⃣  创建测试 Topic..."
if docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list | grep -q "^${TEST_TOPIC}$"; then
    echo "⏩ Topic 已存在,删除旧 Topic"
    docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --delete --topic "$TEST_TOPIC" >/dev/null 2>&1
    sleep 2
fi

docker exec payment-kafka kafka-topics \
    --bootstrap-server localhost:9092 \
    --create \
    --topic "$TEST_TOPIC" \
    --partitions 3 \
    --replication-factor 1 >/dev/null 2>&1

echo "✅ Topic 创建成功"

# 列出所有 topics
echo ""
echo "3️⃣  当前所有 Topics:"
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list

# 发送测试消息
echo ""
echo "4️⃣  发送测试消息..."
TEST_MESSAGES=(
    '{"event":"payment.created","payment_id":"PAY001","amount":10000,"currency":"USD","timestamp":"2025-01-01T00:00:00Z"}'
    '{"event":"payment.success","payment_id":"PAY001","status":"success","timestamp":"2025-01-01T00:00:05Z"}'
    '{"event":"order.created","order_id":"ORD001","payment_id":"PAY001","timestamp":"2025-01-01T00:00:01Z"}'
)

for msg in "${TEST_MESSAGES[@]}"; do
    echo "$msg" | docker exec -i payment-kafka kafka-console-producer \
        --bootstrap-server localhost:9092 \
        --topic "$TEST_TOPIC" >/dev/null 2>&1
    echo "  ✅ 发送: $(echo $msg | cut -c1-60)..."
done

# 读取消息
echo ""
echo "5️⃣  读取消息 (前 5 条)..."
echo ""
docker exec payment-kafka kafka-console-consumer \
    --bootstrap-server localhost:9092 \
    --topic "$TEST_TOPIC" \
    --from-beginning \
    --max-messages 5 \
    --timeout-ms 3000 2>/dev/null | jq -C . || true

# 获取 topic 详情
echo ""
echo "6️⃣  Topic 详细信息:"
docker exec payment-kafka kafka-topics \
    --bootstrap-server localhost:9092 \
    --describe \
    --topic "$TEST_TOPIC"

# 获取消费者组信息
echo ""
echo "7️⃣  消费者组列表:"
docker exec payment-kafka kafka-consumer-groups \
    --bootstrap-server localhost:9092 \
    --list

# 性能测试 (可选)
echo ""
read -p "是否执行性能测试? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "8️⃣  性能测试 (发送 10000 条消息)..."
    docker exec payment-kafka kafka-producer-perf-test \
        --topic "$TEST_TOPIC" \
        --num-records 10000 \
        --record-size 256 \
        --throughput -1 \
        --producer-props bootstrap.servers=localhost:9092

    echo ""
    echo "9️⃣  性能测试 (消费测试)..."
    docker exec payment-kafka kafka-consumer-perf-test \
        --bootstrap-server localhost:9092 \
        --topic "$TEST_TOPIC" \
        --messages 10000 \
        --threads 1
fi

echo ""
echo "========================================="
echo "✅ Kafka 测试完成"
echo "========================================="
echo ""
echo "有用的命令:"
echo ""
echo "# 查看所有 topics"
echo "docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list"
echo ""
echo "# 查看 topic 详情"
echo "docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic $TEST_TOPIC"
echo ""
echo "# 消费消息 (从头开始)"
echo "docker exec payment-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic $TEST_TOPIC --from-beginning"
echo ""
echo "# 生产消息"
echo 'echo "test message" | docker exec -i payment-kafka kafka-console-producer --bootstrap-server localhost:9092 --topic '$TEST_TOPIC
echo ""
echo "# 查看消费者组"
echo "docker exec payment-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list"
echo ""
echo "# 访问 Kafka UI"
echo "http://localhost:40080"
echo ""
