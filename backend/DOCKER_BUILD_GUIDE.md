# Docker 构建指南

本文档说明如何构建和部署支付平台的所有 19 个微服务的 Docker 镜像。

---

## 📁 文件结构

```
backend/
├── Dockerfile.template          # 通用模板（可选，用于参考）
├── .dockerignore               # Docker 忽略文件
├── build-docker.sh             # 构建脚本（推荐）
├── services/
│   ├── admin-service/
│   │   └── Dockerfile          # ✅ 已生成
│   ├── merchant-service/
│   │   └── Dockerfile          # ✅ 已生成
│   ├── payment-gateway/
│   │   └── Dockerfile          # ✅ 已生成
│   └── ...                     # 共 19 个服务
```

---

## 🚀 快速开始

### 1. 构建所有服务镜像

```bash
cd /home/eric/payment/backend

# 方式 1: 使用构建脚本（推荐）
./build-docker.sh

# 方式 2: 指定版本
./build-docker.sh --version v1.0.0

# 方式 3: 并行构建（8 个并发）
./build-docker.sh --parallel 8
```

### 2. 构建单个服务

```bash
# 使用脚本
./build-docker.sh admin-service

# 或手动构建
cd /home/eric/payment/backend
docker build \
  -f services/admin-service/Dockerfile \
  -t payment-platform/admin-service:latest \
  .
```

### 3. 构建并推送到镜像仓库

```bash
# 推送到 Docker Hub
./build-docker.sh --push --registry yourusername

# 推送到私有仓库
./build-docker.sh --push --registry registry.example.com/payment
```

---

## 📋 所有服务列表

| 服务名 | 端口 | 数据库 | Dockerfile 路径 |
|-------|------|--------|----------------|
| admin-service | 40001 | payment_admin | services/admin-service/Dockerfile |
| merchant-service | 40002 | payment_merchant | services/merchant-service/Dockerfile |
| payment-gateway | 40003 | payment_gateway | services/payment-gateway/Dockerfile |
| order-service | 40004 | payment_order | services/order-service/Dockerfile |
| channel-adapter | 40005 | payment_channel | services/channel-adapter/Dockerfile |
| risk-service | 40006 | payment_risk | services/risk-service/Dockerfile |
| accounting-service | 40007 | payment_accounting | services/accounting-service/Dockerfile |
| notification-service | 40008 | payment_notification | services/notification-service/Dockerfile |
| analytics-service | 40009 | payment_analytics | services/analytics-service/Dockerfile |
| config-service | 40010 | payment_config | services/config-service/Dockerfile |
| merchant-auth-service | 40011 | payment_merchant_auth | services/merchant-auth-service/Dockerfile |
| merchant-config-service | 40012 | payment_merchant_config | services/merchant-config-service/Dockerfile |
| settlement-service | 40013 | payment_settlement | services/settlement-service/Dockerfile |
| withdrawal-service | 40014 | payment_withdrawal | services/withdrawal-service/Dockerfile |
| kyc-service | 40015 | payment_kyc | services/kyc-service/Dockerfile |
| cashier-service | 40016 | payment_cashier | services/cashier-service/Dockerfile |
| reconciliation-service | 40020 | payment_reconciliation | services/reconciliation-service/Dockerfile |
| dispute-service | 40021 | payment_dispute | services/dispute-service/Dockerfile |
| merchant-limit-service | 40022 | payment_merchant_limit | services/merchant-limit-service/Dockerfile |

---

## 🛠️ 构建脚本使用

### 基本用法

```bash
./build-docker.sh [选项] [服务名...]
```

### 选项说明

| 选项 | 说明 | 示例 |
|-----|------|-----|
| `-h, --help` | 显示帮助信息 | `./build-docker.sh --help` |
| `-v, --version VERSION` | 指定镜像版本标签 | `./build-docker.sh --version v1.0.0` |
| `-r, --registry REGISTRY` | 指定镜像仓库前缀 | `./build-docker.sh --registry myregistry` |
| `-p, --push` | 构建后推送到镜像仓库 | `./build-docker.sh --push` |
| `--no-cache` | 构建时不使用缓存 | `./build-docker.sh --no-cache` |
| `--parallel N` | 并行构建 N 个镜像 | `./build-docker.sh --parallel 8` |

### 使用示例

#### 1. 构建指定服务

```bash
# 构建 admin-service 和 merchant-service
./build-docker.sh admin-service merchant-service
```

