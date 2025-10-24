# 收银台快速启动指南

## 环境配置

### 1. 配置 Stripe 公钥

编辑 `frontend/merchant-portal/.env.development`:

```env
# Stripe 测试公钥 (从 https://dashboard.stripe.com/test/apikeys 获取)
VITE_STRIPE_PUBLIC_KEY=pk_test_51xxxxxxxxxxxxx

# API 基础URL (可选,Kong网关)
VITE_API_PREFIX=http://localhost:40080/api/v1
```

### 2. 启动服务

#### 启动后端服务

```bash
cd backend

# 启动所有服务 (包括 cashier-service 在端口 40016)
./scripts/start-all-services.sh

# 检查服务状态
./scripts/status-all-services.sh

# 验证 cashier-service
curl http://localhost:40016/health
```

#### 启动商户前端

```bash
cd frontend/merchant-portal

# 安装依赖 (首次运行)
pnpm install

# 启动开发服务器
pnpm dev

# 访问: http://localhost:5174
```

## 使用流程

### 方式一: 通过商户后台快捷生成

1. **登录商户后台**
   ```
   URL: http://localhost:5174/login
   用户名: test@test.com
   密码: password123
   ```

2. **进入收银台配置**
   ```
   导航: 侧边栏 → "收银台配置"
   ```

3. **配置收银台外观** (可选)
   - 外观设置: Logo URL, 主题颜色, 背景图
   - 支付方式: 启用 Stripe
   - 安全设置: 会话超时时间

4. **生成支付链接**
   - 切换到 "快捷工具" 标签
   - 填写:
     - 金额: `99.99` (美元)
     - 货币: `USD`
     - 描述: `测试商品`
     - 客户邮箱: `customer@example.com`
   - 点击 "生成链接"
   - 复制生成的链接,或扫描二维码

5. **测试支付**
   - 在新标签页打开生成的链接
   - 填写支付信息:
     - 持卡人姓名: `John Doe`
     - 卡号: `4242 4242 4242 4242`
     - 有效期: `12/30`
     - CVV: `123`
   - 点击 "立即支付"

### 方式二: 通过 API 创建会话

#### 创建支付会话

```bash
# 获取 JWT Token (先登录)
TOKEN=$(curl -s -X POST http://localhost:40002/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test.com",
    "password": "password123"
  }' | jq -r '.data.token')

# 创建收银台会话
curl -X POST http://localhost:40016/api/v1/cashier/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_no": "ORDER-'$(date +%s)'",
    "amount": 9999,
    "currency": "USD",
    "description": "测试商品购买",
    "customer_email": "customer@example.com",
    "expires_in_minutes": 30
  }'
```

响应示例:
```json
{
  "code": 0,
  "data": {
    "session_token": "abc123xyz...",
    "cashier_url": "/cashier/checkout/abc123xyz...",
    "session": {
      "id": "uuid",
      "order_no": "ORDER-1729789234",
      "amount": 9999,
      "currency": "USD",
      "status": "pending",
      "expires_at": "2025-10-24T18:30:00Z"
    }
  },
  "message": "success"
}
```

#### 访问收银台

```
http://localhost:5174/cashier/checkout?token=abc123xyz...
```

## Stripe 测试卡号

### 成功支付
```
卡号: 4242 4242 4242 4242
有效期: 任意未来日期 (如 12/30)
CVV: 任意3位数字 (如 123)
```

### 3D Secure 验证
```
卡号: 4000 0025 0000 3155
有效期: 12/30
CVV: 123
→ 会弹出 3DS 验证窗口,点击 "Complete" 完成验证
```

### 支付失败 (测试错误处理)
```
卡号: 4000 0000 0000 0002
有效期: 12/30
CVV: 123
→ 返回 "Card declined" 错误
```

### 其他测试卡号
```
# 余额不足
4000 0000 0000 9995

# 过期卡片
4000 0000 0000 0069

# 错误的CVC
4000 0000 0000 0127
```

完整测试卡号列表: https://stripe.com/docs/testing

## 常见问题

### 1. 收银台页面空白

**原因**: Session token 无效或已过期

**解决**:
- 检查 token 是否正确
- 重新生成支付链接
- 检查会话是否过期 (默认30分钟)

```bash
# 验证会话
curl http://localhost:40016/api/v1/cashier/sessions/$TOKEN \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### 2. Stripe 支付失败

**原因**: Stripe 公钥未配置或无效

**解决**:
```bash
# 检查环境变量
cd frontend/merchant-portal
cat .env.development | grep STRIPE

# 重启开发服务器
pnpm dev
```

### 3. API 请求 401 未授权

**原因**: JWT Token 过期

**解决**:
```bash
# 重新登录获取新 token
curl -X POST http://localhost:40002/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test.com",
    "password": "password123"
  }'
```

### 4. CORS 错误

**原因**: API 跨域配置问题

**解决**:
- 确保使用 Kong 网关 (http://localhost:40080)
- 或者后端服务启用了 CORS

## 收银台功能清单

### 已实现 ✅

- [x] 会话管理 (创建、获取、完成、取消)
- [x] 商户配置 (主题、Logo、支付渠道)
- [x] Stripe 支付集成
- [x] 信用卡表单验证 (Luhn 算法)
- [x] 支付成功/失败处理
- [x] 会话过期检测
- [x] 多语言支持 (英文、简体中文)
- [x] 响应式设计
- [x] 支付链接生成
- [x] 二维码支付链接

### 待实现 ⏳

- [ ] PayPal 支付集成
- [ ] Alipay (支付宝) 集成
- [ ] WeChat Pay (微信支付) 集成
- [ ] 支付日志记录
- [ ] 支付数据分析
- [ ] 保存卡片功能
- [ ] 多卡支付
- [ ] 分期付款

## 目录结构

```
frontend/merchant-portal/
├── src/
│   ├── components/
│   │   ├── StripePaymentForm.tsx     # Stripe 支付表单 ✨
│   │   └── ...
│   ├── pages/
│   │   ├── CashierConfig.tsx         # 收银台配置页面
│   │   ├── CashierCheckout.tsx       # 收银台结账页面 ✨
│   │   └── ...
│   ├── services/
│   │   ├── cashierService.ts         # 收银台 API 服务
│   │   └── ...
│   ├── utils/
│   │   ├── cardValidation.ts         # 卡片验证工具 ✨
│   │   └── ...
│   └── i18n/
│       └── locales/
│           ├── en-US.json            # 英文翻译 (已更新)
│           └── zh-CN.json            # 中文翻译 (已更新)
└── .env.development                  # 环境变量配置
```

## 相关文档

- [收银台整合总结](./CASHIER_INTEGRATION_SUMMARY.md)
- [前端架构说明](./README.md)
- [Stripe 官方文档](https://stripe.com/docs)
- [Stripe 测试指南](https://stripe.com/docs/testing)

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite 5** - 构建工具
- **Ant Design 5.15** - UI 组件库
- **@stripe/react-stripe-js** - Stripe React 集成
- **@stripe/stripe-js** - Stripe JavaScript SDK
- **react-i18next** - 国际化
- **React Router** - 路由管理

---

*Last updated: 2025-10-24*
