# Air 热重载使用说明

本项目使用 [Air](https://github.com/cosmtrek/air) 进行 Go 微服务的热重载开发。

## 端口分配

### 后端微服务端口（40001-40010）
| 服务 | 端口 | 说明 |
|------|------|------|
| admin-service | 40001 | 管理后台服务 |
| merchant-service | 40002 | 商户服务 |
| payment-gateway | 40003 | 支付网关 |
| order-service | 40004 | 订单服务 |
| channel-adapter | 40005 | 渠道适配器 |
| risk-service | 40006 | 风控服务 |
| accounting-service | 40007 | 账务服务 |
| notification-service | 40008 | 通知服务 |
| analytics-service | 40009 | 分析服务 |
| config-service | 40010 | 配置服务 |

### 前端应用端口（40100+）
| 应用 | 端口 | API代理目标 |
|------|------|-------------|
| Admin Portal | 40101 | http://localhost:40001 |
| Merchant Portal | 40200 | http://localhost:40002 |

## 快速启动

### 1. 启动所有微服务

```bash
cd /home/eric/payment/backend
./scripts/start-all-services.sh
```

这个脚本会：
- 加载环境变量（.env文件）
- 按正确顺序启动所有服务
- 先启动config-service（其他服务可能依赖）
- 所有服务使用air进行热重载

### 2. 查看服务状态

```bash
./scripts/status-all-services.sh
```

输出示例：
```
========================================
支付平台微服务状态
========================================

accounting-service       运行中  PID: 1150739  端口: 40007
merchant-service         运行中  PID: 1150873  端口: 40002
channel-adapter          运行中  PID: 1151058  端口: 40005
...

========================================
总计: 10 个服务运行中, 0 个服务已停止
========================================
```

### 3. 停止所有服务

```bash
./scripts/stop-all-services.sh
```

这个脚本会：
- 优雅停止所有服务
- 清理air进程
- 删除临时文件（tmp目录）

## 查看日志

### 所有日志存放位置
```
/home/eric/payment/backend/logs/
```

### 查看实时日志

```bash
# 查看特定服务的实时日志
tail -f /home/eric/payment/backend/logs/admin-service.log

# 查看最近的日志
tail -50 /home/eric/payment/backend/logs/admin-service.log

# 搜索错误日志
grep ERROR /home/eric/payment/backend/logs/*.log
```

## Air 工作原理

### 热重载流程

1. **监控文件变化**：Air监控Go源文件（*.go）的变化
2. **自动编译**：检测到文件改动后自动编译
3. **重启服务**：编译成功后自动重启服务
4. **保持端口**：服务重启时保持相同的端口

### 配置文件

每个服务都有自己的`.air.toml`配置文件：
```
services/
├── admin-service/.air.toml
├── merchant-service/.air.toml
├── payment-gateway/.air.toml
...
```

### 关键配置项

```toml
[build]
  # 编译命令（包含GOWORK环境变量）
  cmd = "GOWORK=/home/eric/payment/backend/go.work go build -o ./tmp/main ./cmd/main.go"

  # 编译产物存放路径
  bin = "./tmp/main"

  # 监控的文件扩展名
  include_ext = ["go", "tpl", "tmpl", "html"]

  # 排除的目录
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
```

## 环境变量配置

环境变量存放在 `/home/eric/payment/backend/.env`：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=40432
DB_USER=postgres
DB_PASSWORD=postgres

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT配置
JWT_SECRET=your-256-bit-secret-key-change-this-in-production
JWT_EXPIRE_HOURS=24
```

启动脚本会自动加载这些环境变量。

## 开发工作流

### 1. 修改代码
在任意服务中修改Go代码，例如：
```bash
vi /home/eric/payment/backend/services/admin-service/internal/handler/admin.go
```

### 2. 自动重载
保存文件后，Air会：
- 自动检测文件变化
- 重新编译服务
- 重启服务

### 3. 查看日志
检查日志确认重启成功：
```bash
tail -f /home/eric/payment/backend/logs/admin-service.log
```

### 4. 测试API
```bash
# 健康检查
curl http://localhost:40001/health

# 测试具体API
curl -X POST http://localhost:40001/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}'
```

## 常见问题

### 1. 端口被占用

**现象**：服务启动失败，日志显示 `address already in use`

**解决**：
```bash
# 停止所有服务
./scripts/stop-all-services.sh

# 查找占用端口的进程
lsof -i :40001

# 杀掉占用进程
kill -9 <PID>

# 重新启动
./scripts/start-all-services.sh
```

### 2. 服务启动失败

**现象**：status脚本显示服务已停止

**解决**：
```bash
# 查看服务日志
tail -50 /home/eric/payment/backend/logs/<service-name>.log

# 检查数据库连接
docker ps | grep postgres

# 检查环境变量
cat /home/eric/payment/backend/.env
```

### 3. 编译错误

**现象**：Air显示编译失败

**解决**：
```bash
# 手动编译检查错误
cd /home/eric/payment/backend/services/admin-service
GOWORK=/home/eric/payment/backend/go.work go build -o ./tmp/main ./cmd/main.go

# 检查依赖
go mod tidy

# 检查go.work配置
cat /home/eric/payment/backend/go.work
```

### 4. Air进程卡住

**现象**：修改代码后不重载

**解决**：
```bash
# 重启该服务
cd /home/eric/payment/backend/services/<service-name>
pkill -f "air.*<service-name>"
nohup ~/go/bin/air -c .air.toml > /home/eric/payment/backend/logs/<service-name>.log 2>&1 &
```

## 性能优化

### 1. 调整编译延迟

修改`.air.toml`中的`delay`参数：
```toml
[build]
  delay = 1000  # 文件变化后等待1秒再编译（避免频繁编译）
```

### 2. 排除不必要的文件

在`.air.toml`中添加排除规则：
```toml
[build]
  exclude_regex = ["_test.go", ".*_gen.go"]
  exclude_dir = ["tmp", "vendor", "docs"]
```

## 访问服务

### Swagger文档

每个服务都提供Swagger API文档：

```
http://localhost:40001/swagger/index.html  # Admin Service
http://localhost:40002/swagger/index.html  # Merchant Service
http://localhost:40003/swagger/index.html  # Payment Gateway
... (其他服务类似)
```

### 健康检查

```bash
# 检查所有服务健康状态
for port in {40001..40010}; do
  echo "Port $port:"
  curl -s http://localhost:$port/health | python3 -m json.tool
done
```

### Metrics（Prometheus）

如果配置了Prometheus，可以通过以下端点获取指标：
```
http://localhost:40001/metrics
http://localhost:40002/metrics
...
```

## 生产部署

**注意**：Air是开发工具，不应在生产环境使用。

生产环境应该：
1. 编译二进制文件
2. 使用systemd或supervisord管理进程
3. 配置反向代理（Nginx）
4. 使用Docker容器化部署

## 相关资源

- [Air官方文档](https://github.com/cosmtrek/air)
- [Go Workspace文档](https://go.dev/doc/tutorial/workspaces)
- [项目架构文档](../docs/ARCHITECTURE.md)
