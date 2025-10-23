# 安全功能设计文档

## 概述

本文档详细说明支付平台的安全功能实现，包括密码管理、双因素认证（2FA）、登录活动追踪、会话管理和用户偏好设置。

## 功能特性

### 1. 密码管理

#### 1.1 密码强度要求

- 最少8个字符，最多128个字符
- 必须包含至少1个大写字母
- 必须包含至少1个小写字母
- 必须包含至少1个数字
- 必须包含至少1个特殊字符（!@#$%^&*()_+-=[]{}|;:,.<>?）

#### 1.2 密码历史

- 记录最近5次使用的密码
- 防止用户重复使用近期密码
- 使用bcrypt加密存储

#### 1.3 密码修改

**API端点：** `POST /api/v1/security/change-password`

**请求示例：**
```json
{
  "old_password": "OldPassword@123",
  "new_password": "NewPassword@456"
}
```

**响应示例：**
```json
{
  "message": "密码修改成功"
}
```

#### 1.4 密码过期策略

- 可配置密码过期天数（默认90天）
- 密码过期后要求用户强制修改
- 记录最后修改密码时间

---

### 2. 双因素认证（2FA）

#### 2.1 TOTP实现

- 使用基于时间的一次性密码算法（TOTP）
- 兼容Google Authenticator、Microsoft Authenticator等应用
- 6位数字验证码
- 30秒时间窗口，允许±1窗口偏差（容错60秒）

#### 2.2 设置2FA流程

**步骤1：生成密钥**

`POST /api/v1/security/2fa/setup`

```json
{
  "account_name": "user@example.com"
}
```

**响应：**
```json
{
  "message": "2FA设置成功，请使用Google Authenticator或其他TOTP应用扫描二维码",
  "data": {
    "secret": "JBSWY3DPEHPK3PXP",
    "qr_code_url": "otpauth://totp/Payment Platform:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Payment%20Platform",
    "backup_codes": [
      "ABCD-1234",
      "EFGH-5678",
      "IJKL-9012",
      "MNOP-3456",
      "QRST-7890",
      "UVWX-1234",
      "YZAB-5678",
      "CDEF-9012"
    ]
  }
}
```

**步骤2：验证并启用**

`POST /api/v1/security/2fa/verify`

```json
{
  "code": "123456"
}
```

#### 2.3 备用恢复代码

- 生成8个备用恢复代码
- 每个代码为8字符Base32编码，格式：xxxx-xxxx
- 使用后自动失效
- 可重新生成备用代码

**API：** `POST /api/v1/security/2fa/backup-codes`

#### 2.4 禁用2FA

**API：** `POST /api/v1/security/2fa/disable`

```json
{
  "password": "YourPassword@123"
}
```

需要提供密码验证以确保安全。

---

### 3. 登录活动追踪

#### 3.1 记录信息

每次登录尝试都会记录以下信息：

- **用户信息**：用户ID、用户类型（admin/merchant）
- **登录类型**：密码登录、2FA验证、API密钥
- **状态**：成功、失败、被阻止
- **网络信息**：IP地址、User-Agent
- **设备信息**：设备类型（桌面/手机/平板）、浏览器、操作系统
- **地理位置**：国家、城市（可选）
- **时间戳**：登录时间、登出时间
- **会话ID**：关联的会话标识

#### 3.2 异常登录检测

系统会自动检测以下异常情况：

| 异常类型 | 说明 | 检测方式 |
|---------|------|---------|
| 新设备 | 从未使用过的设备 | User-Agent匹配 |
| 新IP | 从未使用过的IP地址 | IP地址匹配 |
| 新位置 | 从未登录过的城市 | GeoIP查询 |
| 新国家 | 从未登录过的国家 | GeoIP查询 |
| VPN/代理 | 使用VPN或代理服务器 | IP信誉检查 |
| 高频登录 | 短时间内多次登录 | 频率限制 |
| 时区异常 | 登录时区与常用时区差异大 | 时区分析 |

#### 3.3 查看登录活动

**API：** `GET /api/v1/security/login-activities?limit=50`

