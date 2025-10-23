# Settlement Service

> 结算处理服务

## 职责

从 `accounting-service` 拆分出来的结算模块，负责商户的批量结算、对账和清算。

## 核心功能

- 💰 自动结算（每日/每周/每月）
- 📊 交易汇总
- 🧾 费用计算
- ✅ 结算审批
- 📈 结算报表

## 数据库

**Database**: `payment_settlement`

**Tables**:
- settlements（结算单）
- settlement_items（结算明细）
- settlement_approvals（结算审批）

## 端口

**Port**: `8012`

## 状态

📋 **预留中** - 待从 accounting-service 拆分

## 依赖服务

- accounting-service（读取交易记录）
- withdrawal-service（触发提现）
- notification-service（结算通知）

## API端点

```
POST   /api/v1/settlements             # 创建结算单
GET    /api/v1/settlements             # 结算单列表
GET    /api/v1/settlements/:id         # 结算单详情
POST   /api/v1/settlements/:id/approve # 审批结算
POST   /api/v1/settlements/:id/execute # 执行结算
GET    /api/v1/settlements/reports     # 结算报表
```

## 启动命令

```bash
PORT=8012 \
DB_NAME=payment_settlement \
go run ./cmd/main.go
```

## 拆分计划

- 预计工作量：3周
- 优先级：P1（第二批拆分）
- 开始时间：待定
