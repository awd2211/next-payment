# 微服务架构改进实施路线图

**版本**: v1.0  
**创建时间**: 2025-10-24  
**预计完成**: 2025-12-24 (2个月)

---

## 📋 快速概览

| 阶段 | 任务数 | 预计时间 | 优先级 | 状态 |
|------|-------|---------|--------|------|
| 阶段1 | 4个任务 | 4-5周 | 🔴 高 | ⏳ 待开始 |
| 阶段2 | 3个任务 | 3-4周 | 🟡 中 | ⏳ 待开始 |
| 阶段3 | 3个任务 | 长期进行 | 🟢 低 | ⏳ 待开始 |

**总体目标**: 从 4.2/5.0 → 4.8/5.0

---

## 🚀 阶段1: 核心基础设施 (4-5周)

### 任务1.1: API网关部署 (2周)

#### 📊 当前问题
- 前端直接调用15个微服务端口 (40001-40010)
- 缺少统一认证、限流、监控
- 服务端口直接暴露,安全风险高

#### 🎯 目标
部署Kong API网关,统一入口

#### 📝 实施步骤

**Week 1: 环境搭建**
```bash
# Day 1-2: 安装Kong (使用Docker Compose)
cd /home/eric/payment
mkdir -p deployments/kong

# deployments/kong/docker-compose.yml
cat > deployments/kong/docker-compose.yml << 'EOF'
version: '3.8'

services:
  kong-database:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: kong
      POSTGRES_PASSWORD: kong
      POSTGRES_DB: kong
    ports:
      - "5433:5432"
    volumes:
      - kong_data:/var/lib/postgresql/data
    networks:
      - kong-net

  kong-migration:
    image: kong:3.4
    command: kong migrations bootstrap
    environment:
      KONG_DATABASE: postgres
      KONG_PG_HOST: kong-database
      KONG_PG_USER: kong
      KONG_PG_PASSWORD: kong
    depends_on:
      - kong-database
    networks:
      - kong-net

  kong:
    image: kong:3.4
    environment:
      KONG_DATABASE: postgres
      KONG_PG_HOST: kong-database
      KONG_PG_USER: kong
      KONG_PG_PASSWORD: kong
      KONG_PROXY_ACCESS_LOG: /dev/stdout
      KONG_ADMIN_ACCESS_LOG: /dev/stdout
      KONG_PROXY_ERROR_LOG: /dev/stderr
      KONG_ADMIN_ERROR_LOG: /dev/stderr
      KONG_ADMIN_LISTEN: '0.0.0.0:8001'
      KONG_PROXY_LISTEN: '0.0.0.0:8000'
    ports:
      - "8000:8000"   # Kong Proxy (API Gateway)
      - "8001:8001"   # Kong Admin API
      - "8443:8443"   # Kong Proxy SSL
      - "8444:8444"   # Kong Admin API SSL
    depends_on:
      - kong-database
      - kong-migration
    networks:
      - kong-net
      - payment-network  # 连接到微服务网络
    extra_hosts:
      - "host.docker.internal:host-gateway"

  konga:
    image: pantsel/konga:latest
    environment:
      NODE_ENV: production
      DB_ADAPTER: postgres
      DB_HOST: kong-database
      DB_USER: kong
      DB_PASSWORD: kong
      DB_DATABASE: konga
    ports:
      - "1337:1337"  # Konga UI
    depends_on:
      - kong-database
    networks:
      - kong-net

volumes:
  kong_data:

networks:
  kong-net:
    driver: bridge
  payment-network:
    external: true  # 使用现有的payment-network
EOF

# Day 2: 启动Kong
docker-compose -f deployments/kong/docker-compose.yml up -d

# 验证
curl http://localhost:8001/status
# 访问Konga UI: http://localhost:1337
```

