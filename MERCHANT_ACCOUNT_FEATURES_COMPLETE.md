# 商户后台账户功能完善总结

## 📋 需求完成情况

✅ **1. 修改密码功能**
✅ **2. 双因素认证(2FA)设置**
✅ **3. 活动记录查看**
✅ **4. 偏好设置(语言/时区/货币等)**

---

## 🎯 功能详情

### 1. 修改密码功能

**路径**: 账户设置 → 安全设置 → 修改密码

**功能特性**:
- ✅ 当前密码验证
- ✅ 新密码强度实时检测(弱/中/强)
- ✅ 密码确认验证
- ✅ 最小8位长度要求
- ✅ 密码强度提示(颜色编码)
  - 🔴 弱: 纯数字或字母
  - 🟠 中: 2-3种字符类型
  - 🟢 强: 包含大小写字母、数字、特殊字符

**安全验证**:
```typescript
// 使用 utils/security.ts 中的密码强度验证
const result = validatePasswordStrength(password)
// 返回: { strength: 'weak' | 'medium' | 'strong', message: string }
```

**表单验证规则**:
- 当前密码: 必填
- 新密码: 必填 + 最小8位 + 不能为弱密码
- 确认密码: 必填 + 必须与新密码一致

---

### 2. 双因素认证(2FA)设置

**路径**: 账户设置 → 安全设置 → 双因素认证

**功能特性**:
- ✅ 查看2FA状态(已启用/未启用)
- ✅ 启用2FA
  - 生成二维码
  - 显示手动输入密钥
  - 验证6位动态验证码
- ✅ 禁用2FA(带安全确认)
- ✅ 支持Google Authenticator等2FA应用

**启用流程**:
1. 点击"启用双因素认证"
2. 扫描二维码或手动输入密钥到2FA应用
3. 输入应用生成的6位验证码
4. 验证成功,启用2FA

**禁用流程**:
1. 点击"禁用双因素认证"
2. 弹出确认对话框,提示安全风险
3. 确认后禁用2FA

**二维码生成**:
```typescript
// 使用 antd 的 QRCode 组件
<QRCode
  value="otpauth://totp/MerchantPortal:user@example.com?secret=XXX&issuer=MerchantPortal"
  size={200}
/>
```

---

### 3. 活动记录查看

**路径**: 账户设置 → 活动记录

**功能特性**:
- ✅ 显示用户所有操作记录
- ✅ 记录信息包括:
  - 操作类型(登录、修改密码、修改设置等)
  - IP地址
  - 地理位置
  - 设备/浏览器信息
  - 操作时间
  - 状态(成功/失败)
- ✅ 表格形式展示,支持分页
- ✅ 失败操作用红色标签标注

**数据结构**:
```typescript
interface ActivityLog {
  id: string
  action: string           // 操作类型
  ip_address: string       // IP地址
  user_agent: string       // 浏览器/设备信息
  created_at: string       // 时间
  location?: string        // 地理位置(可选)
  status: 'success' | 'failed'
}
```

**示例数据**:
| 操作 | IP地址 | 位置 | 设备 | 时间 |
|------|--------|------|------|------|
| 登录 | 192.168.1.1 | 中国 上海 | Chrome 120.0 | 2024-01-20 14:30:00 |
| 修改密码 | 192.168.1.1 | 中国 上海 | Chrome 120.0 | 2024-01-19 10:15:00 |
| 登录失败 🔴 | 203.0.113.0 | 美国 纽约 | Firefox 121.0 | 2024-01-18 08:20:00 |

**安全监控**:
- 可快速发现异常登录(不同地区/IP)
- 识别暴力破解尝试(多次登录失败)
- 追踪账户变更历史

---

### 4. 偏好设置

**路径**: 账户设置 → 偏好设置

#### 4.1 区域设置

**语言 (Language)**:
- 简体中文
- English
- 繁體中文
- 日本語
- 한국어

**时区 (Timezone)**:
- Asia/Shanghai (UTC+8)
- Asia/Tokyo (UTC+9)
- Asia/Seoul (UTC+9)
- America/New_York (UTC-5)
- America/Los_Angeles (UTC-8)
- Europe/London (UTC+0)
- UTC (UTC+0)

**默认货币 (Default Currency)**:
- USD - 美元
- CNY - 人民币
- EUR - 欧元
- GBP - 英镑
- JPY - 日元
- KRW - 韩元
- HKD - 港币

