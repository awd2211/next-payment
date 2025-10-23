# Payment Platform - 集成测试文档

## 概述

完整的端到端集成测试框架，覆盖支付平台所有核心业务流程。测试使用Go标准测试库，提供自动化验证和回归测试。

## 测试架构

```
tests/integration/
├── go.mod                          # 测试模块定义
├── testutil.go                     # 测试工具和辅助函数
├── payment_flow_test.go            # 支付流程测试（4个测试）
├── withdrawal_flow_test.go         # 提现流程测试（4个测试）
├── settlement_kyc_test.go          # 结算和KYC测试（6个测试）
└── README.md                       # 详细测试文档
```

**总计**: 14个完整的集成测试用例

## 测试覆盖范围

### 1. 支付流程 (Payment Flow) - 4个测试

| 测试用例 | 描述 | 覆盖功能 |
|---------|------|---------|
| **TestPaymentFlowComplete** | 完整支付流程 | 创建支付 → 风控 → 订单 → 查询 → 幂等性 → 取消 |
| **TestPaymentFlowWithInvalidData** | 错误处理 | 负金额验证、无效货币、缺失字段 |
| **TestPaymentRefund** | 退款流程 | 创建退款、查询退款状态 |
| **TestPaymentListQuery** | 列表查询 | 分页查询、响应格式验证 |

**涉及服务**: payment-gateway, order-service, risk-service, channel-adapter

### 2. 提现流程 (Withdrawal Flow) - 4个测试

| 测试用例 | 描述 | 审批级别 |
|---------|------|---------|
| **TestWithdrawalFlowComplete** | 小额提现（单级审批） | 5,000元 → 1级审批 |
| **TestWithdrawalMultiLevelApproval** | 大额提现（三级审批） | 100万元 → 3级审批 |
| **TestWithdrawalRejection** | 提现拒绝 | 审批拒绝流程 |
| **TestWithdrawalBatch** | 批量提现 | 批次管理 |

**审批规则**:
- < 10万元: 1级审批（经理）
- 10万-100万: 2级审批（经理 + 总监）
- >= 100万: 3级审批（经理 + 总监 + CEO）

**涉及服务**: withdrawal-service

### 3. 结算流程 (Settlement Flow) - 2个测试

| 测试用例 | 描述 | 覆盖功能 |
|---------|------|---------|
| **TestSettlementFlowComplete** | 完整结算流程 | 创建 → 审批 → 执行 → 完成 |
| **TestSettlementAutoGeneration** | 自动结算 | 定时自动生成结算单 |

**涉及服务**: settlement-service, accounting-service

### 4. KYC认证流程 (KYC Flow) - 4个测试

| 测试用例 | 描述 | 认证等级 |
|---------|------|---------|
| **TestKYCFlowComplete** | 完整KYC流程 | Basic → Intermediate |
| **TestKYCDocumentRejection** | 文档拒绝 | 审核拒绝流程 |
| **TestKYCStatistics** | 统计信息 | KYC统计数据 |
| **TestSettlementFlowComplete** | 结算流程 | 与KYC集成测试 |

**KYC等级体系**:
```
Basic (基础)
├── 交易限额: 10,000元
├── 日限额: 50,000元
└── 月限额: 1,000,000元

Intermediate (中级)
├── 交易限额: 100,000元
├── 日限额: 500,000元
└── 月限额: 10,000,000元

Advanced (高级)
├── 交易限额: 1,000,000元
├── 日限额: 5,000,000元
└── 月限额: 100,000,000元

Enterprise (企业)
├── 交易限额: 10,000,000元
├── 日限额: 50,000,000元
└── 月限额: 1,000,000,000元
```

**涉及服务**: kyc-service

## 快速开始

### 1. 准备环境

```bash
# 启动基础设施
cd /home/eric/payment
docker-compose up -d

# 启动所有服务
cd backend
./scripts/start-all.sh

# 验证服务状态
./scripts/health-check.sh
```

### 2. 运行测试

```bash
# 运行所有测试
./scripts/run-integration-tests.sh

# 运行特定测试套件
./scripts/run-integration-tests.sh -s payment      # 支付测试
./scripts/run-integration-tests.sh -s withdrawal   # 提现测试
./scripts/run-integration-tests.sh -s settlement   # 结算测试
./scripts/run-integration-tests.sh -s kyc          # KYC测试

# 运行单个测试
./scripts/run-integration-tests.sh -t TestPaymentFlowComplete

# 详细输出 + 生成报告
./scripts/run-integration-tests.sh -v -r
```

