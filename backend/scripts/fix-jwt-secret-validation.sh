#!/bin/bash
# 批量为所有微服务添加JWT_SECRET强制验证
# 日期: 2025-10-26
# 目的: 修复严重安全漏洞 - 硬编码JWT密钥

set -e

BACKEND_DIR="/home/eric/payment/backend"
SERVICE_DIRS=$(find "$BACKEND_DIR/services" -maxdepth 1 -type d -name "*-service" -o -name "*-bff-service")

echo "🔐 开始批量修复JWT_SECRET验证问题..."
echo "============================================"

FIXED_COUNT=0
SKIPPED_COUNT=0

for SERVICE_DIR in $SERVICE_DIRS; do
    SERVICE_NAME=$(basename "$SERVICE_DIR")
    MAIN_FILE="$SERVICE_DIR/cmd/main.go"

    if [ ! -f "$MAIN_FILE" ]; then
        echo "⚠️  跳过 $SERVICE_NAME: main.go不存在"
        ((SKIPPED_COUNT++))
        continue
    fi

    # 检查是否包含JWT_SECRET
    if ! grep -q "JWT_SECRET" "$MAIN_FILE"; then
        echo "⚠️  跳过 $SERVICE_NAME: 不使用JWT_SECRET"
        ((SKIPPED_COUNT++))
        continue
    fi

    # 检查是否已经修复
    if grep -q "JWT_SECRET environment variable is required" "$MAIN_FILE"; then
        echo "✅ $SERVICE_NAME: 已经修复"
        ((SKIPPED_COUNT++))
        continue
    fi

    echo "🔧 修复 $SERVICE_NAME..."

    # 备份原文件
    cp "$MAIN_FILE" "$MAIN_FILE.backup"

    # 使用Perl进行多行替换
    # 将 jwtSecret := getConfig("JWT_SECRET", "payment-platform-secret-key-2024")
    # 替换为带验证的版本
    perl -i -0pe 's/(\tjwtSecret := getConfig\("JWT_SECRET", "payment-platform-secret-key-2024"\))/\t\/\/ ⚠️ 安全要求: JWT_SECRET必须在生产环境中设置，不能使用默认值\n\tjwtSecret := getConfig("JWT_SECRET", "")\n\tif jwtSecret == "" {\n\t\tlogger.Fatal("JWT_SECRET environment variable is required and cannot be empty")\n\t}\n\tif len(jwtSecret) < 32 {\n\t\tlogger.Fatal("JWT_SECRET must be at least 32 characters for security",\n\t\t\tzap.Int("current_length", len(jwtSecret)),\n\t\t\tzap.Int("minimum_length", 32))\n\t}\n\tlogger.Info("JWT_SECRET validation passed", zap.Int("length", len(jwtSecret)))/g' "$MAIN_FILE"

    # 验证修改是否成功
    if grep -q "JWT_SECRET environment variable is required" "$MAIN_FILE"; then
        echo "✅ $SERVICE_NAME 修复成功"
        ((FIXED_COUNT++))
        # 删除备份
        rm "$MAIN_FILE.backup"
    else
        echo "❌ $SERVICE_NAME 修复失败，恢复备份"
        mv "$MAIN_FILE.backup" "$MAIN_FILE"
    fi

    echo ""
done

echo "============================================"
echo "📊 修复统计:"
echo "  ✅ 成功修复: $FIXED_COUNT 个服务"
echo "  ⚠️  跳过/已修复: $SKIPPED_COUNT 个服务"
echo ""
echo "🎉 JWT_SECRET验证修复完成！"
echo ""
echo "⚠️  重要提示:"
echo "  1. 所有服务启动前必须设置JWT_SECRET环境变量"
echo "  2. JWT_SECRET长度必须至少32个字符"
echo "  3. 建议使用以下命令生成强密钥:"
echo "     openssl rand -base64 32"
echo ""