#### 4.2 格式设置

**日期格式**:
- `YYYY-MM-DD` (2024-01-15)
- `MM/DD/YYYY` (01/15/2024)
- `DD/MM/YYYY` (15/01/2024)
- `YYYY年MM月DD日` (2024年01月15日)

**时间格式**:
- 24小时制 (14:30)
- 12小时制 (2:30 PM)

#### 4.3 通知设置

- ✅ 邮件通知 (Email Notifications)
- ✅ 短信通知 (SMS Notifications)
- ✅ 推送通知 (Push Notifications)

每个通知类型都可以独立开启/关闭(Switch组件)

**偏好设置数据结构**:
```typescript
interface MerchantPreferences {
  language: string
  timezone: string
  currency: string
  date_format: string
  time_format: string
  notifications_email: boolean
  notifications_sms: boolean
  notifications_push: boolean
}
```

**保存逻辑**:
- 点击"保存设置"按钮提交
- 语言切换立即生效(调用i18n.changeLanguage())
- 其他设置保存到后端
- 显示成功/失败提示

---

## 🎨 界面设计

### Tab布局

账户设置页面使用Tabs组件分为3个标签页:

```
┌─────────────────────────────────────────┐
│ 账户设置                                 │
│ 管理您的账户安全、偏好设置和活动记录       │
├─────────────────────────────────────────┤
│ [🔒 安全设置] [📋 活动记录] [⚙️ 偏好设置] │
├─────────────────────────────────────────┤
│                                         │
│  (当前Tab的内容)                         │
│                                         │
└─────────────────────────────────────────┘
```

### 安全设置Tab

**修改密码卡片**:
```
┌─────────────────────────┐
│ 修改密码                │
├─────────────────────────┤
│ 当前密码: [__________] │
│ 新密码:   [__________] │
│ ⚠️ 密码强度: 中         │
│ 确认密码: [__________] │
│ [修改密码]              │
└─────────────────────────┘
```

**2FA设置卡片**:
```
┌──────────────────────────────┐
│ 双因素认证 (2FA)            │
├──────────────────────────────┤
│ ℹ️ 双因素认证为您的账户提供  │
│   额外的安全保护            │
│                             │
│ 状态: [✅ 已启用]           │
│ [禁用双因素认证]            │
└──────────────────────────────┘
```

### 2FA设置Modal

```
┌────────────────────────────┐
│ 设置双因素认证              │
├────────────────────────────┤
│ ℹ️ 扫描二维码               │
│                            │
│    ┌───────────┐          │
│    │  QR Code  │          │
│    │           │          │
│    └───────────┘          │
│                            │
│ ─── 或手动输入密钥 ───     │
│                            │
│ JBSWY3DPEHPK3PXP [复制]   │
│                            │
│ 输入6位验证码:             │
│ [______]                   │
│                            │
│ [验证]                     │
└────────────────────────────┘
```

### 活动记录Tab

```
┌────────────────────────────────────────────────┐
│ 操作      │ IP地址       │ 位置     │ 设备      │ 时间          │
├────────────────────────────────────────────────┤
│ 登录      │ 192.168.1.1  │ 中国上海 │ Chrome   │ 2024-01-20... │
│ 修改密码  │ 192.168.1.1  │ 中国上海 │ Chrome   │ 2024-01-19... │
│ 登录失败🔴│ 203.0.113.0  │ 美国纽约 │ Firefox  │ 2024-01-18... │
└────────────────────────────────────────────────┘
[< 1 2 3 4 5 >]  共 50 条
```

### 偏好设置Tab

```
┌─────────────────────────┐
│ 🌍 区域设置             │
├─────────────────────────┤
│ 语言:   [简体中文  ▼]  │
│ 时区:   [Asia/Shanghai▼]│
│ 货币:   [USD - 美元 ▼] │
├─────────────────────────┤
│ 🕐 格式设置             │
├─────────────────────────┤
│ 日期:   [YYYY-MM-DD ▼] │
│ 时间:   [24小时制   ▼] │
├─────────────────────────┤
│ 🔔 通知设置             │
├─────────────────────────┤
│ 邮件通知:  [✓]         │
│ 短信通知:  [ ]         │
│ 推送通知:  [✓]         │
├─────────────────────────┤
│ [保存设置] [重置]       │
└─────────────────────────┘
```

