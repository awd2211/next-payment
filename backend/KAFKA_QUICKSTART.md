# Kafka 快速开始

## 🚀 一键启动

```bash
# 1. 启动 Kafka 基础设施
cd /home/eric/payment
docker compose up -d zookeeper kafka kafka-ui

# 2. 等待 Kafka 就绪 (约 30 秒)
docker compose ps

# 3. 初始化所有 Topics
./scripts/init-kafka-topics.sh

# 4. 测试 Kafka 功能
./scripts/test-kafka.sh
```

## 📊 访问 Kafka UI

打开浏览器访问: **http://localhost:40080**

可以看到：
- 所有 35 个 topics
- 实时消息流
- 消费者组状态
- Broker 信息

## 🔍 验证 Kafka 状态

```bash
# 检查容器状态
docker compose ps | grep kafka

# 查看所有 topics
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list

# 查看 topic 详情
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic payment.created
```

## 📨 发送测试消息

```bash
# 方式 1: 使用脚本
./scripts/test-kafka.sh

# 方式 2: 手动发送
echo '{"payment_id":"TEST001","amount":10000}' | \
  docker exec -i payment-kafka kafka-console-producer \
  --bootstrap-server localhost:9092 \
  --topic payment.created
```

## 📥 消费消息

```bash
# 从头消费所有消息
docker exec payment-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic payment.created \
  --from-beginning

# 实时消费新消息
docker exec payment-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic payment.created

# 带格式化输出
docker exec payment-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic payment.created \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

## 🏗️ 已创建的 Topics

总计 **35 个 topics**，分为以下类别：

### 支付相关 (6个)
- `payment.created` - 支付创建
- `payment.success` - 支付成功
- `payment.failed` - 支付失败
- `payment.refund.created` - 退款创建
- `payment.refund.success` - 退款成功
- `payment.refund.failed` - 退款失败

### 订单相关 (4个)
- `order.created` - 订单创建
- `order.updated` - 订单更新
- `order.cancelled` - 订单取消
- `order.completed` - 订单完成

### 账务相关 (4个)
- `accounting.transaction.created` - 交易创建
- `accounting.balance.updated` - 余额更新
- `accounting.settlement.created` - 结算创建
- `accounting.settlement.completed` - 结算完成

### 风控相关 (4个)
- `risk.check.started` - 风控检查开始
- `risk.check.completed` - 风控检查完成
- `risk.alert.high` - 高风险告警
- `risk.alert.critical` - 严重风险告警

### 通知相关 (3个)
- `notification.email` - 邮件通知
- `notification.sms` - 短信通知
- `notification.webhook` - Webhook 通知

### 商户相关 (4个)
- `merchant.created` - 商户创建
- `merchant.updated` - 商户更新
- `merchant.approved` - 商户审核通过
- `merchant.frozen` - 商户冻结

### 提现相关 (4个)
- `withdrawal.created` - 提现创建
- `withdrawal.approved` - 提现审核通过
- `withdrawal.rejected` - 提现拒绝
- `withdrawal.completed` - 提现完成

### Saga 事务 (2个)
- `saga.payment.start` - 支付 Saga 开始
- `saga.payment.compensate` - 支付 Saga 补偿

### 其他 (4个)
- `analytics.events` - 分析事件
- `audit.logs` - 审计日志 (6 分区)
- `dlq.payment` - 支付死信队列
- `dlq.notification` - 通知死信队列

## 🔧 常用命令

### 管理 Topics

```bash
# 创建 topic
docker exec payment-kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --create \
  --topic my-topic \
  --partitions 3 \
  --replication-factor 1

# 删除 topic
docker exec payment-kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --delete \
  --topic my-topic

# 查看 topic 配置
docker exec payment-kafka kafka-configs \
  --bootstrap-server localhost:9092 \
  --describe \
  --entity-type topics \
  --entity-name payment.created
```

### 查看消费者组

```bash
# 列出所有消费者组
docker exec payment-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --list

# 查看消费者组详情 (包括 lag)
docker exec payment-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe \
  --group payment-gateway-group
```

### 性能测试

```bash
# 生产者性能测试 (发送 10000 条消息)
docker exec payment-kafka kafka-producer-perf-test \
  --topic payment.created \
  --num-records 10000 \
  --record-size 256 \
  --throughput -1 \
  --producer-props bootstrap.servers=localhost:9092

# 消费者性能测试
docker exec payment-kafka kafka-consumer-perf-test \
  --bootstrap-server localhost:9092 \
  --topic payment.created \
  --messages 10000
```

## 🌐 连接配置

### 从主机连接 (Go 服务)

```go
import "github.com/payment-platform/pkg/kafka"

// 使用外部端口 40093
kafkaProducer, err := kafka.NewProducer(kafka.Config{
    Brokers: []string{"localhost:40093"},
})
```

### 从 Docker 容器连接

```go
// 使用内部端口和服务名
kafkaProducer, err := kafka.NewProducer(kafka.Config{
    Brokers: []string{"kafka:9092"},
})
```

### 环境变量

```bash
# 添加到 .env 或服务配置
KAFKA_BROKERS=localhost:40093
KAFKA_CONSUMER_GROUP=payment-gateway-group
KAFKA_TOPICS=saga.payment.compensate,order.updated
```

## 📈 监控

### Prometheus 指标

访问: **http://localhost:40308**

Kafka Exporter 提供的指标：
- `kafka_topic_partition_current_offset` - 当前 offset
- `kafka_topic_partition_oldest_offset` - 最老 offset
- `kafka_consumergroup_lag` - 消费者 lag

### Grafana 仪表板

访问: **http://localhost:40300**

导入 Kafka 仪表板查看：
- Topic 吞吐量
- 消费者 Lag
- Broker 状态

## 🛠️ 故障排查

### Kafka 无法启动

```bash
# 查看日志
docker compose logs kafka

# 重启服务
docker compose restart zookeeper
docker compose restart kafka
```

### 无法连接 Kafka

```bash
# 检查健康状态
docker exec payment-kafka kafka-broker-api-versions \
  --bootstrap-server localhost:9092

# 检查端口
lsof -i :40092
lsof -i :40093

# 从主机测试连接 (使用 40093)
# 从容器内测试连接 (使用 kafka:9092)
```

### Topic 不存在

```bash
# 重新运行初始化脚本
./scripts/init-kafka-topics.sh
```

## 🔗 相关资源

- [完整 Kafka 指南](./KAFKA_GUIDE.md) - 详细文档
- [Kafka UI](http://localhost:40080) - Web 管理界面
- [Prometheus](http://localhost:40090) - 指标监控
- [Grafana](http://localhost:40300) - 可视化仪表板

## 💡 下一步

1. 查看 [KAFKA_GUIDE.md](./KAFKA_GUIDE.md) 了解如何在服务中使用 Kafka
2. 访问 Kafka UI 浏览实时消息
3. 在 payment-gateway 中实现事件发布
4. 在 notification-service 中实现事件消费
