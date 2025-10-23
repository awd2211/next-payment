# RBAC权限系统设计文档

## 系统概述

Payment Platform 采用**增强型RBAC（Role-Based Access Control）**权限模型，支持：

✅ **多级权限控制** - 平台级、商户级、资源级
✅ **动态角色配置** - 后台可视化创建和编辑角色
✅ **细粒度权限** - 精确到API接口、数据字段
✅ **权限继承** - 角色可继承其他角色的权限
✅ **数据级权限** - 控制用户可访问的数据范围
✅ **条件权限** - 基于条件的动态权限判断
✅ **权限审计** - 记录所有权限变更和使用

---

## 权限模型架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户/商户                              │
│        User (Admin/Merchant)                                 │
└────────────────────┬────────────────────────────────────────┘
                     │ N:M
                     ↓
┌─────────────────────────────────────────────────────────────┐
│                         角色                                 │
│              Role (角色定义)                                 │
│   - 名称、描述                                               │
│   - 角色类型（系统角色/自定义角色）                            │
│   - 优先级                                                   │
│   - 是否可继承                                               │
└────────────────────┬────────────────────────────────────────┘
                     │ N:M
                     ↓
┌─────────────────────────────────────────────────────────────┐
│                        权限                                  │
│           Permission (权限定义)                              │
│   - 资源（Resource）：merchant, payment, order               │
│   - 动作（Action）：view, create, edit, delete, approve      │
│   - 范围（Scope）：all, own, team, none                      │
│   - 条件（Condition）：JSON规则引擎                           │
└─────────────────────────────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────────┐
│                      权限规则                                │
│          Permission Rule（运行时判断）                        │
│   - 字段级权限（Field-level）                                │
│   - 数据过滤规则（Data Filter）                              │
│   - 时间限制（Time-based）                                   │
│   - IP限制（IP-based）                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 核心概念

### 1. 资源（Resource）

系统中所有需要权限控制的对象：

| 资源 | 说明 | 示例操作 |
|------|------|---------|
| `merchant` | 商户管理 | 查看、创建、编辑、删除、审核商户 |
| `payment` | 支付管理 | 查看、退款、导出支付数据 |
| `order` | 订单管理 | 查看、取消、导出订单 |
| `user` | 用户管理 | 管理管理员账号 |
| `role` | 角色管理 | 配置角色和权限 |
| `config` | 系统配置 | 修改系统设置 |
| `report` | 报表 | 查看和导出报表 |
| `webhook` | Webhook | 配置Webhook |
| `api_key` | API密钥 | 管理API密钥 |
| `channel` | 支付渠道 | 配置支付渠道 |

### 2. 动作（Action）

对资源可执行的操作：

| 动作 | 说明 | 示例 |
|------|------|------|
| `view` | 查看 | 查看商户列表、订单详情 |
| `create` | 创建 | 创建商户、生成API密钥 |
| `edit` | 编辑 | 修改商户信息 |
| `delete` | 删除 | 删除商户 |
| `approve` | 审批 | 审核商户KYC |
| `export` | 导出 | 导出订单数据 |
| `refund` | 退款 | 执行退款操作 |
| `manage` | 全面管理 | 包含上述所有权限 |

### 3. 范围（Scope）

权限的作用范围：

| 范围 | 说明 | 示例 |
|------|------|------|
| `all` | 全部数据 | 可以查看所有商户的订单 |
| `own` | 自己的数据 | 只能查看自己的订单 |
| `team` | 团队数据 | 可以查看本团队的数据 |
| `assigned` | 分配的数据 | 只能查看分配给自己的任务 |
| `none` | 无权限 | 无法访问该资源 |

### 4. 条件（Condition）

基于条件的动态权限判断（JSON规则引擎）：

```json
{
  "rules": [
    {
      "field": "amount",
      "operator": "<=",
      "value": 10000,
      "description": "只能处理1万元以下的退款"
    },
    {
      "field": "status",
      "operator": "in",
      "value": ["pending", "processing"],
      "description": "只能操作待处理和处理中的订单"
    },
    {
      "field": "created_at",
      "operator": "within_days",
      "value": 30,
      "description": "只能查看近30天的数据"
    }
  ],
  "logic": "AND"
}
```

