# Merchant Portal 401 错误修复

## 问题描述

Merchant Portal 前端在未登录状态下访问页面时，出现大量 401 错误。

## 根本原因

1. **Dashboard、Transactions、Orders 页面**在组件加载时（`useEffect`）立即调用 API
2. 此时用户可能还未登录，或者浏览器中存储了过期的 token
3. 导致大量 401 错误出现在控制台

## 修复内容

### 1. 添加 Token 检查
在以下页面的数据加载函数中添加 token 检查：

- ✅ `Dashboard.tsx` - loadDashboardData()
- ✅ `Transactions.tsx` - loadPayments(), loadStats()  
- ✅ `Orders.tsx` - loadOrders(), loadStats()

修改示例：
```typescript
const loadDashboardData = async () => {
  // 检查是否已登录
  const token = useAuthStore.getState().token
  if (!token) {
    console.log('No token found, skipping dashboard data load')
    return
  }
  
  // ... 原有代码
}
```

### 2. 图表优雅降级
在 Dashboard.tsx 中，图表在无数据时显示"暂无数据"而不是报错：

```typescript
{trendData.length > 0 ? (
  <Line {...lineConfig} />
) : (
  <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
    暂无数据
  </div>
)}
```

### 3. WebSocket 暂时禁用
在 WebSocketProvider.tsx 中暂时禁用 WebSocket 连接，避免错误干扰。

## 使用指南

### 方法 1：清除浏览器缓存并重新登录（推荐）

1. 打开浏览器控制台（F12）
2. 执行以下命令：
   ```javascript
   localStorage.clear()
   location.reload()
   ```

3. 访问登录页面：http://localhost:5174/login

4. 使用测试账号登录：
   - Email: `test@test.com`
   - Password: `password123`

5. 登录成功后，将自动跳转到 Dashboard，不再出现 401 错误

### 方法 2：直接访问登录页

如果页面自动跳转到登录页（ProtectedRoute 生效），直接使用测试账号登录即可。

## 验证修复

登录成功后，以下操作应该正常工作：

1. ✅ Dashboard 页面正常显示统计数据
2. ✅ Transactions 页面可以查看交易列表
3. ✅ Orders 页面可以查看订单列表
4. ✅ 不再出现 401 错误（除非 token 过期）
5. ✅ 浏览器控制台不再有大量错误信息

## Kong 配置状态

Kong API Gateway 已正确配置：

- ✅ Merchant Service 路由：`/api/v1/merchant`, `/api/v1/dashboard`
- ✅ Payment Gateway 路由：`/api/v1/payments`
- ✅ Order Service 路由：`/api/v1/orders`
- ✅ Admin Service 路由：`/api/v1/admin`, `/api/v1/merchants`
- ✅ CORS 插件已启用
- ✅ 认证由后端服务处理（Kong 不拦截）

## 测试 API

### 登录测试
```bash
curl -X POST http://localhost:40080/api/v1/merchant/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123"}'
```

### Dashboard API 测试（需要 token）
```bash
TOKEN="你的_TOKEN"
curl http://localhost:40080/api/v1/dashboard \
  -H "Authorization: Bearer $TOKEN"
```

## 已知问题

1. **404 错误**：`/api/v1/merchant/payments/stats` 接口可能还未实现，返回 404
   - 影响：Transactions 页面统计数据不显示
   - 状态：前端已做容错处理，不会影响页面使用

2. **PWA 图标缺失**：`pwa-192x192.png` 文件不存在
   - 影响：仅影响 PWA 功能，不影响正常使用
   - 状态：可以后续添加图标文件

## 相关脚本

- `scripts/setup-kong.sh` - 配置 Kong 路由
- `scripts/reset-kong.sh` - 重置 Kong 配置

## 技术栈

- **Kong Gateway**: 3.9 (API Gateway)
- **React**: 18 (前端框架)
- **Vite**: 5 (开发服务器，支持代理)
- **Ant Design**: 5.15 (UI 组件库)
- **Zustand**: 4.5 (状态管理)

---

最后更新：2025-10-24

