# Grafana Dashboard 配置指南

## BFF Services Dashboard

监控 Admin BFF 和 Merchant BFF 服务的完整 Grafana Dashboard。

### 功能概览

**15 个监控面板**:
1. **Service Status** - 服务运行状态（UP/DOWN）
2. **Request Rate** - 请求速率（req/s）
3. **Error Rate** - 错误率（5xx）
4. **P95/P99 Latency** - 响应延迟
5. **Rate Limit Violations** - 限流违规（429）
6. **Authentication Failures** - 认证失败（401/403）
7. **HTTP Status Distribution** - 状态码分布（饼图）
8. **Memory Usage** - 内存使用
9. **CPU Usage** - CPU 使用
10. **Active Goroutines** - 协程数量
11. **Request by Endpoint** - Top 10 请求端点
12. **2FA Failures** - 2FA 验证失败（仅 Admin BFF）
13. **Tenant Metrics** - 租户指标（仅 Merchant BFF）
14. **Request Size** - 平均请求大小
15. **Response Size** - 平均响应大小

### 快速导入

#### 方法 1: Grafana UI 导入

1. 登录 Grafana: http://localhost:40300 (admin/admin)

2. 导航到 **Dashboards** → **Import**

3. 点击 **Upload JSON file**，选择:
   ```
   monitoring/grafana/dashboards/bff-services-dashboard.json
   ```

4. 选择 Prometheus 数据源

5. 点击 **Import**

#### 方法 2: 使用 API 导入

```bash
# 设置 Grafana API Key (在 Configuration → API Keys 创建)
export GRAFANA_API_KEY="your-api-key"

# 导入 Dashboard
curl -X POST http://localhost:40300/api/dashboards/db \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana/dashboards/bff-services-dashboard.json
```

#### 方法 3: 自动配置（推荐）

在 `docker-compose.yml` 中配置自动加载:

```yaml
grafana:
  image: grafana/grafana:latest
  volumes:
    - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
```

### Dashboard 配置

#### 变量（Variables）

- **service**: 选择要监控的服务（admin-bff, merchant-bff, 或全部）
- **interval**: 时间间隔（30s, 1m, 5m, 10m, 30m, 1h）

#### 告警注解（Annotations）

Dashboard 会自动显示触发的 Prometheus 告警。

红色图标表示告警正在触发。

### 关键指标说明

#### 1. 服务状态（Service Status）
- **绿色 (1)**: 服务正常运行
- **红色 (0)**: 服务宕机
- **来源**: `up{job=~"admin-bff|merchant-bff"}`

#### 2. 错误率（Error Rate）
- **阈值**: >5% 触发告警
- **计算**: (5xx 请求数 / 总请求数) * 100
- **建议**: 保持 <1%

#### 3. P95 延迟（P95 Latency）
- **目标**: <500ms
- **警告**: >1s
- **严重**: >3s

#### 4. 限流违规（Rate Limit Violations）
- **Admin BFF**: 60 req/min (normal), 5 req/min (sensitive)
- **Merchant BFF**: 300 req/min (relaxed), 60 req/min (normal)
- **高频违规**: 可能是攻击或需要调整限流策略

#### 5. 认证失败（Authentication Failures）
- **401**: JWT Token 无效或缺失
- **403**: 权限不足或 2FA 失败
- **建议**: 高频失败需要调查是否有暴力破解攻击

#### 6. 2FA 失败（Admin BFF）
- **监控**: 财务操作的 2FA 验证失败
- **路径**: /payments, /settlements, /withdrawals
- **异常**: 高频失败可能是未授权访问尝试

#### 7. 资源使用
- **Memory**:
  - Admin BFF: 正常 <500MB
  - Merchant BFF: 正常 <1GB
- **CPU**: 正常 <50%
- **Goroutines**: 正常 <1000

### 告警配置

Dashboard 包含以下内置告警:

1. **High Error Rate** (错误率 >5%)
   - 严重性: Warning
   - 通知渠道: Slack, Email

2. **High Latency** (P95 >1s)
   - 严重性: Warning
   - 通知渠道: Slack

3. **Service Down** (up == 0)
   - 严重性: Critical
   - 通知渠道: PagerDuty, Slack, Email

### 自定义 Dashboard

#### 添加新面板

1. 点击右上角 **Add Panel**
2. 选择 **Add new panel**
3. 输入 PromQL 查询
4. 配置可视化选项
5. 点击 **Apply**

#### 常用 PromQL 查询

```promql
# 平均响应时间
avg(http_request_duration_seconds{job="admin-bff"})

# 请求成功率
sum(rate(http_requests_total{job="admin-bff",status="200"}[5m]))
/
sum(rate(http_requests_total{job="admin-bff"}[5m])) * 100

# Top 10 最慢端点
topk(10,
  histogram_quantile(0.95,
    rate(http_request_duration_seconds_bucket[5m])
  )
) by (path)

# 每秒限流次数
sum(rate(http_requests_total{status="429"}[5m])) by (job)

# 活跃租户数（Merchant BFF）
count(count(http_requests_total{job="merchant-bff"}) by (merchant_id))
```

### 最佳实践

#### 1. 定期审查
- 每日查看 Dashboard 了解服务健康状况
- 每周审查趋势，识别性能退化

#### 2. 设置告警阈值
- 根据历史数据调整告警阈值
- 避免告警疲劳（过多误报）

#### 3. 性能基线
- 建立性能基线（P50, P95, P99）
- 监控基线偏离情况

#### 4. 容量规划
- 监控资源使用趋势
- 提前规划扩容

#### 5. 安全监控
- 密切关注认证失败和限流违规
- 设置异常访问模式告警

### 故障排查

#### 高错误率

1. 检查 **Error Rate** 面板
2. 查看 **HTTP Status Distribution** 确定错误类型
3. 检查 **Request by Endpoint** 定位问题端点
4. 查看服务日志获取详细错误信息

#### 高延迟

1. 检查 **P95 Latency** 面板
2. 对比 **Memory Usage** 和 **CPU Usage**
3. 检查是否有数据库或 Redis 连接问题
4. 查看 **Active Goroutines** 是否异常增长

#### 频繁限流

1. 检查 **Rate Limit Violations** 面板
2. 确认是否是正常流量增长
3. 检查是否有恶意攻击
4. 考虑调整限流策略或扩容

#### 认证失败激增

1. 检查 **Authentication Failures** 面板
2. 区分 401（无效 Token）和 403（权限不足）
3. 检查是否有 Token 过期或密钥轮换
4. 调查可能的暴力破解攻击

### 访问链接

- **Grafana**: http://localhost:40300
- **Prometheus**: http://localhost:40090
- **Admin BFF Metrics**: http://localhost:40001/metrics
- **Merchant BFF Metrics**: http://localhost:40023/metrics

### 相关文档

- [BFF Security Summary](../../BFF_SECURITY_COMPLETE_SUMMARY.md)
- [Prometheus Alerts](../prometheus/alerts/bff-alerts.yml)
- [Admin BFF Documentation](../../backend/services/admin-bff-service/ADVANCED_SECURITY_COMPLETE.md)
- [Merchant BFF Documentation](../../backend/services/merchant-bff-service/MERCHANT_BFF_SECURITY.md)

---

**更新日期**: 2025-10-26
**维护**: Payment Platform Team
