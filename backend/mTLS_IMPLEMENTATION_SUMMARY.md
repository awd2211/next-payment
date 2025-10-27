# mTLS实现总结

## 已完成的工作

### 1. ServiceClient mTLS支持

**文件**: `backend/services/admin-bff-service/internal/client/service_client.go`

**关键变更**:
- 完全重写ServiceClient,使用原生`http.Client`而非httpclient包装器
- 添加`createMTLSConfig()`函数,从环境变量加载客户端证书
- 当`ENABLE_MTLS=true`时,自动配置TLS客户端

**mTLS配置**:
```go
// 读取环境变量
TLS_CLIENT_CERT=/path/to/client.crt
TLS_CLIENT_KEY=/path/to/client.key
TLS_CA_FILE=/path/to/ca-cert.pem

// 创建TLS配置
tls.Config{
    Certificates: []tls.Certificate{cert},  // 客户端证书
    RootCAs:      caCertPool,                // CA证书池
    MinVersion:   tls.VersionTLS12,
}
```

### 2. 环境变量配置

**文件**: `backend/scripts/start-all-services.sh`

**为admin-bff-service添加的HTTPS服务URL** (18个微服务):
```bash
CONFIG_SERVICE_URL=https://localhost:40010
RISK_SERVICE_URL=https://localhost:40006
KYC_SERVICE_URL=https://localhost:40015
MERCHANT_SERVICE_URL=https://localhost:40002
ANALYTICS_SERVICE_URL=https://localhost:40009
LIMIT_SERVICE_URL=https://localhost:40022
CHANNEL_SERVICE_URL=https://localhost:40005
CASHIER_SERVICE_URL=https://localhost:40016
ORDER_SERVICE_URL=https://localhost:40004
ACCOUNTING_SERVICE_URL=https://localhost:40007
DISPUTE_SERVICE_URL=https://localhost:40021
MERCHANT_AUTH_SERVICE_URL=https://localhost:40011
MERCHANT_CONFIG_SERVICE_URL=https://localhost:40012
NOTIFICATION_SERVICE_URL=https://localhost:40008
PAYMENT_SERVICE_URL=https://localhost:40003
RECONCILIATION_SERVICE_URL=https://localhost:40020
SETTLEMENT_SERVICE_URL=https://localhost:40013
WITHDRAWAL_SERVICE_URL=https://localhost:40014
```

### 3. Kong Gateway mTLS配置

**Kong客户端证书**:
- `/backend/certs/services/kong-gateway/kong-gateway.crt`
- `/backend/certs/services/kong-gateway/kong-gateway.key`

**配置脚本**: `/scripts/configure-kong-for-mtls.sh`

**Kong Service配置**:
```json
{
  "name": "admin-bff-service",
  "protocol": "https",
  "host": "host.docker.internal",
  "port": 40001,
  "client_certificate": "4f38e1b0-dcb3-424e-ad7e-8b6fba0e7982",
  "tls_verify": true,
  "ca_certificates": ["88861154-5ec4-479a-be8c-801321f63955"]
}
```

**JWT插件**: 已移除(admin-bff-service自己处理JWT认证)

### 4. 证书架构

**CA证书**: `/backend/certs/ca/ca-cert.pem`

**服务证书** (每个服务既作为服务端证书也作为客户端证书):
- `/backend/certs/services/{service-name}/{service-name}.crt`
- `/backend/certs/services/{service-name}/{service-name}.key`

**证书用途**:
- **服务端**: 服务监听HTTPS时使用的TLS证书
- **客户端**: 调用其他服务时用于mTLS认证的客户端证书

## 当前状态

### ✅ 已成功

1. **admin-bff-service已启用mTLS服务端**
   - 监听端口: `:40001` (IPv6所有接口)
   - 服务器证书: admin-bff-service.crt/key
   - 要求客户端提供证书(mTLS)

2. **ServiceClient支持mTLS客户端**
   - 代码已完成并编译成功
   - 可以从环境变量加载证书配置
   - 使用原生http.Client + TLS配置

3. **Kong Gateway mTLS客户端配置**
   - Kong有客户端证书: kong-gateway.crt/key
   - Kong服务配置指向: https://host.docker.internal:40001
   - TLS验证已启用

4. **所有19个微服务运行中**
   - 全部启用ENABLE_MTLS=true
   - 服务间通信URL配置为HTTPS

5. **直接mTLS连接测试成功**
   ```bash
   curl -sk --cert kong-gateway.crt --key kong-gateway.key \
     --cacert ca-cert.pem https://localhost:40001/health
   # 返回: {"error":"rate limit exceeded","retry_after":60}
   # ✅ mTLS握手成功,服务返回了业务错误(rate limit)
   ```