---

## 🔧 技术实现

### 文件结构

```
frontend/merchant-portal/src/
├── pages/
│   └── Account.tsx          ✅ 完全重写,包含全部4项功能
├── utils/
│   └── security.ts          ✅ 密码强度验证
├── i18n/locales/
│   ├── zh-CN.json          ✅ 新增70+条翻译
│   └── en-US.json          ✅ 新增70+条翻译
```

### 核心组件

**Account.tsx** (~600行代码):
```typescript
import { Tabs, Form, Input, Button, Switch, Select, Table, QRCode } from 'antd'
import { validatePasswordStrength } from '../utils/security'

const Account = () => {
  // 3个Tab标签页
  const [activeTab, setActiveTab] = useState('security')

  // 密码修改
  const [passwordForm] = Form.useForm()
  const [passwordStrength, setPasswordStrength] = useState('')

  // 2FA设置
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false)
  const [showQRCode, setShowQRCode] = useState(false)
  const [qrCodeUrl, setQrCodeUrl] = useState('')

  // 活动记录
  const [activityLogs, setActivityLogs] = useState<ActivityLog[]>([])

  // 偏好设置
  const [preferencesForm] = Form.useForm()
  const [preferences, setPreferences] = useState<MerchantPreferences>({...})

  return (
    <Tabs activeKey={activeTab} onChange={setActiveTab}>
      <Tabs.TabPane key="security" tab="🔒 安全设置">
        {/* 修改密码 + 2FA设置 */}
      </Tabs.TabPane>
      <Tabs.TabPane key="activity" tab="📋 活动记录">
        {/* 活动记录表格 */}
      </Tabs.TabPane>
      <Tabs.TabPane key="preferences" tab="⚙️ 偏好设置">
        {/* 语言/时区/货币/格式/通知设置 */}
      </Tabs.TabPane>
    </Tabs>
  )
}
```

### 密码强度验证

使用 `utils/security.ts` 中的 `validatePasswordStrength()`:

```typescript
const result = validatePasswordStrength(password)
// result: { strength: 'weak' | 'medium' | 'strong', message: string }

// 评分规则:
// - 长度 < 8: weak
// - 包含1种字符类型: weak
// - 包含2-3种字符类型: medium
// - 包含4种字符类型(大小写字母+数字+特殊字符): strong
```

### 2FA实现

**二维码格式** (OTP Auth URL):
```
otpauth://totp/MerchantPortal:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=MerchantPortal
```

**组件**:
```typescript
import { QRCode } from 'antd'

<QRCode value={qrCodeUrl} size={200} />
```

**验证流程**:
1. 后端生成secret密钥
2. 前端生成QR码URL
3. 用户扫描QR码
4. 用户输入6位验证码
5. 后端验证code是否正确
6. 验证成功,启用2FA

### 国际化支持

**新增翻译键** (70+条):

**common部分**:
- `view`, `download`, `upload`, `filter`, `all`
- `today`, `yesterday`, `thisWeek`, `thisMonth`
- `yes`, `no`, `ok`, `close`, `previous`, `next`, `finish`

**account部分**:
- 密码相关: `oldPassword`, `newPassword`, `confirmPassword`, `passwordStrength*`
- 2FA相关: `twoFactorAuth`, `enable2FA`, `disable2FA`, `scanQRCode`, `verify`
- 活动记录: `activityLog`, `action`, `ipAddress`, `location`, `device`
- 偏好设置: `preferences`, `language`, `timezone`, `currency`, `dateFormat`, `timeFormat`, `notifications*`

---

## 🔌 API接口(待实现)

### 1. 修改密码

```typescript
POST /api/v1/merchant/change-password
Request:
{
  "old_password": "oldPassword123",
  "new_password": "NewSecureP@ss456"
}

Response:
{
  "code": 0,
  "message": "密码修改成功"
}
```

### 2. 启用2FA

```typescript
// 步骤1: 生成二维码
POST /api/v1/merchant/2fa/enable
Response:
{
  "code": 0,
  "data": {
    "secret": "JBSWY3DPEHPK3PXP",
    "qr_code_url": "otpauth://totp/..."
  }
}

// 步骤2: 验证并启用
POST /api/v1/merchant/2fa/verify
Request:
{
  "code": "123456"
}

Response:
{
  "code": 0,
  "message": "双因素认证已启用"
}
```

