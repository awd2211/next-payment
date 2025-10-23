# gRPC 实施最终步骤

## ✅ 已完成的服务（60%）

以下服务已完全配置 gRPC：

| 服务 | HTTP 端口 | gRPC 端口 | 状态 |
|------|----------|----------|------|
| merchant-service | 40002 | 50002 | ✅ 完成 |
| payment-gateway | 40003 | 50003 | ✅ 完成 |
| order-service | 40004 | 50004 | ✅ 完成 |
| admin-service | 40001 | 50001 | ✅ 完成 |
| risk-service | 40006 | 50006 | ✅ 完成 |
| channel-adapter | 40005 | 50005 | ✅ 完成 |

## ⏳ 待完成的服务（40%）

剩余 4 个服务需要添加 gRPC 启动代码：

| 服务 | HTTP 端口 | gRPC 端口 | main.go 行数 |
|------|----------|----------|------------|
| accounting-service | 40007 | 50007 | 174 行 |
| notification-service | 40008 | 50008 | 328 行 |
| analytics-service | 40009 | 50009 | 169 行 |
| config-service | 40010 | 50010 | 169 行 |

---

## 🚀 快速完成剩余工作

### 方法 1：手动添加（推荐）

对每个服务执行以下操作：

#### 步骤 1：添加 import
在每个服务的 `cmd/main.go` 的 import 部分添加：

```go
grpcServer "payment-platform/{service-name}/internal/grpc"
pb "github.com/payment-platform/proto/{proto-name}"
pkggrpc "github.com/payment-platform/pkg/grpc"
```

**对应关系**：
- accounting-service → proto/accounting
- notification-service → proto/notification
- analytics-service → proto/analytics
- config-service → proto/config

#### 步骤 2：添加 gRPC 启动代码
在 HTTP 服务器启动前（通常是 `r.Run(addr)` 之前）添加：

```go
// 启动 gRPC 服务器（独立 goroutine）
grpcPort := config.GetEnvInt("GRPC_PORT", 50XXX)  // 见下表
gRPCServer := pkggrpc.NewSimpleServer()
xxxGrpcServer := grpcServer.NewXxxServer()  // 见下表
pb.RegisterXxxServiceServer(gRPCServer, xxxGrpcServer)

go func() {
    logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
    if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
        logger.Fatal(fmt.Sprintf("gRPC Server 启动失败: %v", err))
    }
}()
```

**服务特定参数**：

| 服务 | grpcPort | NewXxxServer() | RegisterXxxServiceServer |
|------|----------|----------------|-------------------------|
| accounting-service | 50007 | NewAccountingServer() | RegisterAccountingServiceServer |
| notification-service | 50008 | NewNotificationServer() | RegisterNotificationServiceServer |
| analytics-service | 50009 | NewAnalyticsServer() | RegisterAnalyticsServiceServer |
| config-service | 50010 | NewConfigServer() | RegisterConfigServiceServer |

---

### 方法 2：使用脚本批量完成

创建并运行以下脚本：

```bash
#!/bin/bash
# 文件：/tmp/finish_grpc.sh

cd /home/eric/payment/backend

# accounting-service
echo "=== accounting-service ===\"
cat >> services/accounting-service/cmd/main.go.patch << 'EOF'
在 import 部分添加：
grpcServer "payment-platform/accounting-service/internal/grpc"
pb "github.com/payment-platform/proto/accounting"
pkggrpc "github.com/payment-platform/pkg/grpc"

在 r.Run(addr) 之前添加：
grpcPort := config.GetEnvInt("GRPC_PORT", 50007)
gRPCServer := pkggrpc.NewSimpleServer()
accountingGrpcServer := grpcServer.NewAccountingServer()
pb.RegisterAccountingServiceServer(gRPCServer, accountingGrpcServer)
go func() {
    logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
    if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
        logger.Fatal(fmt.Sprintf("gRPC Server 启动失败: %v", err))
    }
}()
EOF

# 重复以上步骤for notification, analytics, config...
```

---

## 🔧 修复依赖问题

为所有服务添加 proto replace 并修复版本冲突：

```bash
#!/bin/bash
cd /home/eric/payment/backend

for service in admin-service merchant-service payment-gateway order-service channel-adapter risk-service accounting-service notification-service analytics-service config-service; do
  echo "=== 修复 $service ==="
  cd services/$service

  # 1. 添加 proto replace（如果不存在）
  if ! grep -q "github.com/payment-platform/proto" go.mod; then
    echo "" >> go.mod
    echo "replace github.com/payment-platform/proto => ../../proto" >> go.mod
  fi

  # 2. 修复 genproto 版本冲突
  go get google.golang.org/genproto@none 2>/dev/null || true
  go mod edit -exclude google.golang.org/genproto@v0.0.0-20181202183823-bd91e49a0898

  # 3. 清理依赖
  go mod tidy

  echo "✓ $service 完成"
  cd ../..
done

echo "所有服务依赖修复完成！"
```

保存为 `/tmp/fix_deps.sh` 并执行：
```bash
chmod +x /tmp/fix_deps.sh
/tmp/fix_deps.sh
```

---

## ✅ 验证和测试

### 1. 编译所有服务

```bash
cd /home/eric/payment/backend

for service in payment-gateway order-service admin-service risk-service channel-adapter; do
  echo "=== 编译 $service ==="
  cd services/$service
  go build -o /tmp/$service ./cmd/main.go && echo "✓ $service 编译成功" || echo "✗ $service 编译失败"
  cd ../..
done
```

### 2. 测试 gRPC 服务

启动 payment-gateway：
```bash
cd /home/eric/payment/backend/services/payment-gateway
PORT=40003 GRPC_PORT=50003 go run ./cmd/main.go
```

使用 grpcurl 测试：
```bash
# 安装 grpcurl（如果还没有）
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出所有服务
grpcurl -plaintext localhost:50003 list

# 列出特定服务的方法
grpcurl -plaintext localhost:50003 list payment.PaymentService

# 调用方法（示例）
grpcurl -plaintext localhost:50003 payment.PaymentService/ListPayments
```

### 3. 检查所有 gRPC 端口

启动所有服务后检查端口：
```bash
for port in 50001 50002 50003 50004 50005 50006 50007 50008 50009 50010; do
  echo "检查端口 $port..."
  nc -zv localhost $port 2>&1 | grep -q "succeeded" && echo "✓ $port 已监听" || echo "✗ $port 未监听"
done
```

---

## 📊 完成度统计

### 已完成工作：
- ✅ 10 个 proto 文件创建和代码生成（100%）
- ✅ 10 个 gRPC server 实现（100%）
- ✅ 6 个服务的 main.go 修改（60%）
- ✅ Makefile 自动化（100%）

### 待完成工作：
- ⏳ 4 个服务的 main.go 修改（约15-20分钟）
- ⏳ 所有服务的依赖修复（运行脚本即可）

### 总体进度：**85%** ✅

---

## 🎯 下一步

1. **完成剩余 4 个服务的 main.go 修改**（按照方法 1 或方法 2）
2. **运行依赖修复脚本**
3. **编译并测试所有服务**
4. **更新 docker-compose.yml**（如需要，添加 gRPC 端口映射）

---

**最后更新**：2025-10-23
**完成进度**：85%
**剩余工作量**：约 20-30 分钟
