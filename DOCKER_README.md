# 🐳 Docker 部署快速指南

## 🚀 一键启动

```bash
# 快速部署整个系统（推荐）
cd /home/eric/payment
./scripts/deploy-all.sh
```

这将自动完成所有步骤：检查系统、生成证书、启动服务、健康检查。

---

## 📋 前置要求

- **Docker**: 24.0+
- **Docker Compose**: 2.20+
- **系统资源**: CPU 4核+, 内存 8GB+, 磁盘 50GB+

验证：
```bash
docker --version
docker-compose --version
docker info
```

---

## 🏗️ 系统架构

```
19 个微服务 + 2 个 BFF + 基础设施（PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger）

内网域名: <service>.payment-network
mTLS 启用: HTTPS + 双向认证
```

### 服务端口

| 服务 | 端口 | 访问 |
|------|------|------|
| Admin BFF | 40001 | http://localhost:40001/swagger/index.html |
| Merchant BFF | 40023 | http://localhost:40023/swagger/index.html |
| Payment Gateway | 40003 | http://localhost:40003/health |
| Order Service | 40004 | http://localhost:40004/health |
| Prometheus | 40090 | http://localhost:40090 |
| Grafana | 40300 | http://localhost:40300 (admin/admin) |
| Jaeger | 50686 | http://localhost:50686 |

