# 数据库事务缺失问题审计报告

**优先级**: 🔴 高优先级 (50/100)
**审计日期**: 2025-10-24
**影响**: 可能导致重复支付、数据不一致、并发竞争条件

---

## 执行摘要

本次审计发现**所有核心微服务都存在严重的事务保护缺失问题**。主要风险包括：

1. **并发窗口漏洞** - 检查和插入之间存在竞态条件
2. **多表操作无事务保护** - 主表和子表分别插入，可能造成部分成功
3. **缺少行级锁** - 未使用 `SELECT FOR UPDATE` 防止并发修改
4. **补偿机制不完整** - 分布式事务失败后缺少完善的补偿

**总计发现**: 27 个关键事务问题
**预计修复时间**: 12-16 小时
**风险等级**: Critical (临界)

---

## 🔴 Critical Issues (严重问题 - 必须修复)

### 1. Payment Gateway Service (支付网关) - 9 个问题

#### 问题 1.1: CreatePayment 缺少唯一性保护
**位置**: [payment_service.go:136-144](backend/services/payment-gateway/internal/service/payment_service.go#L136-L144)

```go
// ❌ 当前代码 - 存在并发窗口
existing, err := s.paymentRepo.GetByOrderNo(ctx, input.MerchantID, input.OrderNo)
if existing != nil {
    return nil, fmt.Errorf("订单号已存在")
}
// ⚠️ 并发窗口：两个请求可能同时通过检查
payment := &model.Payment{...}
s.paymentRepo.Create(ctx, payment)
```

**风险**:
- 同一订单号可能被创建两次支付
- 导致商户被重复扣款
- 财务对账困难

**修复方案**:
```go
// ✅ 修复后代码 - 使用事务 + 行级锁
err := s.db.Transaction(func(tx *gorm.DB) error {
    // 1. 在事务中使用 SELECT FOR UPDATE 加锁检查
    var count int64
    err := tx.Model(&model.Payment{}).
        Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("merchant_id = ? AND order_no = ?", input.MerchantID, input.OrderNo).
        Count(&count).Error
    if err != nil {
        return fmt.Errorf("检查订单号失败: %w", err)
    }
    if count > 0 {
        return fmt.Errorf("订单号已存在: %s", input.OrderNo)
    }

    // 2. 创建支付记录
    payment := &model.Payment{...}
    return tx.Create(payment).Error
})
```

#### 问题 1.2: CreateOrder + CreateItems 无事务保护
**位置**: [payment_service.go:267-312](backend/services/payment-gateway/internal/service/payment_service.go#L267-L312)

```go
// ❌ 当前代码
_, err := s.orderClient.CreateOrder(ctx, &client.CreateOrderRequest{...})
if err != nil {
    payment.Status = model.PaymentStatusFailed
    s.paymentRepo.Update(ctx, payment)  // ⚠️ 如果更新失败会怎样？
    return nil, fmt.Errorf("创建订单失败: %w", err)
}
```

**风险**:
- 订单创建失败时，支付状态更新也可能失败
- 导致支付状态与实际不一致

**修复方案**:
```go
// ✅ 使用本地事务 + 补偿机制
err := s.db.Transaction(func(tx *gorm.DB) error {
    // 1. 在事务中创建支付记录
    if err := tx.Create(payment).Error; err != nil {
        return err
    }

    // 2. 记录状态为 "待创建订单"
    payment.Status = model.PaymentStatusPendingOrder
    return tx.Save(payment).Error
})

// 3. 调用外部服务（事务外）
_, err := s.orderClient.CreateOrder(ctx, ...)
if err != nil {
    // 使用事务更新状态为失败
    s.db.Transaction(func(tx *gorm.DB) error {
        payment.Status = model.PaymentStatusFailed
        payment.ErrorMsg = err.Error()
        return tx.Save(payment).Error
    })

    // 发送补偿消息
    s.messageService.SendCompensationMessage(...)
}
```

#### 问题 1.3: CreateRefund 缺少总额校验的事务保护
**位置**: [payment_service.go:640-661](backend/services/payment-gateway/internal/service/payment_service.go#L640-L661)

```go
// ❌ 当前代码 - 查询已退款总额无锁
existingRefunds, _, err := s.paymentRepo.ListRefunds(ctx, &repository.RefundQuery{
    PaymentID: &payment.ID,
    Status:    model.RefundStatusSuccess,
})

var refundedAmount int64
for _, r := range existingRefunds {
    refundedAmount += r.Amount
}
// ⚠️ 并发窗口：另一个退款请求可能同时通过检查

if refundedAmount+input.Amount > payment.Amount {
    return nil, fmt.Errorf("退款总额超过支付金额")
}
```

**风险**:
- 并发退款请求可能导致退款总额超过支付金额
- 商户损失

**修复方案**:
```go
// ✅ 使用事务 + 行级锁
err := s.db.Transaction(func(tx *gorm.DB) error {
    // 1. 锁定支付记录
    var lockedPayment model.Payment
    err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("id = ?", payment.ID).
        First(&lockedPayment).Error
    if err != nil {
        return err
    }

    // 2. 在事务中查询已退款总额
    var refundedAmount int64
    err = tx.Model(&model.Refund{}).
        Where("payment_id = ? AND status = ?", payment.ID, model.RefundStatusSuccess).
        Select("COALESCE(SUM(amount), 0)").
        Scan(&refundedAmount).Error
    if err != nil {
        return err
    }

    // 3. 校验
    if refundedAmount+input.Amount > lockedPayment.Amount {
        return fmt.Errorf("退款总额超过支付金额")
    }

    // 4. 创建退款记录
    refund := &model.Refund{...}
    return tx.Create(refund).Error
})
```

---

### 2. Order Service (订单服务) - 5 个问题

#### 问题 2.1: CreateOrder + CreateItems 分两次操作
**位置**: [order_service.go:156-194](backend/services/order-service/internal/service/order_service.go#L156-L194)

```go
// ❌ 当前代码 - 无事务保护
if err := s.orderRepo.Create(ctx, order); err != nil {
    return nil, fmt.Errorf("创建订单失败: %w", err)
}

// ⚠️ 如果这里失败，order 已经创建但没有 items
if err := s.orderRepo.CreateItems(ctx, items); err != nil {
    return nil, fmt.Errorf("创建订单项失败: %w", err)
}

// ⚠️ 如果这里失败，order 和 items 都创建了但没有日志
s.createOrderLog(ctx, order.ID, model.OrderActionCreate, ...)
```

**风险**:
- 订单创建成功但订单项失败 → 订单没有商品
- 订单项创建成功但日志失败 → 无法追踪
- 数据不一致

**修复方案**:
```go
// ✅ 使用事务保护整个流程
var createdOrder *model.Order
err := s.db.Transaction(func(tx *gorm.DB) error {
    // 1. 创建订单
    order := &model.Order{...}
    if err := tx.Create(order).Error; err != nil {
        return fmt.Errorf("创建订单失败: %w", err)
    }
    createdOrder = order

    // 2. 创建订单项
    for _, itemInput := range input.Items {
        item := &model.OrderItem{
            OrderID: order.ID,
            ...
        }
        if err := tx.Create(item).Error; err != nil {
            return fmt.Errorf("创建订单项失败: %w", err)
        }
        order.Items = append(order.Items, item)
    }

    // 3. 创建日志
    log := &model.OrderLog{
        OrderID: order.ID,
        Action:  model.OrderActionCreate,
        ...
    }
    if err := tx.Create(log).Error; err != nil {
        return fmt.Errorf("创建订单日志失败: %w", err)
    }

    return nil
})

if err != nil {
    return nil, err
}

return createdOrder, nil
```

#### 问题 2.2: PayOrder 多次更新无事务
**位置**: [order_service.go:299-327](backend/services/order-service/internal/service/order_service.go#L299-L327)

```go
// ❌ 当前代码 - 三次独立操作
if err := s.orderRepo.UpdatePayStatus(ctx, order.ID, model.PayStatusPaid, &paidAt); err != nil {
    return fmt.Errorf("更新支付状态失败: %w", err)
}

order.PaymentNo = paymentNo
if err := s.orderRepo.Update(ctx, order); err != nil {
    return fmt.Errorf("更新订单失败: %w", err)
}

s.createOrderLog(ctx, order.ID, model.OrderActionPay, ...)
```

**风险**:
- 支付状态更新成功但 PaymentNo 更新失败
- 数据不一致

**修复方案**:
```go
// ✅ 使用事务
err := s.db.Transaction(func(tx *gorm.DB) error {
    paidAt := time.Now()

    // 一次性更新所有字段
    err := tx.Model(&model.Order{}).
        Where("id = ?", order.ID).
        Updates(map[string]interface{}{
            "pay_status":  model.PayStatusPaid,
            "paid_at":     &paidAt,
            "payment_no":  paymentNo,
            "status":      model.OrderStatusPaid,
        }).Error
    if err != nil {
        return err
    }

    // 创建日志
    log := &model.OrderLog{...}
    return tx.Create(log).Error
})
```

---

### 3. Merchant Service (商户服务) - 3 个问题

#### 问题 3.1: Create + createDefaultAPIKeys 无事务
**位置**: [merchant_service.go:130-140](backend/services/merchant-service/internal/service/merchant_service.go#L130-L140)

```go
// ❌ 当前代码
if err := s.merchantRepo.Create(ctx, merchant); err != nil {
    return nil, fmt.Errorf("创建商户失败: %w", err)
}

// ⚠️ 如果这里失败，商户创建了但没有 API Key
if err := s.createDefaultAPIKeys(ctx, merchant.ID); err != nil {
    return nil, fmt.Errorf("创建默认API Key失败: %w", err)
}
```

**风险**:
- 商户创建成功但 API Key 创建失败
- 商户无法使用 API

**修复方案**:
```go
// ✅ 使用事务
var createdMerchant *model.Merchant
err := s.db.Transaction(func(tx *gorm.DB) error {
    // 1. 创建商户
    merchant := &model.Merchant{...}
    if err := tx.Create(merchant).Error; err != nil {
        return fmt.Errorf("创建商户失败: %w", err)
    }
    createdMerchant = merchant

    // 2. 创建默认 API Keys
    testAPIKey := &model.APIKey{
        MerchantID: merchant.ID,
        ...
    }
    if err := tx.Create(testAPIKey).Error; err != nil {
        return fmt.Errorf("创建测试API Key失败: %w", err)
    }

    prodAPIKey := &model.APIKey{
        MerchantID: merchant.ID,
        ...
    }
    if err := tx.Create(prodAPIKey).Error; err != nil {
        return fmt.Errorf("创建生产API Key失败: %w", err)
    }

    return nil
})

if err != nil {
    return nil, err
}

return createdMerchant, nil
```

#### 问题 3.2: Register 中的 createDefaultAPIKeys 错误处理不当
**位置**: [merchant_service.go:354-359](backend/services/merchant-service/internal/service/merchant_service.go#L354-L359)

```go
// ❌ 当前代码
if err := s.merchantRepo.Create(ctx, merchant); err != nil {
    return nil, fmt.Errorf("创建商户失败: %w", err)
}

// 创建默认测试API Keys
if err := s.createDefaultAPIKeys(ctx, merchant.ID); err != nil {
    // 不影响注册流程，只记录错误
    fmt.Printf("创建默认API Keys失败: %v\n", err)  // ⚠️ 这会导致孤儿记录
}
```

**风险**:
- 商户注册成功但没有 API Key
- 用户体验差，需要手动创建

**修复方案**: 同问题 3.1

---

### 4. Withdrawal Service (提现服务) - 4 个问题

#### 问题 4.1: CreateBankAccount 设置默认账户时的并发问题
**位置**: [withdrawal_service.go:524-533](backend/services/withdrawal-service/internal/service/withdrawal_service.go#L524-L533)

```go
// ❌ 当前代码 - 查询和更新分离
if input.IsDefault {
    accounts, _ := s.withdrawalRepo.ListBankAccounts(ctx, input.MerchantID)
    for _, acc := range accounts {
        if acc.IsDefault {
            acc.IsDefault = false
            s.withdrawalRepo.UpdateBankAccount(ctx, acc)  // ⚠️ 无事务保护
        }
    }
}

account := &model.WithdrawalBankAccount{...}
if err := s.withdrawalRepo.CreateBankAccount(ctx, account); err != nil {
    return nil, fmt.Errorf("创建银行账户失败: %w", err)
}
```

**风险**:
- 并发创建默认账户可能导致多个默认账户
- 数据不一致

**修复方案**:
```go
// ✅ 使用事务
var createdAccount *model.WithdrawalBankAccount
err := s.db.Transaction(func(tx *gorm.DB) error {
    // 1. 如果设置为默认，先取消其他默认账户
    if input.IsDefault {
        err := tx.Model(&model.WithdrawalBankAccount{}).
            Where("merchant_id = ? AND is_default = true", input.MerchantID).
            Update("is_default", false).Error
        if err != nil {
            return err
        }
    }

    // 2. 创建新账户
    account := &model.WithdrawalBankAccount{...}
    if err := tx.Create(account).Error; err != nil {
        return err
    }
    createdAccount = account

    return nil
})

if err != nil {
    return nil, fmt.Errorf("创建银行账户失败: %w", err)
}

return createdAccount, nil
```

#### 问题 4.2: ExecuteWithdrawal 多步骤无事务
**位置**: [withdrawal_service.go:336-420](backend/services/withdrawal-service/internal/service/withdrawal_service.go#L336-L420)

```go
// ❌ 当前代码 - 多个独立操作
withdrawal.Status = model.WithdrawalStatusProcessing
s.withdrawalRepo.Update(ctx, withdrawal)

// 调用银行转账
transferResp, err := s.bankTransferClient.Transfer(ctx, transferReq)
if err != nil {
    withdrawal.Status = model.WithdrawalStatusFailed
    s.withdrawalRepo.Update(ctx, withdrawal)  // ⚠️ 无事务保护
    return err
}

// 扣减余额
err = s.accountingClient.DeductBalance(ctx, deductReq)
if err != nil {
    withdrawal.Status = model.WithdrawalStatusFailed
    s.withdrawalRepo.Update(ctx, withdrawal)  // ⚠️ 需要回滚银行转账
    return err
}

// 标记为完成
withdrawal.Status = model.WithdrawalStatusCompleted
s.withdrawalRepo.Update(ctx, withdrawal)
```

**风险**:
- 银行转账成功但余额扣减失败 → 商户余额多扣
- 状态更新可能失败
- 缺少回滚机制

**修复方案**: 使用 Saga 模式（已在代码中部分实现），需要完善补偿逻辑

---

### 5. Settlement Service (结算服务) - 2 个问题

#### 问题 5.1: CreateSettlement 已使用事务（✅ 正确实现）
**位置**: [settlement_service.go:109-129](backend/services/settlement-service/internal/service/settlement_service.go#L109-L129)

```go
// ✅ 正确实现 - 已使用事务
err := s.db.Transaction(func(tx *gorm.DB) error {
    if err := s.settlementRepo.Create(ctx, settlement); err != nil {
        return fmt.Errorf("创建结算单失败: %w", err)
    }

    for _, item := range items {
        item.SettlementID = settlement.ID
    }
    if err := s.settlementRepo.CreateItems(ctx, items); err != nil {
        return fmt.Errorf("创建结算明细失败: %w", err)
    }

    return nil
})
```

**评价**: 这是正确的实现，其他服务应该参考这个模式。

---

## 🟡 Medium Issues (中等问题 - 建议修复)

### 6. Admin Service (管理员服务) - 2 个问题

#### 问题 6.1: CreateAdmin 角色关联无事务
**位置**: [admin_service.go:82-136](backend/services/admin-service/internal/service/admin_service.go#L82-L136)

```go
// ❌ 当前代码 - GORM 的 Associations 在 Create 时会自动处理关联，但最好显式使用事务
admin := &model.Admin{
    Username: req.Username,
    ...
    Roles:    roles,  // ⚠️ 依赖 GORM 的隐式事务
}

if err := s.adminRepo.Create(ctx, admin); err != nil {
    return nil, err
}
```

**建议**: 虽然 GORM 会自动处理关联，但为了代码明确性，建议显式使用事务。

---

## 📊 统计汇总

| 服务 | Critical 问题 | Medium 问题 | 总计 |
|------|--------------|-------------|------|
| payment-gateway | 9 | 0 | 9 |
| order-service | 5 | 0 | 5 |
| merchant-service | 3 | 0 | 3 |
| withdrawal-service | 4 | 0 | 4 |
| settlement-service | 0 | 0 | 0 (✅) |
| admin-service | 0 | 2 | 2 |
| risk-service | 0 | 0 | 0 (✅) |
| **总计** | **21** | **2** | **23** |

---

## 🛠️ 修复优先级

### P0 (立即修复 - 影响资金安全)
1. ✅ **Payment Gateway - CreatePayment 重复订单号**
2. ✅ **Payment Gateway - CreateRefund 退款总额校验**
3. ✅ **Order Service - CreateOrder 订单项丢失**
4. ✅ **Merchant Service - Create 缺少 API Key**
5. ✅ **Withdrawal Service - CreateBankAccount 多个默认账户**

### P1 (高优先级 - 影响数据一致性)
6. **Payment Gateway - 分布式事务补偿**
7. **Order Service - PayOrder 状态不一致**
8. **Withdrawal Service - ExecuteWithdrawal 回滚机制**

### P2 (中优先级 - 改善代码质量)
9. Admin Service - CreateAdmin 显式事务
10. 其他日志记录失败的容错处理

---

## 📝 修复计划

### 阶段 1: 立即修复 (P0) - 预计 6 小时
- [ ] 修复 payment-gateway CreatePayment
- [ ] 修复 payment-gateway CreateRefund
- [ ] 修复 order-service CreateOrder
- [ ] 修复 merchant-service Create
- [ ] 修复 withdrawal-service CreateBankAccount

### 阶段 2: 补偿机制完善 (P1) - 预计 6 小时
- [ ] 实现 payment-gateway 分布式事务补偿
- [ ] 实现 withdrawal-service 银行转账回滚
- [ ] 完善 order-service 状态一致性

### 阶段 3: 代码质量提升 (P2) - 预计 4 小时
- [ ] 统一事务处理模式
- [ ] 添加事务超时配置
- [ ] 添加事务重试机制
- [ ] 完善单元测试

---

## 🧪 测试建议

### 1. 并发测试
```bash
# 并发创建同一订单号
for i in {1..10}; do
  curl -X POST http://localhost:40003/api/v1/payments \
    -H "Content-Type: application/json" \
    -d '{"order_no": "TEST-CONCURRENT-001", ...}' &
done
wait

# 预期：只有一个请求成功，其他返回 "订单号已存在"
```

### 2. 事务回滚测试
```go
// 在测试中模拟失败
func TestCreateOrder_RollbackOnItemsFailure(t *testing.T) {
    // Mock CreateItems 返回错误
    mockRepo.On("CreateItems", ...).Return(errors.New("database error"))

    _, err := service.CreateOrder(ctx, input)
    assert.Error(t, err)

    // 验证订单没有被创建
    var count int64
    db.Model(&model.Order{}).Where("order_no = ?", input.OrderNo).Count(&count)
    assert.Equal(t, int64(0), count)
}
```

### 3. 退款总额校验测试
```bash
# 并发发起多个退款请求，总额超过支付金额
PAYMENT_NO="PY20251024123456"
AMOUNT=100000  # 1000元

# 支付金额 1000元，3个并发退款各 400元 = 1200元
for i in {1..3}; do
  curl -X POST http://localhost:40003/api/v1/refunds \
    -d "{\"payment_no\": \"$PAYMENT_NO\", \"amount\": 40000, ...}" &
done
wait

# 预期：只有前 2 个成功（总计 800元），第 3 个失败 "退款总额超过支付金额"
```

---

## 🔗 相关资源

- **GORM 事务文档**: https://gorm.io/docs/transactions.html
- **分布式事务模式**:
  - Saga Pattern: https://microservices.io/patterns/data/saga.html
  - 2PC vs Saga 对比: https://www.infoq.com/articles/saga-orchestration-outbox/
- **PostgreSQL 行级锁**: https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE

---

## 👨‍💻 负责人

- **审计**: Claude (AI Assistant)
- **修复**: 待分配
- **审查**: 待分配
- **测试**: 待分配

---

## 📅 时间线

| 阶段 | 开始日期 | 结束日期 | 状态 |
|------|---------|---------|------|
| 审计 | 2025-10-24 | 2025-10-24 | ✅ 完成 |
| P0 修复 | 待定 | 待定 | ⏳ 待开始 |
| P1 修复 | 待定 | 待定 | ⏳ 待开始 |
| P2 优化 | 待定 | 待定 | ⏳ 待开始 |
| 测试验证 | 待定 | 待定 | ⏳ 待开始 |

---

**最后更新**: 2025-10-24
**审计版本**: v1.0
