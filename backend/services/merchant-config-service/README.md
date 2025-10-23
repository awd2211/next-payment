# Merchant Config Service

> 商户配置管理服务

## 职责

从 `merchant-service` 拆分出来的配置管理模块，负责商户的API密钥、渠道配置、费率和限额管理。

## 核心功能

- 🔑 API密钥管理
- 🔌 支付渠道配置
- 💰 费率配置管理
- 📊 交易限额设置
- 🔄 配置版本管理

## 数据库

**Database**: `payment_merchant_config`

**Tables**:
- api_keys（API密钥）
- channel_configs（渠道配置）
- merchant_fee_configs（费率配置）
- merchant_transaction_limits（交易限额）

## 端口

**Port**: `8015`

## 状态

📋 **预留中** - 待从 merchant-service 拆分

## 依赖服务

- merchant-service（验证商户状态）
- channel-adapter（验证渠道配置）

## API端点

```
POST   /api/v1/api-keys                # 创建API密钥
GET    /api/v1/api-keys                # API密钥列表
DELETE /api/v1/api-keys/:id            # 删除API密钥
POST   /api/v1/channels                # 配置支付渠道
GET    /api/v1/channels                # 渠道配置列表
PUT    /api/v1/channels/:id            # 更新渠道配置
POST   /api/v1/fee-configs             # 创建费率配置
GET    /api/v1/limits                  # 获取交易限额
PUT    /api/v1/limits/:id              # 更新交易限额
```

## 启动命令

```bash
PORT=8015 \
DB_NAME=payment_merchant_config \
go run ./cmd/main.go
```

## 拆分计划

- 预计工作量：3周
- 优先级：P2（第五批拆分）
- 开始时间：待定
