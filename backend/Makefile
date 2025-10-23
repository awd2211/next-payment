.PHONY: proto clean test build run-all docker-build

# 生成所有Proto文件
proto:
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/admin/*.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/merchant/*.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/payment/*.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/order/*.proto
	@echo "Proto generation completed!"

# 清理生成的文件
clean:
	@echo "Cleaning generated files..."
	@find . -name "*.pb.go" -type f -delete
	@find . -name "*_grpc.pb.go" -type f -delete
	@echo "Clean completed!"

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 构建所有服务
build:
	@echo "Building all services..."
	@cd services/admin-service && go build -o ../../bin/admin-service ./cmd
	@cd services/merchant-service && go build -o ../../bin/merchant-service ./cmd
	@cd services/payment-gateway && go build -o ../../bin/payment-gateway ./cmd
	@cd services/order-service && go build -o ../../bin/order-service ./cmd
	@cd services/channel-adapter && go build -o ../../bin/channel-adapter ./cmd
	@echo "Build completed!"

# 运行所有服务（开发环境）
run-all:
	@echo "Starting all services..."
	@make -j5 run-admin run-merchant run-payment run-order run-channel

run-admin:
	@cd services/admin-service && go run ./cmd

run-merchant:
	@cd services/merchant-service && go run ./cmd

run-payment:
	@cd services/payment-gateway && go run ./cmd

run-order:
	@cd services/order-service && go run ./cmd

run-channel:
	@cd services/channel-adapter && go run ./cmd

# Docker构建
docker-build:
	@echo "Building Docker images..."
	@docker build -t payment-platform/admin-service:latest -f services/admin-service/Dockerfile .
	@docker build -t payment-platform/merchant-service:latest -f services/merchant-service/Dockerfile .
	@docker build -t payment-platform/payment-gateway:latest -f services/payment-gateway/Dockerfile .
	@docker build -t payment-platform/order-service:latest -f services/order-service/Dockerfile .
	@docker build -t payment-platform/channel-adapter:latest -f services/channel-adapter/Dockerfile .
	@echo "Docker build completed!"

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed!"

# 初始化数据库
init-db:
	@echo "Initializing database..."
	@./scripts/init-db.sh
	@echo "Database initialization completed!"

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format completed!"

# 代码检查
lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Lint completed!"
