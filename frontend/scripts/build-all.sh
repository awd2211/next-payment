#!/bin/bash

# 构建所有前端项目

set -e

echo "📦 开始构建所有前端项目..."

cd /home/eric/payment/frontend

# 清理旧的构建文件
echo "🧹 清理旧的构建文件..."
rm -rf admin-portal/dist
rm -rf merchant-portal/dist
rm -rf website/dist

# 构建admin-portal
echo ""
echo "📦 构建 admin-portal..."
cd admin-portal
pnpm build
echo "✅ admin-portal构建完成"

# 构建merchant-portal
echo ""
echo "📦 构建 merchant-portal..."
cd ../merchant-portal
pnpm build
echo "✅ merchant-portal构建完成"

# 构建website
echo ""
echo "📦 构建 website..."
cd ../website
pnpm build
echo "✅ website构建完成"

cd ..

# 显示构建结果
echo ""
echo "🎉 所有项目构建完成！"
echo ""
echo "📁 构建产物:"
echo "  - admin-portal/dist"
echo "  - merchant-portal/dist"
echo "  - website/dist"
echo ""
echo "📝 下一步:"
echo "  1. 预览构建: cd admin-portal && pnpm preview"
echo "  2. 部署到服务器"
echo ""


