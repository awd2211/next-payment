# Grafana Saga 监控仪表盘导入指南

## 📊 仪表盘概览

**Saga 分布式事务监控** 仪表盘提供了完整的 Saga 执行监控视图，包括：

1. **Saga 执行速率 (QPS)** - 实时监控 Saga 执行的 QPS
2. **Saga 成功率** - 总体成功率仪表（目标: >99.5%）
3. **Saga 执行延迟** - P95 / P99 延迟曲线
4. **Saga 补偿执行频率** - 补偿触发监控（⚠️ 应该 <1%）
5. **Saga 重试频率** - 重试次数分布
6. **分布式锁失败总数** - 并发控制监控
7. **Recovery Worker 错误总数** - 后台恢复服务监控
8. **Saga 步骤执行延迟** - 各步骤详细性能分析

---

## 🚀 导入步骤

### 方式 1: 通过 Grafana UI 导入（推荐）

1. **访问 Grafana**
   ```bash
   open http://localhost:40300
   ```
   - 用户名: `admin`
   - 密码: `admin`

2. **导入仪表盘**
   - 点击左侧菜单 `+` → `Import`
   - 点击 `Upload JSON file`
   - 选择文件: `grafana/dashboards/saga-monitoring.json`
   - 或者直接粘贴 JSON 内容

3. **配置数据源**
   - 选择 Prometheus 数据源（应该已经配置好）
   - 如果没有，需要先添加 Prometheus 数据源:
     - URL: `http://prometheus:9090` (Docker) 或 `http://localhost:40090` (本地)

4. **保存并查看**
   - 点击 `Import` 按钮
   - 仪表盘将自动打开

### 方式 2: 自动部署（通过 docker-compose）

如果使用 docker-compose 部署，可以将仪表盘配置文件放入 Grafana 的 provisioning 目录：

```bash
# 1. 创建 provisioning 目录（如果不存在）
mkdir -p docker/grafana/provisioning/dashboards

# 2. 复制仪表盘配置
cp grafana/dashboards/saga-monitoring.json docker/grafana/provisioning/dashboards/

# 3. 创建 dashboard provider 配置
cat > docker/grafana/provisioning/dashboards/dashboard.yml <<EOF
apiVersion: 1

providers:
  - name: 'Saga Monitoring'
    orgId: 1
    folder: 'Saga'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

# 4. 重启 Grafana 容器
docker-compose restart grafana
```

---

## 📈 仪表盘面板详解

### 1. Saga 执行速率 (QPS)

**指标**: `saga_execution_total`

**查询**:
```promql
# 成功
rate(saga_execution_total{status="success"}[5m])

# 失败
rate(saga_execution_total{status="failed"}[5m])
```

**解读**:
- 绿色线: 成功的 Saga 执行速率
- 红色线: 失败的 Saga 执行速率
- **正常**: 失败线接近 0
- **异常**: 失败线突然上升

---

### 2. Saga 成功率

**指标**: 成功率 = 成功数 / 总数

**查询**:
```promql
sum(rate(saga_execution_total{status="success"}[5m]))
/
sum(rate(saga_execution_total[5m]))
```

**解读**:
- 绿色 (>99.5%): 正常
- 黄色 (95-98%): 警告
- 红色 (<95%): 严重

---

### 3. Saga 执行延迟

**指标**: `saga_execution_duration_seconds`

**查询**:
```promql
# P95
histogram_quantile(0.95, rate(saga_execution_duration_seconds_bucket[5m]))

# P99
histogram_quantile(0.99, rate(saga_execution_duration_seconds_bucket[5m]))
```

**解读**:
- P95: 95% 的请求延迟低于此值
- P99: 99% 的请求延迟低于此值
- **正常**: P95 < 5s, P99 < 10s
- **异常**: P95 > 10s 或 P99 > 30s

---

### 4. Saga 补偿执行频率 ⚠️

**指标**: `saga_compensation_total`

**查询**:
```promql
rate(saga_compensation_total[5m])
```

**解读**:
- **正常**: 补偿频率 < 1% (偶尔触发)
- **警告**: 补偿频率 1-5% (频繁触发)
- **严重**: 补偿频率 > 5% (需要排查原因)

