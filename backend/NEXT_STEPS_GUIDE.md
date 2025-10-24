# 🚀 Next Steps Guide - Phase 9 & 10 实施指南

**目标读者**: 负责数据迁移和代码清理的工程师
**前置条件**: Phase 1-8 已完成，所有新服务编译成功
**预计耗时**: 5-7 小时

---

## 📋 快速检查清单

在开始之前，请确认以下条件：

- [ ] 所有新服务编译成功（merchant-auth-service, merchant-config-service, settlement-service）
- [ ] 数据库已创建（payment_merchant_auth, payment_merchant_config, payment_settlement）
- [ ] 阅读过完整总结报告 [MERCHANT_SERVICE_REFACTORING_COMPLETE.md](./MERCHANT_SERVICE_REFACTORING_COMPLETE.md)
- [ ] 了解重构架构 [REFACTORING_FINAL_SUMMARY.txt](./REFACTORING_FINAL_SUMMARY.txt)
- [ ] 准备好数据库备份策略

---

## 🎯 Phase 9: 数据迁移（P0 优先级）

**目标**: 将现有数据从 merchant-service 迁移到新服务
**优先级**: P0（高优先级，必须完成）
**预计耗时**: 2-3小时

### 9.1 迁移准备

#### Step 1: 备份所有数据

```bash
# 创建备份目录
mkdir -p /home/eric/payment/backend/backups/$(date +%Y%m%d)

# 备份 merchant-service 数据库
PGPASSWORD=postgres pg_dump -h localhost -p 40432 -U postgres \
  payment_merchant > /home/eric/payment/backend/backups/$(date +%Y%m%d)/merchant_service_backup.sql

# 验证备份文件
ls -lh /home/eric/payment/backend/backups/$(date +%Y%m%d)/
```

#### Step 2: 验证目标数据库存在

```bash
# 检查新数据库是否已创建
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -c "\l" | grep -E "payment_merchant_auth|payment_merchant_config"

# 检查表结构是否已创建（通过 AutoMigrate）
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_auth -c "\dt"
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_config -c "\dt"
```

### 9.2 迁移 APIKey 数据

**源表**: `payment_merchant.api_keys`
**目标表**: `payment_merchant_auth.api_keys`

```bash
# 导出数据（CSV格式）
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant <<EOF
\copy (SELECT * FROM api_keys ORDER BY created_at) TO '/tmp/api_keys_export.csv' WITH CSV HEADER;
EOF

# 查看导出数据
head -5 /tmp/api_keys_export.csv

# 导入到新数据库
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_auth <<EOF
\copy api_keys FROM '/tmp/api_keys_export.csv' WITH CSV HEADER;
EOF

# 验证数据
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_auth <<EOF
SELECT COUNT(*) AS total_api_keys FROM api_keys;
SELECT environment, COUNT(*) FROM api_keys GROUP BY environment;
EOF
```

### 9.3 迁移 SettlementAccount 数据

**源表**: `payment_merchant.settlement_accounts`
**目标表**: `payment_settlement.settlement_accounts`

```bash
# 导出数据
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant <<EOF
\copy (SELECT * FROM settlement_accounts ORDER BY created_at) TO '/tmp/settlement_accounts_export.csv' WITH CSV HEADER;
EOF

# 导入到新数据库
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_settlement <<EOF
\copy settlement_accounts FROM '/tmp/settlement_accounts_export.csv' WITH CSV HEADER;
EOF

# 验证数据
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_settlement <<EOF
SELECT COUNT(*) AS total_accounts FROM settlement_accounts;
SELECT account_type, COUNT(*) FROM settlement_accounts GROUP BY account_type;
SELECT status, COUNT(*) FROM settlement_accounts GROUP BY status;
EOF
```

### 9.4 迁移配置数据

#### 9.4.1 MerchantFeeConfig