**响应示例：**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "user_type": "admin",
      "login_type": "password",
      "status": "success",
      "ip": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "device_type": "desktop",
      "browser": "Chrome",
      "os": "Windows",
      "country": "United States",
      "city": "New York",
      "is_abnormal": false,
      "abnormal_reason": null,
      "session_id": "session_123",
      "login_at": "2024-01-15T10:30:00Z",
      "logout_at": null
    }
  ]
}
```

#### 3.4 查看异常活动

**API：** `GET /api/v1/security/abnormal-activities`

返回所有被标记为异常的登录记录。

---

### 4. 安全设置

#### 4.1 可配置选项

| 设置项 | 说明 | 默认值 |
|-------|------|--------|
| password_expiry_days | 密码过期天数（0=永不过期） | 90 |
| session_timeout_minutes | 会话超时分钟数 | 60 |
| max_concurrent_sessions | 最大并发会话数 | 5 |
| ip_whitelist | IP白名单（JSON数组） | [] |
| allowed_countries | 允许的国家列表 | [] |
| blocked_countries | 禁止的国家列表 | [] |
| login_notification | 新设备登录通知 | true |
| abnormal_notification | 异常活动通知 | true |

#### 4.2 获取安全设置

**API：** `GET /api/v1/security/settings`

#### 4.3 更新安全设置

**API：** `PUT /api/v1/security/settings`

```json
{
  "password_expiry_days": 60,
  "session_timeout_minutes": 30,
  "max_concurrent_sessions": 3,
  "ip_whitelist": ["192.168.1.0/24", "10.0.0.0/8"],
  "allowed_countries": ["US", "CN", "JP"],
  "blocked_countries": ["XX"],
  "login_notification": true,
  "abnormal_notification": true
}
```

---

### 5. 会话管理

#### 5.1 会话创建

- 登录成功后自动创建会话
- 会话ID使用32字节随机数生成（Base32编码）
- 记录IP、User-Agent、过期时间

#### 5.2 会话追踪

- 记录最后活跃时间（last_seen_at）
- 自动延长活跃会话的过期时间
- 会话超时后自动失效

#### 5.3 查看活跃会话

**API：** `GET /api/v1/security/sessions`

**响应示例：**
```json
{
  "data": [
    {
      "id": "uuid",
      "session_id": "session_123",
      "user_id": "uuid",
      "ip": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "expires_at": "2024-01-15T12:00:00Z",
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z",
      "last_seen_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

#### 5.4 停用会话

**停用指定会话：** `POST /api/v1/security/sessions/deactivate`

```json
{
  "session_id": "session_123"
}
```

**停用其他所有会话：** `POST /api/v1/security/sessions/deactivate-others`

保留当前会话，停用所有其他活跃会话。

#### 5.5 会话清理

- 定期清理过期会话（建议每小时运行一次）
- 清理已停用的会话
- 可通过后台任务自动执行

---

### 6. 用户偏好设置

#### 6.1 支持的偏好选项

**语言（Language）**
- 支持12种语言：en, zh-CN, zh-TW, ja, ko, es, fr, de, pt, ru, ar, hi

**货币（Currency）**
- 支持20种货币：USD, EUR, GBP, CNY, JPY, KRW, HKD, SGD, AUD, CAD, INR, BRL, MXN, RUB, TRY, ZAR, CHF, SEK, NOK, DKK

**时区（Timezone）**
- 支持常用时区：UTC, Asia/Shanghai, America/New_York, Europe/London等

**日期格式（Date Format）**
- YYYY-MM-DD（2024-01-15）
- DD/MM/YYYY（15/01/2024）
- MM/DD/YYYY（01/15/2024）
- DD-Mon-YYYY（15-Jan-2024）

**时间格式（Time Format）**
- 12小时制（12h）- 带AM/PM
- 24小时制（24h）

**数字格式（Number Format）**
- 1,234.56（英语）
- 1.234,56（欧洲）
- 1 234,56（法语）
- 1'234.56（瑞士）

**主题（Theme）**
- light - 浅色主题
- dark - 深色主题
- auto - 自动（跟随系统）

#### 6.2 获取偏好设置

**API：** `GET /api/v1/preferences`

**响应示例：**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "user_type": "admin",
    "language": "zh-CN",
    "currency": "USD",
    "timezone": "Asia/Shanghai",
    "date_format": "YYYY-MM-DD",
    "time_format": "24h",
    "number_format": "1,234.56",
    "theme": "light",
    "dashboard_layout": {},
    "notification_prefs": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

#### 6.3 更新偏好设置

**API：** `PUT /api/v1/preferences`

```json
{
  "language": "zh-CN",
  "currency": "CNY",
  "timezone": "Asia/Shanghai",
  "date_format": "YYYY-MM-DD",
  "time_format": "24h",
  "number_format": "1,234.56",
  "theme": "dark"
}
```

---

## 数据库架构

### 表结构

#### two_factor_auth（双因素认证表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户ID |
| user_type | VARCHAR(20) | 用户类型（admin/merchant） |
| secret | VARCHAR(256) | TOTP密钥（加密存储） |
| is_enabled | BOOLEAN | 是否启用 |
| is_verified | BOOLEAN | 是否已验证 |
| backup_codes | JSONB | 备用恢复代码（加密存储） |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |
| verified_at | TIMESTAMPTZ | 验证时间 |

**唯一索引：** (user_id, user_type)

#### login_activities（登录活动记录表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户ID |
| user_type | VARCHAR(20) | 用户类型 |
| login_type | VARCHAR(20) | 登录类型 |
| status | VARCHAR(20) | 状态（success/failed/blocked） |
| ip | VARCHAR(50) | IP地址 |
| user_agent | TEXT | User-Agent |
| device_type | VARCHAR(20) | 设备类型 |
| browser | VARCHAR(50) | 浏览器 |
| os | VARCHAR(50) | 操作系统 |
| country | VARCHAR(50) | 国家 |
| city | VARCHAR(100) | 城市 |
| location | VARCHAR(200) | 完整位置 |
| is_abnormal | BOOLEAN | 是否异常 |
| abnormal_reason | TEXT | 异常原因 |
| failed_reason | VARCHAR(200) | 失败原因 |
| session_id | VARCHAR(128) | 会话ID |
| login_at | TIMESTAMPTZ | 登录时间 |
| logout_at | TIMESTAMPTZ | 登出时间 |

**索引：**
- (user_id, user_type)
- login_at DESC
- status
- is_abnormal
- ip
- session_id

#### security_settings（安全设置表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户ID |
| user_type | VARCHAR(20) | 用户类型 |
| password_changed_at | TIMESTAMPTZ | 密码修改时间 |
| require_password_change | BOOLEAN | 是否要求修改密码 |
| password_expiry_days | INTEGER | 密码过期天数 |
| session_timeout_minutes | INTEGER | 会话超时分钟 |
| max_concurrent_sessions | INTEGER | 最大并发会话数 |
| ip_whitelist | JSONB | IP白名单 |
| allowed_countries | JSONB | 允许的国家 |
| blocked_countries | JSONB | 禁止的国家 |
| login_notification | BOOLEAN | 登录通知 |
| abnormal_notification | BOOLEAN | 异常通知 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

**唯一索引：** (user_id, user_type)

#### password_history（密码历史表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户ID |
| user_type | VARCHAR(20) | 用户类型 |
| password_hash | VARCHAR(255) | 密码哈希 |
| created_at | TIMESTAMPTZ | 创建时间 |

**索引：**
- (user_id, user_type)
- created_at DESC

#### sessions（会话表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| session_id | VARCHAR(128) | 会话ID（唯一） |
| user_id | UUID | 用户ID |
| user_type | VARCHAR(20) | 用户类型 |
| ip | VARCHAR(50) | IP地址 |
| user_agent | TEXT | User-Agent |
| data | JSONB | 会话数据 |
| expires_at | TIMESTAMPTZ | 过期时间 |
| is_active | BOOLEAN | 是否活跃 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |
| last_seen_at | TIMESTAMPTZ | 最后活跃时间 |

**索引：**
- session_id（唯一）
- (user_id, user_type)
- expires_at
- is_active

#### user_preferences（用户偏好设置表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户ID |
| user_type | VARCHAR(20) | 用户类型 |
| language | VARCHAR(10) | 语言 |
| currency | VARCHAR(10) | 货币 |
| timezone | VARCHAR(50) | 时区 |
| date_format | VARCHAR(20) | 日期格式 |
| time_format | VARCHAR(20) | 时间格式 |
| number_format | VARCHAR(20) | 数字格式 |
| theme | VARCHAR(20) | 主题 |
| dashboard_layout | JSONB | 仪表板布局 |
| notification_prefs | JSONB | 通知偏好 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

**唯一索引：** (user_id, user_type)

---

## 安全最佳实践

### 1. 密码安全

- ✅ 使用bcrypt加密（cost=12）
- ✅ 强制密码复杂度要求
- ✅ 防止密码重用（最近5次）
- ✅ 定期密码过期提醒
- ✅ 密码重置后强制修改

### 2. 2FA安全

- ✅ TOTP标准实现（RFC 6238）
- ✅ 备用恢复代码
- ✅ 禁用2FA需要密码验证
- ✅ 使用后失效的备用代码

### 3. 会话安全

- ✅ 随机会话ID生成
- ✅ 会话超时自动失效
- ✅ 限制并发会话数
- ✅ 支持远程注销会话
- ✅ 记录会话活动

### 4. 登录安全

- ✅ 记录所有登录尝试
- ✅ 异常登录检测
- ✅ 失败次数限制（建议配合rate limiting）
- ✅ IP白名单/黑名单
- ✅ 地理位置限制

### 5. 通知安全

- ✅ 新设备登录通知
- ✅ 异常活动通知
- ✅ 密码修改通知
- ✅ 2FA启用/禁用通知

---

## API完整列表

### 密码管理
- `POST /api/v1/security/change-password` - 修改密码

### 2FA管理
- `POST /api/v1/security/2fa/setup` - 设置2FA
- `POST /api/v1/security/2fa/verify` - 验证并启用2FA
- `POST /api/v1/security/2fa/disable` - 禁用2FA
- `POST /api/v1/security/2fa/backup-codes` - 重新生成备用代码

### 登录活动
- `GET /api/v1/security/login-activities` - 获取登录活动记录
- `GET /api/v1/security/abnormal-activities` - 获取异常登录活动

### 安全设置
- `GET /api/v1/security/settings` - 获取安全设置
- `PUT /api/v1/security/settings` - 更新安全设置

### 会话管理
- `GET /api/v1/security/sessions` - 获取活跃会话
- `POST /api/v1/security/sessions/deactivate` - 停用指定会话
- `POST /api/v1/security/sessions/deactivate-others` - 停用其他所有会话

### 偏好设置
- `GET /api/v1/preferences` - 获取用户偏好设置
- `PUT /api/v1/preferences` - 更新用户偏好设置

---

## 未来改进

### 短期（1-3个月）
- [ ] 生物识别支持（WebAuthn/FIDO2）
- [ ] 基于风险的自适应认证
- [ ] IP信誉检查集成
- [ ] GeoIP精确定位

### 中期（3-6个月）
- [ ] 行为分析异常检测
- [ ] 设备指纹识别
- [ ] 登录模式学习
- [ ] 智能验证码

### 长期（6-12个月）
- [ ] 零信任架构
- [ ] 持续认证
- [ ] AI驱动的威胁检测
- [ ] 生物行为识别

---

## 合规性

本安全实现符合以下标准：

- ✅ **OWASP** - 遵循OWASP Top 10安全最佳实践
- ✅ **PCI DSS** - 满足支付卡行业数据安全标准
- ✅ **GDPR** - 支持用户数据隐私保护
- ✅ **ISO 27001** - 信息安全管理体系

---

## 技术栈

- **加密库**：golang.org/x/crypto
- **JWT**：github.com/golang-jwt/jwt/v5
- **TOTP**：github.com/pquerna/otp
- **User-Agent解析**：github.com/ua-parser/uap-go
- **数据库**：PostgreSQL 15+
- **缓存**：Redis 7+
