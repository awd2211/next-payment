# 收银台系统快速启动指南

## 一分钟快速启动

### 1. 启动后端服务（Port 40016）

```bash
cd /home/eric/payment/backend
./scripts/start-cashier-service.sh
```

### 2. 启动收银台前端（Port 5176）

```bash
cd /home/eric/payment/frontend/cashier-portal

# 首次运行：配置Stripe密钥
cp .env.example .env
# 编辑 .env，添加: VITE_STRIPE_PUBLIC_KEY=pk_test_xxx

npm run dev
```

### 3. 访问各个门户

| 应用 | URL | 用途 |
|------|-----|------|
| 收银台页面 | http://localhost:5176/checkout?token=xxx | 客户支付 |
| 商户门户 | http://localhost:5174/cashier-config | 配置收银台 |
| 管理员门户 | http://localhost:5173/cashier | 管理模板 |

---

## 完整测试流程

### Step 1: 启动所有服务

```bash
# 1. 确保基础设施运行
docker ps | grep payment-postgres  # PostgreSQL
docker ps | grep payment-redis     # Redis

# 2. 启动cashier-service
cd /home/eric/payment/backend
./scripts/start-cashier-service.sh

# 3. 启动admin-portal
cd /home/eric/payment/frontend/admin-portal
npm run dev  # Port 5173

# 4. 启动merchant-portal
cd /home/eric/payment/frontend/merchant-portal
npm run dev  # Port 5174

# 5. 启动cashier-portal
cd /home/eric/payment/frontend/cashier-portal
npm run dev  # Port 5176
```

### Step 2: 商户配置收银台

1. 打开 http://localhost:5174
2. 登录商户账号
3. 点击左侧菜单"收银台配置"
4. 在"外观设置"Tab：
   - Logo URL: `https://via.placeholder.com/150x50?text=MyShop`
   - 主题颜色: `#1890ff`
5. 在"支付方式"Tab：
   - 启用渠道: `stripe`
   - 默认渠道: `stripe`
6. 在"安全设置"Tab：
   - 会话超时: `30` 分钟
   - 需要CVV: `开启`
   - 启用3DS: `开启`
   - 成功回调: `https://yoursite.com/success`
7. 点击"保存配置"

### Step 3: 创建支付会话

使用curl或Postman调用API（需要商户JWT token）：

```bash
# 获取商户JWT token（先登录merchant-service）
TOKEN="your_merchant_jwt_token"

# 创建支付会话
curl -X POST http://localhost:40016/api/v1/cashier/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "your-merchant-uuid",
    "order_no": "ORDER-20241024-001",
    "amount": 9999,
    "currency": "USD",
    "description": "Premium Subscription - 1 Year",
    "customer_email": "john@example.com",
    "customer_name": "John Doe",
    "customer_ip": "192.168.1.100",
    "allowed_channels": ["stripe"],
    "expires_in_minutes": 30
  }'
```

响应示例：
```json
{
  "code": 0,
  "data": {
    "session_token": "vZGJhYzEyMzQ1Njc4OTBhYmNkZWY...",
    "expires_at": "2024-10-24T11:30:00Z"
  },
  "message": "success"
}
```

### Step 4: 用户访问支付页面

打开浏览器访问：
```
http://localhost:5176/checkout?token=vZGJhYzEyMzQ1Njc4OTBhYmNkZWY...
```

你会看到：
- 商户Logo
- 订单摘要（订单号、金额$99.99）
- Stripe支付表单
- "立即支付"按钮

### Step 5: 填写测试支付信息

使用Stripe测试卡：

| 场景 | 卡号 | 有效期 | CVV | 结果 |
|------|------|--------|-----|------|
| 成功支付 | 4242 4242 4242 4242 | 12/25 | 123 | ✅ 支付成功 |
| 3DS验证 | 4000 0025 0000 3155 | 12/25 | 123 | ✅ 触发3DS验证 |
| 支付失败 | 4000 0000 0000 0002 | 12/25 | 123 | ❌ 卡片被拒 |
| 余额不足 | 4000 0000 0000 9995 | 12/25 | 123 | ❌ 余额不足 |

填写示例：
```
卡号: 4242 4242 4242 4242
持卡人: John Doe
有效期: 12/25
CVV: 123
邮箱: john@example.com
```

