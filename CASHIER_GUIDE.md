# 收银台管理系统使用指南

## 📚 目录
1. [系统架构](#系统架构)
2. [快速开始](#快速开始)
3. [商户后台管理](#商户后台管理)
4. [管理员后台管理](#管理员后台管理)
5. [API使用说明](#api使用说明)
6. [常见问题](#常见问题)

---

## 系统架构

### 组件说明

```
┌─────────────────────────────────────────────────────┐
│  Cashier Service (后端微服务 - 40016端口)           │
│  - 收银台配置管理                                    │
│  - 支付会话管理                                      │
│  - 访问日志记录                                      │
│  - 数据统计分析                                      │
└─────────────────────────────────────────────────────┘
                    ↓ API调用
┌─────────────────────────────────────────────────────┐
│  Merchant Portal (商户后台 - 5174端口)              │
│  - 收银台外观配置                                    │
│  - 支付方式管理                                      │
│  - 数据分析查看                                      │
│  - 支付链接生成                                      │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Admin Portal (管理员后台 - 5173端口)               │
│  - 平台级模板管理                                    │
│  - 全局配置管理                                      │
│  - 所有商户监控                                      │
└─────────────────────────────────────────────────────┘
```

### 数据库表结构

- **cashier_configs** - 商户收银台配置
- **cashier_sessions** - 支付会话记录
- **cashier_logs** - 收银台访问日志
- **cashier_templates** - 平台模板(管理员管理)

---

## 快速开始

### 1. 启动后端服务

```bash
# 1. 确保数据库和Redis已启动
docker-compose up -d postgres redis

# 2. 启动 cashier-service
cd /home/eric/payment/backend/services/cashier-service

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=payment_cashier
export REDIS_HOST=localhost
export REDIS_PORT=40379
export PORT=40016
export JWT_SECRET=your-secret-key

# 启动服务
/tmp/cashier-service

# 或者使用air热重载
~/go/bin/air
```

### 2. 启动前端应用

```bash
# 启动 Merchant Portal
cd /home/eric/payment/frontend/merchant-portal
npm run dev
# 访问: http://localhost:5174
```

### 3. 访问收银台配置

1. 登录 Merchant Portal (http://localhost:5174)
2. 在左侧菜单中点击 **"收银台配置"**
3. 开始配置您的收银台

---

## 商户后台管理

### 如何通过网页后台管理收银台?

登录商户后台后,进入 **"收银台配置"** 页面,您可以管理以下内容:

### 📌 Tab 1: 外观设置

配置收银台的视觉效果,提供白标解决方案。

**可配置项**:
- ✅ **Logo上传**: 上传您的品牌Logo (推荐尺寸 200x60px)
- ✅ **主题颜色**: 使用颜色选择器自定义主题色
- ✅ **背景图片**: 设置收银台背景图片URL
- ✅ **自定义CSS**: 高级用户可编写CSS进一步定制样式
- ✅ **实时预览**: 查看配置效果的实时预览

**操作步骤**:
```
1. 进入"外观设置" Tab
2. 填写 Logo URL: https://your-cdn.com/logo.png
3. 选择主题颜色: 点击颜色选择器,选择您的品牌色
4. 点击页面顶部的"保存配置"按钮
5. 预览效果会实时更新
```

**示例配置**:
```json
{
  "logo_url": "https://example.com/logo.png",
  "theme_color": "#1890ff",
  "background_image_url": "https://example.com/bg.jpg",
  "custom_css": ".checkout-page { padding: 20px; }"
}
```

---

### 📌 Tab 2: 支付方式

管理支持的支付渠道和语言。

**可配置项**:
- ✅ **启用的支付渠道**: 选择 Stripe/PayPal/支付宝/微信支付
- ✅ **默认支付渠道**: 设置默认显示的渠道
- ✅ **允许用户切换渠道**: 开关,控制用户是否可以切换支付方式
- ✅ **支持的语言**: 多语言支持 (英文/简中/繁中/日语/韩语等)
- ✅ **默认语言**: 设置默认显示语言

**操作步骤**:
```
1. 进入"支付方式" Tab
2. 在"启用的支付渠道"中,勾选 Stripe 和 PayPal
3. 设置"默认支付渠道"为 Stripe
4. 在"支持的语言"中,选择 英文、简体中文、日语
5. 设置"默认语言"为英文
6. 点击"保存配置"
```

**支持的渠道**:
- ✅ Stripe (已集成)
- ⏳ PayPal (开发中)
- ⏳ 支付宝 (规划中)
- ⏳ 微信支付 (规划中)

---

### 📌 Tab 3: 安全设置

配置收银台安全策略。

**可配置项**:
- ✅ **会话超时时间**: 5-120分钟 (默认30分钟)
- ✅ **强制要求CVV验证**: 开关
- ✅ **启用3D Secure验证**: 开关,增强安全性
- ✅ **允许的国家/地区**: 留空表示允许所有国家
- ✅ **支付成功跳转URL**: 用户完成支付后跳转的页面
- ✅ **支付取消跳转URL**: 用户取消支付后跳转的页面

**操作步骤**:
```
1. 进入"安全设置" Tab
2. 设置"会话超时时间"为 30 分钟
3. 开启"强制要求CVV验证"
4. 开启"启用3D Secure验证"
5. 在"允许的国家/地区"中输入: US, CN, JP (仅允许这三个国家)
6. 设置回调URL:
   - 成功: https://yoursite.com/payment/success
   - 取消: https://yoursite.com/payment/cancel
7. 点击"保存配置"
```

**安全建议**:
- ✅ 生产环境务必开启 CVV 验证
- ✅ 启用 3D Secure 可显著降低拒付风险
- ✅ 设置会话超时防止恶意占用

---

### 📌 Tab 4: 数据分析

查看收银台使用数据和转化率。

**数据指标**:
- 📊 **转化率**: 完成支付的会话占总会话的百分比
- 📊 **总会话数**: 最近7天创建的支付会话总数
- 📊 **平均完成时间**: 用户从打开收银台到完成支付的平均时长
- 📊 **渠道偏好**: 饼图显示各支付渠道的使用占比
- 📊 **支付漏斗**: 漏斗图展示支付流程各环节的转化情况

**如何使用数据分析**:
```
1. 进入"数据分析" Tab
2. 查看"转化率"指标,如果低于60%,需要优化流程
3. 查看"渠道偏好",了解用户更喜欢用哪种支付方式
4. 分析"支付漏斗",找出用户流失最多的环节:
   - 如果"选择渠道"环节流失高,可能是渠道太多导致选择困难
   - 如果"填写信息"环节流失高,可能是表单太复杂
   - 如果"提交支付"环节流失高,可能是支付失败率高
```

**优化建议**:
- 转化率 < 60%: 简化支付流程,减少必填字段
- 平均时长 > 120秒: 优化页面加载速度,简化表单
- 某渠道使用率 < 5%: 考虑移除该渠道,简化选择

---

### 📌 Tab 5: 快捷工具

生成支付链接和测试收银台。

#### 5.1 生成支付链接

**使用场景**:
- 电商订单支付
- 发票支付
- 捐赠/众筹
- 临时收款

**操作步骤**:
```
1. 进入"快捷工具" Tab
2. 在"生成支付链接"卡片中:
   - 输入金额: 99.99
   - 选择货币: USD
   - 输入描述: 订单 #12345 支付
   - (可选) 输入客户邮箱: customer@example.com
3. 点击"生成链接"按钮
4. 弹窗显示:
   - 支付链接 (可复制)
   - 会话Token (可复制)
   - 二维码 (可扫码支付)
5. 将链接发送给客户,客户点击即可进入收银台支付
```

**生成的链接示例**:
```
https://yourplatform.com/cashier/checkout/ABC123XYZ...
```

**使用方式**:
- 📧 发送邮件给客户
- 💬 通过聊天工具分享
- 📱 生成二维码供扫描
- 🔗 嵌入到网页中

#### 5.2 测试收银台

**操作步骤**:
```
1. 点击"打开测试收银台"按钮
2. 在新窗口中查看收银台效果
3. 测试完整支付流程
4. 验证配置是否生效
```

---

## 管理员后台管理

### Admin Portal 功能 (待实现)

管理员可以通过 Admin Portal 进行平台级管理:

**功能**:
- 📋 **收银台模板管理**: 创建预设模板供商户使用
- ⚙️ **全局配置**: 设置默认渠道顺序、安全策略
- 📊 **平台监控**: 查看所有商户的收银台使用统计
- 🔍 **日志查询**: 查询收银台访问日志,排查问题

*(Admin Portal 页面将在后续开发)*

---

## API使用说明

### 服务端集成

商户后端可以直接调用 Cashier Service API 创建支付会话。

#### 1. 创建支付会话

```bash
POST http://localhost:40016/api/v1/cashier/sessions
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "order_no": "ORDER-20250124-001",
  "amount": 9999,  // 金额(分) 99.99 USD
  "currency": "USD",
  "description": "订单支付",
  "customer_email": "customer@example.com",
  "customer_name": "张三",
  "allowed_channels": ["stripe", "paypal"],
  "expires_in_minutes": 30,
  "metadata": {
    "order_id": "12345",
    "product": "Premium Plan"
  }
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "session_token": "ABC123XYZ...",
    "session": {
      "id": "uuid-1234",
      "session_token": "ABC123XYZ...",
      "order_no": "ORDER-20250124-001",
      "amount": 9999,
      "currency": "USD",
      "status": "pending",
      "expires_at": "2025-01-24T12:30:00Z"
    },
    "cashier_url": "/cashier/checkout/ABC123XYZ..."
  },
  "message": "success"
}
```

#### 2. 查询会话状态

```bash
GET http://localhost:40016/api/v1/cashier/sessions/ABC123XYZ...
Authorization: Bearer YOUR_JWT_TOKEN
```

#### 3. 完成会话 (支付成功后调用)

```bash
POST http://localhost:40016/api/v1/cashier/sessions/ABC123XYZ.../complete
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "payment_no": "PAY-20250124-001"
}
```

#### 4. 获取配置

```bash
GET http://localhost:40016/api/v1/cashier/configs
Authorization: Bearer YOUR_JWT_TOKEN
```

#### 5. 获取统计数据

```bash
GET http://localhost:40016/api/v1/cashier/analytics?start_time=2025-01-17T00:00:00Z&end_time=2025-01-24T23:59:59Z
Authorization: Bearer YOUR_JWT_TOKEN
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "conversion_rate": 72.5,
    "channel_stats": {
      "stripe": 450,
      "paypal": 180
    },
    "total_sessions": 630
  },
  "message": "success"
}
```

---

## 常见问题

### Q1: 如何快速测试收银台?

**A**:
1. 登录 Merchant Portal
2. 进入"收银台配置" → "快捷工具"
3. 填写测试金额,点击"生成链接"
4. 复制链接在新标签页打开
5. 完成测试支付流程

### Q2: 为什么我的Logo没有显示?

**A**:
- 检查 Logo URL 是否可访问
- 确认URL是 https:// 开头
- 推荐使用CDN托管图片
- 确保图片尺寸合理 (推荐200x60px)

### Q3: 如何提高转化率?

**A**:
1. **简化流程**: 减少不必要的表单字段
2. **优化加载**: 使用CDN加速静态资源
3. **信任建设**: 显示安全认证标志
4. **移动优先**: 确保移动端体验流畅
5. **多种支付**: 提供用户熟悉的支付方式

### Q4: 会话过期后怎么办?

**A**:
- 默认会话30分钟后过期
- 用户需要重新生成支付链接
- 建议设置合理的过期时间(根据业务场景)
- 可以在"安全设置"中调整超时时间

### Q5: 如何集成到我的网站?

**A**:
**方式1: 重定向方式**
```javascript
// 后端创建会话
const response = await fetch('http://localhost:40016/api/v1/cashier/sessions', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    order_no: 'ORDER-001',
    amount: 9999,
    currency: 'USD'
  })
})

const { data } = await response.json()

// 跳转到收银台
window.location.href = `https://pay.yourplatform.com${data.cashier_url}`
```

**方式2: Iframe嵌入**
```html
<iframe
  src="https://pay.yourplatform.com/cashier/checkout/ABC123..."
  width="600"
  height="800"
  frameborder="0">
</iframe>
```

### Q6: 如何保证支付安全?

**A**:
- ✅ 启用 CVV 验证
- ✅ 启用 3D Secure
- ✅ 使用 HTTPS
- ✅ 设置会话超时
- ✅ 验证回调签名
- ✅ 记录所有操作日志

### Q7: 支持哪些货币?

**A**:
目前支持 32+ 种货币,包括:
- 主流货币: USD, EUR, GBP, JPY, CNY, AUD, CAD
- 加密货币: BTC, ETH, USDT (规划中)

完整列表见 `pkg/validator/currency.go`

### Q8: 如何获取技术支持?

**A**:
- 📧 Email: support@yourplatform.com
- 💬 在线客服: 商户后台右下角
- 📖 文档: https://docs.yourplatform.com
- 🐛 Bug反馈: https://github.com/yourrepo/issues

---

## 总结

通过 **Merchant Portal 的收银台配置页面**,您可以:

✅ **自定义外观** - Logo/颜色/背景图/CSS
✅ **管理支付方式** - 启用/禁用渠道,设置默认渠道
✅ **配置安全策略** - CVV/3DS/超时/国家限制
✅ **查看数据分析** - 转化率/渠道统计/支付漏斗
✅ **生成支付链接** - 快速创建收款链接和二维码
✅ **测试收银台** - 实时预览配置效果

**下一步**:
1. 登录 Merchant Portal
2. 进入"收银台配置"
3. 按照本指南完成配置
4. 生成测试链接验证效果
5. 集成到您的业务系统

**需要帮助?** 请联系技术支持团队!