**Week 2: 配置路由和插件**
```bash
# Day 3-4: 配置服务和路由

# 1. 注册admin-service
curl -i -X POST http://localhost:8001/services \
  --data "name=admin-service" \
  --data "url=http://host.docker.internal:40001"

curl -i -X POST http://localhost:8001/services/admin-service/routes \
  --data "paths[]=/api/v1/admins" \
  --data "paths[]=/api/v1/roles" \
  --data "paths[]=/api/v1/permissions"

# 2. 注册merchant-service
curl -i -X POST http://localhost:8001/services \
  --data "name=merchant-service" \
  --data "url=http://host.docker.internal:40002"

curl -i -X POST http://localhost:8001/services/merchant-service/routes \
  --data "paths[]=/api/v1/merchants"

# 3. 注册payment-gateway
curl -i -X POST http://localhost:8001/services \
  --data "name=payment-gateway" \
  --data "url=http://host.docker.internal:40003"

curl -i -X POST http://localhost:8001/services/payment-gateway/routes \
  --data "paths[]=/api/v1/payments" \
  --data "paths[]=/api/v1/refunds" \
  --data "paths[]=/api/v1/webhooks"

# ... 继续注册其他服务

# Day 5-7: 配置插件

# 1. JWT认证插件 (全局)
curl -X POST http://localhost:8001/plugins \
  --data "name=jwt"

# 2. 限流插件 (每个服务)
curl -X POST http://localhost:8001/services/payment-gateway/plugins \
  --data "name=rate-limiting" \
  --data "config.second=100" \
  --data "config.minute=1000"

# 3. CORS插件
curl -X POST http://localhost:8001/plugins \
  --data "name=cors" \
  --data "config.origins=*" \
  --data "config.methods=GET,POST,PUT,DELETE,OPTIONS" \
  --data "config.headers=Accept,Authorization,Content-Type"

# 4. 日志插件 (HTTP Log)
curl -X POST http://localhost:8001/plugins \
  --data "name=http-log" \
  --data "config.http_endpoint=http://host.docker.internal:40090/api/logs"

# 5. Prometheus插件
curl -X POST http://localhost:8001/plugins \
  --data "name=prometheus"

# Day 8-10: 前端迁移

# 修改前端代理配置
# frontend/admin-portal/vite.config.ts
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:8000',  // Kong Gateway
      changeOrigin: true,
    },
  },
}
```

#### ✅ 验证标准
- [ ] Kong成功启动,访问 http://localhost:8001
- [ ] Konga UI可访问 http://localhost:1337
- [ ] 所有15个服务注册成功
- [ ] 前端通过Kong网关访问后端
- [ ] JWT认证插件生效
- [ ] 限流插件生效 (超过100req/s返回429)
- [ ] Prometheus指标导出 http://localhost:8000/metrics

---

### 任务1.2: 服务发现 (Consul) (2周)

#### 📊 当前问题
- 服务URL硬编码在环境变量
- 服务扩缩容需要手动修改配置
- 无自动故障摘除

#### 🎯 目标
部署Consul集群,实现动态服务发现

#### 📝 实施步骤

**Week 1: Consul搭建**
```bash
# Day 1-2: 部署Consul

# deployments/consul/docker-compose.yml
cat > deployments/consul/docker-compose.yml << 'EOF'
version: '3.8'

services:
  consul-server:
    image: consul:1.16
    container_name: consul-server
    command: agent -server -bootstrap-expect=1 -ui -client=0.0.0.0
    ports:
      - "8500:8500"  # HTTP API & UI
      - "8600:8600/udp"  # DNS
    environment:
      - CONSUL_BIND_INTERFACE=eth0
    volumes:
      - consul_data:/consul/data
    networks:
      - payment-network

volumes:
  consul_data:

networks:
  payment-network:
    external: true
EOF

docker-compose -f deployments/consul/docker-compose.yml up -d

# 验证Consul UI: http://localhost:8500
```

