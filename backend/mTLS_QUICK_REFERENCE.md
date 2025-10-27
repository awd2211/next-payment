# mTLS配置快速参考

## 当前状态 ✅

### 已完成
- ✅ ServiceClient支持mTLS客户端 (service_client.go)
- ✅ 所有服务URL配置为HTTPS (start-all-services.sh)
- ✅ Kong Gateway配置mTLS客户端证书
- ✅ Kong连接到admin-bff-service成功 (172.28.0.1:40001)
- ✅ 所有19个微服务运行并启用mTLS

### 待完成
- ⏳ 解决rate limiting问题
- ⏳ 测试admin-bff → config-service mTLS调用
- ⏳ 前端ConfigManagement测试

## 关键文件

### 1. ServiceClient (mTLS客户端)
**文件**: `backend/services/admin-bff-service/internal/client/service_client.go`

```go
// 创建带mTLS的ServiceClient
client := NewServiceClient("https://localhost:40010")

// 自动从环境变量加载证书:
// - TLS_CLIENT_CERT
// - TLS_CLIENT_KEY
// - TLS_CA_FILE
```

### 2. 服务启动脚本
**文件**: `backend/scripts/start-all-services.sh`

```bash
# admin-bff-service环境变量
export CONFIG_SERVICE_URL="https://localhost:40010"
export RISK_SERVICE_URL="https://localhost:40006"
# ... 18个微服务全部HTTPS URLs
```

### 3. Kong配置
**当前配置**:
```json
{
  "name": "admin-bff-service",
  "protocol": "https",
  "host": "172.28.0.1",  // Docker网关IP
  "port": 40001,
  "client_certificate": "4f38e1b0-dcb3-424e-ad7e-8b6fba0e7982",
  "tls_verify": true
}
```

**更新命令**:
```bash
/tmp/update-kong-host.sh
```

### 4. 证书位置
```
/home/eric/payment/backend/certs/
├── ca/
│   ├── ca-cert.pem           # CA根证书
│   └── ca-key.pem
└── services/
    ├── admin-bff-service/
    │   ├── admin-bff-service.crt  # 服务端+客户端证书
    │   └── admin-bff-service.key
    ├── kong-gateway/
    │   ├── kong-gateway.crt       # Kong客户端证书
    │   └── kong-gateway.key
    └── config-service/
        ├── config-service.crt
        └── config-service.key
```

## 测试命令

### 测试Kong连接 (HTTP → Kong → admin-bff mTLS)
```bash
# 登录
curl -X POST http://localhost:40080/api/v1/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'

# 获取配置
TOKEN="<JWT token>"
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:40080/api/v1/admin/configs?page=1&page_size=3"
```

### 测试直接mTLS连接
```bash
# 健康检查
curl -sk --cert /path/to/kong-gateway.crt \
  --key /path/to/kong-gateway.key \
  --cacert /path/to/ca-cert.pem \
  https://localhost:40001/health

# 使用测试脚本
/tmp/test-direct-mtls.sh
```

### 检查服务状态
```bash
# 查看所有服务
cd /home/eric/payment/backend && ./scripts/status-all-services.sh

# 查看Kong配置
curl -s http://localhost:40081/services/admin-bff-service | jq .

# 查看Kong日志
docker logs kong-gateway --tail 50
```

## 环境变量 (每个服务)

```bash
# 服务端mTLS
ENABLE_MTLS=true
TLS_CERT_FILE=/path/to/service.crt
TLS_KEY_FILE=/path/to/service.key

# 客户端mTLS (调用其他服务时)
TLS_CLIENT_CERT=/path/to/service.crt  # 同一个证书
TLS_CLIENT_KEY=/path/to/service.key
TLS_CA_FILE=/path/to/ca-cert.pem

# 下游服务URLs (HTTPS)
CONFIG_SERVICE_URL=https://localhost:40010
PAYMENT_SERVICE_URL=https://localhost:40003
# ...
```

## 常见问题

### 1. Kong无法连接 - "Connection refused"
**原因**: Docker网络隔离
**解决**: 使用`172.28.0.1` (Docker网关IP) 而不是`host.docker.internal`

```bash
/tmp/update-kong-host.sh
```

### 2. "Invalid signature" 错误
**原因**: Kong的JWT插件与admin-bff-service冲突
**解决**: 移除Kong JWT插件

```bash
curl -X DELETE http://localhost:40081/routes/admin-bff-service-route/plugins/ea1af531-83fe-4ced-9e3f-08c0a65d352c
```

### 3. Rate limit exceeded
**原因**: admin-bff-service限制100 req/min
**临时解决**: 等待60秒或重启服务

### 4. Certificate verification failed
**原因**: CA证书未配置
**检查**:
```bash
# 验证环境变量
echo $TLS_CA_FILE
# 应该指向: /home/eric/payment/backend/certs/ca/ca-cert.pem
```

## 架构流程

```
Client (HTTP)
  ↓
Kong Gateway (HTTP → HTTPS with mTLS)
  ↓ [使用kong-gateway.crt客户端证书]
admin-bff-service:40001 (HTTPS server with mTLS)
  ↓ [使用admin-bff-service.crt客户端证书]
config-service:40010 (HTTPS server with mTLS)
  ↓
返回配置数据
```

## Docker网络

- Kong容器网络: `payment_payment-network`
- 网关IP: `172.28.0.1`
- Kong访问宿主机服务: 使用网关IP
- `/etc/hosts` in Kong: `172.28.0.1 host.docker.internal`

## 下次启动

```bash
# 1. 启动基础设施
docker-compose up -d

# 2. 启动所有微服务 (已配置mTLS)
cd /home/eric/payment/backend
./scripts/start-all-services.sh

# 3. Kong已经配置好,不需要额外操作

# 4. 测试
/tmp/test-mtls-config.sh
```