### Step 6: 提交支付

1. 点击"立即支付"按钮
2. Stripe处理支付（可能弹出3DS验证窗口）
3. 支付成功后显示成功页面
4. 2秒后自动重定向到商户的成功页面

### Step 7: 查看分析数据

#### 在商户门户查看
http://localhost:5174/cashier-config → "数据分析"Tab

你会看到：
- 转化率饼图
- 渠道统计（Stripe 100%）
- 转化漏斗（访问→填写→提交→成功）

#### 在管理员门户查看
http://localhost:5173/cashier → "监控"Tab

你会看到：
- 活跃商户数
- 今日会话数
- 今日完成数
- 平均转化率
- 渠道分布图
- 商户转化率排行

---

## API快速参考

### 商户API（需要JWT）

#### 1. 配置管理

```bash
# 创建/更新配置
POST /api/v1/cashier/configs
{
  "theme_color": "#1890ff",
  "logo_url": "https://example.com/logo.png",
  "enabled_channels": ["stripe", "paypal"],
  "default_channel": "stripe",
  "session_timeout_minutes": 30,
  "require_cvv": true,
  "enable_3d_secure": true,
  "success_redirect_url": "https://example.com/success",
  "cancel_redirect_url": "https://example.com/cancel"
}

# 获取配置
GET /api/v1/cashier/configs

# 删除配置
DELETE /api/v1/cashier/configs
```

#### 2. 会话管理

```bash
# 创建会话
POST /api/v1/cashier/sessions
{
  "merchant_id": "uuid",
  "order_no": "ORDER-001",
  "amount": 10000,
  "currency": "USD",
  "description": "订单支付",
  "customer_email": "user@example.com",
  "expires_in_minutes": 30
}

# 获取会话（公开API，不需要JWT）
GET /api/v1/cashier/sessions/:token

# 完成会话
POST /api/v1/cashier/sessions/:token/complete
{
  "payment_no": "PAY-xxx"
}

# 取消会话
DELETE /api/v1/cashier/sessions/:token
```

#### 3. 统计分析

```bash
# 获取分析数据
GET /api/v1/cashier/analytics?start_time=2024-10-01T00:00:00Z&end_time=2024-10-24T23:59:59Z

响应：
{
  "code": 0,
  "data": {
    "conversion_rate": 85.5,
    "channel_stats": {
      "stripe": 120,
      "paypal": 35
    },
    "total_sessions": 155
  }
}
```

### 管理员API（需要admin JWT）

```bash
# 列出模板
GET /api/v1/admin/cashier/templates

# 创建模板
POST /api/v1/admin/cashier/templates
{
  "name": "电商标准模板",
  "template_type": "ecommerce",
  "description": "适用于电商场景",
  "is_active": true,
  "config": {
    "theme_color": "#1890ff",
    "enabled_channels": ["stripe", "paypal"]
  }
}

# 更新模板
PUT /api/v1/admin/cashier/templates/:id
{...}

# 删除模板
DELETE /api/v1/admin/cashier/templates/:id

# 获取平台统计
GET /api/v1/admin/cashier/stats

响应：
{
  "code": 0,
  "data": {
    "total_merchants": 45,
    "active_cashiers": 42,
    "total_sessions": 15680,
    "avg_conversion_rate": 82.3,
    "total_sessions_today": 156,
    "completed_sessions_today": 132
  }
}
```

### 公开API（不需要认证）

```bash
# 记录用户行为日志
POST /api/v1/cashier/logs
{
  "session_token": "xxx",
  "user_ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "device_type": "mobile",
  "browser": "Chrome",
  "selected_channel": "stripe",
  "form_filled": true,
  "payment_submitted": true,
  "page_load_time": 1500,
  "time_to_submit": 45000
}
```

---

## 环境变量配置

### Cashier Service (后端)

```bash
# .env 或 export
DB_HOST=localhost
DB_PORT=40432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_cashier
DB_SSL_MODE=disable

REDIS_HOST=localhost
REDIS_PORT=40379
REDIS_PASSWORD=

PORT=40016
JWT_SECRET=your-super-secret-key-change-in-production

ENV=development
```

### Cashier Portal (前端)