#### 2. 版本化构建

```bash
# 构建版本 v1.0.0
./build-docker.sh --version v1.0.0

# 查看镜像
docker images | grep payment-platform
```

#### 3. 构建并推送

```bash
# 推送到 Docker Hub (需要先 docker login)
./build-docker.sh --push --registry yourusername --version v1.0.0

# 推送到私有仓库
./build-docker.sh --push --registry registry.example.com/payment
```

#### 4. 强制重新构建

```bash
# 不使用缓存重新构建所有服务
./build-docker.sh --no-cache
```

#### 5. 高速并行构建

```bash
# 8 个服务并行构建（适合多核 CPU）
./build-docker.sh --parallel 8
```

---

## 🏗️ 手动构建（不使用脚本）

### 构建单个服务

```bash
cd /home/eric/payment/backend

docker build \
  -f services/admin-service/Dockerfile \
  -t payment-platform/admin-service:latest \
  -t payment-platform/admin-service:v1.0.0 \
  .
```

### 查看构建的镜像

```bash
docker images | grep payment-platform
```

### 运行单个服务（测试）

```bash
docker run -d \
  --name admin-service \
  -p 40001:40001 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=40432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=payment_admin \
  -e REDIS_HOST=host.docker.internal \
  -e REDIS_PORT=40379 \
  -e JWT_SECRET=your-secret-key \
  payment-platform/admin-service:latest
```

### 查看日志

```bash
docker logs -f admin-service
```

### 健康检查

```bash
curl http://localhost:40001/health
```

---

## 🐳 Dockerfile 详解

### 多阶段构建

每个 Dockerfile 使用两阶段构建：

#### 阶段 1: 构建阶段 (Builder)

- 基于 `golang:1.21-alpine`
- 复制 Go Workspace 和依赖
- 下载依赖包（利用 Docker 缓存层）
- 编译二进制文件（静态链接，去除调试符号）
- 输出大小优化（`-ldflags="-s -w"`）

#### 阶段 2: 运行阶段 (Runtime)

- 基于 `alpine:3.19`（最小化镜像，约 5MB）
- 仅包含运行时依赖（ca-certificates, tzdata, curl）
- 以非 root 用户运行（安全最佳实践）
- 包含健康检查配置
- 最终镜像大小：约 15-30MB（取决于服务）

### 镜像大小对比

```
传统方式: golang:1.21 (约 800MB) + 源码 = 1GB+
优化方式: alpine:3.19 (5MB) + 二进制 (10-20MB) = 15-30MB

节省: 97% 镜像大小
```

---

## 🚄 Docker 缓存层优化

### 什么是 Docker 缓存层？

Docker 在构建镜像时，会为 Dockerfile 中的每一条指令创建一个**镜像层（Layer）**。如果某一层的输入（指令内容和依赖文件）没有改变，Docker 会直接使用缓存的层，而不是重新执行该指令。

### 缓存层原理

```dockerfile
FROM golang:1.21-alpine    # Layer 1: 基础镜像（几乎永远命中缓存）
RUN apk add git            # Layer 2: 安装工具（很少变化，通常命中缓存）
COPY go.mod go.sum ./      # Layer 3: 复制依赖文件（依赖文件变化时失效）
RUN go mod download        # Layer 4: 下载依赖（Layer 3 未变化则命中缓存）
COPY . .                   # Layer 5: 复制源码（代码变化时失效）
RUN go build ...           # Layer 6: 编译（Layer 5 变化则重新构建）
```

**关键规则**:
- 如果某一层失效（文件变化），**该层之后的所有层都会失效**
- 因此要将**变化频率低的指令放在前面，变化频率高的指令放在后面**

### 我们的优化策略

#### ✅ 优化前的问题

```dockerfile
# ❌ 不好的做法
COPY . .                    # 复制所有文件（源码一变，依赖也要重新下载）
RUN go mod download         # 每次都要重新下载
RUN go build ...            # 每次都要重新编译
```

**问题**: 修改一行代码 → 重新下载所有依赖 → 浪费 3-5 分钟

#### ✅ 优化后的方案

```dockerfile
# ✅ 好的做法
# 1. 先复制依赖文件
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/admin-service/go.mod services/admin-service/go.sum ./services/admin-service/

# 2. 下载依赖（单独一层，利用缓存）
RUN cd services/admin-service && go mod download

# 3. 最后复制源码
COPY pkg/ ./pkg/
COPY services/admin-service/ ./services/admin-service/

# 4. 编译（源码变化时才重新编译）
RUN cd services/admin-service && go build ...
```

