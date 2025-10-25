# ELK Stack 集成指南

## 概述

本支付平台已完整集成 ELK Stack (Elasticsearch + Logstash + Kibana)，提供统一的日志收集、分析和可视化能力。

## 架构

```
微服务日志 (backend/logs/*.log)
    ↓
Logstash (收集 + 解析 + 过滤)
    ↓
Elasticsearch (存储 + 索引)
    ↓
Kibana (可视化 + 查询)
```

## 端口配置

| 组件 | 端口 | 说明 |
|------|------|------|
| **Elasticsearch** | 40920 | HTTP API |
| **Elasticsearch** | 40930 | TCP 通信 |
| **Kibana** | 40561 | Web UI |
| **Logstash** | 40514 | TCP 输入 |
| **Logstash** | 40515 | UDP 输入 |
| **Logstash** | 40944 | 监控 API |

## 访问方式

### 1. Kibana Web UI

```bash
# 浏览器访问
http://localhost:40561

# 首次访问可能需要等待 30-60 秒初始化
```

### 2. Elasticsearch API

```bash
# 检查集群健康
curl http://localhost:40920/_cluster/health

# 查看所有索引
curl http://localhost:40920/_cat/indices?v

# 搜索日志（最近 10 条）
curl http://localhost:40920/payment-logs-*/_search?size=10&sort=@timestamp:desc
```

### 3. Logstash 监控

```bash
# 查看 Logstash 状态
curl http://localhost:40944/_node/stats?pretty
```

## Logstash 配置详解

配置文件：`/home/eric/payment/config/logstash/logstash.conf`

### 输入源 (Input)

1. **文件输入** - 自动收集所有微服务日志
   - 路径：`/var/log/payment/*.log`（映射自 `backend/logs/`）
   - 格式：JSON
   - 标签：`microservice`

2. **TCP 输入** - 实时日志流（可选）
   - 端口：5014 (外部 40514)
   - 格式：JSON Lines
   - 标签：`tcp`

3. **UDP 输入** - 快速日志传输（可选）
   - 端口：5015 (外部 40515)
   - 格式：JSON Lines
   - 标签：`udp`

### 过滤器 (Filter)

#### 自动提取字段：

| 字段 | 来源 | 说明 |
|------|------|------|
| `service_name` | 文件路径 | 从 `payment-gateway.log` 提取 `payment-gateway` |
| `level` | JSON | 日志级别 (INFO/WARN/ERROR/FATAL) |
| `@timestamp` | JSON `timestamp` | 时间戳（自动解析 ISO8601） |
| `microservice` | JSON `service` | 微服务名称 |
| `trace` | JSON `trace_id` | 分布式追踪 ID（关联 Jaeger） |
| `http_request` | JSON `method` + `path` | HTTP 请求信息 |

#### 自动标记：

- **错误日志** - 标签 `error`（级别为 ERROR 或 FATAL）
- **慢查询** - 标签 `slow_query`（duration > 1000ms）

### 输出 (Output)

- **索引命名**：`payment-logs-YYYY.MM.dd`（每天一个索引）
- **可选配置**：按服务名称分索引 `payment-{service_name}-YYYY.MM.dd`

## Docker Compose 配置

### Elasticsearch

```yaml
elasticsearch:
  image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
  environment:
    - discovery.type=single-node      # 单节点模式
    - ES_JAVA_OPTS=-Xms512m -Xmx512m  # 内存限制 512MB
    - xpack.security.enabled=false     # 禁用安全认证（开发环境）
  volumes:
    - elasticsearch_data:/usr/share/elasticsearch/data
  healthcheck:
    test: curl -f http://localhost:9200/_cluster/health
```

### Kibana

```yaml
kibana:
  image: docker.elastic.co/kibana/kibana:8.11.0
  environment:
    - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
  depends_on:
    - elasticsearch
  healthcheck:
    test: curl -f http://localhost:5601/api/status
```

### Logstash