**Week 2: 服务集成**
```go
// Day 3-5: 创建Consul客户端封装
// backend/pkg/consul/client.go

package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

type ServiceDiscovery struct {
	client *api.Client
}

func NewServiceDiscovery(consulAddr string) (*ServiceDiscovery, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr
	
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	
	return &ServiceDiscovery{client: client}, nil
}

// 服务注册
func (sd *ServiceDiscovery) Register(name, id, address string, port int) error {
	registration := &api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", address, port),
			Interval: "10s",
			Timeout:  "2s",
		},
	}
	
	return sd.client.Agent().ServiceRegister(registration)
}

// 服务发现
func (sd *ServiceDiscovery) Discover(serviceName string) (string, error) {
	services, _, err := sd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}
	
	if len(services) == 0 {
		return "", fmt.Errorf("service not found: %s", serviceName)
	}
	
	// 简单轮询 (可替换为更复杂的负载均衡)
	service := services[0]
	url := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)
	return url, nil
}

// 服务注销
func (sd *ServiceDiscovery) Deregister(id string) error {
	return sd.client.Agent().ServiceDeregister(id)
}

// Day 6-10: 修改服务启动代码

// backend/pkg/app/bootstrap.go
func Bootstrap(cfg ServiceConfig) (*Application, error) {
	// ... 现有代码 ...
	
	// 注册到Consul
	if cfg.EnableConsul {
		consulAddr := config.GetEnv("CONSUL_ADDR", "localhost:8500")
		sd, err := consul.NewServiceDiscovery(consulAddr)
		if err != nil {
			return nil, err
		}
		
		serviceID := fmt.Sprintf("%s-%s", cfg.ServiceName, uuid.New().String()[:8])
		err = sd.Register(
			cfg.ServiceName,
			serviceID,
			getLocalIP(),
			cfg.Port,
		)
		if err != nil {
			return nil, err
		}
		
		// 优雅关闭时注销
		app.onShutdown = append(app.onShutdown, func() {
			sd.Deregister(serviceID)
		})
		
		app.ServiceDiscovery = sd
	}
	
	return app, nil
}

// backend/services/payment-gateway/cmd/main.go
func main() {
	application, _ := app.Bootstrap(app.ServiceConfig{
		ServiceName: "payment-gateway",
		// ...
		EnableConsul: true,  // 启用Consul
	})
	
	// 使用服务发现
	orderServiceURL, err := application.ServiceDiscovery.Discover("order-service")
	channelServiceURL, err := application.ServiceDiscovery.Discover("channel-adapter")
	riskServiceURL, err := application.ServiceDiscovery.Discover("risk-service")
	
	// 创建客户端
	orderClient := client.NewOrderClient(orderServiceURL)
	channelClient := client.NewChannelClient(channelServiceURL)
	riskClient := client.NewRiskClient(riskServiceURL)
	
	// ...
}
```

#### ✅ 验证标准
- [ ] Consul成功启动,UI可访问
- [ ] 所有服务自动注册到Consul
- [ ] 健康检查正常 (绿色状态)
- [ ] 服务间调用通过Consul发现
- [ ] 停止服务自动从Consul注销
- [ ] 故障服务自动标记为不健康

---

### 任务1.3: 日志聚合 (Loki) (1周)

#### 📊 当前问题
- 日志分散在各服务本地文件
- 跨服务问题排查困难
- 无法关联Trace ID查询日志

#### 🎯 目标
部署Grafana Loki,集中收集和查询日志

#### 📝 实施步骤

```bash
# Day 1-3: 部署Loki + Promtail

# deployments/loki/docker-compose.yml
cat > deployments/loki/docker-compose.yml << 'EOF'
version: '3.8'

services:
  loki:
    image: grafana/loki:2.9.0
    container_name: loki
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki_data:/loki
      - ./loki-config.yaml:/etc/loki/local-config.yaml
    networks:
      - payment-network

  promtail:
    image: grafana/promtail:2.9.0
    container_name: promtail
    volumes:
      - /var/log:/var/log
      - ../logs:/app/logs  # 微服务日志目录
      - ./promtail-config.yaml:/etc/promtail/config.yaml
    command: -config.file=/etc/promtail/config.yaml
    depends_on:
      - loki
    networks:
      - payment-network

volumes:
  loki_data:

networks:
  payment-network:
    external: true
EOF

# loki-config.yaml
cat > deployments/loki/loki-config.yaml << 'EOF'
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    address: 127.0.0.1
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
  chunk_idle_period: 5m
  chunk_retain_period: 30s

schema_config:
  configs:
    - from: 2020-05-15
      store: boltdb
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb:
    directory: /loki/index
  filesystem:
    directory: /loki/chunks

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: false
  retention_period: 0s
EOF

# promtail-config.yaml
cat > deployments/loki/promtail-config.yaml << 'EOF'
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # 微服务日志
  - job_name: payment-services
    static_configs:
      - targets:
          - localhost
        labels:
          job: payment-services
          __path__: /app/logs/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            service: service
            trace_id: trace_id
            timestamp: ts
      - labels:
          level:
          service:
          trace_id:
EOF

# Day 4-5: 修改日志输出为JSON格式

# backend/pkg/logger/logger.go
func InitLogger() {
	config := zap.NewProductionConfig()
	config.Encoding = "json"  // JSON格式
	config.EncoderConfig.TimeKey = "ts"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "msg"
	
	config.InitialFields = map[string]interface{}{
		"service": os.Getenv("SERVICE_NAME"),
	}
	
	logger, _ := config.Build()
	zap.ReplaceGlobals(logger)
}

# Day 6-7: 配置Grafana数据源

# 访问 Grafana: http://localhost:40300
# 添加Loki数据源: http://loki:3100
# 创建日志面板,按Trace ID关联
```

