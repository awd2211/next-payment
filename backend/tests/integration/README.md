# Payment Platform - Integration Tests

完整的端到端集成测试套件，覆盖所有核心业务流程。

## 测试覆盖

### 1. 支付流程测试 (payment_flow_test.go)

#### TestPaymentFlowComplete
完整的支付流程测试：
- ✅ 创建支付请求
- ✅ 风控评估
- ✅ 创建订单
- ✅ 处理支付
- ✅ 查询支付状态
- ✅ 幂等性验证
- ✅ 支付取消

#### TestPaymentFlowWithInvalidData
错误处理测试：
- ✅ 负金额验证
- ✅ 无效货币验证
- ✅ 缺失必填字段验证

#### TestPaymentRefund
退款流程测试：
- ✅ 创建退款
- ✅ 查询退款状态

#### TestPaymentListQuery
支付列表查询测试：
- ✅ 分页查询
- ✅ 响应格式验证

### 2. 提现流程测试 (withdrawal_flow_test.go)

#### TestWithdrawalFlowComplete
完整提现流程（小额，单级审批）：
- ✅ 创建银行账户
- ✅ 创建提现申请（5,000元）
- ✅ 查询提现状态
- ✅ 一级审批
- ✅ 验证状态更新

#### TestWithdrawalMultiLevelApproval
多级审批流程（大额）：
- ✅ 创建大额提现（100万元）
- ✅ 三级审批流程
  - Level 1: 经理审批
  - Level 2: 总监审批
  - Level 3: CEO审批
- ✅ 每级审批后状态验证
- ✅ 最终状态确认

#### TestWithdrawalRejection
提现拒绝测试：
- ✅ 创建提现
- ✅ 审批拒绝
- ✅ 验证拒绝状态

#### TestWithdrawalBatch
批量提现测试：
- ✅ 创建提现批次

### 3. 结算流程测试 (settlement_kyc_test.go)

#### TestSettlementFlowComplete
完整结算流程：
- ✅ 创建结算单
- ✅ 查询结算详情
- ✅ 审批结算
- ✅ 执行结算
- ✅ 验证最终状态

#### TestSettlementAutoGeneration
自动结算生成：
- ✅ 触发自动结算
- ✅ 验证生成结果

### 4. KYC认证测试 (settlement_kyc_test.go)

#### TestKYCFlowComplete
完整KYC认证流程：
- ✅ 查询初始KYC等级
- ✅ 提交身份证文件
- ✅ 提交营业执照
- ✅ 审核身份证
- ✅ 审核营业执照
- ✅ 升级KYC等级
- ✅ 验证新等级和限额

KYC等级系统：
- **Basic**: 交易限额 10,000元，日限额 50,000元
- **Intermediate**: 交易限额 100,000元，日限额 500,000元
- **Advanced**: 交易限额 1,000,000元，日限额 5,000,000元
- **Enterprise**: 交易限额 10,000,000元，日限额 50,000,000元

#### TestKYCDocumentRejection
KYC文档拒绝测试：
- ✅ 提交文档
- ✅ 审核拒绝
- ✅ 验证拒绝状态

#### TestKYCStatistics
KYC统计信息：
- ✅ 获取统计数据
- ✅ 验证字段完整性

## 测试框架 (testutil.go)

### 核心组件

#### TestConfig
测试配置管理：
```go
type TestConfig struct {
    AdminServiceURL       string
    MerchantServiceURL    string
    PaymentGatewayURL     string
    // ... 其他服务URL
    JWTSecret             string
    SignatureSecret       string
}
```

#### HTTPClient
HTTP客户端封装：
```go
client := NewHTTPClient(config)
resp, body, err := client.POST(url, data, headers)
resp, body, err := client.GET(url, headers)
resp, body, err := client.PUT(url, data, headers)
resp, body, err := client.DELETE(url, headers)
```

#### 辅助函数

**签名生成**：
```go
signature := SignRequest(secret, merchantID, nonce, timestamp, body)
```

**测试数据生成**：
```go
merchantData := GenerateTestMerchant()
paymentData := GenerateTestPayment(merchantID, amount)
```

**断言函数**：
```go
AssertEqual(t, expected, actual, "message")
AssertNotEmpty(t, value, "message")
AssertStatusCode(t, expected, actual, "message")
AssertJSONField(t, data, "field", expectedValue)
```

**服务等待**：
```go
err := WaitForService(url, maxRetries)
```

## 快速开始

### 1. 启动服务

```bash
# 启动所有服务
cd /home/eric/payment/backend
./scripts/start-all.sh

# 检查服务状态
./scripts/health-check.sh
```

### 2. 运行测试

```bash
# 运行所有集成测试
./scripts/run-integration-tests.sh

# 运行特定测试套件
./scripts/run-integration-tests.sh -s payment      # 支付测试
./scripts/run-integration-tests.sh -s withdrawal   # 提现测试
./scripts/run-integration-tests.sh -s settlement   # 结算测试
./scripts/run-integration-tests.sh -s kyc          # KYC测试

# 运行特定测试
./scripts/run-integration-tests.sh -t TestPaymentFlowComplete

# 详细输出 + 生成报告
./scripts/run-integration-tests.sh -v -r

# 仅检查服务状态
./scripts/run-integration-tests.sh -c
```