完整端口列表请查看 [DOCKER_DEPLOYMENT_GUIDE.md](DOCKER_DEPLOYMENT_GUIDE.md#服务端口映射)

---

## 📦 手动部署（分步）

### 1. 生成 Dockerfile（如果需要）

```bash
cd backend
./scripts/generate-dockerfiles.sh
```

### 2. 生成 mTLS 证书

```bash
cd backend/certs

# 生成 CA 证书
./generate-ca-cert.sh

# 为所有服务生成证书
for service in payment-gateway order-service channel-adapter risk-service \
               accounting-service notification-service analytics-service \
               config-service merchant-auth-service settlement-service \
               withdrawal-service kyc-service cashier-service \
               reconciliation-service dispute-service merchant-policy-service \
               merchant-quota-service admin-bff-service merchant-bff-service; do
    ./generate-service-cert.sh $service
done
```

### 3. 配置环境变量

```bash
cd /home/eric/payment

# 创建 .env 文件
cat > .env << 'EOF'
DB_PASSWORD=your-password
REDIS_PASSWORD=
JWT_SECRET=your-super-secret-jwt-key-256-bits-minimum
STRIPE_API_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
EOF

chmod 600 .env
```

### 4. 启动基础设施

```bash
docker-compose up -d
```

等待约 30 秒，确保 PostgreSQL、Redis、Kafka 就绪。

### 5. 初始化数据库

```bash
cd backend
./scripts/init-db.sh
```

### 6. 构建镜像

```bash
# 方式1: 使用自动化脚本
cd backend
./scripts/build-all-docker-images.sh

# 方式2: 使用 docker-compose
cd ..
docker-compose -f docker-compose.services.yml build
docker-compose -f docker-compose.bff.yml build
```

### 7. 启动所有服务

```bash
# 启动 17 个核心服务
docker-compose -f docker-compose.services.yml up -d

# 启动 2 个 BFF 服务
docker-compose -f docker-compose.bff.yml up -d
```

### 8. 验证部署

```bash
# 使用验证脚本
./scripts/verify-deployment.sh

# 或手动检查
curl http://localhost:40003/health  # Payment Gateway
curl http://localhost:40001/health  # Admin BFF
```

---

## 🛠️ 常用命令

### 查看状态

```bash
# 所有容器
docker ps

# 特定服务
docker-compose -f docker-compose.services.yml ps
```

### 查看日志

```bash
# 实时日志
docker-compose -f docker-compose.services.yml logs -f payment-gateway

# 最后100行
docker logs --tail 100 payment-payment-gateway
```

### 重启服务

```bash
# 重启特定服务
docker-compose -f docker-compose.services.yml restart payment-gateway

# 重启所有服务
docker-compose -f docker-compose.services.yml restart
```

### 停止服务

```bash
# 使用脚本
./scripts/stop-all.sh

# 或手动停止
docker-compose -f docker-compose.bff.yml down
docker-compose -f docker-compose.services.yml down
docker-compose down
```

### 扩展服务

```bash
# 扩展到 3 个实例
docker-compose -f docker-compose.services.yml up -d --scale payment-gateway=3
```

---

## 🔍 故障排查

### 服务无法启动

```bash
# 查看日志
docker logs payment-payment-gateway

# 查看退出代码
docker inspect payment-payment-gateway --format='{{.State.ExitCode}}'
```

### 数据库连接失败

```bash
# 测试数据库
docker exec -it payment-postgres psql -U postgres -c "SELECT 1"

# 检查网络
docker exec payment-payment-gateway ping postgres.payment-network
```

### mTLS 证书问题

```bash
# 验证证书
docker exec payment-payment-gateway \
  openssl x509 -in /app/certs/services/payment-gateway/payment-gateway.crt -text -noout

# 测试 HTTPS 连接
docker exec payment-payment-gateway curl -v \
  --cacert /app/certs/ca/ca-cert.pem \
  --cert /app/certs/services/payment-gateway/payment-gateway.crt \
  --key /app/certs/services/payment-gateway/payment-gateway.key \
  https://order-service.payment-network:40004/health
```

更多故障排查请查看 [DOCKER_DEPLOYMENT_GUIDE.md](DOCKER_DEPLOYMENT_GUIDE.md#故障排查)

---

## 📊 监控访问

- **Prometheus**: http://localhost:40090
- **Grafana**: http://localhost:40300 (admin/admin)
- **Jaeger**: http://localhost:50686
- **Kafka UI**: http://localhost:40084
- **Kong Admin**: http://localhost:40081

---

## 🔒 安全特性

✅ **mTLS 双向认证** - 所有服务间通信加密
✅ **非 root 用户** - 容器以普通用户运行
✅ **最小权限** - 只读证书挂载
✅ **资源限制** - CPU/内存配额
✅ **日志轮转** - 自动清理旧日志
✅ **健康检查** - 自动重启不健康容器

---

## 📚 完整文档

- **部署指南**: [DOCKER_DEPLOYMENT_GUIDE.md](DOCKER_DEPLOYMENT_GUIDE.md) - 完整的部署步骤和配置说明
- **打包总结**: [DOCKER_PACKAGE_SUMMARY.md](DOCKER_PACKAGE_SUMMARY.md) - 交付成果和关键特性
- **项目文档**: [CLAUDE.md](CLAUDE.md) - 项目架构和开发指南
- **主 README**: [README.md](README.md) - 项目总览

---

## 🎯 快速链接

### 自动化脚本

| 脚本 | 功能 | 位置 |
|------|------|------|
| `deploy-all.sh` | 一键部署 | `scripts/deploy-all.sh` |
| `stop-all.sh` | 停止所有服务 | `scripts/stop-all.sh` |
| `verify-deployment.sh` | 验证部署 | `scripts/verify-deployment.sh` |
| `build-all-docker-images.sh` | 构建所有镜像 | `backend/scripts/build-all-docker-images.sh` |
| `generate-dockerfiles.sh` | 生成 Dockerfile | `backend/scripts/generate-dockerfiles.sh` |

### 配置文件

| 文件 | 用途 |
|------|------|
| `docker-compose.yml` | 基础设施（PostgreSQL, Redis, Kafka, 监控）|
| `docker-compose.services.yml` | 17个核心微服务 |
| `docker-compose.bff.yml` | 2个BFF服务 |
| `.env` | 环境变量 |

---

## 💡 使用技巧

### 开发环境

```bash
# 仅启动基础设施（本地开发服务）
docker-compose up -d postgres redis kafka

# 查看服务日志（带颜色）
docker-compose logs -f --tail=100 payment-gateway | bat -l log
```

### 生产环境

```bash
# 使用生产配置
ENV=production docker-compose -f docker-compose.services.yml up -d

# 启用 Jaeger 低采样率（10%）
JAEGER_SAMPLING_RATE=10 docker-compose -f docker-compose.services.yml up -d
```

### 性能调优

```bash
# 增加资源限制（修改 docker-compose.services.yml）
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 1024M

# 重启生效
docker-compose -f docker-compose.services.yml up -d payment-gateway
```

---

## ❓ 常见问题

**Q: 端口冲突怎么办？**
A: 修改 `docker-compose.yml` 中的端口映射，例如 `"40003:40003"` 改为 `"50003:40003"`

**Q: 证书过期了？**
A: 重新生成证书：`cd backend/certs && ./generate-service-cert.sh <service-name>`

**Q: 服务启动慢？**
A: 检查系统资源（`docker stats`），考虑增加内存或减少并发服务数

**Q: 如何备份数据？**
A:
```bash
# 备份 PostgreSQL
docker exec payment-postgres pg_dumpall -U postgres > backup.sql

# 备份卷
docker run --rm -v payment-logs:/data -v $(pwd):/backup alpine tar czf /backup/logs-backup.tar.gz /data
```

---

## 🆘 获取帮助

如遇到问题：

1. 查看 [DOCKER_DEPLOYMENT_GUIDE.md](DOCKER_DEPLOYMENT_GUIDE.md) 的故障排查章节
2. 运行 `./scripts/verify-deployment.sh` 检查部署状态
3. 查看服务日志 `docker logs <container-name>`
4. 提交 GitHub Issue

---

**🎉 祝您使用愉快！**
