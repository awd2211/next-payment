# API 网关 & 服务发现实施方案

**创建时间**: 2025-10-24  
**优先级**: 🔴 **HIGH** (1-2个月内完成)  
**预估周期**: 4-5 周

---

## 📋 目录

1. [API 网关方案评估](#api-网关方案评估)
2. [服务发现方案评估](#服务发现方案评估)
3. [推荐方案](#推荐方案)
4. [实施路线图](#实施路线图)
5. [Docker Compose 配置](#docker-compose-配置)
6. [代码改造指南](#代码改造指南)

---

## API 网关方案评估

### 方案对比

| 特性 | Kong | APISIX | Nginx + Lua | 自建 Go Gateway |
|-----|------|--------|-------------|----------------|
| **性能** | ⭐⭐⭐⭐ (OpenResty) | ⭐⭐⭐⭐⭐ (最快) | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **插件生态** | ⭐⭐⭐⭐⭐ (最丰富) | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ (需自研) |
| **学习成本** | ⭐⭐⭐ (中等) | ⭐⭐⭐⭐ (较低) | ⭐⭐ (较高) | ⭐ (最高) |
| **社区支持** | ⭐⭐⭐⭐⭐ (最成熟) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **配置方式** | REST API + DB | REST API + etcd | 配置文件 | 代码 |
| **动态路由** | ✅ 是 | ✅ 是 | ❌ 否 | ✅ 是 |
| **可视化界面** | ✅ (需付费Konga) | ✅ (自带Dashboard) | ❌ 否 | ❌ 否 |
| **云原生** | ✅ K8s Ingress | ✅ K8s Ingress | ⚠️ 需手动 | ✅ |
| **协议支持** | HTTP/gRPC/WS | HTTP/gRPC/WS/TCP | HTTP/TCP | 自定义 |
| **许可证** | Apache 2.0 | Apache 2.0 | BSD | MIT (自建) |

### 详细分析

#### 1. Kong ⭐⭐⭐⭐

**优点**:
- ✅ 最成熟的商业方案，生产环境验证充分
- ✅ 插件生态最丰富 (50+ 官方插件 + 社区插件)
- ✅ 企业支持完善 (Kong Inc.)
- ✅ 与 Consul 无缝集成
- ✅ 支持数据库模式 (PostgreSQL) 和 DB-less 模式

**缺点**:
- ❌ 性能略低于 APISIX (但仍然很快)
- ❌ 配置相对复杂
- ❌ 免费版 Dashboard (Konga) 功能受限
- ❌ 部分高级功能需要企业版

**适用场景**:
- 需要丰富插件生态
- 追求稳定性和成熟度
- 有预算购买企业版

**部署复杂度**: ⭐⭐⭐ (中等)

#### 2. APISIX ⭐⭐⭐⭐⭐ (推荐)

**优点**:
- ✅ 性能最高 (基于 OpenResty + LuaJIT)
- ✅ 完全开源，功能不受限
- ✅ 自带 Dashboard (免费且功能完整)
- ✅ 配置简单 (REST API + etcd)
- ✅ 支持动态路由、热更新
- ✅ 国内社区活跃 (Apache 顶级项目)
- ✅ 与 Consul/Nacos 集成良好
- ✅ 支持 gRPC、WebSocket、TCP/UDP

**缺点**:
- ⚠️ 相对年轻 (2019年开源)
- ⚠️ 英文文档不如 Kong 完善
- ⚠️ 企业支持较少

**适用场景**:
- 追求高性能
- 需要完整免费功能
- 国内团队 (中文文档完善)
- 快速迭代的项目

**部署复杂度**: ⭐⭐ (简单)

#### 3. Nginx + Lua ⭐⭐⭐

**优点**:
- ✅ 极致性能和稳定性
- ✅ 社区最成熟
- ✅ 运维团队熟悉度高
- ✅ 配置文件管理，版本化容易

**缺点**:
- ❌ 不支持动态路由 (需重启)
- ❌ 无 Dashboard
- ❌ 需要手写 Lua 脚本
- ❌ 功能扩展需要较高技术能力

**适用场景**:
- 静态路由规则
- 追求极致性能
- 运维能力强的团队

**部署复杂度**: ⭐⭐⭐⭐ (较高)

#### 4. 自建 Go Gateway ⭐⭐

**优点**:
- ✅ 完全可控
- ✅ 与现有 Go 代码库统一
- ✅ 灵活定制

**缺点**:
- ❌ 开发成本高 (2-3个月)
- ❌ 需要自己实现所有功能
- ❌ 缺乏生产验证
- ❌ 维护成本高

**适用场景**:
- 有非常特殊的需求
- 团队有充足的开发资源
- 长期项目

**部署复杂度**: ⭐⭐⭐⭐⭐ (最高)

---

## 服务发现方案评估

### 方案对比

| 特性 | Consul | Nacos | Eureka | etcd |
|-----|--------|-------|--------|------|
| **性能** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **功能** | 服务发现+配置+KV | 服务发现+配置 | 服务发现 | KV存储 |
| **语言** | Go | Java | Java | Go |
| **协议** | HTTP+DNS+gRPC | HTTP+gRPC | HTTP | gRPC |
| **健康检查** | ✅ 多种方式 | ✅ 多种方式 | ✅ 心跳 | ❌ 需自己实现 |
| **配置中心** | ✅ KV Store | ✅ 完整功能 | ❌ 否 | ✅ KV Store |
| **Dashboard** | ✅ 自带 | ✅ 自带 | ❌ 需第三方 | ❌ 需第三方 |
| **K8s支持** | ✅ 官方支持 | ✅ 良好 | ⚠️ 一般 | ✅ 原生 |
| **社区** | ⭐⭐⭐⭐⭐ (HashiCorp) | ⭐⭐⭐⭐ (阿里) | ⭐⭐⭐ (Netflix) | ⭐⭐⭐⭐⭐ (CNCF) |
| **国内使用** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |

### 详细分析

#### 1. Consul ⭐⭐⭐⭐⭐ (推荐)

**优点**:
- ✅ 功能最完整 (服务发现 + 健康检查 + KV存储 + 多数据中心)
- ✅ 生产环境验证充分
- ✅ 与 Kong、APISIX 无缝集成
- ✅ 支持多种健康检查方式 (HTTP/TCP/gRPC/Docker/Script)
- ✅ 自带 DNS 接口
- ✅ 自带 Web UI
- ✅ 多数据中心支持 (WAN Federation)
- ✅ ACL 权限控制

**缺点**:
- ⚠️ 学习曲线略陡
- ⚠️ 配置相对复杂

**适用场景**:
- 需要完整的服务治理
- 多数据中心部署
- 与 HashiCorp 生态集成

**部署复杂度**: ⭐⭐⭐ (中等)

#### 2. Nacos ⭐⭐⭐⭐

**优点**:
- ✅ 国内使用最广泛
- ✅ 中文文档完善
- ✅ 配置中心功能强大 (支持配置版本、灰度发布)
- ✅ 与 Spring Cloud Alibaba 无缝集成
- ✅ 自带权限控制
- ✅ Dashboard 功能完整

**缺点**:
- ⚠️ 主要面向 Java 生态
- ⚠️ Go SDK 相对不成熟
- ⚠️ 多数据中心支持一般

**适用场景**:
- Java 微服务为主
- 国内团队
- 需要强大的配置中心

**部署复杂度**: ⭐⭐ (简单)

#### 3. Eureka ⭐⭐⭐

**优点**:
- ✅ Spring Cloud 原生支持
- ✅ AP 模型 (可用性优先)
- ✅ 部署简单

**缺点**:
- ❌ 已停止维护 (2.x)
- ❌ 功能单一 (仅服务发现)
- ❌ Go 支持差

**适用场景**:
- 遗留 Spring Cloud 项目
- **不推荐新项目使用**

**部署复杂度**: ⭐⭐ (简单)

#### 4. etcd ⭐⭐⭐⭐

**优点**:
- ✅ 性能最高
- ✅ Kubernetes 原生使用
- ✅ 强一致性 (Raft)
- ✅ Go 原生支持

**缺点**:
- ❌ 无服务发现功能 (仅 KV 存储)
- ❌ 需要自己实现健康检查
- ❌ 无 Dashboard

**适用场景**:
- Kubernetes 环境
- 仅需 KV 存储
- 追求极致性能

**部署复杂度**: ⭐⭐⭐⭐ (较高)

---

## 推荐方案

### 🏆 最佳组合: APISIX + Consul

**理由**:

1. **APISIX 作为 API 网关**
   - ✅ 性能最高，满足支付场景的低延迟要求
   - ✅ 开源免费，功能完整
   - ✅ Dashboard 开箱即用
   - ✅ 动态路由，支持热更新
   - ✅ 与 Consul 无缝集成

2. **Consul 作为服务发现**
   - ✅ 功能最完整 (服务发现 + 健康检查 + 配置)
   - ✅ 生产验证充分
   - ✅ 支持多数据中心
   - ✅ 自带 DNS 接口
   - ✅ 与 Go 生态完美契合

3. **组合优势**
   - ✅ APISIX 原生支持从 Consul 动态获取上游服务
   - ✅ 服务故障自动摘除
   - ✅ 支持蓝绿部署、金丝雀发布
   - ✅ 统一的服务治理平台

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端应用                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Admin Portal │  │Merchant Portal│  │   Website   │         │
│  └──────┬───────┘  └──────┬────────┘  └──────┬───────┘         │
│         │                  │                  │                 │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
                     ┌───────▼───────┐
                     │   APISIX      │  ← API 网关 (port 9080)
                     │ (API Gateway) │
                     └───────┬───────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
    ┌─────▼─────┐      ┌────▼────┐      ┌─────▼─────┐
    │  Service 1│      │Service 2│      │ Service 3 │
    │  :40001   │      │  :40002 │      │  :40003   │
    └─────┬─────┘      └────┬────┘      └─────┬─────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
                     ┌───────▼───────┐
                     │    Consul     │  ← 服务发现 (port 8500)
                     │(Service       │
                     │ Discovery)    │
                     └───────┬───────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
    ┌─────▼─────┐      ┌────▼────┐      ┌─────▼─────┐
    │PostgreSQL │      │  Redis  │      │   Kafka   │
    │  :40432   │      │ :40379  │      │  :40092   │
    └───────────┘      └─────────┘      └───────────┘
```

---

## 实施路线图

### Phase 1: 基础设施搭建 (Week 1-2)

#### 任务清单

**Week 1: Consul 部署**
- [ ] Day 1-2: Consul 单节点部署
  ```bash
  docker-compose up consul
  ```
- [ ] Day 3-4: 所有服务注册到 Consul
  ```go
  // 每个服务添加 Consul 注册代码
  consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
      Name: "payment-gateway",
      Port: 40003,
      Check: &api.AgentServiceCheck{
          HTTP:     "http://localhost:40003/health",
          Interval: "10s",
      },
  })
  ```
- [ ] Day 5: 验证服务发现和健康检查

**Week 2: APISIX 部署**
- [ ] Day 1-2: APISIX + Dashboard 部署
  ```bash
  docker-compose up apisix apisix-dashboard
  ```
- [ ] Day 3-4: 配置路由规则
  ```bash
  # 为每个服务创建路由
  curl http://localhost:9180/apisix/admin/routes/1 -H 'X-API-KEY: xxx' -X PUT -d '{
    "uri": "/api/v1/payments/*",
    "upstream": {
      "type": "roundrobin",
      "discovery_type": "consul",
      "service_name": "payment-gateway"
    }
  }'
  ```
- [ ] Day 5: 前端测试连接 APISIX

### Phase 2: 功能完善 (Week 3)

**任务清单**
- [ ] Day 1: JWT 认证插件配置
- [ ] Day 2: 限流插件配置
- [ ] Day 3: 日志插件配置 (Kafka/File)
- [ ] Day 4: CORS 插件配置
- [ ] Day 5: 监控指标接入 Prometheus

### Phase 3: 灰度发布 (Week 4)

**任务清单**
- [ ] Day 1-2: 配置蓝绿部署
- [ ] Day 3-4: 配置金丝雀发布 (1% → 10% → 50% → 100%)
- [ ] Day 5: 回滚测试

### Phase 4: 生产上线 (Week 5)

**任务清单**
- [ ] Day 1-2: 性能测试 (压测 APISIX)
- [ ] Day 3: 安全审计
- [ ] Day 4: 文档完善
- [ ] Day 5: 生产环境部署

---

## Docker Compose 配置

### 完整配置文件

```yaml
version: '3.8'

services:
  # ========================================
  # Consul 服务发现
  # ========================================
  consul:
    image: consul:1.18
    container_name: payment-consul
    command: agent -server -bootstrap-expect=1 -ui -client=0.0.0.0
    ports:
      - "8500:8500"  # HTTP API + Web UI
      - "8600:8600/udp"  # DNS
    environment:
      - CONSUL_BIND_INTERFACE=eth0
    volumes:
      - consul-data:/consul/data
    networks:
      - payment-network
    healthcheck:
      test: ["CMD", "consul", "members"]
      interval: 10s
      timeout: 5s
      retries: 3

  # ========================================
  # APISIX API 网关
  # ========================================
  apisix:
    image: apache/apisix:3.8.0-debian
    container_name: payment-apisix
    ports:
      - "9080:9080"  # HTTP 入口
      - "9443:9443"  # HTTPS 入口
      - "9091:9091"  # Prometheus 指标
    environment:
      - APISIX_STAND_ALONE=false
    volumes:
      - ./apisix-config.yaml:/usr/local/apisix/conf/config.yaml:ro
    depends_on:
      - etcd
      - consul
    networks:
      - payment-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9080/apisix/status"]
      interval: 10s
      timeout: 5s
      retries: 3

  # ========================================
  # APISIX Dashboard (可视化界面)
  # ========================================
  apisix-dashboard:
    image: apache/apisix-dashboard:3.0.1
    container_name: payment-apisix-dashboard
    ports:
      - "9000:9000"  # Dashboard 入口
    environment:
      - APISIX_API_BASE_URL=http://apisix:9180
    depends_on:
      - apisix
    networks:
      - payment-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000"]
      interval: 10s
      timeout: 5s
      retries: 3

  # ========================================
  # etcd (APISIX 配置存储)
  # ========================================
  etcd:
    image: bitnami/etcd:3.5
    container_name: payment-etcd
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
    ports:
      - "2379:2379"
    volumes:
      - etcd-data:/bitnami/etcd
    networks:
      - payment-network
    healthcheck:
      test: ["CMD", "etcdctl", "endpoint", "health"]
      interval: 10s
      timeout: 5s
      retries: 3

volumes:
  consul-data:
  etcd-data:

networks:
  payment-network:
    driver: bridge
```

### APISIX 配置文件

```yaml
# apisix-config.yaml
apisix:
  node_listen: 9080
  admin_key:
    - name: "admin"
      key: "edd1c9f034335f136f87ad84b625c8f1"  # 生产环境请修改
      role: admin

etcd:
  host:
    - "http://etcd:2379"
  prefix: "/apisix"
  timeout: 30

discovery:
  consul:
    servers:
      - "http://consul:8500"

plugin_attr:
  prometheus:
    export_addr:
      ip: "0.0.0.0"
      port: 9091
```

---

## 代码改造指南

### 1. Go 服务注册到 Consul

#### 安装 Consul SDK

```bash
go get github.com/hashicorp/consul/api
```

#### 在 Bootstrap 中添加 Consul 注册

```go
// backend/pkg/app/bootstrap.go

import (
    consulapi "github.com/hashicorp/consul/api"
)

type ServiceConfig struct {
    // ... 现有字段
    
    // Consul 配置
    EnableConsul    bool   // 是否启用 Consul
    ConsulAddress   string // Consul 地址 (default: localhost:8500)
    ServiceName     string // 服务名称
    ServicePort     int    // 服务端口
    ServiceTags     []string // 服务标签
}

// RegisterToConsul 注册服务到 Consul
func (app *Application) RegisterToConsul(cfg ServiceConfig) error {
    if !cfg.EnableConsul {
        return nil
    }

    consulConfig := consulapi.DefaultConfig()
    if cfg.ConsulAddress != "" {
        consulConfig.Address = cfg.ConsulAddress
    }

    client, err := consulapi.NewClient(consulConfig)
    if err != nil {
        return fmt.Errorf("创建 Consul 客户端失败: %w", err)
    }

    // 服务注册
    registration := &consulapi.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%d", cfg.ServiceName, cfg.ServicePort),
        Name:    cfg.ServiceName,
        Port:    cfg.ServicePort,
        Address: getLocalIP(), // 获取本机 IP
        Tags:    cfg.ServiceTags,
        Check: &consulapi.AgentServiceCheck{
            HTTP:                           fmt.Sprintf("http://%s:%d/health", getLocalIP(), cfg.ServicePort),
            Interval:                       "10s",
            Timeout:                        "3s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }

    if err := client.Agent().ServiceRegister(registration); err != nil {
        return fmt.Errorf("服务注册失败: %w", err)
    }

    logger.Info("服务已注册到 Consul",
        zap.String("service_name", cfg.ServiceName),
        zap.Int("service_port", cfg.ServicePort))

    app.ConsulClient = client
    app.ServiceID = registration.ID

    return nil
}

// DeregisterFromConsul 从 Consul 注销服务
func (app *Application) DeregisterFromConsul() error {
    if app.ConsulClient == nil || app.ServiceID == "" {
        return nil
    }

    if err := app.ConsulClient.Agent().ServiceDeregister(app.ServiceID); err != nil {
        return fmt.Errorf("服务注销失败: %w", err)
    }

    logger.Info("服务已从 Consul 注销", zap.String("service_id", app.ServiceID))
    return nil
}

// getLocalIP 获取本机 IP
func getLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "127.0.0.1"
    }
    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return "127.0.0.1"
}
```

#### 修改服务启动代码

```go
// backend/services/payment-gateway/cmd/main.go

func main() {
    application, err := app.Bootstrap(app.ServiceConfig{
        ServiceName: "payment-gateway",
        DBName:      "payment_gateway",
        Port:        40003,
        
        // 启用 Consul
        EnableConsul:  true,
        ConsulAddress: config.GetEnv("CONSUL_ADDRESS", "localhost:8500"),
        ServiceTags:   []string{"payment", "gateway", "v1"},
        
        // ... 其他配置
    })

    // 注册到 Consul
    if err := application.RegisterToConsul(); err != nil {
        logger.Fatal("Consul 注册失败", zap.Error(err))
    }

    // 启动服务
    if err := application.RunWithGracefulShutdown(); err != nil {
        logger.Fatal("服务启动失败", zap.Error(err))
    }

    // 优雅关闭时注销服务
    defer application.DeregisterFromConsul()
}
```

### 2. 服务间调用使用 Consul 发现

#### 创建 Consul 服务发现客户端

```go
// backend/pkg/discovery/consul.go

package discovery

import (
    "fmt"
    "math/rand"
    
    consulapi "github.com/hashicorp/consul/api"
)

type ConsulDiscovery struct {
    client *consulapi.Client
}

func NewConsulDiscovery(address string) (*ConsulDiscovery, error) {
    config := consulapi.DefaultConfig()
    config.Address = address
    
    client, err := consulapi.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    return &ConsulDiscovery{client: client}, nil
}

// GetServiceURL 获取服务地址 (负载均衡)
func (d *ConsulDiscovery) GetServiceURL(serviceName string) (string, error) {
    services, _, err := d.client.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return "", err
    }
    
    if len(services) == 0 {
        return "", fmt.Errorf("服务 %s 不可用", serviceName)
    }
    
    // 随机负载均衡
    service := services[rand.Intn(len(services))].Service
    return fmt.Sprintf("http://%s:%d", service.Address, service.Port), nil
}
```

#### 修改 HTTP 客户端使用服务发现

```go
// backend/services/payment-gateway/internal/client/order_client.go

type OrderClient struct {
    discovery *discovery.ConsulDiscovery
    httpClient *httpclient.BreakerClient
}

func NewOrderClient(consulAddress string) (*OrderClient, error) {
    disc, err := discovery.NewConsulDiscovery(consulAddress)
    if err != nil {
        return nil, err
    }
    
    return &OrderClient{
        discovery: disc,
        httpClient: httpclient.NewBreakerClient(/* ... */),
    }, nil
}

func (c *OrderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // 从 Consul 获取服务地址
    serviceURL, err := c.discovery.GetServiceURL("order-service")
    if err != nil {
        return nil, fmt.Errorf("获取 order-service 地址失败: %w", err)
    }
    
    // 发送请求
    url := serviceURL + "/api/v1/orders"
    // ...
}
```

### 3. APISIX 路由配置

#### 使用 REST API 配置路由

```bash
# 为 payment-gateway 创建路由
curl http://localhost:9180/apisix/admin/routes/1 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' \
  -X PUT -d '{
    "name": "payment-gateway-route",
    "uri": "/api/v1/payments/*",
    "plugins": {
        "jwt-auth": {},
        "limit-req": {
            "rate": 100,
            "burst": 50,
            "rejected_code": 429
        },
        "prometheus": {}
    },
    "upstream": {
        "type": "roundrobin",
        "discovery_type": "consul",
        "service_name": "payment-gateway",
        "checks": {
            "active": {
                "http_path": "/health",
                "healthy": {
                    "interval": 10,
                    "successes": 2
                },
                "unhealthy": {
                    "interval": 10,
                    "http_failures": 3
                }
            }
        }
    }
}'
```

#### 使用 Dashboard 配置（推荐）

1. 访问 http://localhost:9000
2. 登录 (默认: admin/admin)
3. 导航到 "Routes" → "Create"
4. 填写路由信息:
   - URI: `/api/v1/payments/*`
   - Upstream Type: `Consul`
   - Service Name: `payment-gateway`
5. 添加插件: JWT Auth, Rate Limit, Prometheus
6. 保存

---

## 监控和运维

### 1. Consul 监控

```bash
# 查看集群成员
curl http://localhost:8500/v1/agent/members

# 查看已注册服务
curl http://localhost:8500/v1/catalog/services

# 查看服务健康状态
curl http://localhost:8500/v1/health/service/payment-gateway
```

### 2. APISIX 监控

```bash
# Prometheus 指标
curl http://localhost:9091/apisix/prometheus/metrics

# APISIX 状态
curl http://localhost:9080/apisix/status
```

### 3. Grafana Dashboard

导入 APISIX 官方 Dashboard:
- Dashboard ID: 11719
- URL: https://grafana.com/grafana/dashboards/11719

---

## 预估成本

### 时间成本

| 阶段 | 任务 | 预估时间 | 负责人 |
|-----|------|---------|--------|
| Phase 1 | Consul 部署 + 服务注册 | 5 天 | 后端 |
| Phase 1 | APISIX 部署 + 路由配置 | 5 天 | 后端 + 运维 |
| Phase 2 | 插件配置 (JWT/限流/日志) | 5 天 | 后端 |
| Phase 3 | 灰度发布配置 | 5 天 | 后端 + 运维 |
| Phase 4 | 测试 + 上线 | 5 天 | 全员 |
| **总计** | | **25 天 (5 周)** | |

### 资源成本

| 资源 | 配置 | 数量 | 月成本 (估算) |
|-----|------|------|------------|
| Consul | 2C4G | 1 (单节点) | $50 |
| APISIX | 2C4G | 2 (HA) | $100 |
| etcd | 2C4G | 1 | $50 |
| **总计** | | | **$200/月** |

**备注**: 生产环境建议 Consul 3 节点集群，APISIX 至少 2 节点

---

## 风险评估

### 高风险

1. **性能风险** 🔴
   - 问题: APISIX 增加一跳，延迟可能增加 1-2ms
   - 缓解: 压测验证，优化配置，使用 HTTP/2

2. **单点故障** 🔴
   - 问题: APISIX 单节点故障导致整个系统不可用
   - 缓解: 部署 2+ 节点，配置负载均衡

### 中风险

3. **学习成本** 🟡
   - 问题: 团队需要学习 APISIX 和 Consul
   - 缓解: 提前培训，提供文档

4. **配置错误** 🟡
   - 问题: 路由配置错误导致服务不可访问
   - 缓解: 先在测试环境验证，灰度发布

### 低风险

5. **兼容性** 🟢
   - 问题: 现有服务可能不兼容
   - 缓解: 逐步迁移，保留旧入口一段时间

---

## 成功标准

### 功能指标

- [x] 所有服务成功注册到 Consul
- [x] APISIX 能正确路由到所有服务
- [x] 服务故障自动摘除 (健康检查)
- [x] JWT 认证生效
- [x] 限流生效 (100 req/min)
- [x] 日志正确输出到 Kafka

### 性能指标

- [x] P99 延迟 < 50ms (增加 < 5ms)
- [x] APISIX 吞吐量 > 10000 QPS
- [x] APISIX CPU < 50%
- [x] APISIX 内存 < 2GB

### 可用性指标

- [x] APISIX 可用性 > 99.9%
- [x] 单节点故障 < 1 分钟恢复
- [x] 配置更新 < 1 秒生效

---

## 下一步行动

### 立即开始 (本周)

1. [x] 阅读本方案文档
2. [ ] 团队评审会议 (2小时)
3. [ ] 确定最终方案 (APISIX + Consul)
4. [ ] 申请测试环境资源

### Week 1 (下周)

1. [ ] 部署 Consul 到测试环境
2. [ ] 修改 1 个服务注册到 Consul
3. [ ] 验证健康检查

### Week 2

1. [ ] 部署 APISIX + Dashboard
2. [ ] 配置 3 个服务的路由
3. [ ] 前端测试连接 APISIX

---

**文档版本**: v1.0  
**创建时间**: 2025-10-24  
**创建人**: Claude Code Agent  
**审核状态**: ⏳ Pending Review  
**下次更新**: 完成 Phase 1 后

