#!/bin/bash

# 启动所有前端开发服务器（使用PM2）

echo "🚀 启动前端开发服务器..."

# 检查PM2是否安装
if ! command -v pm2 &> /dev/null; then
    echo "⚠️  PM2未安装，正在安装..."
    npm install -g pm2
fi

# 停止已有的进程
echo "🛑 停止旧的进程..."
pm2 delete all 2>/dev/null || true

# 启动新进程
echo "▶️  启动新进程..."
cd /home/eric/payment/frontend
pm2 start ecosystem.config.js

# 显示状态
echo ""
pm2 status

echo ""
echo "✅ 前端服务已启动！"
echo ""
echo "📊 查看日志:"
echo "  pm2 logs"
echo ""
echo "🌐 访问地址:"
echo "  - Admin Portal:    http://localhost:5173"
echo "  - Merchant Portal: http://localhost:5174"
echo "  - Website:         http://localhost:5175"
echo ""
echo "⚙️  常用命令:"
echo "  pm2 status       - 查看状态"
echo "  pm2 logs         - 查看日志"
echo "  pm2 restart all  - 重启所有"
echo "  pm2 stop all     - 停止所有"
echo ""





