# Config Client SDK

配置中心客户端SDK,用于从 config-service 自动获取和更新配置。

## 特性

- ✅ 自动配置拉取和缓存
- ✅ 配置热更新(可配置刷新频率)
- ✅ 类型转换(string, int, bool)
- ✅ 更新通知hook
- ✅ 线程安全
- ✅ 容错重试

## 快速开始

### 1. 初始化客户端

```go
import "github.com/payment-platform/pkg/configclient"

// 在 main.go 或 Bootstrap 中初始化
configClient, err := configclient.NewClient(configclient.ClientConfig{
    ServiceName: "payment-gateway",
    Environment: "production",
    ConfigURL:   "http://localhost:40010",
    RefreshRate: 30 * time.Second,
})
if err != nil {
    log.Fatal(err)
}
defer configClient.Stop()
```

### 2. 获取配置

```go
// 获取字符串配置
jwtSecret := configClient.Get("JWT_SECRET")
stripeKey := configClient.Get("STRIPE_API_KEY")

// 获取整数配置
port := configClient.GetInt("PORT", 40003)
timeout := configClient.GetInt("PAYMENT_TIMEOUT", 300)

// 获取布尔配置
enableMTLS := configClient.GetBool("ENABLE_MTLS", true)

// 带默认值的获取
smtpHost := configClient.GetWithDefault("SMTP_HOST", "localhost")
```

### 3. 监听配置更新

```go
// 注册更新回调
configClient.OnUpdate(func(key, value string) {
    log.Printf("Config updated: %s = %s", key, value)

    // 根据配置key做不同处理
    switch key {
    case "PAYMENT_TIMEOUT":
        updateTimeout(value)
    case "RISK_SCORE_THRESHOLD":
        updateRiskThreshold(value)
    }
})
```

## 集成到 Bootstrap

在 `pkg/app/bootstrap.go` 中集成:

```go
// Bootstrap 添加配置客户端支持
func Bootstrap(cfg ServiceConfig) (*App, error) {
    // ... 现有代码 ...

    // 初始化配置客户端
    if cfg.EnableConfigClient {
        configClient, err := configclient.NewClient(configclient.ClientConfig{
            ServiceName: cfg.ServiceName,
            Environment: env,
            ConfigURL:   config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"),
            RefreshRate: 30 * time.Second,
        })
        if err != nil {
            return nil, fmt.Errorf("初始化配置客户端失败: %w", err)
        }

        app.ConfigClient = configClient

        // 从配置中心获取敏感配置
        if jwtSecret := configClient.Get("JWT_SECRET"); jwtSecret != "" {
            app.JWTSecret = jwtSecret
        }
    }

    return app, nil
}
```

## 配置优先级

1. 配置中心 (最高优先级)
2. 环境变量
3. 默认值 (最低优先级)

示例:
```go
// 优先使用配置中心,其次环境变量,最后默认值
func getConfig(configClient *configclient.Client, envKey, defaultValue string) string {
    if val := configClient.Get(envKey); val != "" {
        return val
    }
    return config.GetEnv(envKey, defaultValue)
}
```

## 性能优化

- ✅ 本地缓存,读取无网络开销
- ✅ 异步刷新,不阻塞业务
- ✅ HTTP客户端带连接池和重试
- ✅ 熔断器保护,避免雪崩

## 故障处理

### 配置中心不可用

- 使用本地缓存的配置继续运行
- 后台持续重试连接
- 记录告警日志

### 配置加载失败

- 初始化时加载失败会记录警告
- 后续定时刷新会自动重试
- 业务代码应处理配置缺失情况

## 最佳实践

### 1. 使用有意义的配置key

```go
// ✅ 好的命名
"STRIPE_API_KEY"
"PAYMENT_TIMEOUT"
"RISK_SCORE_THRESHOLD"

// ❌ 差的命名
"KEY1"
"TIMEOUT"
"THRESHOLD"
```

### 2. 总是提供默认值

```go
// ✅ 推荐
timeout := configClient.GetInt("PAYMENT_TIMEOUT", 300)

// ❌ 不推荐
timeout := configClient.GetInt("PAYMENT_TIMEOUT", 0)
if timeout == 0 {
    timeout = 300
}
```

### 3. 敏感配置加密存储

在 config-service 中创建配置时设置 `is_encrypted: true`:

```sql
INSERT INTO configs (service_name, config_key, config_value, is_encrypted, ...)
VALUES ('payment-gateway', 'JWT_SECRET', 'encrypted_value', true, ...);
```

### 4. 使用功能开关控制特性

```go
// 检查功能开关
if configClient.GetBool("enable_stripe_payments", false) {
    // Stripe 支付逻辑
}
```

## API 参考

### Client 方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `Get(key)` | 获取字符串配置 | `Get("JWT_SECRET")` |
| `GetWithDefault(key, default)` | 获取配置,带默认值 | `GetWithDefault("HOST", "localhost")` |
| `GetInt(key, default)` | 获取整数配置 | `GetInt("PORT", 8080)` |
| `GetBool(key, default)` | 获取布尔配置 | `GetBool("ENABLE_MTLS", true)` |
| `OnUpdate(hook)` | 注册更新回调 | `OnUpdate(func(k,v){...})` |
| `Stop()` | 停止客户端 | `defer client.Stop()` |
| `GetAllConfigs()` | 获取所有配置(调试用) | `GetAllConfigs()` |

## 故障排查

### 问题: 配置获取为空

```go
// 检查服务名和环境是否正确
configClient, _ := configclient.NewClient(configclient.ClientConfig{
    ServiceName: "payment-gateway", // 确保与数据库中一致
    Environment: "production",       // 确保环境匹配
})

// 检查配置是否存在
allConfigs := configClient.GetAllConfigs()
log.Printf("All configs: %+v", allConfigs)
```

### 问题: 配置更新不生效

- 检查 RefreshRate 设置
- 确认 config-service 可访问
- 查看客户端日志

## 示例

完整示例见: `/home/eric/payment/backend/examples/configclient_example.go`
