#!/bin/bash
set -e

# Kong 配置备份脚本
# 用途: 备份 Kong 的所有配置 (Services, Routes, Plugins, Consumers)

KONG_ADMIN="http://localhost:40081"
BACKUP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../backups/kong" && pwd)"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/kong_backup_${TIMESTAMP}.json"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

echo ""
echo "=========================================="
echo "  Kong 配置备份工具"
echo "=========================================="
echo ""

# 创建备份目录
mkdir -p "$BACKUP_DIR"
log_info "备份目录: $BACKUP_DIR"

# 检查 Kong 是否可访问
log_info "检查 Kong 连接..."
if ! curl -s -f $KONG_ADMIN/ > /dev/null 2>&1; then
    log_error "无法连接到 Kong Admin API ($KONG_ADMIN)"
    log_info "请确保 Kong 正在运行: docker compose ps kong"
    exit 1
fi
log_success "Kong 连接正常"

echo ""
log_info "开始备份 Kong 配置..."
echo ""

# 备份 Services
log_info "备份 Services..."
services=$(curl -s $KONG_ADMIN/services)
services_count=$(echo "$services" | jq '.data | length')
log_success "备份 $services_count 个 Services"

# 备份 Routes
log_info "备份 Routes..."
routes=$(curl -s $KONG_ADMIN/routes)
routes_count=$(echo "$routes" | jq '.data | length')
log_success "备份 $routes_count 个 Routes"

# 备份 Plugins
log_info "备份 Plugins..."
plugins=$(curl -s $KONG_ADMIN/plugins)
plugins_count=$(echo "$plugins" | jq '.data | length')
log_success "备份 $plugins_count 个 Plugins"

# 备份 Consumers
log_info "备份 Consumers..."
consumers=$(curl -s $KONG_ADMIN/consumers)
consumers_count=$(echo "$consumers" | jq '.data | length')
log_success "备份 $consumers_count 个 Consumers"

# 备份 Consumer JWT Credentials
log_info "备份 JWT Credentials..."
jwt_credentials="[]"
if [ "$consumers_count" -gt 0 ]; then
    consumer_usernames=$(echo "$consumers" | jq -r '.data[].username')
    all_jwt_creds="[]"

    for username in $consumer_usernames; do
        jwt_creds=$(curl -s $KONG_ADMIN/consumers/$username/jwt)
        if [ "$(echo "$jwt_creds" | jq '.data | length')" -gt 0 ]; then
            # 添加 consumer 信息到每个 credential
            jwt_creds=$(echo "$jwt_creds" | jq --arg username "$username" '.data[] | . + {consumer: $username}')
            all_jwt_creds=$(echo "$all_jwt_creds" | jq --argjson new "$jwt_creds" '. + [$new]')
        fi
    done

    jwt_credentials="$all_jwt_creds"
fi
jwt_count=$(echo "$jwt_credentials" | jq 'length')
log_success "备份 $jwt_count 个 JWT Credentials"

# 备份 Consumer API Keys
log_info "备份 API Keys..."
api_keys="[]"
if [ "$consumers_count" -gt 0 ]; then
    consumer_usernames=$(echo "$consumers" | jq -r '.data[].username')
    all_api_keys="[]"

    for username in $consumer_usernames; do
        keys=$(curl -s $KONG_ADMIN/consumers/$username/key-auth)
        if [ "$(echo "$keys" | jq '.data | length')" -gt 0 ]; then
            # 添加 consumer 信息到每个 API key
            keys=$(echo "$keys" | jq --arg username "$username" '.data[] | . + {consumer: $username}')
            all_api_keys=$(echo "$all_api_keys" | jq --argjson new "$keys" '. + [$new]')
        fi
    done

    api_keys="$all_api_keys"
fi
api_keys_count=$(echo "$api_keys" | jq 'length')
log_success "备份 $api_keys_count 个 API Keys"

echo ""
log_info "生成备份文件..."

# 组合所有数据到单一 JSON 文件
backup_data=$(jq -n \
    --argjson services "$services" \
    --argjson routes "$routes" \
    --argjson plugins "$plugins" \
    --argjson consumers "$consumers" \
    --argjson jwt_credentials "$jwt_credentials" \
    --argjson api_keys "$api_keys" \
    '{
        version: "1.0",
        timestamp: "'$TIMESTAMP'",
        kong_version: "3.9",
        services: $services.data,
        routes: $routes.data,
        plugins: $plugins.data,
        consumers: $consumers.data,
        jwt_credentials: $jwt_credentials,
        api_keys: $api_keys,
        statistics: {
            services_count: ($services.data | length),
            routes_count: ($routes.data | length),
            plugins_count: ($plugins.data | length),
            consumers_count: ($consumers.data | length),
            jwt_credentials_count: ($jwt_credentials | length),
            api_keys_count: ($api_keys | length)
        }
    }')

# 保存到文件
echo "$backup_data" > "$BACKUP_FILE"
log_success "备份文件已保存: $BACKUP_FILE"

# 计算文件大小
file_size=$(du -h "$BACKUP_FILE" | cut -f1)
log_info "文件大小: $file_size"

echo ""
log_success "备份完成!"
echo ""

# 显示备份摘要
echo "=========================================="
echo "  备份摘要"
echo "=========================================="
echo ""
echo "$backup_data" | jq '.statistics'
echo ""

# 列出最近的备份文件
echo "=========================================="
echo "  最近的备份文件 (最多显示 5 个)"
echo "=========================================="
echo ""
ls -lht "$BACKUP_DIR" | head -6
echo ""

# 清理旧备份 (保留最近 10 个)
log_info "清理旧备份文件 (保留最近 10 个)..."
backup_count=$(ls -1 "$BACKUP_DIR"/kong_backup_*.json 2>/dev/null | wc -l)
if [ "$backup_count" -gt 10 ]; then
    ls -1t "$BACKUP_DIR"/kong_backup_*.json | tail -n +11 | xargs rm -f
    removed_count=$((backup_count - 10))
    log_success "已删除 $removed_count 个旧备份文件"
else
    log_info "备份文件数量未超过限制,无需清理"
fi

echo ""
echo "=========================================="
echo "  恢复备份"
echo "=========================================="
echo ""
echo "使用以下命令恢复备份:"
echo ""
echo "  bash backend/scripts/kong-restore.sh $BACKUP_FILE"
echo ""
echo "=========================================="
echo ""