#### ✅ 验证标准
- [ ] Loki成功启动
- [ ] Promtail收集日志
- [ ] Grafana可查询日志
- [ ] 按Trace ID关联日志
- [ ] 按服务名筛选日志
- [ ] 日志保留7天

---

### 任务1.4: CI/CD流程 (GitHub Actions) (1周)

#### 📝 实施步骤

```yaml
# Day 1-3: 编写GitHub Actions工作流

# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  # 任务1: 代码质量检查
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run golangci-lint
        run: |
          cd backend
          make lint
  
  # 任务2: 运行测试
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          cd backend
          make test
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          REDIS_HOST: localhost
          REDIS_PORT: 6379
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out
  
  # 任务3: 构建Docker镜像
  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service:
          - payment-gateway
          - order-service
          - merchant-service
          # ... 其他服务
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: ./backend/services/${{ matrix.service }}
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: yourorg/payment-${{ matrix.service }}:${{ github.sha }}
  
  # 任务4: 部署到测试环境 (仅develop分支)
  deploy-staging:
    needs: build
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to staging
        run: |
          # 这里可以使用kubectl或SSH部署
          echo "Deploying to staging..."
  
  # 任务5: 部署到生产环境 (仅main分支,需要手动审批)
  deploy-production:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://payment.example.com
    steps:
      - name: Deploy to production
        run: |
          echo "Deploying to production..."

# Day 4-7: 配置Makefile命令

# backend/Makefile
.PHONY: lint test coverage

lint:
	golangci-lint run ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

coverage:
	go tool cover -html=coverage.out -o coverage.html
```

#### ✅ 验证标准
- [ ] Push代码自动触发CI
- [ ] 所有测试通过
- [ ] 代码覆盖率上传到Codecov
- [ ] Docker镜像成功构建
- [ ] develop分支自动部署到测试环境
- [ ] main分支需要审批才能部署生产

---

## 🎯 阶段2: 功能完善 (3-4周)

### 任务2.1: Kubernetes部署配置 (3周)
### 任务2.2: 配置中心完全迁移 (1周)
### 任务2.3: 提升测试覆盖率 (持续进行)

*(详细步骤见完整文档)*

---

## 📊 进度跟踪

### Week 1-2: API网关
- [ ] Day 1-2: Kong环境搭建
- [ ] Day 3-7: 服务注册和路由配置
- [ ] Day 8-10: 前端迁移

### Week 3-4: 服务发现
- [ ] Day 1-2: Consul部署
- [ ] Day 3-10: 服务集成

### Week 5: 日志聚合
- [ ] Day 1-3: Loki + Promtail部署
- [ ] Day 4-5: JSON日志格式
- [ ] Day 6-7: Grafana配置

### Week 6: CI/CD
- [ ] Day 1-3: GitHub Actions配置
- [ ] Day 4-7: Makefile和测试优化

---

## 📞 需要帮助?

每完成一个任务,建议:
1. ✅ 验证所有检查项
2. 📸 截图关键配置
3. 📝 记录遇到的问题
4. 🔄 Code Review

**下次审查**: 建议2个月后评估进度

---

**创建人**: AI架构师  
**文档版本**: v1.0