```bash
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant <<EOF
\copy (SELECT * FROM merchant_fee_configs ORDER BY created_at) TO '/tmp/merchant_fee_configs_export.csv' WITH CSV HEADER;
EOF

PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_config <<EOF
\copy merchant_fee_configs FROM '/tmp/merchant_fee_configs_export.csv' WITH CSV HEADER;
EOF
```

#### 9.4.2 MerchantTransactionLimit

```bash
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant <<EOF
\copy (SELECT * FROM merchant_transaction_limits ORDER BY created_at) TO '/tmp/merchant_transaction_limits_export.csv' WITH CSV HEADER;
EOF

PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_config <<EOF
\copy merchant_transaction_limits FROM '/tmp/merchant_transaction_limits_export.csv' WITH CSV HEADER;
EOF
```

#### 9.4.3 ChannelConfig

```bash
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant <<EOF
\copy (SELECT * FROM channel_configs ORDER BY created_at) TO '/tmp/channel_configs_export.csv' WITH CSV HEADER;
EOF

PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_config <<EOF
\copy channel_configs FROM '/tmp/channel_configs_export.csv' WITH CSV HEADER;
EOF
```

### 9.5 验证迁移完整性

创建验证脚本：

```bash
cat > /home/eric/payment/backend/scripts/verify_migration.sh <<'EOF'
#!/bin/bash

echo "========================================"
echo "数据迁移验证脚本"
echo "========================================"

# 源数据库计数
echo ""
echo "【源数据库 - payment_merchant】"
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant -t -c "
SELECT
  'api_keys: ' || COUNT(*) FROM api_keys
  UNION ALL
  SELECT 'settlement_accounts: ' || COUNT(*) FROM settlement_accounts
  UNION ALL
  SELECT 'merchant_fee_configs: ' || COUNT(*) FROM merchant_fee_configs
  UNION ALL
  SELECT 'merchant_transaction_limits: ' || COUNT(*) FROM merchant_transaction_limits
  UNION ALL
  SELECT 'channel_configs: ' || COUNT(*) FROM channel_configs;
"

# 目标数据库计数
echo ""
echo "【目标数据库 - merchant-auth-service】"
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_auth -t -c "
SELECT 'api_keys: ' || COUNT(*) FROM api_keys;
"

echo ""
echo "【目标数据库 - settlement-service】"
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_settlement -t -c "
SELECT 'settlement_accounts: ' || COUNT(*) FROM settlement_accounts;
"

echo ""
echo "【目标数据库 - merchant-config-service】"
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_merchant_config -t -c "
SELECT
  'merchant_fee_configs: ' || COUNT(*) FROM merchant_fee_configs
  UNION ALL
  SELECT 'merchant_transaction_limits: ' || COUNT(*) FROM merchant_transaction_limits
  UNION ALL
  SELECT 'channel_configs: ' || COUNT(*) FROM channel_configs;
"

echo ""
echo "========================================"
echo "验证完成！请对比源数据和目标数据的数量"
echo "========================================"
EOF

chmod +x /home/eric/payment/backend/scripts/verify_migration.sh
./scripts/verify_migration.sh
```

### 9.6 更新应用配置

启动新服务并测试：

```bash
# 启动 merchant-auth-service
cd /home/eric/payment/backend/services/merchant-auth-service
export DB_NAME=payment_merchant_auth PORT=40011
go run cmd/main.go &

# 启动 merchant-config-service
cd /home/eric/payment/backend/services/merchant-config-service
export DB_NAME=payment_merchant_config PORT=40012
go run cmd/main.go &

# 测试健康检查
curl http://localhost:40011/health
curl http://localhost:40012/health

# 测试 API（需要有数据）
# 例如：列出 API Keys
curl http://localhost:40011/api/v1/api-keys -H "Authorization: Bearer <token>"
```

### 9.7 灰度切换

在 payment-gateway 中启用新服务：

