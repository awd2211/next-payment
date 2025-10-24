# 前端 Kong Gateway 迁移指南

## 一、迁移概述

### 变更内容
将前端应用的 API 调用从 **直连微服务端口** 迁移到 **通过 Kong Gateway 统一入口**。

### 影响范围
- Admin Portal (frontend/admin-portal)
- Merchant Portal (frontend/merchant-portal)
- Website (frontend/website) - 如有 API 调用

---

## 二、Admin Portal 迁移

### Step 1: 更新环境变量

**文件**: `frontend/admin-portal/.env.local` (新建或修改)

```bash
# Kong Gateway 统一入口
VITE_API_BASE_URL=http://localhost:40080

# API 超时时间
VITE_API_TIMEOUT=30000

# 环境标识
VITE_ENV=development
```

**文件**: `frontend/admin-portal/.env.production` (生产环境)

```bash
# 生产环境使用真实域名
VITE_API_BASE_URL=https://api.yourdomain.com

VITE_API_TIMEOUT=30000

VITE_ENV=production
```

---

### Step 2: 更新 API 配置

**文件**: `frontend/admin-portal/src/services/api.ts`

```typescript
import axios, { AxiosInstance, AxiosError, AxiosResponse } from 'axios';
import { message } from 'antd';

// 创建 Axios 实例
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:40080',
  timeout: Number(import.meta.env.VITE_API_TIMEOUT) || 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加认证 Token
api.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 JWT token
    const token = localStorage.getItem('admin_token');

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // 添加请求追踪 ID (可选,Kong 会自动生成)
    // config.headers['X-Request-ID'] = generateUUID();

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误
api.interceptors.response.use(
  (response: AxiosResponse) => {
    // Kong 会添加 X-Request-ID 响应头,可用于日志追踪
    const requestId = response.headers['x-request-id'];
    if (requestId && import.meta.env.VITE_ENV === 'development') {
      console.log(`[API] Request ID: ${requestId}`);
    }

    return response;
  },
  (error: AxiosError) => {
    const { response } = error;

    if (!response) {
      // 网络错误
      message.error('网络连接失败,请检查网络设置');
      return Promise.reject(error);
    }

    switch (response.status) {
      case 401:
        // JWT 认证失败 (Kong JWT 插件返回)
        message.error('登录已过期,请重新登录');
        localStorage.removeItem('admin_token');
        localStorage.removeItem('admin_user');
        window.location.href = '/login';
        break;

      case 403:
        // 权限不足
        message.error('权限不足,无法访问该资源');
        break;

      case 429:
        // Kong Rate Limiting 插件返回
        const retryAfter = response.headers['retry-after'];
        const waitTime = retryAfter ? `${retryAfter}秒` : '稍后';
        message.error(`请求过于频繁,请${waitTime}再试`);
        break;

      case 502:
      case 503:
      case 504:
        // Kong 上游服务不可用
        message.error('服务暂时不可用,请稍后再试');
        break;

      default:
        // 其他错误
        const errorMessage = response.data?.message || '请求失败';
        message.error(errorMessage);
    }

    return Promise.reject(error);
  }
);

export default api;
```

---

### Step 3: 更新登录逻辑

**文件**: `frontend/admin-portal/src/pages/Login/index.tsx`

```typescript
import { useState } from 'react';
import { Form, Input, Button, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';

interface LoginForm {
  username: string;
  password: string;
}

const Login: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const onFinish = async (values: LoginForm) => {
    setLoading(true);

    try {
      // 通过 Kong Gateway 调用登录接口
      // Kong 会转发到 admin-service:40001
      const response = await api.post('/api/v1/admin/login', {
        username: values.username,
        password: values.password,
      });

      const { token, admin } = response.data;

      // 保存 Token 和用户信息
      localStorage.setItem('admin_token', token);
      localStorage.setItem('admin_user', JSON.stringify(admin));

      message.success('登录成功');
      navigate('/dashboard');
    } catch (error) {
      // 错误处理已在 api.ts 拦截器中处理
      console.error('Login failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <Form
        name="login"
        onFinish={onFinish}
        autoComplete="off"
        size="large"
      >
        <Form.Item
          name="username"
          rules={[{ required: true, message: '请输入用户名' }]}
        >
          <Input
            prefix={<UserOutlined />}
            placeholder="用户名"
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[{ required: true, message: '请输入密码' }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="密码"
          />
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
          >
            登录
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
};

export default Login;
```

---

### Step 4: 测试迁移

