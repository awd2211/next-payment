# Cashier Portal

收银台前端应用 - 为终端用户提供的安全支付页面。

## 功能特性

### 核心功能
- **安全支付页面** - 符合PCI DSS标准的支付表单
- **多支付渠道** - 支持Stripe、PayPal等多种支付方式
- **实时验证** - 信用卡号Luhn算法验证、有效期、CVV验证
- **会话管理** - 基于token的安全会话机制
- **主题定制** - 商户可自定义logo、颜色、背景图
- **多语言支持** - 英文、简体中文（可扩展）
- **响应式设计** - 支持手机、平板、桌面设备

### 支付集成
- **Stripe Elements** - 集成Stripe官方支付组件
- **3D Secure** - 支持3DS验证增强安全性
- **保存卡片** - 支持保存卡片信息用于未来支付
- **实时反馈** - 支付状态实时更新

### 用户体验
- **页面加载分析** - 记录页面加载时间
- **行为追踪** - 记录用户选择的支付渠道、表单填写时间
- **转化漏斗** - 追踪用户在支付流程中的drop-off点
- **错误记录** - 记录支付失败原因用于后续分析

## 技术栈

- **React 18** - UI框架
- **TypeScript** - 类型安全
- **Vite 5** - 构建工具
- **Ant Design 5.15** - UI组件库
- **Stripe React** - Stripe支付集成
- **react-i18next** - 国际化
- **Axios** - HTTP客户端

## 快速开始

### 安装依赖

```bash
npm install
```

### 配置环境变量

复制 `.env.example` 为 `.env` 并配置：

```bash
cp .env.example .env
```

编辑 `.env`：

```env
VITE_STRIPE_PUBLIC_KEY=pk_test_your_stripe_public_key
```

### 启动开发服务器

```bash
npm run dev
```

访问：http://localhost:5176

### 构建生产版本

```bash
npm run build
```

## 使用方式

### 创建支付会话

商户后端调用 cashier-service API 创建支付会话：

```bash
curl -X POST http://localhost:40016/api/v1/cashier/sessions \
  -H "Authorization: Bearer $MERCHANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "uuid",
    "order_no": "ORDER-001",
    "amount": 10000,
    "currency": "USD",
    "description": "订单支付",
    "customer_email": "customer@example.com",
    "customer_name": "John Doe",
    "expires_in_minutes": 30
  }'
```

响应：

```json
{
  "code": 0,
  "data": {
    "session_token": "xyz123...",
    "expires_at": "2024-10-24T10:30:00Z"
  }
}
```

### 生成支付链接

```
http://localhost:5176/checkout?token=xyz123...
```

将此链接发送给用户，用户打开后即可看到支付页面。

### 支付流程

1. 用户访问支付链接
2. 收银台加载会话信息和商户配置
3. 用户选择支付渠道（如果有多个）
4. 填写支付信息（卡号、有效期、CVV等）
5. 提交支付
6. 对于Stripe：使用Stripe Elements处理，支持3DS
7. 支付成功后重定向到商户的成功页面

## 项目结构

```
cashier-portal/
├── src/
│   ├── components/         # React组件
│   │   └── StripePaymentForm.tsx  # Stripe支付表单
│   ├── pages/              # 页面
│   │   └── Checkout.tsx    # 收银台主页面
│   ├── services/           # API服务
│   │   └── cashierApi.ts   # 收银台API调用
│   ├── types/              # TypeScript类型
│   │   └── index.ts
│   ├── utils/              # 工具函数
│   │   └── validation.ts   # 表单验证（Luhn算法等）
│   ├── i18n/               # 国际化
│   │   ├── index.ts
│   │   └── locales/
│   │       ├── en.json
│   │       └── zh-CN.json
│   ├── App.tsx             # 主应用
│   ├── main.tsx            # 入口文件
│   └── index.css           # 全局样式
├── package.json
├── vite.config.ts
├── tsconfig.json
└── README.md
```

## 关键组件

### Checkout.tsx

主收银台页面，包含：
- 会话加载和验证
- 商户配置应用（主题、logo等）
- 订单摘要显示
- 支付渠道选择
- 支付表单（原生或Stripe Elements）
- 用户行为日志记录

### StripePaymentForm.tsx

Stripe支付表单，包含：
- Stripe CardElement集成
- 支付意图创建
- 3D Secure验证
- 支付确认

### validation.ts

支付表单验证工具：
- `validateCardNumber()` - Luhn算法验证卡号
- `validateExpiryDate()` - 有效期验证
- `validateCVV()` - CVV验证
- `formatCardNumber()` - 卡号格式化（每4位加空格）
- `formatExpiryDate()` - 有效期格式化（MM/YY）
- `getCardType()` - 识别卡类型（Visa, Mastercard等）

## 安全特性

- **HTTPS强制** - 生产环境仅支持HTTPS
- **会话token** - 一次性token，使用后失效
- **会话过期** - 可配置过期时间（默认30分钟）
- **Stripe Elements** - 不接触原始卡号，符合PCI DSS
- **CSP头** - Content Security Policy防止XSS
- **表单验证** - 前端+后端双重验证
- **日志记录** - 记录所有支付尝试用于审计

## 环境变量

| 变量名 | 描述 | 默认值 | 必需 |
|--------|------|--------|------|
| `VITE_STRIPE_PUBLIC_KEY` | Stripe公钥 | - | 是（使用Stripe时）|
| `VITE_API_BASE_URL` | API基础URL | `/api/v1` | 否 |

## API集成

### 获取会话

```typescript
const session = await cashierApi.getSession(sessionToken)
```

### 获取配置

```typescript
const config = await cashierApi.getConfig(merchantId)
```

### 记录日志

```typescript
await cashierApi.recordLog({
  session_token: 'xyz',
  selected_channel: 'stripe',
  form_filled: true,
})
```

### 创建支付

```typescript
const result = await cashierApi.createPayment({
  session_token: 'xyz',
  channel: 'stripe',
  payment_method: 'card',
})
```

## 浏览器支持

- Chrome (最新版)
- Firefox (最新版)
- Safari (最新版)
- Edge (最新版)
- Mobile Safari (iOS 12+)
- Chrome Mobile (Android 8+)

## 性能优化

- **代码分割** - 动态import减小初始加载
- **懒加载** - Stripe Elements按需加载
- **CDN** - 静态资源使用CDN
- **缓存策略** - Service Worker缓存静态资源
- **图片优化** - WebP格式，响应式图片

## 开发指南

### 本地测试Stripe支付

1. 注册Stripe测试账号：https://stripe.com/
2. 获取测试公钥（pk_test_xxx）
3. 配置到 `.env` 文件
4. 使用Stripe测试卡号：
   - 成功：`4242 4242 4242 4242`
   - 3DS验证：`4000 0025 0000 3155`
   - 失败：`4000 0000 0000 0002`

### 添加新支付渠道

1. 在 `types/index.ts` 添加渠道类型
2. 创建新组件（如 `PayPalPaymentForm.tsx`）
3. 在 `Checkout.tsx` 添加渠道选择逻辑
4. 更新 `cashierApi.ts` 添加对应API调用

## 故障排查

### 页面空白
- 检查会话token是否有效
- 检查浏览器控制台是否有错误
- 确认cashier-service是否运行

### Stripe支付失败
- 检查VITE_STRIPE_PUBLIC_KEY是否配置
- 检查Stripe测试卡号是否正确
- 查看浏览器Network标签查看API响应

### 样式不显示
- 确认Ant Design样式已正确引入
- 检查商户自定义CSS是否有语法错误

## License

MIT