```bash
# 设置环境变量
export USE_AUTH_SERVICE=true
export MERCHANT_AUTH_SERVICE_URL=http://localhost:40011
export MERCHANT_CONFIG_SERVICE_URL=http://localhost:40012

# 重启 payment-gateway
# ...
```

---

## 🧹 Phase 10: 代码清理（P1 优先级）

**目标**: 清理 merchant-service 中已迁移的代码
**优先级**: P1（中优先级，建议完成）
**预计耗时**: 3-4小时

### 10.1 删除已迁移的模型

编辑 `/home/eric/payment/backend/services/merchant-service/internal/model/`：

#### Step 1: 删除 merchant.go 中的模型

删除以下内容：
- `APIKey` struct (行 40-52)
- `ChannelConfig` struct (行 60-75)
- 相关常量（行 98-111）

保留：
- `Merchant` struct ✅

#### Step 2: 删除 business.go 中的模型

删除以下内容：
- `SettlementAccount` struct (行 11-34)
- `KYCDocument` struct (行 42-60)
- `BusinessQualification` struct (行 68-85)
- `MerchantFeeConfig` struct (行 93-116)
- `MerchantTransactionLimit` struct (行 152-169)

保留：
- `MerchantUser` struct ✅
- `MerchantContract` struct ✅

#### Step 3: 删除相关常量

删除 business.go 中的常量定义（行 204-298），除了保留 MerchantUser 和 MerchantContract 相关的常量。

### 10.2 删除 Repository 层

删除以下文件（如果存在）：
```bash
rm -f /home/eric/payment/backend/services/merchant-service/internal/repository/api_key_repository.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/repository/settlement_account_repository.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/repository/fee_config_repository.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/repository/transaction_limit_repository.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/repository/channel_config_repository.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/repository/kyc_repository.go
```

### 10.3 删除 Service 层

删除以下文件（如果存在）：
```bash
rm -f /home/eric/payment/backend/services/merchant-service/internal/service/api_key_service.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/service/settlement_account_service.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/service/fee_config_service.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/service/transaction_limit_service.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/service/channel_config_service.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/service/kyc_service.go
```

### 10.4 删除 Handler 层

删除以下文件（如果存在）：
```bash
rm -f /home/eric/payment/backend/services/merchant-service/internal/handler/api_key_handler.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/handler/settlement_account_handler.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/handler/fee_config_handler.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/handler/transaction_limit_handler.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/handler/channel_config_handler.go
rm -f /home/eric/payment/backend/services/merchant-service/internal/handler/kyc_handler.go
```

### 10.5 更新 main.go

编辑 `/home/eric/payment/backend/services/merchant-service/cmd/main.go`：

删除 AutoMigrate 中的已迁移模型：
```go
// Before
if err := database.AutoMigrate(
    &model.Merchant{},
    &model.APIKey{},              // ❌ 删除
    &model.ChannelConfig{},       // ❌ 删除
    &model.SettlementAccount{},   // ❌ 删除
    &model.KYCDocument{},         // ❌ 删除
    &model.BusinessQualification{}, // ❌ 删除
    &model.MerchantFeeConfig{},   // ❌ 删除
    &model.MerchantTransactionLimit{}, // ❌ 删除
    &model.MerchantUser{},
    &model.MerchantContract{},
); err != nil {

// After
if err := database.AutoMigrate(
    &model.Merchant{},
    &model.MerchantUser{},
    &model.MerchantContract{},
); err != nil {
```

删除相关的 repository, service, handler 初始化代码。

### 10.6 更新 API 文档

更新 Swagger 注释，移除已迁移的端点说明。

### 10.7 更新前端

#### Admin Portal 更新

编辑 `/home/eric/payment/frontend/admin-portal/src/services/api.ts`：