### ❌ 待解决

1. **Kong到admin-bff-service的连接问题**
   - **症状**: Kong报"Connection refused"错误
   - **原因**: Docker网络配置问题
     - Kong在Docker容器内,解析`host.docker.internal`为`172.17.0.1`
     - admin-bff-service运行在宿主机上(Air热加载)
     - Docker bridge网络无法从容器内访问宿主机的localhost服务

   - **Kong日志**:
     ```
     connect() failed (111: Connection refused) while connecting to upstream
     upstream: "https://172.17.0.1:40001/api/v1/admin/login"
     ```

2. **两个解决方案**:

   **方案A**: 使用Docker的host网络模式
   ```yaml
   # docker-compose.yml
   kong:
     network_mode: "host"
     # 注意: 端口映射将失效,需直接使用8000/8001
   ```

   **方案B**: 将admin-bff-service也运行在Docker内
   ```yaml
   # docker-compose.yml
   admin-bff-service:
     build: ./backend/services/admin-bff-service
     ports:
       - "40001:40001"
     environment:
       - ENABLE_MTLS=true
       - TLS_CERT_FILE=/certs/admin-bff-service.crt
       - TLS_KEY_FILE=/certs/admin-bff-service.key
     volumes:
       - ./backend/certs:/certs:ro
     networks:
       - payment-network
   ```

   **方案C**: 修改Kong配置使用宿主机IP
   ```bash
   # 获取宿主机在Docker网络中的IP
   HOST_IP=$(ip addr show docker0 | grep -Po 'inet \K[\d.]+')

   # 更新Kong服务
   curl -X PATCH http://localhost:40081/services/admin-bff-service \
     -d "host=$HOST_IP" \
     -d "port=40001"
   ```

## 测试命令

### 直接mTLS测试 (✅ 成功)
```bash
curl -sk --cert /path/to/kong-gateway.crt \
  --key /path/to/kong-gateway.key \
  --cacert /path/to/ca-cert.pem \
  https://localhost:40001/health
```

### 通过Kong测试 (❌ 失败 - Connection refused)
```bash
# 登录
curl -X POST http://localhost:40080/api/v1/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'

# 获取token并测试API
TOKEN="<获取的JWT token>"
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:40080/api/v1/admin/configs?page=1&page_size=3"
```

## 代码统计

- **ServiceClient重写**: 261行 (service_client.go)
- **start-all-services.sh修改**: +40行 (环境变量配置)
- **Kong配置脚本**: configure-kong-for-mtls.sh (194行)
- **生成Kong证书**: generate-kong-client-cert.sh

## 下一步

### ✅ 已完成 (2025-10-27)

1. **Kong配置已更新** - 使用Docker网关IP `172.28.0.1`
   ```bash
   # Kong Service现在配置为:
   {
     "host": "172.28.0.1",  # 之前是 host.docker.internal
     "port": 40001,
     "protocol": "https",
     "client_certificate": "4f38e1b0-dcb3-424e-ad7e-8b6fba0e7982",
     "tls_verify": true
   }
   ```

2. **Kong可以成功连接到admin-bff-service**
   - Kong日志显示: `200` 登录成功
   - Kong日志显示: `401` config API (说明已经到达admin-bff-service)
   - 不再有"Connection refused"错误

3. **移除了Kong的JWT插件** - 因为admin-bff-service自己处理JWT认证

### ⏳ 待完成

1. **解决rate limiting问题** - admin-bff-service配置了很严格的rate limit
   - 当前: 100 req/min
   - 建议: 开发环境调整为1000 req/min或临时禁用

2. **验证完整的mTLS流程**:
   - ✅ Kong → admin-bff-service (mTLS成功)
   - ⏳ admin-bff-service → config-service (mTLS) - 需要等待rate limit重置后测试
   - ⏳ config-service返回配置数据

3. **在前端测试ConfigManagement页面**

4. **将相同的mTLS配置应用到merchant-bff-service**

## 技术要点

1. **每个服务的证书既是服务端证书也是客户端证书** (同一个)
2. **CA证书用于双向验证**:
   - 服务端验证客户端证书是否由CA签发
   - 客户端验证服务端证书是否由CA签发
3. **Docker网络隔离**:
   - 容器内无法直接访问宿主机localhost
   - 需要使用`host.docker.internal`或宿主机IP
4. **TLS版本**: 最低TLS 1.2
5. **证书有效期**: 10年 (测试环境)