```yaml
logstash:
  image: docker.elastic.co/logstash/logstash:8.11.0
  volumes:
    - ./config/logstash/logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    - ./backend/logs:/var/log/payment:ro  # 只读挂载日志目录
  environment:
    - LS_JAVA_OPTS=-Xms256m -Xmx256m      # 内存限制 256MB
    - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
```

## 使用场景

### 1. 查看微服务日志

#### 在 Kibana 中创建索引模式：

1. 访问 http://localhost:40561
2. 导航：Management → Stack Management → Index Patterns
3. 创建索引模式：`payment-logs-*`
4. 时间字段：`@timestamp`

#### 使用 Discover 查看日志：

- **按服务过滤**：`service_name: "payment-gateway"`
- **按级别过滤**：`level: "ERROR"`
- **按追踪 ID 查询**：`trace: "trace-id-123"`（关联分布式追踪）
- **查询慢查询**：`tags: "slow_query"`

### 2. 常用 KQL 查询

```kql
# 查看支付网关的错误日志（最近 1 小时）
service_name: "payment-gateway" AND level: "ERROR"

# 查看所有慢查询（duration > 1s）
tags: "slow_query"

# 按追踪 ID 查询完整请求链路
trace: "specific-trace-id"

# 查询特定接口的日志
http_request: "POST /api/v1/payments"

# 查询包含特定关键词的日志
message: *timeout*
```

### 3. 创建可视化仪表板

#### 推荐图表：

1. **日志量趋势图** - Line chart (时间 vs 日志数量)
   - 按 `service_name` 分组

2. **错误率饼图** - Pie chart
   - 按 `level` 分组

3. **服务错误统计** - Bar chart
   - X 轴：`service_name`
   - Y 轴：错误日志数量（`level: "ERROR"`）

4. **慢查询 Top 10** - Data table
   - 过滤：`tags: "slow_query"`
   - 显示：service_name, http_request, duration

5. **请求方法分布** - Donut chart
   - 按 `method` 分组（GET/POST/PUT/DELETE）

### 4. 设置告警（Kibana Alerting）

#### 示例：错误日志告警

1. 导航：Stack Management → Rules and Connectors
2. 创建规则：
   - **名称**：High Error Rate Alert
   - **类型**：Elasticsearch query
   - **索引**：payment-logs-*
   - **查询**：`level: "ERROR"`
   - **阈值**：count > 50 in last 5 minutes
   - **操作**：发送邮件/Slack 通知

## 与 Jaeger 集成

### 分布式追踪关联

通过 `trace_id` 字段，可以关联 Kibana 日志和 Jaeger 追踪：

1. **在 Kibana 中查询日志**：
   ```kql
   service_name: "payment-gateway" AND level: "ERROR"
   ```

2. **复制 `trace` 字段值**：如 `abc123-trace-id`

3. **在 Jaeger UI 中搜索**：
   - 访问：http://localhost:50686
   - 搜索追踪 ID：`abc123-trace-id`
   - 查看完整请求链路

### 完整故障排查流程

```
1. Kibana 发现错误日志
   ↓
2. 提取 trace_id
   ↓
3. Jaeger 查看分布式追踪
   ↓
4. Grafana 查看性能指标
   ↓
5. 定位问题根因
```

## 性能优化建议

### 生产环境配置

1. **索引生命周期管理（ILM）**
   - 自动删除 30 天前的索引
   - 热-温-冷架构

2. **日志采样**
   - INFO 级别：10% 采样
   - WARN/ERROR 级别：100% 保留

3. **资源配置**
   - Elasticsearch：4GB+ 堆内存
   - Logstash：1GB+ 堆内存
   - 使用 SSD 存储

4. **安全加固**
   - 启用 X-Pack Security
   - TLS/SSL 加密
   - RBAC 权限控制

### 开发环境（当前配置）

- Elasticsearch：512MB 堆内存（足够测试使用）
- Logstash：256MB 堆内存
- 禁用安全认证（简化开发流程）
- 单节点集群（yellow 状态正常）