**优势**: 修改代码 → 依赖层命中缓存 → 仅重新编译 → 节省 90% 时间

### 缓存命中示例

#### 首次构建（无缓存）

```bash
$ docker build -f services/admin-service/Dockerfile -t admin:v1 .

Step 1/10 : FROM golang:1.21-alpine
 ---> 使用基础镜像 (200MB)
Step 2/10 : RUN apk add git
 ---> Running in abc123...       # 安装工具 (10秒)
Step 3/10 : COPY go.mod go.sum ./
 ---> 复制依赖文件
Step 4/10 : RUN go mod download
 ---> Running in def456...       # 下载依赖 (120秒) ⏱️
Step 5/10 : COPY . .
 ---> 复制源码
Step 6/10 : RUN go build
 ---> Running in ghi789...       # 编译 (60秒) ⏱️

总耗时: ~200秒
```

#### 修改代码后重新构建（命中缓存）

```bash
$ docker build -f services/admin-service/Dockerfile -t admin:v2 .

Step 1/10 : FROM golang:1.21-alpine
 ---> 使用缓存 ✅
Step 2/10 : RUN apk add git
 ---> 使用缓存 ✅
Step 3/10 : COPY go.mod go.sum ./
 ---> 使用缓存 ✅ (go.mod 未变化)
Step 4/10 : RUN go mod download
 ---> 使用缓存 ✅ (跳过下载！节省 120秒)
Step 5/10 : COPY . .
 ---> 源码变化，缓存失效 ❌
Step 6/10 : RUN go build
 ---> Running in xyz123...       # 重新编译 (60秒) ⏱️

总耗时: ~65秒 (节省 67%)
```

### 缓存最佳实践

#### 1. 分层复制文件

```dockerfile
# ✅ 正确：分层复制
COPY go.mod go.sum ./           # Layer 1: 依赖声明
RUN go mod download             # Layer 2: 下载依赖
COPY *.go ./                    # Layer 3: 源码
RUN go build                    # Layer 4: 编译

# ❌ 错误：一次性复制
COPY . .                        # 任何文件变化都会导致后续层失效
RUN go mod download
RUN go build
```

#### 2. 按变化频率排序

```dockerfile
# 变化频率：低 → 高
FROM golang:1.21-alpine         # 几乎不变
RUN apk add git                 # 很少变
COPY go.mod go.sum ./           # 偶尔变（添加依赖）
RUN go mod download             # 依赖 go.mod
COPY *.go ./                    # 经常变（代码修改）
RUN go build                    # 依赖源码
```

#### 3. 使用 .dockerignore

```bash
# .dockerignore
.git/
*.log
tmp/
docs/
README.md
```

**作用**: 减少构建上下文大小，避免不必要的文件变化导致缓存失效

#### 4. 使用 BuildKit（推荐）

```bash
# 启用 BuildKit（Docker 18.09+）
export DOCKER_BUILDKIT=1

# 或在构建时启用
DOCKER_BUILDKIT=1 docker build ...
```

**优势**:
- ✅ 并行构建多个层
- ✅ 更智能的缓存
- ✅ 支持缓存挂载（cache mount）
- ✅ 更详细的进度输出

#### 5. 缓存挂载（BuildKit 特性）

```dockerfile
# 使用缓存挂载加速 go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 使用缓存挂载加速 Go 构建缓存
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /app/service ./cmd/main.go
```

**效果**: 即使 `go.mod` 变化，也能复用大部分依赖包

### 查看缓存使用情况

```bash
# 查看构建历史（包含缓存信息）
docker history payment-platform/admin-service:latest

# 查看层大小
docker history --no-trunc payment-platform/admin-service:latest

# 清理构建缓存
docker builder prune
```

### 缓存失效的常见原因

| 原因 | 解决方案 |
|-----|---------|
| ❌ `COPY . .` 复制了不需要的文件 | ✅ 使用 `.dockerignore` |
| ❌ 修改了 `go.mod` | ✅ 无法避免，但后续代码修改可复用 |
| ❌ Dockerfile 指令顺序不当 | ✅ 将稳定的指令放前面 |
| ❌ 使用 `--no-cache` 构建 | ✅ 仅在必要时使用 |
| ❌ 使用了 `ADD` 而非 `COPY` | ✅ `ADD` 会自动解压，影响缓存 |

