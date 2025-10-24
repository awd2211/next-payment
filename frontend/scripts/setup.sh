#!/bin/bash

# 前端项目初始化脚本

set -e

echo "🚀 开始初始化前端项目..."

# 1. 检查必要工具
echo "📋 检查必要工具..."

if ! command -v node &> /dev/null; then
    echo "❌ Node.js未安装，请先安装Node.js >= 18.0"
    exit 1
fi

if ! command -v pnpm &> /dev/null; then
    echo "⚠️  pnpm未安装，正在安装..."
    npm install -g pnpm
fi

echo "✅ Node.js版本: $(node -v)"
echo "✅ pnpm版本: $(pnpm -v)"

# 2. 清理旧的lock文件
echo "🧹 清理旧的lock文件..."
find . -name "package-lock.json" -type f -delete
echo "✅ 已清理package-lock.json文件"

# 3. 安装依赖
echo "📦 安装依赖..."
pnpm install

# 4. 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p logs
mkdir -p admin-portal/logs
mkdir -p merchant-portal/logs
mkdir -p website/logs

# 5. 创建.env文件（如果不存在）
echo "⚙️  创建环境变量文件..."

# Admin Portal
if [ ! -f "admin-portal/.env.development" ]; then
    cat > admin-portal/.env.development << EOF
VITE_APP_TITLE=支付平台管理后台
VITE_PORT=5173
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=10000
VITE_ENABLE_MOCK=false
EOF
    echo "✅ 已创建 admin-portal/.env.development"
fi

if [ ! -f "admin-portal/.env.production" ]; then
    cat > admin-portal/.env.production << EOF
VITE_APP_TITLE=支付平台管理后台
VITE_PORT=5173
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=30000
VITE_ENABLE_MOCK=false
EOF
    echo "✅ 已创建 admin-portal/.env.production"
fi

# Merchant Portal
if [ ! -f "merchant-portal/.env.development" ]; then
    cat > merchant-portal/.env.development << EOF
VITE_APP_TITLE=支付平台商户中心
VITE_PORT=5174
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=10000
VITE_ENABLE_MOCK=false
EOF
    echo "✅ 已创建 merchant-portal/.env.development"
fi

if [ ! -f "merchant-portal/.env.production" ]; then
    cat > merchant-portal/.env.production << EOF
VITE_APP_TITLE=支付平台商户中心
VITE_PORT=5174
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=30000
VITE_ENABLE_MOCK=false
EOF
    echo "✅ 已创建 merchant-portal/.env.production"
fi

# Website
if [ ! -f "website/.env.development" ]; then
    cat > website/.env.development << EOF
VITE_APP_TITLE=支付平台官网
VITE_PORT=5175
EOF
    echo "✅ 已创建 website/.env.development"
fi

# 6. 复制配置文件到其他项目
echo "📋 复制配置文件到其他项目..."

# ESLint配置
if [ -f "admin-portal/.eslintrc.json" ]; then
    cp admin-portal/.eslintrc.json merchant-portal/.eslintrc.json
    cp admin-portal/.eslintrc.json website/.eslintrc.json
    echo "✅ 已复制.eslintrc.json"
fi

# Prettier配置
if [ -f "admin-portal/.prettierrc.json" ]; then
    cp admin-portal/.prettierrc.json merchant-portal/.prettierrc.json
    cp admin-portal/.prettierrc.json website/.prettierrc.json
    echo "✅ 已复制.prettierrc.json"
fi

# 7. 类型检查
echo "🔍 TypeScript类型检查..."
cd admin-portal && pnpm type-check && cd ..
echo "✅ admin-portal类型检查通过"

cd merchant-portal && pnpm type-check && cd ..
echo "✅ merchant-portal类型检查通过"

cd website && pnpm type-check && cd ..
echo "✅ website类型检查通过"

echo ""
echo "🎉 前端项目初始化完成！"
echo ""
echo "📝 下一步操作："
echo "  1. 确保后端服务运行在 40001-40010 端口"
echo "  2. 启动开发服务器："
echo "     - 方式1: pm2 start ecosystem.config.js"
echo "     - 方式2: cd admin-portal && pnpm dev"
echo ""
echo "🌐 访问地址："
echo "  - Admin Portal:    http://localhost:5173"
echo "  - Merchant Portal: http://localhost:5174"
echo "  - Website:         http://localhost:5175"
echo ""