### 3. 手动运行（开发模式）

```bash
cd tests/integration

# 初始化依赖
go mod tidy

# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestPaymentFlowComplete

# 运行特定测试套件
go test -v -run TestPayment     # 所有支付相关测试
go test -v -run TestWithdrawal  # 所有提现相关测试
go test -v -run TestSettlement  # 所有结算相关测试
go test -v -run TestKYC         # 所有KYC相关测试
```

## 测试策略

### 前置条件

1. **基础设施必须运行**：
   - PostgreSQL (localhost:40432)
   - Redis (localhost:40379)
   - Kafka (localhost:40092)

2. **必需服务必须启动**：
   - payment-gateway (8003)
   - order-service (8004)
   - risk-service (8006)
   - settlement-service (8012)
   - withdrawal-service (8013)
   - kyc-service (8014)

3. **数据库已初始化**：
   ```bash
   ./scripts/migrate.sh
   ```

### 测试隔离

- 每个测试使用独立的UUID标识
- 测试数据互不干扰
- 测试可以并发运行

### 测试跳过

如果服务未运行，测试会自动跳过：
```go
if err := WaitForService(config.PaymentGatewayURL, 3); err != nil {
    t.Skipf("Payment gateway not ready: %v", err)
}
```

## CI/CD集成

### GitHub Actions示例

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

      redis:
        image: redis:7
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start services
        run: |
          ./scripts/start-all.sh
          sleep 10

      - name: Run tests
        run: ./scripts/run-integration-tests.sh -v -r

      - name: Upload test report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-results
          path: tests/integration/test-results.json
```

## 测试报告

### JSON格式报告

测试报告保存在 `tests/integration/test-results.json`：

```json
{
  "Time": "2024-01-01T10:00:00Z",
  "Action": "pass",
  "Package": "payment-platform/tests/integration",
  "Test": "TestPaymentFlowComplete",
  "Elapsed": 2.5
}
```

### 生成HTML报告

```bash
# 安装go-junit-report
go install github.com/jstemmer/go-junit-report@latest

# 生成JUnit格式报告
go test -v 2>&1 | go-junit-report > test-report.xml

# 使用工具转换为HTML
# 或集成到CI/CD查看
```

## 性能基准

### 预期响应时间

| 操作 | 预期时间 | 说明 |
|------|----------|------|
| 创建支付 | < 500ms | 包含风控和订单创建 |
| 查询支付 | < 100ms | 数据库查询 |
| 创建提现 | < 300ms | 包含费用计算 |
| 审批操作 | < 200ms | 状态更新 |
| 创建结算 | < 500ms | 包含数据聚合 |
| KYC文档提交 | < 300ms | 文件上传 |

### 性能测试

```bash
# 使用-benchmem标志运行基准测试
go test -bench=. -benchmem

# 测试并发性能
go test -run=XXX -bench=. -benchtime=10s -cpu=1,2,4,8
```

## 常见问题

### 1. 测试失败：服务未运行

**问题**：`Service not ready after 3 retries`

**解决**：
```bash
./scripts/start-all.sh
./scripts/health-check.sh
```

### 2. 测试失败：连接拒绝

**问题**：`connection refused`

**解决**：检查端口是否正确，防火墙设置

### 3. 测试超时

**问题**：测试运行时间过长

**解决**：
- 检查服务响应时间
- 增加超时配置
- 检查数据库性能

### 4. 幂等性测试失败

**问题**：重复请求返回不同结果

**解决**：
- 检查Redis缓存是否正常
- 验证幂等性key生成逻辑
- 检查merchant_order_no唯一性

## 扩展测试

### 添加新测试

1. 创建测试文件：
   ```bash
   touch tests/integration/my_new_test.go
   ```

2. 编写测试：
   ```go
   package integration

   import "testing"

   func TestMyNewFeature(t *testing.T) {
       config := DefaultTestConfig()
       client := NewHTTPClient(config)

       // 测试逻辑
   }
   ```

3. 运行测试：
   ```bash
   go test -v -run TestMyNewFeature
   ```

### 添加测试套件

在 `run-integration-tests.sh` 中添加新套件：

```bash
myfeature)
    echo -e "${CYAN}Running My Feature Tests${NC}"
    run_tests "TestMyFeature" false
    ;;
```

## 最佳实践

### 1. 测试独立性
- 每个测试应该独立运行
- 不依赖其他测试的执行顺序
- 使用唯一标识符（UUID）避免数据冲突

### 2. 清晰的测试名称
- 使用描述性测试名称
- 遵循命名约定：`Test[Feature][Scenario]`
- 例如：`TestPaymentFlowComplete`, `TestWithdrawalRejection`

### 3. 合理的断言
- 使用辅助函数进行断言
- 提供清晰的错误消息
- 检查所有关键字段

### 4. 错误处理
- 测试正常流程和异常流程
- 验证错误响应格式
- 确保错误消息有意义

### 5. 测试文档
- 在测试函数上方添加注释
- 说明测试目的和步骤
- 记录已知问题和限制

## 下一步

- [ ] 添加压力测试
- [ ] 添加并发测试
- [ ] 添加安全测试（注入、XSS等）
- [ ] 集成性能监控
- [ ] 添加混沌工程测试