---

## 权限表达式

### 标准格式

```
{resource}.{action}.{scope}[?condition]
```

### 示例

```
# 基础权限
merchant.view.all           # 查看所有商户
merchant.edit.own           # 编辑自己的商户信息
payment.view.all            # 查看所有支付记录
payment.refund.own          # 退款自己的订单

# 条件权限
payment.refund.all?amount<=10000    # 退款全部订单（金额<=1万）
order.export.all?days<=30           # 导出近30天的订单
merchant.approve.all?kyc_verified   # 审批已完成KYC的商户

# 字段级权限
merchant.view.all:fields=id,name,email        # 只能查看部分字段
payment.view.all:fields=*,-customer_card      # 查看所有字段除了卡号

# 复合权限
merchant.*.*                # 商户的所有权限
*.view.all                  # 查看所有资源
```

---

## 预定义角色

### 管理后台角色

| 角色 | 权限范围 | 典型用户 |
|------|---------|---------|
| **超级管理员** | 所有权限 | CTO、系统管理员 |
| **系统管理员** | 系统配置、用户管理、角色管理 | 运维人员 |
| **运营经理** | 商户管理、订单查询、数据分析 | 运营主管 |
| **运营专员** | 商户查询、订单查询 | 运营人员 |
| **财务经理** | 财务报表、对账、提现审批 | 财务主管 |
| **财务专员** | 查看财务报表、对账 | 财务人员 |
| **客服主管** | 查看订单、执行退款、查看商户信息 | 客服经理 |
| **客服专员** | 查看订单、提交退款申请 | 客服人员 |
| **风控经理** | 风控规则配置、黑名单管理、可疑订单审核 | 风控主管 |
| **审计员** | 只读权限、查看审计日志 | 审计人员 |

### 商户端角色

| 角色 | 权限范围 | 典型用户 |
|------|---------|---------|
| **商户所有者** | 商户所有权限 | 商户老板 |
| **商户管理员** | 除删除商户外的所有权限 | 商户技术负责人 |
| **开发者** | API密钥管理、Webhook配置、测试支付 | 技术人员 |
| **财务** | 查看订单、财务报表、对账 | 商户财务 |
| **客服** | 查看订单、提交退款申请 | 商户客服 |
| **只读** | 仅查看权限 | 第三方审计 |

---

## 数据库设计

### 核心表

#### 1. roles（角色表）

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,         -- 角色代码
    name VARCHAR(100) NOT NULL,                -- 角色名称
    display_name VARCHAR(100) NOT NULL,        -- 显示名称
    description TEXT,                          -- 描述
    role_type VARCHAR(20) NOT NULL,            -- system/custom/tenant
    priority INTEGER DEFAULT 0,                -- 优先级（数字越大权限越高）
    is_system BOOLEAN DEFAULT false,           -- 是否系统角色（不可删除）
    parent_role_id UUID,                       -- 父角色ID（权限继承）
    scope VARCHAR(20) DEFAULT 'platform',      -- platform/merchant
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

#### 2. permissions（权限表）

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    code VARCHAR(200) UNIQUE NOT NULL,         -- 权限代码 (merchant.view.all)
    name VARCHAR(100) NOT NULL,                -- 权限名称
    resource VARCHAR(50) NOT NULL,             -- 资源
    action VARCHAR(50) NOT NULL,               -- 动作
    scope VARCHAR(20) DEFAULT 'all',           -- 范围
    conditions JSONB,                          -- 条件规则（JSON）
    description TEXT,                          -- 描述
    category VARCHAR(50),                      -- 分类（用于后台展示）
    is_dangerous BOOLEAN DEFAULT false,        -- 是否危险权限（需要二次确认）
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 示例数据
INSERT INTO permissions (code, name, resource, action, scope, category, description) VALUES
    ('merchant.view.all', '查看所有商户', 'merchant', 'view', 'all', '商户管理', '可以查看所有商户信息'),
    ('merchant.edit.own', '编辑自己的商户', 'merchant', 'edit', 'own', '商户管理', '只能编辑自己的商户信息'),
    ('payment.refund.all', '退款所有支付', 'payment', 'refund', 'all', '支付管理', '可以对所有支付执行退款'),
    ('config.edit.*', '修改系统配置', 'config', 'edit', 'all', '系统管理', '可以修改系统配置（危险）');