### 3. 查看结果

```bash
# 查看测试报告
cat backend/tests/integration/test-results.json

# 查看测试日志
# 每个服务的日志在 /tmp/<service-name>.log
tail -f /tmp/payment-gateway.log
```

## 测试框架特性

### 核心工具类 (testutil.go)

#### 1. HTTP客户端
```go
client := NewHTTPClient(config)

// 发送请求
resp, body, err := client.POST(url, data, headers)
resp, body, err := client.GET(url, headers)
resp, body, err := client.PUT(url, data, headers)
resp, body, err := client.DELETE(url, headers)
```

#### 2. API签名
```go
// 生成HMAC-SHA256签名
signature := SignRequest(secret, merchantID, nonce, timestamp, body)

// 设置请求头
headers := map[string]string{
    "X-Merchant-ID": merchantID,
    "X-Nonce":       nonce,
    "X-Timestamp":   timestamp,
    "X-Signature":   signature,
}
```

#### 3. 测试数据生成
```go
// 生成测试商户
merchantData := GenerateTestMerchant()

// 生成测试支付
paymentData := GenerateTestPayment(merchantID, amount)
```

#### 4. 断言函数
```go
AssertEqual(t, expected, actual, "message")
AssertNotEmpty(t, value, "field name")
AssertStatusCode(t, 200, resp.StatusCode, "operation")
AssertJSONField(t, data, "status", "success")
```

#### 5. 服务等待
```go
// 等待服务就绪（最多重试3次）
if err := WaitForService(url, 3); err != nil {
    t.Skipf("Service not ready: %v", err)
}
```

### 测试策略

#### 自动跳过
如果服务未运行，测试会自动跳过而不是失败：
```go
if err := WaitForService(config.PaymentGatewayURL, 3); err != nil {
    t.Skipf("Payment gateway not ready: %v", err)
}
```

#### 测试隔离
- 每个测试使用独立的UUID
- 测试数据互不干扰
- 支持并发运行

#### 幂等性保证
使用唯一的 `merchant_order_no` 确保重复请求返回相同结果

## 测试用例详解

### 示例: 完整支付流程

```go
func TestPaymentFlowComplete(t *testing.T) {
    // Step 1: 创建支付
    paymentData := GenerateTestPayment(merchantID, 10000) // 100.00 USD
    resp, body, err := client.POST(
        config.PaymentGatewayURL+"/api/v1/payments",
        paymentData,
        headers,
    )
    AssertStatusCode(t, 200, resp.StatusCode, "Create payment")

    // Step 2: 验证支付创建成功
    result := ParseJSONResponse(t, body)
    paymentNo := result["payment_no"].(string)
    AssertNotEmpty(t, paymentNo, "payment_no")

    // Step 3: 查询支付状态
    resp, body, err = client.GET(
        fmt.Sprintf("%s/api/v1/payments/%s", config.PaymentGatewayURL, paymentNo),
        headers,
    )
    AssertStatusCode(t, 200, resp.StatusCode, "Query payment")

    // Step 4: 测试幂等性（重复请求应返回相同结果）
    resp, _, err = client.POST(
        config.PaymentGatewayURL+"/api/v1/payments",
        paymentData, // 相同的 merchant_order_no
        headers,
    )
    // 应该返回 200 (已存在) 或 409 (冲突)
}
```

### 示例: 多级审批

```go
func TestWithdrawalMultiLevelApproval(t *testing.T) {
    // 创建大额提现（100万元，需要3级审批）
    withdrawalData := map[string]interface{}{
        "amount": 100000000, // 100万元
    }

    // Level 1: 经理审批
    approvalData := map[string]interface{}{
        "level": 1,
        "action": "approve",
    }
    client.POST(url+"/approve", approvalData, nil)

    // Level 2: 总监审批
    approvalData["level"] = 2
    client.POST(url+"/approve", approvalData, nil)

    // 验证状态: 应该还是 pending（等待第3级）
    AssertEqual(t, "pending", status, "Status after 2 levels")

    // Level 3: CEO审批
    approvalData["level"] = 3
    client.POST(url+"/approve", approvalData, nil)

    // 最终状态: approved 或 processing
    AssertEqual(t, "approved", finalStatus, "Final status")
}
```

## CI/CD集成

