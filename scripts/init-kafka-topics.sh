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
echo "创建支付相关 Topics"
echo "========================================="

# 支付事件
create_topic "payment.created"
create_topic "payment.success"
create_topic "payment.failed"
create_topic "payment.refund.created"
create_topic "payment.refund.success"
create_topic "payment.refund.failed"

# 订单事件
create_topic "order.created"
create_topic "order.updated"
create_topic "order.cancelled"
create_topic "order.completed"

# 账务事件
create_topic "accounting.transaction.created"
create_topic "accounting.balance.updated"
create_topic "accounting.settlement.created"
create_topic "accounting.settlement.completed"

# 风控事件
create_topic "risk.check.started"
create_topic "risk.check.completed"
create_topic "risk.alert.high"
create_topic "risk.alert.critical"

# 通知事件
create_topic "notification.email"
create_topic "notification.sms"
create_topic "notification.webhook"

# 商户事件
create_topic "merchant.created"
create_topic "merchant.updated"
create_topic "merchant.approved"
create_topic "merchant.frozen"

# 提现事件
create_topic "withdrawal.created"
create_topic "withdrawal.approved"
create_topic "withdrawal.rejected"
create_topic "withdrawal.completed"

# Saga 分布式事务
create_topic "saga.payment.start"
create_topic "saga.payment.compensate"

# 分析事件
create_topic "analytics.events"

# 审计日志
create_topic "audit.logs" 6

# 死信队列
create_topic "dlq.payment" 1
create_topic "dlq.notification" 1

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
