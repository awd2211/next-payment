# Phase 1 实施指南：APIKey 迁移到 merchant-auth-service

## 已完成的工作 ✅

### 1. merchant-auth-service 新增功能
- ✅ 模型：`internal/model/api_key.go`
- ✅ 仓储层：`internal/repository/api_key_repository.go`
- ✅ 服务层：`internal/service/api_key_service.go`
- ✅ HTTP API：`internal/handler/api_key_handler.go`
- ✅ 路由注册：修改 `cmd/main.go`
- ✅ 编译测试：通过 ✅

### 2. payment-gateway 新增客户端
- ✅ 客户端：`internal/client/merchant_auth_client.go`
- ✅ 简化中间件：`internal/middleware/signature_v2.go`

## 下一步：集成和测试

### 步骤 1：修改 payment-gateway/cmd/main.go（渐进式迁移）

在 `cmd/main.go` 中添加环境变量切换：

```go
// 9. 初始化签名验证中间件
useAuthService := config.GetEnv("USE_AUTH_SERVICE", "false") == "true"

var signatureMiddleware gin.HandlerFunc

if useAuthService {
    // 新方案：调用 merchant-auth-service
    logger.Info("使用 merchant-auth-service 进行签名验证")
    authServiceURL := config.GetEnv("MERCHANT_AUTH_SERVICE_URL", "http://localhost:40011")
    authClient := client.NewMerchantAuthClient(authServiceURL)
    signatureMW := localMiddleware.NewSignatureMiddlewareV2(authClient)
    signatureMiddleware = signatureMW.Verify()
} else {
    // 旧方案：本地验证（向后兼容）
    logger.Info("使用本地 API Key 进行签名验证")
    signatureMW := localMiddleware.NewSignatureMiddleware(
        func(apiKey string) (*localMiddleware.APIKeyData, error) {
            ctx := context.Background()
            key, err := apiKeyRepo.GetByAPIKey(ctx, apiKey)
            if err != nil {
                return nil, err
            }
            return &localMiddleware.APIKeyData{
                Secret:       key.APISecret,
                MerchantID:   key.MerchantID,
                IsActive:     key.IsActive,
                ExpiresAt:    key.ExpiresAt,
                Environment:  key.Environment,
                IPWhitelist:  key.IPWhitelist,
                ShouldRotate: key.ShouldRotate(),
            }, nil
        },
        application.Redis,
    )
    signatureMW.SetAPIKeyUpdater(apiKeyRepo)
    signatureMiddleware = signatureMW.Verify()
}

// 应用到路由
// payments.Use(signatureMiddleware) // 保持原样
```

### 步骤 2：数据迁移

```bash
# 1. 创建迁移脚本
cat > /tmp/migrate_api_keys.sh << 'EOF'
#!/bin/bash
set -e

echo "开始迁移 API Keys..."

# 从 payment_gateway 导出
docker exec payment-postgres psql -U postgres -d payment_gateway -c "
COPY (SELECT * FROM api_keys ORDER BY created_at)
TO STDOUT CSV HEADER
" > /tmp/api_keys.csv

# 导入到 payment_merchant_auth
docker exec -i payment-postgres psql -U postgres -d payment_merchant_auth -c "
COPY api_keys(id, merchant_id, api_key, api_secret, name, environment, is_active, last_used_at, expires_at, created_at, updated_at)
FROM STDIN CSV HEADER
" < /tmp/api_keys.csv

# 验证数据
echo "验证数据..."
PG_COUNT=$(docker exec payment-postgres psql -U postgres -d payment_gateway -t -c "SELECT COUNT(*) FROM api_keys")
MA_COUNT=$(docker exec payment-postgres psql -U postgres -d payment_merchant_auth -t -c "SELECT COUNT(*) FROM api_keys")

echo "payment_gateway: $PG_COUNT rows"
echo "payment_merchant_auth: $MA_COUNT rows"

if [ "$PG_COUNT" == "$MA_COUNT" ]; then
    echo "✅ 数据迁移成功！"
else
    echo "❌ 数据迁移失败，行数不匹配"
    exit 1
fi
EOF

chmod +x /tmp/migrate_api_keys.sh
```

### 步骤 3：集成测试流程