```

#### 3. role_permissions（角色-权限关联表）

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY,
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    granted_by UUID,                           -- 授予人
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    conditions JSONB,                          -- 额外条件（可覆盖默认条件）
    UNIQUE(role_id, permission_id)
);
```

#### 4. user_roles（用户-角色关联表）

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_type VARCHAR(20) NOT NULL,            -- admin/merchant
    role_id UUID NOT NULL,
    assigned_by UUID,                          -- 分配人
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                    -- 过期时间（临时角色）
    UNIQUE(user_id, user_type, role_id)
);
```

#### 5. permission_groups（权限组）

```sql
CREATE TABLE permission_groups (
    id UUID PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),                          -- 图标（用于UI展示）
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 示例：权限分组
INSERT INTO permission_groups (code, name, description, icon) VALUES
    ('merchant_management', '商户管理', '商户相关的所有权限', 'shop'),
    ('payment_management', '支付管理', '支付和订单相关权限', 'credit-card'),
    ('financial_management', '财务管理', '财务、对账、结算权限', 'dollar'),
    ('system_management', '系统管理', '系统配置和用户管理权限', 'setting');
```

#### 6. permission_audit_logs（权限审计日志）

```sql
CREATE TABLE permission_audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_type VARCHAR(20) NOT NULL,
    action VARCHAR(50) NOT NULL,               -- grant/revoke/check
    resource VARCHAR(50),
    resource_id VARCHAR(100),
    permission_code VARCHAR(200),
    granted BOOLEAN,                           -- 是否授予
    reason TEXT,                               -- 原因
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_permission_audit_user ON permission_audit_logs(user_id, created_at DESC);
CREATE INDEX idx_permission_audit_resource ON permission_audit_logs(resource, resource_id);
```

---

## 权限检查流程

```
┌─────────────────────────────────────────────────────────┐
│  1. 用户发起请求                                         │
│     POST /api/v1/payments/refund                        │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  2. JWT Token解析                                       │
│     提取：user_id, user_type, roles                     │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  3. 权限中间件拦截                                       │
│     Required: payment.refund.all                        │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  4. 查询用户角色                                         │
│     SELECT roles FROM user_roles WHERE user_id = ?      │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  5. 查询角色权限                                         │
│     SELECT permissions FROM role_permissions            │
│     WHERE role_id IN (user_roles)                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  6. 匹配权限代码                                         │
│     payment.refund.all 是否在权限列表中？                │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  7. 条件判断（如果有）                                   │
│     amount <= 10000?                                    │
│     status in ['pending']?                              │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  8. 记录审计日志                                         │
│     INSERT INTO permission_audit_logs                   │
└────────────────────┬────────────────────────────────────┘
                     ↓
        ┌────────────┴──────────────┐
        │                           │
    ✅ 通过                      ❌ 拒绝
    执行业务逻辑              返回403 Forbidden