## 常见问题

### 1. Kibana 无法连接到 Elasticsearch

**症状**：访问 http://localhost:40561 显示 "Kibana server is not ready yet"

**解决方案**：
```bash
# 检查 Elasticsearch 健康状态
curl http://localhost:40920/_cluster/health

# 查看 Kibana 日志
docker logs payment-kibana

# 等待 30-60 秒让服务完全启动
```

### 2. 日志未出现在 Kibana

**症状**：Discover 中看不到日志

**检查步骤**：
```bash
# 1. 验证日志文件存在
ls -lh /home/eric/payment/backend/logs/

# 2. 检查 Logstash 是否正常
docker logs payment-logstash | tail -20

# 3. 检查索引是否创建
curl http://localhost:40920/_cat/indices?v

# 4. 手动插入测试日志
echo '{"timestamp":"'$(date -Iseconds)'","level":"INFO","service":"test","message":"Test"}' \
  >> /home/eric/payment/backend/logs/test.log

# 5. 等待 Logstash 处理（约 5-10 秒）
```

### 3. 索引模式未显示

**解决方案**：
1. 确认索引存在：`curl http://localhost:40920/_cat/indices?v`
2. 在 Kibana 中手动创建索引模式：Management → Index Patterns
3. 模式名称：`payment-logs-*`

### 4. 内存不足导致容器崩溃

**症状**：`docker ps` 显示 ELK 容器不断重启

**解决方案**：
```yaml
# 减少 JVM 堆内存（docker-compose.yml）
elasticsearch:
  environment:
    - ES_JAVA_OPTS=-Xms256m -Xmx256m  # 从 512MB 降到 256MB

logstash:
  environment:
    - LS_JAVA_OPTS=-Xms128m -Xmx128m  # 从 256MB 降到 128MB
```

## 启动和停止

### 启动 ELK Stack

```bash
cd /home/eric/payment
docker compose up -d elasticsearch kibana logstash

# 等待服务就绪（约 1-2 分钟）
docker compose ps
```

### 停止 ELK Stack

```bash
docker compose stop elasticsearch kibana logstash
```

### 完全删除（包括数据）

```bash
docker compose down -v
# 警告：这会删除所有日志索引数据！
```

## 验证清单

- [ ] Elasticsearch 健康状态：`curl http://localhost:40920/_cluster/health`
- [ ] Kibana 可访问：http://localhost:40561
- [ ] Logstash 正在运行：`docker logs payment-logstash`
- [ ] 日志索引已创建：`curl http://localhost:40920/_cat/indices?v | grep payment-logs`
- [ ] 可以在 Kibana Discover 中看到日志
- [ ] 追踪 ID 关联有效（Kibana ↔ Jaeger）

## 监控端点汇总

| 服务 | 端点 | 说明 |
|------|------|------|
| **Elasticsearch** | http://localhost:40920/_cluster/health | 集群健康状态 |
| **Elasticsearch** | http://localhost:40920/_cat/indices?v | 索引列表 |
| **Kibana** | http://localhost:40561 | Web UI |
| **Kibana** | http://localhost:40561/api/status | API 状态 |
| **Logstash** | http://localhost:40944/_node/stats | 节点统计 |
| **Prometheus** | http://localhost:40090 | 指标收集 |
| **Grafana** | http://localhost:40300 | 可视化（admin/admin） |
| **Jaeger** | http://localhost:50686 | 分布式追踪 |

## 下一步

1. **创建 Kibana 仪表板** - 可视化关键指标
2. **设置告警规则** - 错误率/慢查询告警
3. **配置索引生命周期** - 自动清理旧日志
4. **集成 Slack/邮件通知** - 实时告警推送
5. **优化 Logstash 解析规则** - 针对特定日志格式

---

**文档版本**：1.0
**最后更新**：2025-10-25
**维护者**：Payment Platform Team