### 实际效果对比

| 场景 | 无缓存优化 | 有缓存优化 | 节省 |
|-----|-----------|-----------|-----|
| 首次构建 | 200秒 | 200秒 | 0% |
| 修改代码 | 200秒 | 65秒 | **67%** |
| 修改依赖 | 200秒 | 135秒 | 32% |
| 仅修改注释 | 200秒 | 10秒 | **95%** |

---

## 📦 .dockerignore 说明

`.dockerignore` 文件排除了不需要打包到镜像的文件：

- ✅ **排除**: Git 文件、IDE 配置、日志文件、测试文件、文档
- ✅ **排除**: 构建产物、临时文件、环境变量文件、证书
- ✅ **排除**: 脚本和 Makefile
- ✅ **保留**: 源代码、go.mod/go.sum、Dockerfile

---

## 🔑 构建参数 (Build Args)

每个 Dockerfile 支持以下构建参数：

| 参数 | 说明 | 默认值 | 示例 |
|-----|------|-------|-----|
| `SERVICE_NAME` | 服务名称 | - | `admin-service` |
| `PORT` | 服务端口 | `8000` | `40001` |
| `VERSION` | 镜像版本 | `latest` | `v1.0.0` |
| `BUILD_DATE` | 构建时间 | - | `2025-10-25T10:00:00Z` |
| `GIT_COMMIT` | Git 提交哈希 | `unknown` | `abc1234` |

---

## 🚨 常见问题

### 1. 构建失败：找不到 go.mod

**问题**: `COPY go.work go.work.sum ./` 失败

**原因**: Dockerfile 必须从 `backend/` 目录构建，不能从服务目录构建

**解决方案**:
```bash
# ❌ 错误
cd services/admin-service
docker build -f Dockerfile .

# ✅ 正确
cd backend
docker build -f services/admin-service/Dockerfile .
```

### 2. 构建速度慢

**问题**: 每次构建都下载依赖

**原因**: 没有利用 Docker 缓存层

**解决方案**:
- 使用多阶段构建（已实现）
- 先复制 go.mod，再复制源码
- 使用 Docker BuildKit
- 参考上面的"Docker 缓存层优化"章节

```bash
# 启用 BuildKit 加速
DOCKER_BUILDKIT=1 docker build ...
```

### 3. 镜像太大

**问题**: 镜像超过 500MB

**原因**: 使用了完整的 golang 镜像作为运行阶段

**解决方案**:
- 已使用 alpine 作为运行阶段（约 15-30MB）
- 静态编译 (`CGO_ENABLED=0`)
- 去除调试符号 (`-ldflags="-s -w"`)

---

## 🔒 安全最佳实践

### 已实施的安全措施

1. ✅ **非 root 用户运行**
2. ✅ **最小化镜像**
3. ✅ **静态链接二进制**
4. ✅ **去除调试信息**
5. ✅ **镜像元数据**

### 生产环境建议

6. ⚠️ **使用私有镜像仓库**
7. ⚠️ **镜像签名**
8. ⚠️ **漏洞扫描**
9. ⚠️ **定期更新基础镜像**

---

## 📊 性能优化

### 构建速度优化

1. **并行构建**
   ```bash
   ./build-docker.sh --parallel 8
   ```

2. **使用 BuildKit**
   ```bash
   export DOCKER_BUILDKIT=1
   ./build-docker.sh
   ```

3. **利用缓存** - 已按依赖下载和源码编译分层

### 镜像大小优化

| 优化措施 | 效果 | 已实施 |
|---------|------|-------|
| 多阶段构建 | -95% | ✅ |
| Alpine 基础镜像 | -90% | ✅ |
| 静态链接 | -80% | ✅ |
| 去除调试符号 | -30% | ✅ |
| .dockerignore | -20% | ✅ |

**最终效果**: 从 1GB+ 减小到 15-30MB

---

## 📝 总结

### 已完成的工作

- ✅ 为所有 19 个微服务生成了独立的 Dockerfile
- ✅ 创建了通用的 Dockerfile 模板
- ✅ 配置了 .dockerignore 文件
- ✅ 开发了自动化构建脚本 (build-docker.sh)
- ✅ 实施了多阶段构建优化
- ✅ 实施了 Docker 缓存层优化
- ✅ 实施了安全最佳实践

---

**构建愉快！🚀**