```bash
cd frontend/admin-portal

# 安装依赖 (如果需要)
npm install

# 启动开发服务器
npm run dev

# 访问 http://localhost:5173
# 尝试登录,检查 Network 面板:
# - Request URL 应该是 http://localhost:40080/api/v1/admin/login
# - Response Headers 包含 X-Request-ID (Kong 添加)
# - 登录成功后,访问其他页面,检查 Authorization 头部
```

---

## 三、Merchant Portal 迁移

### 完全相同的步骤

Merchant Portal 的迁移步骤与 Admin Portal 完全一致:

1. ✅ 更新 `.env.local`: `VITE_API_BASE_URL=http://localhost:40080`
2. ✅ 更新 `src/services/api.ts` (复制上述代码,修改 token key 为 `merchant_token`)
3. ✅ 更新登录页面 (API 路径为 `/api/v1/merchant/login`)
4. ✅ 测试登录和 API 调用

**差异点**:

```typescript
// Merchant Portal: src/services/api.ts

// 请求拦截器
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('merchant_token'); // 注意 key 名称
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器 401 处理
api.interceptors.response.use(null, (error) => {
  if (error.response?.status === 401) {
    localStorage.removeItem('merchant_token');
    localStorage.removeItem('merchant_user');
    window.location.href = '/login';
  }
  // ...
});
```

---

## 四、API 调用示例对比

### 变更前 (直连服务)

```typescript
// Admin Portal
const response = await axios.post('http://localhost:40001/api/v1/admin/login', data);

// Merchant Portal
const response = await axios.post('http://localhost:40002/api/v1/merchant/login', data);

// Payment API
const response = await axios.post('http://localhost:40003/api/v1/payments', data);
```

**问题**:
- ❌ 需要配置多个 Base URL
- ❌ CORS 配置繁琐
- ❌ 无统一错误处理
- ❌ 缺少请求追踪

---

### 变更后 (通过 Kong)

```typescript
// 所有服务统一通过 Kong Gateway
const api = axios.create({
  baseURL: 'http://localhost:40080',  // 统一入口
});

// Admin Portal
const response = await api.post('/api/v1/admin/login', data);

// Merchant Portal
const response = await api.post('/api/v1/merchant/login', data);

// Payment API (需要 API Key)
const response = await api.post('/api/v1/payments', data, {
  headers: { 'X-API-Key': 'sk_test_xxx' }
});
```

**优势**:
- ✅ 单一 Base URL 配置
- ✅ Kong 自动处理 CORS
- ✅ 统一错误码 (401, 429, 502 等)
- ✅ 自动添加 X-Request-ID 追踪

---

## 五、Kong 错误码处理

### 常见 Kong 错误响应

#### 1. 401 Unauthorized (JWT 认证失败)

**原因**: JWT token 无效、过期或缺失

**Kong 响应**:
```json
{
  "message": "Unauthorized"
}
```

**前端处理**:
```typescript
if (response.status === 401) {
  // 清除本地 token
  localStorage.removeItem('admin_token');
  // 跳转登录页
  window.location.href = '/login';
}
```

---

#### 2. 429 Too Many Requests (限流)

**原因**: 超过 Kong Rate Limiting 配置的阈值

**Kong 响应**:
```json
{
  "message": "API rate limit exceeded"
}
```

**响应头**:
```
X-RateLimit-Limit-Minute: 10
X-RateLimit-Remaining-Minute: 0
Retry-After: 45
```

**前端处理**:
```typescript
if (response.status === 429) {
  const retryAfter = response.headers['retry-after'];
  message.error(`请求过于频繁,请 ${retryAfter} 秒后再试`);

  // 可选: 自动重试
  setTimeout(() => {
    // Retry request
  }, retryAfter * 1000);
}
```

---

#### 3. 502 Bad Gateway (上游服务不可用)

**原因**: Kong 无法连接到后端服务 (服务未启动或崩溃)

**Kong 响应**:
```json
{
  "message": "An invalid response was received from the upstream server"
}
```

**前端处理**:
```typescript
if (response.status === 502) {
  message.error('服务暂时不可用,请稍后再试');
  // 可选: 显示维护页面
}
```

---

#### 4. 503 Service Unavailable (Kong 过载)

**原因**: Kong 自身过载或上游服务全部不可用

**前端处理**:
```typescript
if (response.status === 503) {
  message.error('系统繁忙,请稍后再试');
}
```

---

## 六、高级功能

### 1. 请求重试 (Retry on 5xx)

```typescript
import axios from 'axios';
import axiosRetry from 'axios-retry';

const api = axios.create({
  baseURL: 'http://localhost:40080',
});

// 配置自动重试
axiosRetry(api, {
  retries: 3,                          // 最多重试 3 次
  retryDelay: axiosRetry.exponentialDelay,  // 指数退避
  retryCondition: (error) => {
    // 仅重试 5xx 错误和网络错误
    return axiosRetry.isNetworkOrIdempotentRequestError(error)
      || (error.response?.status ?? 0) >= 500;
  },
});
```