添加新服务的 API 端点：
```typescript
// 新增：merchant-auth-service
export const authServiceAPI = axios.create({
  baseURL: 'http://localhost:40011/api/v1',
  headers: { 'Authorization': `Bearer ${token}` }
});

// 新增：merchant-config-service
export const configServiceAPI = axios.create({
  baseURL: 'http://localhost:40012/api/v1',
  headers: { 'Authorization': `Bearer ${token}` }
});

// 新增：settlement-service
export const settlementServiceAPI = axios.create({
  baseURL: 'http://localhost:40013/api/v1',
  headers: { 'Authorization': `Bearer ${token}` }
});
```

更新相关页面的 API 调用：
- API Key 管理页面 → 调用 authServiceAPI
- 费率配置页面 → 调用 configServiceAPI
- 交易限额页面 → 调用 configServiceAPI
- 渠道配置页面 → 调用 configServiceAPI
- 结算账户页面 → 调用 settlementServiceAPI

#### Merchant Portal 更新

类似地更新 merchant-portal 的 API 调用。

### 10.8 编译验证

```bash
# 编译 merchant-service（精简版）
cd /home/eric/payment/backend/services/merchant-service
GOWORK=/home/eric/payment/backend/go.work go build -o /tmp/merchant-service ./cmd/main.go

# 检查二进制大小（应该比之前小）
ls -lh /tmp/merchant-service

# 运行测试
go test ./...
```

### 10.9 清理临时文件

```bash
# 删除导出的CSV文件
rm -f /tmp/api_keys_export.csv
rm -f /tmp/settlement_accounts_export.csv
rm -f /tmp/merchant_fee_configs_export.csv
rm -f /tmp/merchant_transaction_limits_export.csv
rm -f /tmp/channel_configs_export.csv
```

---

## ✅ 验收标准

### Phase 9 完成标准

- [ ] 所有数据成功迁移到新数据库
- [ ] 源数据和目标数据数量一致（通过 verify_migration.sh 验证）
- [ ] 新服务启动成功，健康检查通过
- [ ] payment-gateway 可以成功调用新服务
- [ ] 备份文件已保存

### Phase 10 完成标准

- [ ] 已迁移的模型从 merchant-service 删除
- [ ] 相关的 repository/service/handler 代码删除
- [ ] main.go AutoMigrate 更新
- [ ] merchant-service 编译成功
- [ ] 前端页面可以调用新服务API
- [ ] API 文档更新

---

## 🚨 回滚计划

### 如果数据迁移失败

```bash
# 1. 停止所有新服务
killall merchant-auth-service merchant-config-service

# 2. 恢复备份
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres \
  payment_merchant < /home/eric/payment/backend/backups/YYYYMMDD/merchant_service_backup.sql

# 3. 重启 merchant-service（旧版本）
# 4. 在 payment-gateway 中设置 USE_AUTH_SERVICE=false
```

### 如果代码清理后出现问题

```bash
# 使用 git 恢复删除的文件
git checkout HEAD -- services/merchant-service/internal/model/
git checkout HEAD -- services/merchant-service/internal/repository/
git checkout HEAD -- services/merchant-service/internal/service/
git checkout HEAD -- services/merchant-service/internal/handler/
git checkout HEAD -- services/merchant-service/cmd/main.go
```

---

## 📞 需要帮助？

参考以下文档：
- [MERCHANT_SERVICE_REFACTORING_COMPLETE.md](./MERCHANT_SERVICE_REFACTORING_COMPLETE.md) - 完整总结
- [REFACTORING_FINAL_SUMMARY.txt](./REFACTORING_FINAL_SUMMARY.txt) - 快速参考
- [PHASE1_MIGRATION_COMPLETE.md](./PHASE1_MIGRATION_COMPLETE.md) - APIKey 迁移示例
- [PHASE3_MIGRATION_COMPLETE.md](./PHASE3_MIGRATION_COMPLETE.md) - SettlementAccount 迁移示例

---

**最后更新**: 2025-10-24
**作者**: Claude Code Assistant
**状态**: Ready for Phase 9-10 实施

---