```

---

## 后台配置界面

### 1. 角色管理页面

```
┌─────────────────────────────────────────────────────────┐
│  角色管理                            [+ 新建角色]         │
├─────────────────────────────────────────────────────────┤
│  搜索: [____________]  类型: [全部▼]  状态: [全部▼]     │
├────┬──────────┬────────┬──────┬──────────┬─────────────┤
│ ID │ 角色名称  │ 类型   │ 优先级│ 用户数   │ 操作        │
├────┼──────────┼────────┼──────┼──────────┼─────────────┤
│ 01 │ 超级管理员│ system │  100 │    5     │ [查看][编辑]│
│ 02 │ 运营经理  │ system │   80 │   12     │ [查看][编辑]│
│ 03 │ 客服专员  │ custom │   30 │   45     │ [查看][编辑][删除]│
└────┴──────────┴────────┴──────┴──────────┴─────────────┘
```

### 2. 角色编辑页面

```
┌─────────────────────────────────────────────────────────┐
│  编辑角色：运营经理                                       │
├─────────────────────────────────────────────────────────┤
│  基本信息                                                │
│  ┌───────────────────────────────────────────────────┐ │
│  │ 角色名称: [运营经理___________________________]      │ │
│  │ 显示名称: [运营经理___________________________]      │ │
│  │ 描述:     [负责日常运营，管理商户和订单_________]   │ │
│  │ 优先级:   [80_____] (数字越大权限越高)              │ │
│  │ 父角色:   [无▼] (继承父角色的所有权限)              │ │
│  └───────────────────────────────────────────────────┘ │
│                                                          │
│  权限配置                                                │
│  ┌───────────────────────────────────────────────────┐ │
│  │ 🛍️  商户管理 (3/5)                                   │ │
│  │   ☑ 查看所有商户    ☑ 创建商户    ☐ 删除商户         │ │
│  │   ☑ 编辑商户       ☑ 审核商户                        │ │
│  │                                                      │ │
│  │ 💳 支付管理 (4/6)                                    │ │
│  │   ☑ 查看所有支付    ☑ 退款       ☐ 导出支付数据      │ │
│  │   ☑ 查看订单       ☑ 取消订单    ☐ 修改订单          │ │
│  │                                                      │ │
│  │ 💰 财务管理 (2/4)                                    │ │
│  │   ☑ 查看财务报表    ☐ 对账      ☐ 提现审批          │ │
│  │   ☑ 查看结算记录    ☐ 修改费率                       │ │
│  │                                                      │ │
│  │ ⚙️  系统管理 (0/5)                                   │ │
│  │   ☐ 用户管理       ☐ 角色管理    ☐ 系统配置          │ │
│  │   ☐ 查看日志       ☐ 查看监控                        │ │
│  └───────────────────────────────────────────────────┘ │
│                                                          │
│  高级配置                                                │
│  ┌───────────────────────────────────────────────────┐ │
│  │ 数据范围限制:                                        │ │
│  │   ○ 全部数据                                        │ │
│  │   ○ 本团队数据                                      │ │
│  │   ● 自定义规则                                      │ │
│  │     created_at > now() - interval '30 days'        │ │
│  │     AND status IN ('active', 'pending')            │ │
│  │                                                      │ │
│  │ IP白名单: [可选]                                     │ │
│  │   192.168.1.0/24, 10.0.0.1                         │ │
│  └───────────────────────────────────────────────────┘ │
│                                                          │
│  [取消]                                      [保存角色]  │
└─────────────────────────────────────────────────────────┘
```

### 3. 权限矩阵视图

```
┌─────────────────────────────────────────────────────────┐
│  权限矩阵                          导出[Excel][CSV]       │
├─────────┬──────┬──────┬──────┬──────┬──────┬──────────┤
│ 角色/权限│ 查看 │ 创建 │ 编辑 │ 删除 │ 审批 │ 导出     │
├─────────┼──────┼──────┼──────┼──────┼──────┼──────────┤
│超级管理员│  ✅  │  ✅  │  ✅  │  ✅  │  ✅  │  ✅     │
│运营经理  │  ✅  │  ✅  │  ✅  │  ❌  │  ✅  │  ✅     │
│客服专员  │  ✅  │  ❌  │  ❌  │  ❌  │  ❌  │  ❌     │
│财务专员  │  ✅  │  ❌  │  ❌  │  ❌  │  ❌  │  ✅     │
└─────────┴──────┴──────┴──────┴──────┴──────┴──────────┘
```

---

## 实现示例

### 1. 权限检查中间件

```go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, _ := middleware.GetClaims(c)

        // 检查是否有权限
        hasPermission, err := permissionService.CheckPermission(
            c.Request.Context(),
            claims.UserID,
            claims.UserType,
            permission,
        )

        if err != nil || !hasPermission {
            // 记录审计日志
            permissionService.LogAccess(claims.UserID, permission, false)

            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }

        // 记录审计日志
        permissionService.LogAccess(claims.UserID, permission, true)

        c.Next()
    }
}