```bash
# .env
VITE_STRIPE_PUBLIC_KEY=pk_test_51xxxxxxxxxxxxxxxxxxxxx
VITE_API_BASE_URL=http://localhost:40016/api/v1  # 可选
```

---

## 常见问题

### Q1: 页面显示"Invalid Session"

**原因**: 会话token无效或已过期

**解决**:
1. 检查URL中的token是否完整
2. 确认会话未超过过期时间（默认30分钟）
3. 重新创建会话

### Q2: Stripe支付按钮显示灰色

**原因**: Stripe公钥未配置或无效

**解决**:
```bash
cd frontend/cashier-portal
cp .env.example .env
# 编辑 .env，设置正确的 VITE_STRIPE_PUBLIC_KEY
npm run dev  # 重启
```

### Q3: 无法创建支付会话（401 Unauthorized）

**原因**: JWT token缺失或无效

**解决**:
1. 确保在请求头中包含 `Authorization: Bearer <token>`
2. 确认token未过期
3. 使用正确的商户token（不是管理员token）

### Q4: 支付提交后无响应

**原因**: Payment Gateway服务未运行

**解决**:
```bash
# 启动payment-gateway
cd /home/eric/payment/backend/services/payment-gateway
GOWORK=/home/eric/payment/backend/go.work go run ./cmd/main.go
```

### Q5: 数据库连接失败

**原因**: PostgreSQL未运行或数据库不存在

**解决**:
```bash
# 检查PostgreSQL
docker ps | grep payment-postgres

# 创建数据库（如不存在）
docker exec payment-postgres psql -U postgres -c "CREATE DATABASE payment_cashier;"
```

---

## 测试清单

使用以下清单验证系统是否正常工作：

- [ ] 后端服务启动成功 (http://localhost:40016/health)
- [ ] Admin Portal可访问 (http://localhost:5173)
- [ ] Merchant Portal可访问 (http://localhost:5174)
- [ ] Cashier Portal可访问 (http://localhost:5176)
- [ ] 商户可创建配置
- [ ] 商户可创建支付会话
- [ ] 收银台页面加载成功
- [ ] 商户主题正确应用
- [ ] 支付表单验证工作（卡号、有效期、CVV）
- [ ] Stripe支付成功（测试卡4242...）
- [ ] 3DS验证测试通过（测试卡4000 0025...）
- [ ] 支付失败正确处理（测试卡4000 0000 0002）
- [ ] 会话过期检测
- [ ] 重定向到成功页面
- [ ] 日志正确记录
- [ ] 统计数据准确
- [ ] 移动端响应式正常
- [ ] 多语言切换正常
- [ ] 管理员可创建模板
- [ ] 管理员可查看平台统计

---

## 性能测试

```bash
# 测试API响应时间
time curl http://localhost:40016/api/v1/cashier/sessions/test-token

# 测试并发创建会话
ab -n 100 -c 10 -H "Authorization: Bearer $TOKEN" \
   -T application/json -p session.json \
   http://localhost:40016/api/v1/cashier/sessions

# 测试前端加载时间
curl -o /dev/null -s -w "Total: %{time_total}s\n" \
  http://localhost:5176/checkout?token=xxx
```

---

## 生产部署检查

上线前确保完成以下配置：

- [ ] 更改JWT_SECRET为强密码
- [ ] 配置生产Stripe密钥（pk_live_xxx）
- [ ] 启用HTTPS（Let's Encrypt）
- [ ] 配置CORS允许的域名
- [ ] 设置数据库备份策略
- [ ] 配置Redis持久化
- [ ] 设置监控告警（Prometheus + Grafana）
- [ ] 配置日志聚合（ELK）
- [ ] 压力测试（目标: 1000 req/s）
- [ ] 安全扫描（OWASP ZAP）
- [ ] CDN配置（Cloudflare）
- [ ] 设置rate limiting
- [ ] 配置防火墙规则
- [ ] 准备灾难恢复计划

---

## 支持

如有问题，请查看：
- [完整实现文档](CASHIER_SYSTEM_COMPLETE.md)
- [Cashier Portal README](frontend/cashier-portal/README.md)
- [CLAUDE.md](CLAUDE.md) - 项目总体文档

快速启动成功！🚀
