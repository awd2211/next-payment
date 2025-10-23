# Air 热加载开发指南

本文档介绍如何使用 Air 进行微服务的热加载开发。

## 什么是 Air？

Air 是一个用于 Go 应用的热重载工具。当你修改代码时，Air 会自动重新编译并重启应用，无需手动重启。

## 安装 Air

```bash
go install github.com/cosmtrek/air@latest
```

确保 `$GOPATH/bin` 在你的 `PATH` 环境变量中。

## 使用方法

### 方法一：使用脚本启动所有服务

```bash
# 启动所有微服务（带热加载）
./scripts/dev-with-air.sh

# 停止所有微服务
./scripts/stop-services.sh
```

### 方法二：单独启动某个服务

```bash
# 进入服务目录
cd backend/services/accounting-service

# 使用 air 启动
air

# 或者使用配置文件
air -c .air.toml
```

## Air 配置说明

每个微服务都有自己的 `.air.toml` 配置文件，位于服务根目录下。

### 主要配置项

```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/main.go"  # 编译命令
  bin = "./tmp/main"                             # 可执行文件路径
  include_ext = ["go", "tpl", "tmpl", "html"]   # 监视的文件扩展名
  exclude_dir = ["assets", "tmp", "vendor"]      # 排除的目录
  delay = 1000                                   # 延迟重启（毫秒）
```

## 已配置 Air 的服务

所有微服务都已配置 Air 支持：

1. **Accounting Service** (8005) - 账务服务
2. **Risk Service** (8006) - 风控服务
3. **Notification Service** (8007) - 通知服务
4. **Analytics Service** (8008) - 分析服务
5. **Config Service** (8009) - 配置服务

## 开发流程

### 1. 启动依赖服务

首先启动 PostgreSQL、Redis 等基础设施：

```bash
docker-compose up -d postgres redis kafka
```

### 2. 启动微服务

使用 Air 启动你要开发的服务：

```bash
# 单个服务
cd backend/services/accounting-service
air

# 或所有服务
./scripts/dev-with-air.sh
```

### 3. 修改代码

修改 Go 源代码文件，Air 会自动检测并重新编译和重启服务。

### 4. 查看日志

```bash
# 使用脚本启动的服务，日志在 logs 目录
tail -f backend/logs/accounting-service.log

# 直接启动的服务，日志在终端输出
```

## 常见问题

### Q: Air 编译失败怎么办？

A: 检查 `build-errors.log` 文件，里面有详细的编译错误信息：

```bash
cat build-errors.log
```

### Q: 端口被占用怎么办？

A: 停止正在运行的服务：

```bash
# 使用脚本停止
./scripts/stop-services.sh

# 或手动查找并杀死进程
lsof -ti:8005 | xargs kill -9
```

### Q: Air 没有检测到文件变化？

A: 确保你修改的是 `.go` 文件，并且文件不在 `exclude_dir` 列表中。

### Q: 如何修改监视的文件类型？

A: 编辑 `.air.toml` 文件，修改 `include_ext` 配置：

```toml
include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
```

## 性能优化

### 1. 减少编译时间

```toml
[build]
  # 只编译当前服务，不编译依赖
  cmd = "go build -i -o ./tmp/main ./cmd/main.go"
```

### 2. 调整延迟时间

```toml
[build]
  # 延迟 500ms 重启（默认 1000ms）
  delay = 500
```

### 3. 排除不必要的目录

```toml
[build]
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docs"]
```

## 调试支持

### 使用 Delve 调试

修改 `.air.toml` 启用调试模式：

```toml
[build]
  cmd = "go build -gcflags='all=-N -l' -o ./tmp/main ./cmd/main.go"
  full_bin = "dlv exec ./tmp/main --headless --listen=:2345 --api-version=2"
```

然后使用 IDE 连接到 `:2345` 端口进行调试。

## 生产环境注意事项

⚠️ **重要**: Air 仅用于开发环境，请勿在生产环境使用！

生产环境请使用：
```bash
go build -o main ./cmd/main.go
./main
```

或使用 Docker：
```bash
docker-compose up -d
```

## 日志管理

### 日志位置

- 脚本启动：`backend/logs/`
- 直接启动：终端输出

### 日志级别

在 `.air.toml` 中配置：

```toml
[log]
  main_only = false  # 显示所有日志
  time = true        # 显示时间戳
```

## 多服务协作开发

开发时通常只需启动相关服务：

```bash
# 开发账务服务，只需启动相关依赖
cd backend/services/accounting-service
air &

# 如果需要测试服务间调用
cd backend/services/payment-gateway
air &
```

## 总结

Air 大大提高了 Go 微服务的开发效率，主要优势：

- ✅ 自动热重载，无需手动重启
- ✅ 快速编译和重启
- ✅ 支持自定义构建命令
- ✅ 彩色日志输出
- ✅ 支持调试模式

开始使用：
```bash
./scripts/dev-with-air.sh
```

停止服务：
```bash
./scripts/stop-services.sh
```