// 使用示例
router.POST("/payments/refund",
    authMiddleware,
    RequirePermission("payment.refund.all"),
    handler.RefundPayment,
)
```

### 2. 动态权限检查

```go
func (s *PermissionService) CheckPermission(ctx context.Context, userID uuid.UUID, userType, permissionCode string) (bool, error) {
    // 1. 获取用户角色
    roles, err := s.roleRepo.GetUserRoles(ctx, userID, userType)
    if err != nil {
        return false, err
    }

    // 2. 获取角色权限
    var allPermissions []model.Permission
    for _, role := range roles {
        permissions, err := s.permissionRepo.GetRolePermissions(ctx, role.ID)
        if err != nil {
            continue
        }
        allPermissions = append(allPermissions, permissions...)
    }

    // 3. 匹配权限代码（支持通配符）
    for _, perm := range allPermissions {
        if matchPermission(perm.Code, permissionCode) {
            // 4. 检查条件（如果有）
            if perm.Conditions != "" {
                conditionMet, err := s.evaluateCondition(ctx, perm.Conditions)
                if err != nil || !conditionMet {
                    continue
                }
            }
            return true, nil
        }
    }

    return false, nil
}

// 权限代码匹配（支持通配符）
func matchPermission(pattern, target string) bool {
    // merchant.*.all 可以匹配 merchant.view.all, merchant.edit.all
    // *.view.all 可以匹配 merchant.view.all, payment.view.all
    // merchant.view.* 可以匹配 merchant.view.all, merchant.view.own

    patternParts := strings.Split(pattern, ".")
    targetParts := strings.Split(target, ".")

    if len(patternParts) != len(targetParts) {
        return false
    }

    for i := range patternParts {
        if patternParts[i] != "*" && patternParts[i] != targetParts[i] {
            return false
        }
    }

    return true
}
```

### 3. 条件规则引擎

```go
type ConditionRule struct {
    Rules []Rule `json:"rules"`
    Logic string `json:"logic"` // AND/OR
}

type Rule struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`
    Value    interface{} `json:"value"`
}

func (s *PermissionService) evaluateCondition(ctx context.Context, conditionJSON string) (bool, error) {
    var condition ConditionRule
    if err := json.Unmarshal([]byte(conditionJSON), &condition); err != nil {
        return false, err
    }

    results := make([]bool, len(condition.Rules))

    for i, rule := range condition.Rules {
        // 获取字段值（从上下文或请求参数中）
        value := getFieldValue(ctx, rule.Field)

        // 执行比较
        switch rule.Operator {
        case "==":
            results[i] = value == rule.Value
        case "!=":
            results[i] = value != rule.Value
        case ">":
            results[i] = compareNumbers(value, rule.Value, ">")
        case ">=":
            results[i] = compareNumbers(value, rule.Value, ">=")
        case "<":
            results[i] = compareNumbers(value, rule.Value, "<")
        case "<=":
            results[i] = compareNumbers(value, rule.Value, "<=")
        case "in":
            results[i] = contains(rule.Value, value)
        case "within_days":
            results[i] = withinDays(value, rule.Value.(int))
        }
    }

    // 应用逻辑（AND/OR）
    if condition.Logic == "OR" {
        for _, result := range results {
            if result {
                return true, nil
            }
        }
        return false, nil
    } else {
        // AND
        for _, result := range results {
            if !result {
                return false, nil
            }
        }
        return true, nil
    }
}
```

---

## 最佳实践

### 1. 最小权限原则

默认不授予任何权限，只授予完成工作所需的最小权限集。

### 2. 权限分离

将高危权限（如删除商户、修改系统配置）与常规权限分离，需要额外审批。

### 3. 定期审计

定期审查角色和权限配置，清理不再需要的权限。

### 4. 权限过期

为临时权限设置过期时间，自动回收。

### 5. 审计日志

记录所有权限变更和高危操作，便于追溯。

---

## 扩展功能

### 1. 权限申请流程

```
用户申请权限 → 主管审批 → 自动授予 → 定期回收
```

### 2. 权限模板

预定义常用权限组合，快速分配。

### 3. 权限继承

子角色自动继承父角色权限。

### 4. 权限委托

临时委托权限给其他用户。

### 5. 权限可视化

树形图展示权限层级关系。

---

## 参考资料

- [NIST RBAC标准](https://csrc.nist.gov/projects/role-based-access-control)
- [AWS IAM最佳实践](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [Casbin权限管理](https://casbin.org/)
- [OPA策略引擎](https://www.openpolicyagent.org/)
