# Withdrawal Service

> 提现管理服务

## 职责

从 `accounting-service` 拆分出来的提现模块，负责商户提现的审批、风控和银行转账。

## 核心功能

- 💸 提现申请
- 🔍 提现审批（多级审批）
- ⚠️ 风控检查
- 🏦 银行转账集成
- 📋 提现记录查询

## 数据库

**Database**: `payment_withdrawal`

**Tables**:
- withdrawals（提现单）
- withdrawal_approvals（审批记录）
- withdrawal_bank_transfers（银行转账记录）

## 端口

**Port**: `8013`

## 状态

📋 **预留中** - 待从 accounting-service 拆分

## 依赖服务

- accounting-service（扣减账户余额）
- risk-service（提现风控检查）
- notification-service（提现通知）

## API端点

```
POST   /api/v1/withdrawals             # 创建提现申请
GET    /api/v1/withdrawals             # 提现列表
GET    /api/v1/withdrawals/:id         # 提现详情
POST   /api/v1/withdrawals/:id/approve # 审批提现
POST   /api/v1/withdrawals/:id/reject  # 拒绝提现
POST   /api/v1/withdrawals/:id/process # 执行提现
```

## 启动命令

```bash
PORT=8013 \
DB_NAME=payment_withdrawal \
go run ./cmd/main.go
```

## 拆分计划

- 预计工作量：4周
- 优先级：P1（第三批拆分）
- 开始时间：待定
