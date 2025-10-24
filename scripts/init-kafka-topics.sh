#!/bin/bash

# Kafka Topics 初始化脚本
# 创建支付平台所需的所有 Kafka topics

set -e

KAFKA_HOST="${KAFKA_HOST:-localhost:40093}"
PARTITIONS="${PARTITIONS:-3}"
REPLICATION_FACTOR="${REPLICATION_FACTOR:-1}"

echo "========================================="
echo "Kafka Topics 初始化"
echo "========================================="
echo "Kafka 地址: $KAFKA_HOST"
echo "分区数: $PARTITIONS"
echo "副本因子: $REPLICATION_FACTOR"
echo ""

# 等待 Kafka 就绪
echo "等待 Kafka 启动..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker exec payment-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 >/dev/null 2>&1; then
        echo "✅ Kafka 已就绪"
        break
    fi
    attempt=$((attempt + 1))
    echo "等待 Kafka... ($attempt/$max_attempts)"
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Kafka 启动超时"
    exit 1
fi

# Topic 创建函数
create_topic() {
    local topic_name=$1
    local partitions=${2:-$PARTITIONS}
    local replication=${3:-$REPLICATION_FACTOR}

    echo -n "创建 Topic: $topic_name ... "

    if docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list | grep -q "^${topic_name}$"; then
        echo "⏩ 已存在"
    else
        docker exec payment-kafka kafka-topics \
            --bootstrap-server localhost:9092 \
            --create \
            --topic "$topic_name" \
            --partitions "$partitions" \
            --replication-factor "$replication" \
            --config retention.ms=604800000 \
            --config segment.ms=86400000 >/dev/null 2>&1
        echo "✅ 创建成功"
    fi
}

echo ""
echo "========================================="
echo "创建聚合事件 Topics (Event-Driven Architecture)"
echo "========================================="

# 核心业务事件Topics (聚合模式 - 一个领域一个Topic)
create_topic "payment.events" 6               # 支付事件(高吞吐)
create_topic "payment.refund.events" 3        # 退款事件
create_topic "order.events" 3                 # 订单事件
create_topic "accounting.events" 3            # 财务事件
create_topic "settlement.events" 1            # 结算事件(低频)
create_topic "withdrawal.events" 1            # 提现事件(低频)
create_topic "merchant.events" 1              # 商户事件(低频)
create_topic "kyc.events" 1                   # KYC事件(低频)

echo ""
echo "========================================="
echo "创建内部通知 Topics"
echo "========================================="

# 通知服务内部Topics
create_topic "notifications.email" 3
create_topic "notifications.sms" 3
create_topic "notifications.webhook" 3

echo ""
echo "========================================="
echo "创建系统 Topics"
echo "========================================="

# Saga 分布式事务
create_topic "saga.payment.start"
create_topic "saga.payment.compensate"

# 分析事件
create_topic "analytics.events" 6

# 审计日志
create_topic "audit.logs" 6

# 死信队列
create_topic "dlq.payment" 1
create_topic "dlq.notification" 1

echo ""
echo "========================================="
echo "创建旧版单事件 Topics (向后兼容,可选)"
echo "========================================="

# 旧版单事件Topics (如果需要向后兼容)
# create_topic "payment.created"
# create_topic "payment.success"
# create_topic "payment.failed"
# 等等...

echo ""
echo "========================================="
echo "Topic 列表"
echo "========================================="
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list

echo ""
echo "========================================="
echo "Topic 详细信息"
echo "========================================="
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic payment.created

echo ""
echo "✅ Kafka Topics 初始化完成"
echo ""
echo "访问 Kafka UI: http://localhost:40080"
echo "使用 kafka-console-consumer 消费消息:"
echo "  docker exec -it payment-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic payment.created --from-beginning"
echo ""
