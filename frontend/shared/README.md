# @payment/shared

支付平台前端共享库 - 包含工具函数、类型定义、自定义 Hooks 和共享组件

## 📦 包含内容

### Utils (工具函数)
- `format.ts` - 格式化工具（金额、日期、脱敏等）
- `validate.ts` - 数据验证（邮箱、手机、银行卡等）
- `debounce.ts` - 防抖/节流功能
- `storage.ts` - 本地存储（支持过期时间）

### Types (类型定义)
- API 响应类型
- 业务实体类型（Admin, Merchant, Payment, Order 等）
- 完整的 TypeScript 类型覆盖

### Hooks (自定义 Hooks)
- `useRequest` - 简化异步请求处理
- `useDebounce` - 防抖值处理

### Components (共享组件)
- `ErrorBoundary` - 错误边界组件

## 🚀 使用方法

在 workspace 项目中引用：

```typescript
// 导入工具函数
import { formatAmount, isEmail, debounce } from '@payment/shared/utils'

// 导入类型
import type { Payment, ApiResponse } from '@payment/shared/types'

// 导入 Hooks
import { useRequest, useDebounce } from '@payment/shared/hooks'

// 导入组件
import { ErrorBoundary } from '@payment/shared/components'

// 或者全部导入
import { formatAmount, type Payment, useRequest, ErrorBoundary } from '@payment/shared'
```

## 📝 配置

在各个前端项目的 `package.json` 中添加依赖：

```json
{
  "dependencies": {
    "@payment/shared": "workspace:*"
  }
}
```

## 🔧 开发

共享库使用 TypeScript，无需编译，直接引用源码。

## ⚠️ 注意事项

1. 修改共享库后，所有使用该库的项目都会受到影响
2. 保持向后兼容性，避免破坏性更改
3. 新增功能时注意文档完整性
