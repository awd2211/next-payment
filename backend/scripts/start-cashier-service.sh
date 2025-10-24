#!/bin/bash

# 启动 cashier-service 的脚本

set -e

echo "🚀 Starting Cashier Service..."

# 切换到服务目录
cd "$(dirname "$0")/../services/cashier-service"

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=payment_cashier
export REDIS_HOST=localhost
export REDIS_PORT=40379
export PORT=40016
export JWT_SECRET=your-secret-key
export ENV=development

# 检查数据库是否存在
if ! docker exec payment-postgres psql -U postgres -lqt | cut -d \| -f 1 | grep -qw payment_cashier; then
    echo "📦 Creating database payment_cashier..."
    docker exec payment-postgres psql -U postgres -c "CREATE DATABASE payment_cashier;"
fi

# 启动服务
echo "✅ Starting service on port 40016..."
GOWORK=/home/eric/payment/backend/go.work go run ./cmd/main.go