---

### 2. 请求追踪

```typescript
// 在请求拦截器中记录 Request ID
api.interceptors.response.use((response) => {
  const requestId = response.headers['x-request-id'];

  // 记录到日志服务 (如 Sentry, LogRocket)
  if (window.Sentry) {
    window.Sentry.setContext('api_request', {
      request_id: requestId,
      url: response.config.url,
      method: response.config.method,
    });
  }

  return response;
});
```

---

### 3. 性能监控

```typescript
// 记录 API 调用性能
api.interceptors.request.use((config) => {
  config.metadata = { startTime: Date.now() };
  return config;
});

api.interceptors.response.use((response) => {
  const duration = Date.now() - response.config.metadata.startTime;

  // 上报性能指标
  if (window.gtag) {
    window.gtag('event', 'api_request', {
      event_category: 'API',
      event_label: response.config.url,
      value: duration,
    });
  }

  // 警告慢请求
  if (duration > 3000) {
    console.warn(`Slow API request: ${response.config.url} (${duration}ms)`);
  }

  return response;
});
```

---

## 七、测试清单

### 功能测试

- [ ] 登录成功 (Admin Portal)
- [ ] 登录成功 (Merchant Portal)
- [ ] 登录失败 (错误密码)
- [ ] JWT 过期自动跳转登录页
- [ ] 创建资源 (POST)
- [ ] 查询资源 (GET)
- [ ] 更新资源 (PUT)
- [ ] 删除资源 (DELETE)
- [ ] 权限不足 (403) 提示
- [ ] Rate Limit (429) 提示

---

### 网络测试

- [ ] CORS 预检请求成功
- [ ] 请求头包含 Authorization
- [ ] 响应头包含 X-Request-ID
- [ ] 后端服务停止时显示 502 错误
- [ ] 网络断开时显示连接失败

---

### 性能测试

- [ ] API 调用延迟 <200ms (P95)
- [ ] 页面加载时间无明显变化
- [ ] 大量并发请求无卡顿

---

## 八、回滚方案

如果迁移后发现问题,可快速回滚:

### 方式 1: 修改环境变量 (推荐)

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:40001  # 回滚到 admin-service 直连
```

重启前端开发服务器:
```bash
npm run dev
```

---

### 方式 2: Git 回滚代码

```bash
# 回滚到迁移前的 commit
git revert <migration-commit-hash>
git push
```

---

## 九、常见问题

### Q1: CORS 错误 "blocked by CORS policy"

**原因**: Kong CORS 插件未配置前端域名

**解决**:
```bash
curl -X PATCH http://localhost:40081/plugins/{cors-plugin-id} \
  --data "config.origins=http://localhost:5173" \
  --data "config.origins=http://localhost:5174"
```

---

### Q2: 401 错误但 Token 有效

**原因**: JWT `iss` (issuer) 与 Kong Consumer key 不匹配

**解决**:
1. 检查后端 JWT 签发时的 `iss` 值
2. 运行 `bash backend/scripts/kong-jwt-setup.sh` 创建对应的 Consumer

---

### Q3: 429 限流错误频繁出现

**原因**: 开发环境限流阈值过低 (10/分钟)

**解决**:
```bash
# 临时提高限流阈值
curl -X PATCH http://localhost:40081/plugins/{rate-limiting-plugin-id} \
  --data "config.minute=1000"
```

---

### Q4: Request ID 未出现在响应头

**原因**: Kong correlation-id 插件未启用

**解决**:
```bash
# 运行配置脚本
bash backend/scripts/kong-setup.sh
```

---

## 十、总结

### 迁移步骤回顾

1. ✅ 更新 `.env.local`: `VITE_API_BASE_URL=http://localhost:40080`
2. ✅ 更新 `src/services/api.ts` (添加 401/429 错误处理)
3. ✅ 测试登录流程
4. ✅ 测试所有 CRUD 操作
5. ✅ 检查 Network 面板 (Request ID, Rate Limit 头部)

### 预期收益

- **开发体验**: 单一 API Base URL,配置简化
- **错误处理**: 统一 401/429/502 错误码,用户体验改善
- **可观测性**: X-Request-ID 自动追踪,问题排查更容易
- **安全性**: Kong 统一认证和限流,减少前端配置错误

---

**文档版本**: v1.0
**最后更新**: 2025-10-24
**维护者**: Payment Platform Frontend Team
