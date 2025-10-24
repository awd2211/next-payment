# Kafka 使用指南

本文档介绍如何在支付平台中使用 Kafka 消息队列。

## 目录

- [快速开始](#快速开始)
- [Topics 说明](#topics-说明)
- [在服务中使用 Kafka](#在服务中使用-kafka)
- [监控和管理](#监控和管理)
- [故障排查](#故障排查)

## 快速开始

### 1. 启动 Kafka

```bash
# 启动所有基础设施 (包括 Kafka)
docker-compose up -d

# 仅启动 Kafka 相关服务
docker-compose up -d zookeeper kafka kafka-ui
```

### 2. 检查 Kafka 状态

```bash
# 检查容器状态
docker-compose ps

# 查看 Kafka 日志
docker-compose logs -f kafka

# 检查健康状态
docker exec payment-kafka kafka-broker-api-versions --bootstrap-server localhost:9092
```

### 3. 初始化 Topics

```bash
# 创建所有预定义的 topics
./scripts/init-kafka-topics.sh
```

### 4. 测试 Kafka

```bash
# 运行测试脚本
./scripts/test-kafka.sh
```

## Topics 说明

### 支付相关 Topics

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `payment.created` | 3 | 支付创建事件 | payment-gateway | analytics, notification |
| `payment.success` | 3 | 支付成功事件 | payment-gateway | accounting, analytics, notification |
| `payment.failed` | 3 | 支付失败事件 | payment-gateway | analytics, notification |
| `payment.refund.created` | 3 | 退款创建事件 | payment-gateway | analytics |
| `payment.refund.success` | 3 | 退款成功事件 | payment-gateway | accounting, analytics, notification |
| `payment.refund.failed` | 3 | 退款失败事件 | payment-gateway | analytics, notification |

### 订单相关 Topics

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `order.created` | 3 | 订单创建事件 | order-service | analytics |
| `order.updated` | 3 | 订单更新事件 | order-service | analytics, notification |
| `order.cancelled` | 3 | 订单取消事件 | order-service | accounting, analytics |
| `order.completed` | 3 | 订单完成事件 | order-service | accounting, analytics, notification |

### 账务相关 Topics

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `accounting.transaction.created` | 3 | 账务交易创建 | accounting-service | analytics |
| `accounting.balance.updated` | 3 | 余额更新 | accounting-service | analytics, notification |
| `accounting.settlement.created` | 3 | 结算创建 | settlement-service | analytics |
| `accounting.settlement.completed` | 3 | 结算完成 | settlement-service | accounting, notification |

### 风控相关 Topics

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `risk.check.started` | 3 | 风控检查开始 | risk-service | analytics |
| `risk.check.completed` | 3 | 风控检查完成 | risk-service | analytics |
| `risk.alert.high` | 3 | 高风险告警 | risk-service | notification, admin-service |
| `risk.alert.critical` | 3 | 严重风险告警 | risk-service | notification, admin-service |

### 通知相关 Topics

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `notification.email` | 3 | 邮件通知 | 所有服务 | notification-service |
| `notification.sms` | 3 | 短信通知 | 所有服务 | notification-service |
| `notification.webhook` | 3 | Webhook 通知 | 所有服务 | notification-service |

### Saga 分布式事务

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `saga.payment.start` | 3 | 支付 Saga 开始 | payment-gateway | order-service, channel-adapter |
| `saga.payment.compensate` | 3 | 支付 Saga 补偿 | 所有 Saga 参与者 | payment-gateway |

### 其他 Topics

| Topic | 分区 | 描述 | 生产者 | 消费者 |
|-------|------|------|--------|--------|
| `analytics.events` | 3 | 分析事件 | 所有服务 | analytics-service |
| `audit.logs` | 6 | 审计日志 | 所有服务 | admin-service, analytics-service |
| `dlq.payment` | 1 | 支付死信队列 | Kafka Connect | admin-service |
| `dlq.notification` | 1 | 通知死信队列 | Kafka Connect | admin-service |

## 在服务中使用 Kafka

### 1. 使用 pkg/kafka 包

所有服务应使用 `pkg/kafka` 包来操作 Kafka：

```go
import (
    "github.com/payment-platform/pkg/kafka"
)

// 在 main.go 或服务初始化时
kafkaProducer, err := kafka.NewProducer(kafka.Config{
    Brokers: []string{"localhost:40093"},
    // 或从环境变量
    // Brokers: strings.Split(config.GetEnv("KAFKA_BROKERS", "localhost:40093"), ","),
})
if err != nil {
    logger.Fatal("Failed to create Kafka producer", zap.Error(err))
}
defer kafkaProducer.Close()

kafkaConsumer, err := kafka.NewConsumer(kafka.Config{
    Brokers: []string{"localhost:40093"},
    GroupID: "payment-gateway-consumer-group",
    Topics:  []string{"saga.payment.compensate"},
})
if err != nil {
    logger.Fatal("Failed to create Kafka consumer", zap.Error(err))
}
defer kafkaConsumer.Close()
```

### 2. 发送消息

```go
// 在服务层发送事件
func (s *PaymentService) CreatePayment(ctx context.Context, input *CreatePaymentInput) (*Payment, error) {
    // ... 创建支付逻辑 ...

    // 发送支付创建事件
    event := map[string]interface{}{
        "event_type":   "payment.created",
        "payment_id":   payment.ID.String(),
        "payment_no":   payment.PaymentNo,
        "merchant_id":  payment.MerchantID.String(),
        "amount":       payment.Amount,
        "currency":     payment.Currency,
        "status":       payment.Status,
        "created_at":   payment.CreatedAt,
    }

    eventJSON, _ := json.Marshal(event)
    err = s.kafkaProducer.SendMessage(ctx, "payment.created", payment.ID.String(), eventJSON)
    if err != nil {
        logger.Error("Failed to send payment.created event",
            zap.Error(err),
            zap.String("payment_id", payment.ID.String()))
        // 注意: Kafka 发送失败不应阻断主流程
    }

    return payment, nil
}
```

### 3. 消费消息

```go
// 在 main.go 启动消费者
go func() {
    for {
        select {
        case msg := <-kafkaConsumer.Messages():
            handleKafkaMessage(msg)
        case err := <-kafkaConsumer.Errors():
            logger.Error("Kafka consumer error", zap.Error(err))
        case <-ctx.Done():
            return
        }
    }
}()

func handleKafkaMessage(msg *kafka.Message) {
    logger.Info("Received Kafka message",
        zap.String("topic", msg.Topic),
        zap.String("key", msg.Key),
        zap.ByteString("value", msg.Value))

    switch msg.Topic {
    case "saga.payment.compensate":
        handlePaymentCompensation(msg)
    default:
        logger.Warn("Unknown topic", zap.String("topic", msg.Topic))
    }
}
```

### 4. Bootstrap 框架集成

如果服务使用 Bootstrap 框架，Kafka 会自动配置：

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "payment-gateway",
    // ... 其他配置 ...

    EnableKafka: true,  // 启用 Kafka
    KafkaConfig: &app.KafkaConfig{
        Brokers:      []string{"localhost:40093"},
        ConsumerGroup: "payment-gateway-group",
        Topics:       []string{"saga.payment.compensate"},
    },
})

// Kafka Producer 和 Consumer 已自动初始化
producer := application.KafkaProducer
consumer := application.KafkaConsumer
```

## 监控和管理

### 1. Kafka UI (推荐)

访问可视化管理界面：

```
http://localhost:40080
```

功能：
- 查看所有 topics 和消息
- 监控消费者组状态
- 查看 broker 配置
- 实时消息追踪
- 性能监控

### 2. 命令行工具

**查看所有 topics:**

```bash
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

**查看 topic 详情:**

```bash
docker exec payment-kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --describe \
  --topic payment.created
```

**消费消息 (从头开始):**

```bash
docker exec payment-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic payment.created \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

**发送测试消息:**

```bash
echo '{"test": "message"}' | docker exec -i payment-kafka kafka-console-producer \
  --bootstrap-server localhost:9092 \
  --topic payment.created
```

**查看消费者组:**

```bash
docker exec payment-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --list
```

**查看消费者组详情:**

```bash
docker exec payment-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe \
  --group payment-gateway-consumer-group
```

### 3. Prometheus 监控

Kafka Exporter 暴露指标到 Prometheus (端口 40308)：

- Topic 消息速率
- 消费者 Lag
- Broker 状态
- 分区分布

在 Grafana 中导入 Kafka 仪表板。

## 配置说明

### 环境变量

服务使用以下环境变量连接 Kafka：

```bash
# Kafka 地址 (多个 broker 用逗号分隔)
KAFKA_BROKERS=localhost:40093

# 消费者组 ID
KAFKA_CONSUMER_GROUP=payment-gateway-group

# 要订阅的 topics (逗号分隔)
KAFKA_TOPICS=saga.payment.compensate,order.updated
```

### Docker Compose 端口

| 服务 | 内部端口 | 外部端口 | 说明 |
|------|---------|---------|------|
| Zookeeper | 2181 | 42181 | Zookeeper 客户端端口 |
| Kafka (内部) | 9092 | 40092 | 容器间通信 |
| Kafka (外部) | 9093 | 40093 | 主机访问 (推荐) |
| Kafka UI | 8080 | 40080 | Web 管理界面 |
| Kafka Exporter | 9308 | 40308 | Prometheus 指标 |

### 监听器配置

Kafka 配置了两个监听器：

1. **INTERNAL** (kafka:9092) - 容器间通信
   - 用于 Docker 网络内的服务
   - 使用服务名 `kafka`

2. **EXTERNAL** (localhost:40093) - 外部访问
   - 用于主机上的开发环境
   - 使用 `localhost:40093`

### 性能配置

- **分区数**: 默认 3 个分区 (提高并发)
- **副本因子**: 1 (单节点测试环境)
- **消息保留**: 7 天 (168 小时)
- **JVM 堆内存**: 512MB
- **网络线程**: 3
- **IO 线程**: 8

生产环境建议：
- 副本因子: 3
- 分区数: 根据吞吐量调整 (6-12)
- JVM 堆内存: 2-4GB

## 故障排查

### 1. Kafka 启动失败

**症状**: 容器反复重启

**检查**:

```bash
# 查看日志
docker-compose logs kafka

# 常见原因
# 1. Zookeeper 未就绪
docker-compose ps zookeeper

# 2. 端口冲突
lsof -i :40092
lsof -i :40093

# 3. 内存不足
docker stats payment-kafka
```

**解决**:

```bash
# 重启服务
docker-compose restart zookeeper
docker-compose restart kafka

# 如果无法解决，清理数据重建
docker-compose down -v
docker-compose up -d zookeeper kafka
```

### 2. 无法连接 Kafka

**症状**: 客户端连接超时

**检查连接**:

```bash
# 从主机连接 (应使用 40093 端口)
docker exec payment-kafka kafka-broker-api-versions \
  --bootstrap-server localhost:9092

# 从容器内连接
docker exec payment-kafka kafka-broker-api-versions \
  --bootstrap-server kafka:9092
```

**常见问题**:

- 使用了错误的端口 (主机应用用 40093, 容器内用 9092)
- 防火墙阻止连接
- Kafka 未完全启动 (等待 30 秒)

### 3. 消息发送失败

**检查 topic 是否存在**:

```bash
docker exec payment-kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --list | grep payment.created
```

**检查生产者配置**:

```go
// 确保使用正确的 broker 地址
kafkaProducer, err := kafka.NewProducer(kafka.Config{
    Brokers: []string{"localhost:40093"},  // 主机环境
    // 或
    // Brokers: []string{"kafka:9092"},    // 容器环境
})
```

### 4. 消费者 Lag 过高

**查看消费者 lag**:

```bash
docker exec payment-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe \
  --group payment-gateway-group
```

**解决方法**:

1. 增加消费者实例数
2. 增加分区数 (需重新分配)
3. 优化消费逻辑性能
4. 检查是否有消费者卡死

### 5. 查看 Kafka 日志

```bash
# 实时日志
docker-compose logs -f kafka

# 最近 100 行
docker-compose logs --tail=100 kafka

# 保存到文件
docker-compose logs kafka > kafka.log
```

## 最佳实践

### 1. 消息设计

```go
// ✅ 好的消息格式
type PaymentEvent struct {
    EventType   string    `json:"event_type"`   // "payment.created"
    EventID     string    `json:"event_id"`     // 唯一事件 ID
    Timestamp   time.Time `json:"timestamp"`    // 事件时间
    PaymentID   string    `json:"payment_id"`   // 业务 ID
    PaymentNo   string    `json:"payment_no"`   // 业务编号
    MerchantID  string    `json:"merchant_id"`  // 关联信息
    Amount      int64     `json:"amount"`       // 业务数据
    Currency    string    `json:"currency"`
    Status      string    `json:"status"`
    Metadata    map[string]interface{} `json:"metadata"` // 扩展字段
}
```

### 2. 幂等性保证

```go
// 消费者需要实现幂等性
func handlePaymentEvent(msg *kafka.Message) error {
    var event PaymentEvent
    if err := json.Unmarshal(msg.Value, &event); err != nil {
        return err
    }

    // 使用 event_id 去重
    exists, _ := redis.Exists(ctx, "processed:"+event.EventID).Result()
    if exists > 0 {
        logger.Info("Event already processed",
            zap.String("event_id", event.EventID))
        return nil // 已处理，跳过
    }

    // 处理业务逻辑
    if err := processPayment(&event); err != nil {
        return err
    }

    // 标记为已处理 (24 小时过期)
    redis.SetEX(ctx, "processed:"+event.EventID, "1", 24*time.Hour)
    return nil
}
```

### 3. 错误处理

```go
// 生产者: 发送失败不应阻断主流程
err = kafkaProducer.SendMessage(ctx, topic, key, message)
if err != nil {
    logger.Error("Failed to send Kafka message",
        zap.Error(err),
        zap.String("topic", topic))
    // 记录到数据库，后续重试
    // 或发送到备用通知渠道
    // 但不返回错误给用户
}

// 消费者: 失败消息进入死信队列
func handleMessage(msg *kafka.Message) error {
    if err := process(msg); err != nil {
        if retryCount < 3 {
            return err // 重试
        }
        // 发送到死信队列
        sendToDLQ(msg, err)
        return nil // 标记为已处理
    }
    return nil
}
```

### 4. 监控告警

在 Prometheus 中配置告警规则：

```yaml
# Kafka consumer lag 过高
- alert: KafkaConsumerLagHigh
  expr: kafka_consumergroup_lag > 1000
  for: 5m
  annotations:
    summary: "Kafka consumer lag is high"
    description: "Consumer group {{ $labels.consumergroup }} lag is {{ $value }}"

# Kafka broker 离线
- alert: KafkaBrokerDown
  expr: kafka_brokers < 1
  for: 1m
  annotations:
    summary: "Kafka broker is down"
```

## 相关资源

- [Apache Kafka 官方文档](https://kafka.apache.org/documentation/)
- [Confluent Platform 文档](https://docs.confluent.io/)
- [Kafka UI 项目](https://github.com/provectus/kafka-ui)
- [pkg/kafka 包文档](../pkg/kafka/README.md)

## 联系支持

如有问题，请：
1. 查看 [故障排查](#故障排查) 章节
2. 运行 `./scripts/test-kafka.sh` 诊断
3. 查看 Kafka UI (http://localhost:40080)
4. 检查 Kafka 日志
