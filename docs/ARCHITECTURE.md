# 系统架构文档

## 1. 整体架构

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                          前端应用层                               │
├──────────────────────────────┬──────────────────────────────────┤
│   Admin Portal (React)       │   Merchant Portal (React)        │
│   运营管理后台 :3000           │   商户自助后台 :3001              │
│   - 管理员管理                 │   - 商户注册登录                  │
│   - 商户审核                   │   - 订单查询                      │
│   - 系统配置                   │   - 财务报表                      │
│   - 数据看板                   │   - API密钥管理                   │
└──────────────────────────────┴──────────────────────────────────┘
                              ↓ HTTPS
┌─────────────────────────────────────────────────────────────────┐
│                    API Gateway (Traefik :80)                     │
│   - 统一入口                                                       │
│   - 负载均衡                                                       │
│   - TLS终止                                                       │
│   - 限流                                                          │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                        微服务层（gRPC + HTTP）                    │
├──────────────────┬──────────────────┬──────────────────┬─────────┤
│  Admin Service   │ Merchant Service │ Payment Gateway  │ Order   │
│  :8001           │ :8002            │ :8003            │ :8004   │
│  - 管理员管理      │ - 商户CRUD       │ - 支付路由        │ - 订单  │
│  - 权限控制       │ - API密钥        │ - 幂等性         │ - 统计  │
│  - 审批流程       │ - Webhook        │ - 状态机         │         │
└──────────────────┴──────────────────┴──────────────────┴─────────┘
           ↓                    ↓                    ↓
┌─────────────────────────────────────────────────────────────────┐
│               Channel Adapter (渠道适配层) :8005                  │
├──────────────────┬──────────────────┬──────────────────────────┤
│  Stripe Adapter  │  PayPal Adapter  │  Crypto Adapter          │
│  - 卡支付         │ - 钱包支付        │ - BTC/ETH/USDT           │
│  - Webhook       │ - Webhook        │ - 链上监听                │
└──────────────────┴──────────────────┴──────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      支撑服务层                                   │
├──────────────────┬──────────────────┬──────────────────────────┤
│ Accounting       │ Risk Service     │ Notification             │
│ Service          │ :8006            │ Service :8007            │
│ :8005            │ - 反欺诈          │ - Webhook                │
│ - 账务           │ - 限额            │ - Email/SMS              │
│ - 清结算         │ - 黑名单          │                          │
├──────────────────┼──────────────────┼──────────────────────────┤
│ Analytics        │ Config Service   │                          │
│ Service :8008    │ :8009            │                          │
│ - 数据统计        │ - 配置中心        │                          │
│ - 报表           │ - 费率管理        │                          │
└──────────────────┴──────────────────┴──────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      数据层                                       │
├──────────────────┬──────────────────┬──────────────────────────┤
│  PostgreSQL      │  Redis           │  Kafka                   │
│  :5432           │  :6379           │  :9092                   │
│  - 主数据库       │ - 缓存            │ - 事件总线                │
│  - 事务          │ - 分布式锁        │ - 异步消息                │
│  - 多租户隔离     │ - 限流            │                          │
└──────────────────┴──────────────────┴──────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    监控与运维层                                   │
├──────────────────┬──────────────────┬──────────────────────────┤
│  Prometheus      │  Grafana         │  Jaeger                  │
│  :9090           │  :3000           │  :16686                  │
│  - 指标收集       │ - 可视化          │ - 分布式追踪              │
└──────────────────┴──────────────────┴──────────────────────────┘
```

---

## 2. 核心设计理念

### 2.1 微服务架构

**优点：**
- ✅ 服务独立部署、弹性扩展
- ✅ 故障隔离、高可用
- ✅ 技术栈灵活
- ✅ 团队独立开发

**挑战：**
- ⚠️ 分布式事务
- ⚠️ 服务间通信
- ⚠️ 运维复杂度

**解决方案：**
- **分布式事务**：Saga模式 + 事件溯源
- **服务通信**：gRPC（内部）+ REST（对外）
- **服务发现**：Consul
- **负载均衡**：Traefik/Kong

### 2.2 多租户架构

**隔离策略：**
- **数据隔离**：行级安全策略（Row-Level Security）
- **每张表增加 `tenant_id` 字段**
- **PostgreSQL RLS自动过滤**

```sql
-- 启用行级安全
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;