**常见原因**:
1. 下游服务不稳定
2. 网络超时
3. 业务逻辑错误
4. 数据库死锁

---

### 5. Saga 重试频率

**指标**: `saga_retry_total`

**查询**:
```promql
rate(saga_retry_total[5m])
```

**解读**:
- 按尝试次数分层显示
- Attempt 1: 第一次重试
- Attempt 2: 第二次重试
- Attempt 3: 第三次重试
- **正常**: 大部分在 Attempt 1
- **异常**: Attempt 2/3 占比过高

---

### 6. 分布式锁失败总数

**指标**: `saga_lock_acquire_failed_total`

**查询**:
```promql
sum(saga_lock_acquire_failed_total)
```

**解读**:
- **正常**: 0-10 (偶尔并发冲突)
- **警告**: 10-50 (并发较高)
- **严重**: >50 (可能存在热点或死锁)

**优化建议**:
1. 增加 Redis 连接池
2. 优化锁超时时间
3. 使用分片策略

---

### 7. Recovery Worker 错误总数

**指标**: `saga_recovery_worker_errors_total`

**查询**:
```promql
sum(saga_recovery_worker_errors_total)
```

**解读**:
- **正常**: 0-10 (偶尔恢复失败)
- **警告**: 10-50 (恢复异常增多)
- **严重**: >50 (Recovery Worker 异常)

**排查步骤**:
1. 查看 Recovery Worker 日志
2. 检查数据库连接
3. 检查 Redis 连接
4. 验证 Saga 状态表数据

---

### 8. Saga 步骤执行延迟

**指标**: `saga_step_duration_seconds`

**查询**:
```promql
histogram_quantile(0.95, rate(saga_step_duration_seconds_bucket[5m]))
```

**解读**:
- 显示每个 Saga 每个步骤的 P95 延迟
- 帮助定位性能瓶颈步骤

**示例分析**:
```
Withdrawal Saga:
  ├─ FreezeAccount: 0.1s ✅
  ├─ CreateBankTransfer: 2.5s ⚠️ (瓶颈)
  └─ SendNotification: 0.3s ✅
```

---

## 🔔 告警配置建议

### 告警规则 1: Saga 失败率过高

```yaml
- alert: SagaHighFailureRate
  expr: |
    sum(rate(saga_execution_total{status="failed"}[5m]))
    / sum(rate(saga_execution_total[5m])) > 0.05
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Saga 失败率超过 5%"
    description: "当前失败率: {{ $value | humanizePercentage }}"
```

### 告警规则 2: Saga 补偿频繁

```yaml
- alert: SagaFrequentCompensation
  expr: rate(saga_compensation_total[5m]) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Saga 补偿频繁 (>0.1/s)"
    description: "Saga: {{ $labels.saga_type }}, Step: {{ $labels.step }}"
```

### 告警规则 3: Saga 执行超时

```yaml
- alert: SagaExecutionTimeout
  expr: |
    histogram_quantile(0.99,
      rate(saga_execution_duration_seconds_bucket[5m])
    ) > 30
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Saga P99 执行时间超过 30s"
    description: "Saga: {{ $labels.saga_type }}, P99: {{ $value }}s"
```

### 告警规则 4: Recovery Worker 错误

```yaml
- alert: SagaRecoveryWorkerFailed
  expr: saga_recovery_worker_errors_total > 10
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Recovery Worker 错误次数过多"
    description: "当前错误数: {{ $value }}"
```

---

## 📝 告警配置步骤

### 1. 在 Prometheus 中配置告警规则

编辑 `prometheus/alerts/saga_alerts.yml`:

```yaml
groups:
  - name: saga_alerts
    interval: 30s
    rules:
      # ... 粘贴上面的告警规则 ...
```

### 2. 在 Prometheus 配置中引用

编辑 `prometheus.yml`:

```yaml
rule_files:
  - "alerts/saga_alerts.yml"
```

### 3. 配置 Alertmanager（可选）