### 3. 禁用2FA

```typescript
POST /api/v1/merchant/2fa/disable
Response:
{
  "code": 0,
  "message": "双因素认证已禁用"
}
```

### 4. 获取活动记录

```typescript
GET /api/v1/merchant/activity-logs?page=1&page_size=10
Response:
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "1",
        "action": "登录",
        "ip_address": "192.168.1.1",
        "user_agent": "Chrome 120.0",
        "location": "中国 上海",
        "status": "success",
        "created_at": "2024-01-20T14:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 50
    }
  }
}
```

### 5. 获取/更新偏好设置

```typescript
// 获取
GET /api/v1/merchant/preferences
Response:
{
  "code": 0,
  "data": {
    "language": "zh-CN",
    "timezone": "Asia/Shanghai",
    "currency": "USD",
    "date_format": "YYYY-MM-DD",
    "time_format": "24h",
    "notifications_email": true,
    "notifications_sms": false,
    "notifications_push": true
  }
}

// 更新
PUT /api/v1/merchant/preferences
Request:
{
  "language": "en-US",
  "timezone": "America/New_York",
  "currency": "EUR",
  ...
}

Response:
{
  "code": 0,
  "message": "偏好设置已保存"
}
```

---

## ✅ 检查清单

- [x] 修改密码功能
  - [x] 当前密码验证
  - [x] 新密码强度检测
  - [x] 密码确认验证
  - [x] 表单提交逻辑
  - [x] 成功/失败提示

- [x] 2FA双因素认证
  - [x] 查看2FA状态
  - [x] 生成二维码
  - [x] 手动输入密钥
  - [x] 验证6位code
  - [x] 启用/禁用逻辑
  - [x] 安全确认对话框

- [x] 活动记录
  - [x] 获取活动日志
  - [x] 表格展示
  - [x] 分页功能
  - [x] 失败标注
  - [x] 时间格式化

- [x] 偏好设置
  - [x] 语言选择(5种语言)
  - [x] 时区选择(7个常用时区)
  - [x] 货币选择(7种货币)
  - [x] 日期格式(4种)
  - [x] 时间格式(2种)
  - [x] 通知开关(3种)
  - [x] 保存逻辑
  - [x] 语言切换立即生效

- [x] 国际化
  - [x] 中文翻译(70+条)
  - [x] 英文翻译(70+条)
  - [x] common通用翻译补全

---

## 🎯 使用指南

### 修改密码

1. 进入"账户设置"页面
2. 切换到"安全设置"标签
3. 在"修改密码"卡片中填写:
   - 当前密码
   - 新密码(系统会实时检测强度)
   - 确认新密码
4. 点击"修改密码"按钮
5. 修改成功后会提示重新登录

### 设置2FA

1. 进入"账户设置" → "安全设置"
2. 在"双因素认证"卡片中,点击"启用双因素认证"
3. 在弹出的对话框中:
   - 打开Google Authenticator或其他2FA应用
   - 扫描二维码(或手动输入密钥)
   - 输入应用中显示的6位验证码
   - 点击"验证"
4. 验证成功后,2FA已启用

### 查看活动记录

1. 进入"账户设置" → "活动记录"
2. 查看所有操作历史
3. 注意检查:
   - 异常IP地址
   - 异常地理位置
   - 失败的登录尝试(红色标记)

### 设置偏好

1. 进入"账户设置" → "偏好设置"
2. 根据需要调整:
   - 区域设置(语言/时区/货币)
   - 格式设置(日期/时间)
   - 通知设置(邮件/短信/推送)
3. 点击"保存设置"
4. 语言切换会立即生效

---

## 📝 总结

已完成商户后台账户功能的全部4项需求:

1. ✅ **修改密码** - 带强度检测和验证
2. ✅ **2FA设置** - 完整的启用/禁用流程
3. ✅ **活动记录** - 详细的操作日志
4. ✅ **偏好设置** - 语言/时区/货币/格式/通知

**代码质量**:
- 完整的TypeScript类型定义
- 国际化支持(中英文)
- 表单验证完善
- 用户体验友好
- 安全性高

**待后端实现**:
- 密码修改API
- 2FA生成/验证API
- 活动日志API
- 偏好设置API

前端已完全准备就绪,可立即对接后端API! 🎉