-- 创建策略
CREATE POLICY tenant_isolation ON payments
    USING (tenant_id = current_setting('app.tenant_id')::UUID);
```

**优点：**
- 成本低（共享资源）
- 维护简单（统一升级）
- 适合SaaS模式

### 2.3 事件驱动架构

**事件流：**
```
Payment Created → Kafka → [Order Service, Accounting Service, Notification Service]
```

**事件类型：**
- `payment.created` - 支付创建
- `payment.success` - 支付成功
- `payment.failed` - 支付失败
- `refund.completed` - 退款完成
- `merchant.approved` - 商户审核通过

**好处：**
- 异步解耦
- 最终一致性
- 易于扩展

---

## 3. 核心模块设计

### 3.1 认证与授权

**JWT Token结构：**
```json
{
  "user_id": "uuid",
  "tenant_id": "uuid",
  "username": "admin",
  "user_type": "admin",  // admin or merchant
  "roles": ["super_admin"],
  "permissions": ["merchant.view", "payment.refund"],
  "exp": 1234567890
}
```

**RBAC权限模型：**
```
管理员 (Admin) ←→ 角色 (Role) ←→ 权限 (Permission)
    N               M:N              N

Permission格式：resource.action
例如：merchant.view, payment.refund, order.export
```

**中间件流程：**
```
Request → AuthMiddleware
        → ValidateToken
        → ExtractClaims
        → CheckPermission
        → Handler
```

### 3.2 幂等性设计

**问题：**
- 网络抖动导致重复请求
- 用户多次点击支付

**解决方案：**
```go
// 请求头携带幂等键
Idempotency-Key: uuid

// Redis存储
Key: idempotency:{key}
Value: {payment_id, status, response}
TTL: 24小时

// 流程
1. 检查Redis是否存在该key
2. 存在 → 返回缓存结果
3. 不存在 → 加分布式锁 → 创建支付 → 缓存结果
```

### 3.3 分布式锁

**场景：**
- 防止订单重复创建
- 防止余额重复扣减

**实现（Redis）：**
```go
// SET key value NX PX 30000
lock := redis.SetNX(ctx, "lock:order:123", "token", 30*time.Second)

// Lua脚本解锁（原子操作）
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
```

### 3.4 限流策略

**令牌桶算法（Token Bucket）：**
```
Rate: 100 req/min
Burst: 200 (允许突发流量)

Redis实现：
Key: rate_limit:user:{user_id}
INCR key
EXPIRE key 60
```

**分级限流：**
- **IP级别**：100 req/min
- **用户级别**：1000 req/min
- **商户级别**：10000 req/min

---

## 4. 数据库设计

### 4.1 分库分表策略

**垂直拆分（按服务）：**
- `payment_admin` - Admin Service
- `payment_merchant` - Merchant Service
- `payment_order` - Order Service
- `payment_accounting` - Accounting Service

**水平拆分（按商户ID）：**
```
payments_0, payments_1, ..., payments_15
分片键：tenant_id
哈希算法：crc32(tenant_id) % 16
```

### 4.2 索引优化

**高频查询索引：**
```sql
-- 复合索引（最左匹配原则）
CREATE INDEX idx_payments_tenant_created
ON payments(tenant_id, created_at DESC);

-- 部分索引（减少索引大小）
CREATE INDEX idx_payments_pending
ON payments(status) WHERE status = 'pending';

-- BRIN索引（时间序列数据）
CREATE INDEX idx_audit_logs_time
ON audit_logs USING BRIN (created_at);
```

### 4.3 事务隔离级别

**PostgreSQL默认：** `Read Committed`

**支付场景：** `Serializable`（强一致性）
```sql
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
-- 扣减余额
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- 创建支付记录
INSERT INTO payments (...) VALUES (...);
COMMIT;
```

---

## 5. 安全设计

### 5.1 数据加密

**传输加密：**
- TLS 1.3（HTTPS）
- gRPC TLS

**存储加密：**
```go
// AES-256-GCM
func Encrypt(plaintext string, key []byte) string {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext)
}
```

**敏感字段加密：**
- API Secret
- 银行卡信息（Token化，不存储）
- 用户密码（bcrypt）

### 5.2 PCI DSS合规

**核心原则：**
1. ❌ **不存储CVV/CVC**
2. ✅ 使用Stripe/PayPal的Token化
3. ✅ TLS 1.3加密通信
4. ✅ 日志脱敏（不记录完整卡号）
5. ✅ 定期安全审计

### 5.3 防止常见攻击

**SQL注入：**
```go
// ❌ 错误
db.Raw("SELECT * FROM users WHERE username = '" + username + "'")

