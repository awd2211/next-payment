# Merchant Policy Service

## 概述

商户策略服务 - 统一管理商户的所有策略配置,包括等级、费率、限额和渠道策略。

## 职责

本服务专注于**策略配置管理**(静态规则定义),不涉及配额消耗追踪(由merchant-quota-service负责)。

### 核心功能

1. **商户等级管理** (Tier Management)
   - 定义商户等级体系 (Starter → Basic → Professional → Enterprise → Premium)
   - 管理等级权益和升级条件
   - 关联默认策略

2. **费率策略管理** (Fee Policy)
   - 按渠道/支付方式配置费率
   - 支持百分比、固定金额、阶梯费率
   - 支持等级默认策略和商户自定义覆盖

3. **限额策略管理** (Limit Policy)
   - 配置单笔/日/月限额
   - 按渠道和币种差异化配置
   - 支持等级默认策略和商户自定义覆盖

4. **渠道策略管理** (Channel Policy)
   - 商户启用的支付渠道
   - 渠道优先级配置
   - 渠道特定参数

5. **策略生效引擎** (Policy Engine)
   - 计算当前生效策略
   - 优先级: 商户自定义 > 等级默认
   - 支持生效时间和过期时间

## 与其他服务的关系

### 与 merchant-quota-service 的区别

| 维度 | merchant-policy-service | merchant-quota-service |
|------|------------------------|------------------------|
| **职责** | 策略配置(静态规则) | 配额追踪(动态消耗) |
| **数据特征** | 低频变更 | 高频读写 |
| **核心模型** | FeePolicy, LimitPolicy | MerchantQuota, UsageLog |
| **调用场景** | 策略变更、查询规则 | 每笔交易消耗配额 |
| **示例API** | GET /limit-policies/effective | POST /quotas/consume |

### 服务交互

```
payment-gateway
    ↓
1. 调用 policy-service: 获取限额策略
   GET /api/v1/limit-policies/effective?merchant_id=xxx
   返回: { daily_limit: 5000000, single_max: 1000000 }
    ↓
2. 调用 quota-service: 检查配额
   POST /api/v1/quotas/check
    ↓
3. 调用 policy-service: 计算费用
   POST /api/v1/fee-policies/calculate
   返回: { total_fee: 1480 }
```

## 数据模型

### 1. MerchantTier (商户等级)
- 定义等级体系 (starter, basic, professional, enterprise, premium)
- 关联默认费率和限额策略
- 升级条件配置

### 2. MerchantPolicyBinding (商户策略绑定)
- 记录商户当前等级
- 可选的自定义策略覆盖

### 3. MerchantFeePolicy (费率策略)
- 按渠道/支付方式定价
- 支持百分比、固定、阶梯费率
- 支持等级默认和商户自定义

### 4. MerchantLimitPolicy (限额策略)
- 单笔/日/月/年限额
- 按渠道和币种差异化
- 支持等级默认和商户自定义

### 5. ChannelPolicy (渠道策略)
- 商户启用的渠道
- 渠道优先级
- 渠道特定配置

## API设计

### 等级管理

```
GET    /api/v1/tiers              # 获取等级列表
POST   /api/v1/tiers              # 创建等级
GET    /api/v1/tiers/:id          # 获取等级详情
PUT    /api/v1/tiers/:id          # 更新等级
DELETE /api/v1/tiers/:id          # 删除等级
```

### 费率策略

```
POST   /api/v1/fee-policies                    # 创建费率策略
GET    /api/v1/fee-policies/:id                # 获取策略详情
GET    /api/v1/fee-policies/merchant/:id       # 获取商户费率策略
PUT    /api/v1/fee-policies/:id                # 更新策略
DELETE /api/v1/fee-policies/:id                # 删除策略
POST   /api/v1/fee-policies/:id/approve        # 审批策略
POST   /api/v1/fee-policies/calculate          # 计算费用
GET    /api/v1/fee-policies/effective          # 获取生效策略
```

### 限额策略

```
POST   /api/v1/limit-policies                  # 创建限额策略
GET    /api/v1/limit-policies/:id              # 获取策略详情
GET    /api/v1/limit-policies/merchant/:id     # 获取商户限额策略
PUT    /api/v1/limit-policies/:id              # 更新策略
DELETE /api/v1/limit-policies/:id              # 删除策略
POST   /api/v1/limit-policies/:id/approve      # 审批策略
GET    /api/v1/limit-policies/effective        # 获取生效策略
```

### 渠道策略

```
POST   /api/v1/channel-policies                # 创建渠道策略
GET    /api/v1/channel-policies/:id            # 获取策略详情
GET    /api/v1/channel-policies/merchant/:id   # 获取商户渠道策略
PUT    /api/v1/channel-policies/:id            # 更新策略
DELETE /api/v1/channel-policies/:id            # 删除策略
POST   /api/v1/channel-policies/:id/enable     # 启用渠道
POST   /api/v1/channel-policies/:id/disable    # 停用渠道
```

## 配置

### 环境变量

```bash
# 服务配置
PORT=40012
DB_NAME=payment_merchant_policy

# 数据库配置
DB_HOST=localhost
DB_PORT=40432
DB_USER=postgres
DB_PASSWORD=postgres

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=40379

# JWT配置
JWT_SECRET=payment-platform-secret-key-2024

# 配置中心
ENABLE_CONFIG_CLIENT=true
CONFIG_SERVICE_URL=http://localhost:40010
```

## 部署

### Docker

```bash
docker build -t merchant-policy-service .
docker run -p 40012:40012 merchant-policy-service
```

### Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## 监控

- **Metrics**: http://localhost:40012/metrics
- **Health**: http://localhost:40012/health
- **Swagger**: http://localhost:40012/swagger/index.html

## 开发

### 本地运行

```bash
cd backend/services/merchant-policy-service
go run cmd/main.go
```

### 热重载

```bash
air
```

### 测试

```bash
go test ./...
```

## 迁移说明

本服务由原 merchant-config-service 重构而来:

### 主要变化

1. **名称变更**: merchant-config-service → merchant-policy-service
2. **职责聚焦**: 专注策略配置,配额追踪移到quota-service
3. **模型优化**:
   - 新增 MerchantTier 模型 (从limit-service迁移)
   - 新增 MerchantPolicyBinding 模型
   - 优化 FeePolicy 和 LimitPolicy 支持等级默认策略

### 数据迁移

详见: `scripts/migrate-to-policy-service.sql`

## 版本

- **v1.0.0**: 初始版本 (重构自merchant-config-service)
- **端口**: 40012
- **数据库**: payment_merchant_policy