### GitHub Actions 配置

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 40432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7
        ports:
          - 40379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Initialize databases
        run: |
          ./scripts/init-db.sh

      - name: Start services
        run: |
          ./scripts/start-all.sh
          sleep 15

      - name: Run integration tests
        run: ./scripts/run-integration-tests.sh -v -r

      - name: Upload test results
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-results
          path: backend/tests/integration/test-results.json

      - name: Cleanup
        if: always()
        run: ./scripts/stop-all.sh
```

## 性能基准

### 预期响应时间

| 操作 | 目标时间 | 可接受时间 |
|------|---------|-----------|
| 创建支付 | < 300ms | < 500ms |
| 查询支付 | < 50ms | < 100ms |
| 创建提现 | < 200ms | < 300ms |
| 审批操作 | < 100ms | < 200ms |
| 创建结算 | < 300ms | < 500ms |
| KYC文档提交 | < 200ms | < 300ms |

### 压力测试

```bash
# 使用hey进行压力测试
hey -n 1000 -c 10 -m POST \
    -H "Content-Type: application/json" \
    -H "X-Merchant-ID: xxx" \
    -H "X-Signature: xxx" \
    -d '{"amount":10000,"currency":"USD",...}' \
    http://localhost:8003/api/v1/payments
```

## 故障排查

### 常见问题

#### 1. 服务未启动
**错误**: `Service not ready after 3 retries`

**解决**:
```bash
# 检查服务状态
./scripts/health-check.sh

# 启动缺失的服务
./scripts/start-all.sh
```

#### 2. 数据库连接失败
**错误**: `connection refused`

**解决**:
```bash
# 检查PostgreSQL
docker ps | grep postgres

# 检查连接
pg_isready -h localhost -p 40432 -U postgres
```

#### 3. 测试超时
**错误**: `test timed out after 30s`

**解决**:
- 检查服务响应时间
- 增加超时配置
- 查看服务日志: `tail -f /tmp/<service>.log`

#### 4. 签名验证失败
**错误**: `invalid signature`

**解决**:
- 确认 `SignatureSecret` 配置正确
- 检查签名算法实现
- 验证时间戳是否在有效范围内

## 扩展测试

### 添加新测试

1. 在 `tests/integration/` 创建新文件：
```go
// my_feature_test.go
package integration

import "testing"

func TestMyFeature(t *testing.T) {
    config := DefaultTestConfig()
    client := NewHTTPClient(config)

    // 测试逻辑
}
```

2. 运行测试：
```bash
go test -v -run TestMyFeature
```

### 添加测试套件

在 `run-integration-tests.sh` 添加：
```bash
my-feature)
    echo -e "${CYAN}Running My Feature Tests${NC}"
    run_tests "TestMyFeature" false
    ;;
```

## 测试覆盖率

### 生成覆盖率报告

```bash
# 运行测试并生成覆盖率
go test -coverprofile=coverage.out

# 查看覆盖率
go tool cover -func=coverage.out

# 生成HTML报告
go tool cover -html=coverage.out -o coverage.html
```

## 最佳实践

### 1. 测试命名
- 使用描述性名称: `TestPaymentFlowComplete`
- 遵循约定: `Test[Feature][Scenario]`
- 清晰表达测试意图

### 2. 测试结构
```go
func TestFeature(t *testing.T) {
    // Arrange: 准备测试数据
    config := DefaultTestConfig()

    // Act: 执行操作
    resp, body, err := client.POST(...)

    // Assert: 验证结果
    AssertStatusCode(t, 200, resp.StatusCode, "operation")
}
```

### 3. 错误处理
- 测试正常和异常场景
- 验证错误响应格式
- 提供清晰的错误消息

### 4. 测试独立性
- 每个测试应该独立运行
- 使用UUID避免数据冲突
- 不依赖其他测试的执行顺序

## 统计信息

- **总测试数**: 14个
- **测试文件**: 3个
- **测试框架代码**: ~400行
- **测试代码**: ~1200行
- **覆盖服务**: 6个核心服务
- **测试二进制大小**: 10MB

## 下一步计划

- [ ] 添加性能测试
- [ ] 添加并发测试
- [ ] 添加安全测试（SQL注入、XSS）
- [ ] 集成Prometheus监控
- [ ] 添加混沌工程测试
- [ ] 实现测试数据管理
- [ ] 添加快照测试
- [ ] 集成Allure测试报告

## 参考资料

- [测试详细文档](backend/tests/integration/README.md)
- [脚本使用指南](backend/scripts/README.md)
- [环境配置](ENVIRONMENT_SETUP.md)
- [项目状态](PROJECT_STATUS.md)