// ✅ 正确（参数化查询）
db.Where("username = ?", username).Find(&user)
```

**XSS防护：**
```go
// 前端自动转义
import DOMPurify from 'dompurify';
const clean = DOMPurify.sanitize(dirty);
```

**CSRF防护：**
```go
// SameSite Cookie + CSRF Token
c.SetSameSite(http.SameSiteStrictMode)
c.SetCookie("csrf_token", token, 3600, "/", "", true, true)
```

---

## 6. 性能优化

### 6.1 缓存策略

**多级缓存：**
```
浏览器缓存 → CDN → Redis → PostgreSQL

热点数据：
- 商户配置（TTL: 1小时）
- 系统配置（TTL: 10分钟）
- 汇率信息（TTL: 5分钟）
```

**缓存更新策略：**
- **Cache Aside**：查询时填充，更新时删除
- **Write Through**：写入时同步更新缓存

### 6.2 数据库优化

**连接池：**
```go
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

**批量操作：**
```go
// ❌ N次查询
for _, id := range ids {
    db.Find(&user, id)
}

// ✅ 1次查询
db.Where("id IN ?", ids).Find(&users)
```

**慢查询监控：**
```sql
-- 启用慢查询日志（>100ms）
ALTER DATABASE payment_platform SET log_min_duration_statement = 100;
```

### 6.3 异步处理

**耗时操作异步化：**
- Webhook发送 → Kafka
- 对账任务 → 定时任务
- 报表生成 → 后台任务

---

## 7. 监控告警

### 7.1 监控指标

**系统指标：**
- CPU、内存、磁盘使用率
- 网络带宽

**应用指标：**
- QPS、延迟（P50/P95/P99）
- 错误率
- 数据库连接数

**业务指标：**
- 支付成功率
- 单日GMV
- 各渠道成功率

### 7.2 告警规则

```yaml
# Prometheus告警规则
groups:
  - name: payment_alerts
    rules:
      # 支付成功率低于95%
      - alert: LowPaymentSuccessRate
        expr: |
          rate(payment_success_total[5m]) / rate(payment_total[5m]) < 0.95
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "支付成功率异常"

      # API响应时间 > 1s
      - alert: HighAPILatency
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 1
        for: 5m
        labels:
          severity: warning
```

---

## 8. 灾备与容错

### 8.1 数据备份

**PostgreSQL：**
- 全量备份：每天凌晨3点
- 增量备份：每小时
- WAL归档：实时
- 异地备份：AWS S3

### 8.2 故障恢复

**RTO（恢复时间目标）：** 30分钟
**RPO（恢复点目标）：** 5分钟

**灾备方案：**
- 主从复制（Streaming Replication）
- 自动故障转移（Patroni + etcd）
- 多可用区部署

---

## 9. 未来规划

### 短期（1-3个月）
- ✅ 完成核心支付功能
- ⏳ 集成Stripe、PayPal
- ⏳ 开发前端后台
- ⏳ 单元测试覆盖率达80%

### 中期（3-6个月）
- ⏳ 加密货币支付
- ⏳ 订阅支付
- ⏳ 多币种支持
- ⏳ Kubernetes部署

### 长期（6-12个月）
- ⏳ 跨境分期支付
- ⏳ 智能路由（AI优化渠道选择）
- ⏳ 全球化部署（多区域）
- ⏳ ISO 27001认证

---

## 10. 参考资料

**官方文档：**
- [Stripe API](https://stripe.com/docs/api)
- [PayPal Developer](https://developer.paypal.com)
- [gRPC Documentation](https://grpc.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

**最佳实践：**
- [12-Factor App](https://12factor.net/)
- [微服务设计模式](https://microservices.io/patterns/)
- [PCI DSS合规指南](https://www.pcisecuritystandards.org/)