```yaml
# alertmanager.yml
route:
  receiver: 'team-saga'
  group_by: ['alertname', 'saga_type']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

receivers:
  - name: 'team-saga'
    webhook_configs:
      - url: 'http://your-webhook-url/alerts'
    # 或者使用邮件
    email_configs:
      - to: 'team-saga@example.com'
        from: 'prometheus@example.com'
```

### 4. 重启 Prometheus

```bash
docker-compose restart prometheus
```

---

## 🎯 监控最佳实践

### 1. 日常监控

**每日检查**:
- ✅ Saga 成功率 > 99.5%
- ✅ 补偿频率 < 1%
- ✅ P99 延迟 < 10s
- ✅ Recovery Worker 错误 = 0

### 2. 性能优化

**关注指标**:
1. **步骤延迟** - 找出最慢的步骤
2. **重试次数** - 降低重试率
3. **补偿频率** - 提高步骤成功率

**优化方向**:
- 优化慢步骤（数据库查询、HTTP 调用）
- 增加超时时间（避免误超时）
- 提高下游服务稳定性

### 3. 容量规划

**监控趋势**:
```promql
# 过去 7 天的 QPS 趋势
rate(saga_execution_total[7d])

# 过去 30 天的成功率趋势
sum(rate(saga_execution_total{status="success"}[30d]))
/ sum(rate(saga_execution_total[30d]))
```

**扩容建议**:
- Saga QPS > 1000: 考虑增加 Redis 节点
- Recovery Worker 错误增多: 增加 Worker 数量
- 分布式锁失败增多: 优化锁策略

---

## 🔍 故障排查流程

### 场景 1: Saga 失败率突然上升

1. **查看仪表盘** → 确定是哪个 Saga 类型失败
2. **查看步骤延迟面板** → 定位是哪个步骤失败
3. **查看应用日志**:
   ```bash
   tail -f logs/payment-gateway.log | grep "Saga 执行失败"
   ```
4. **检查下游服务**:
   ```bash
   curl http://localhost:40004/health  # Order Service
   curl http://localhost:40005/health  # Channel Adapter
   ```

### 场景 2: Saga 执行延迟增加

1. **查看步骤延迟面板** → 找出最慢的步骤
2. **分析步骤类型**:
   - 数据库操作 → 检查数据库性能
   - HTTP 调用 → 检查下游服务
   - Redis 操作 → 检查 Redis 性能
3. **优化建议**:
   - 添加索引（数据库慢查询）
   - 增加超时（避免误超时）
   - 启用缓存（减少重复查询）

### 场景 3: 补偿频繁触发

1. **查看补偿面板** → 确定是哪个 Saga 和步骤
2. **分析补偿原因**:
   ```bash
   # 查看 Saga 状态表
   psql -h localhost -p 40432 -U postgres -d payment_gateway -c \
     "SELECT * FROM saga_logs WHERE status = 'compensated' ORDER BY created_at DESC LIMIT 10;"
   ```
3. **常见原因和解决方案**:
   - 网络超时 → 增加超时时间
   - 服务异常 → 检查下游服务日志
   - 数据库死锁 → 优化事务隔离级别
   - 业务逻辑错误 → 修复代码逻辑

---

## 📚 参考文档

1. **Saga 实现报告**: [SAGA_FINAL_IMPLEMENTATION_REPORT.md](SAGA_FINAL_IMPLEMENTATION_REPORT.md)
2. **Saga 集成报告**: [SAGA_DEEP_INTEGRATION_COMPLETE.md](SAGA_DEEP_INTEGRATION_COMPLETE.md)
3. **Prometheus 指标**: http://localhost:40090/metrics
4. **Grafana 官方文档**: https://grafana.com/docs/

---

## 🎉 总结

通过这个监控仪表盘，你可以：

✅ **实时监控** Saga 执行状态
✅ **快速定位** 性能瓶颈
✅ **及时发现** 异常情况
✅ **数据驱动** 优化决策

**下一步**:
1. 导入仪表盘
2. 配置告警规则
3. 启动服务并观察指标
4. 根据监控数据进行优化

---

*Generated: 2025-10-24*
*Author: Claude Code*
*Version: 1.0.0*