```bash
# 1. 启动 merchant-auth-service
cd /home/eric/payment/backend/services/merchant-auth-service
DB_HOST=localhost DB_PORT=40432 DB_USER=postgres DB_PASSWORD=postgres \
DB_NAME=payment_merchant_auth PORT=40011 \
go run cmd/main.go &

# 2. 等待服务启动
sleep 3

# 3. 创建测试 API Key（通过 merchant-auth-service）
curl -X POST http://localhost:40011/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Key",
    "environment": "test"
  }'

# 保存返回的 api_key 和 api_secret

# 4. 测试签名验证 API
API_KEY="your_api_key"
API_SECRET="your_api_secret"
PAYLOAD='{"amount":100,"currency":"USD"}'

# 计算签名（使用 HMAC-SHA256）
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$API_SECRET" | cut -d' ' -f2)

# 调用验证接口
curl -X POST http://localhost:40011/api/v1/auth/validate-signature \
  -H "Content-Type: application/json" \
  -d "{
    \"api_key\": \"$API_KEY\",
    \"signature\": \"$SIGNATURE\",
    \"payload\": \"$PAYLOAD\"
  }"

# 预期输出：
# {"valid":true,"merchant_id":"...","environment":"test"}

# 5. 启动 payment-gateway（使用新方案）
cd /home/eric/payment/backend/services/payment-gateway
DB_HOST=localhost DB_PORT=40432 DB_USER=postgres DB_PASSWORD=postgres \
DB_NAME=payment_gateway PORT=40003 \
USE_AUTH_SERVICE=true \
MERCHANT_AUTH_SERVICE_URL=http://localhost:40011 \
go run cmd/main.go &

# 6. 测试支付接口
curl -X POST http://localhost:40003/api/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -H "X-Signature: $CALCULATED_SIGNATURE" \
  -d '{
    "merchant_order_no": "ORDER-TEST-001",
    "amount": 10000,
    "currency": "USD",
    "channel": "stripe",
    "payment_method": "card",
    "subject": "Test Payment",
    "body": "Test payment description"
  }'
```

### 步骤 4：灰度发布计划

#### 第一阶段：验证（1周）
- 环境变量：`USE_AUTH_SERVICE=false`（使用旧方案）
- 在测试环境运行 merchant-auth-service
- 验证所有 API 正常工作

#### 第二阶段：灰度（1周）
- 在测试环境切换：`USE_AUTH_SERVICE=true`
- 对比两种方案的性能和准确性
- 监控日志和错误率

#### 第三阶段：全量（1周）
- 生产环境切换：`USE_AUTH_SERVICE=true`
- 监控 7 天无问题

#### 第四阶段：清理（1周后）
- 从 payment_gateway 数据库删除 api_keys 表
- 从 merchant-service 删除 APIKey 模型和代码
- 删除旧的 SignatureMiddleware

### 步骤 5：性能对比测试

```bash
# 测试旧方案（本地查询）
ab -n 1000 -c 10 -H "X-API-Key: test" -H "X-Signature: xxx" \
  http://localhost:40003/api/v1/payments

# 测试新方案（调用 merchant-auth-service）
USE_AUTH_SERVICE=true ab -n 1000 -c 10 -H "X-API-Key: test" -H "X-Signature: xxx" \
  http://localhost:40003/api/v1/payments

# 对比指标：
# - 吞吐量 (requests/sec)
# - P50/P95/P99 延迟
# - 错误率
```

## 监控指标

### merchant-auth-service
- API Key 验证 QPS
- 验证成功率
- P95 延迟
- 缓存命中率（如果添加 Redis 缓存）

### payment-gateway
- 支付请求 QPS
- 签名验证失败率
- 签名验证耗时（本地 vs 远程）

## 回滚方案

如果新方案出现问题：

```bash
# 1. 立即切换回旧方案
export USE_AUTH_SERVICE=false

# 2. 重启 payment-gateway
killall payment-gateway
./payment-gateway &

# 3. 验证服务恢复
curl http://localhost:40003/health
```

## 优化建议（Phase 1.5）

### 1. 添加 Redis 缓存
在 merchant-auth-service 中添加 API Key 缓存：

```go
// 缓存 5 分钟
func (r *apiKeyRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error) {
    // 先从 Redis 查询
    cacheKey := fmt.Sprintf("apikey:%s", apiKey)
    cached, err := r.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var key model.APIKey
        json.Unmarshal([]byte(cached), &key)
        return &key, nil
    }

    // 从数据库查询
    var key model.APIKey
    err = r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&key).Error
    if err != nil {
        return nil, err
    }

    // 写入缓存
    data, _ := json.Marshal(key)
    r.redis.Set(ctx, cacheKey, data, 5*time.Minute)

    return &key, nil
}
```

### 2. 批量验证 API
如果有批量验证需求：

```go
// POST /api/v1/auth/validate-signatures-batch
{
  "validations": [
    {"api_key": "key1", "signature": "sig1", "payload": "payload1"},
    {"api_key": "key2", "signature": "sig2", "payload": "payload2"}
  ]
}
```

### 3. gRPC 版本（可选）
如果 HTTP 性能不够：

```proto
service MerchantAuthService {
  rpc ValidateSignature(ValidateSignatureRequest) returns (ValidateSignatureResponse);
}
```

## 成功标准

- ✅ merchant-auth-service 编译成功
- ✅ payment-gateway 编译成功
- ✅ 数据迁移行数一致
- ✅ 签名验证 API 测试通过
- ✅ 支付接口测试通过
- ✅ P95 延迟 < 100ms
- ✅ 错误率 < 0.1%
- ✅ 运行 7 天无故障

## 当前状态

- [x] merchant-auth-service 代码完成
- [x] payment-gateway 客户端完成
- [ ] main.go 修改（待执行）
- [ ] 数据迁移（待执行）
- [ ] 集成测试（待执行）
- [ ] 性能测试（待执行）
- [ ] 生产发布（待执行）

---

下一步：你需要我继续实施吗？还是你想先审查一下设计？
