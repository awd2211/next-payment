#!/bin/bash
set -e

# Kong 配置恢复脚本
# 用途: 从备份文件恢复 Kong 配置

KONG_ADMIN="http://localhost:40081"
BACKUP_FILE="$1"

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
echo "  Kong 配置恢复工具"
echo "=========================================="
echo ""

# 检查参数
if [ -z "$BACKUP_FILE" ]; then
    log_error "用法: $0 <backup-file>"
    echo ""
    echo "示例:"
    echo "  $0 backend/backups/kong/kong_backup_20231024_120000.json"
    echo ""
    exit 1
fi

# 检查备份文件是否存在
if [ ! -f "$BACKUP_FILE" ]; then
    log_error "备份文件不存在: $BACKUP_FILE"
    exit 1
fi

log_info "备份文件: $BACKUP_FILE"

# 检查 Kong 是否可访问
log_info "检查 Kong 连接..."
if ! curl -s -f $KONG_ADMIN/ > /dev/null 2>&1; then
    log_error "无法连接到 Kong Admin API ($KONG_ADMIN)"
    exit 1
fi
log_success "Kong 连接正常"

# 读取备份文件
log_info "读取备份文件..."
backup_data=$(cat "$BACKUP_FILE")
backup_version=$(echo "$backup_data" | jq -r '.version')
backup_timestamp=$(echo "$backup_data" | jq -r '.timestamp')
log_success "备份版本: $backup_version, 时间: $backup_timestamp"

echo ""
log_warning "警告: 此操作将覆盖当前 Kong 配置!"
echo ""
echo -n "是否继续? (y/N): "
read confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    log_info "已取消恢复操作"
    exit 0
fi

echo ""
log_info "开始恢复 Kong 配置..."
echo ""

# 统计数据
services_count=$(echo "$backup_data" | jq '.statistics.services_count')
routes_count=$(echo "$backup_data" | jq '.statistics.routes_count')
plugins_count=$(echo "$backup_data" | jq '.statistics.plugins_count')
consumers_count=$(echo "$backup_data" | jq '.statistics.consumers_count')
jwt_count=$(echo "$backup_data" | jq '.statistics.jwt_credentials_count')
api_keys_count=$(echo "$backup_data" | jq '.statistics.api_keys_count')

log_info "备份包含:"
echo "  - Services: $services_count"
echo "  - Routes: $routes_count"
echo "  - Plugins: $plugins_count"
echo "  - Consumers: $consumers_count"
echo "  - JWT Credentials: $jwt_count"
echo "  - API Keys: $api_keys_count"
echo ""

# 1. 恢复 Services
log_info "恢复 Services..."
services=$(echo "$backup_data" | jq -c '.services[]')
service_restored=0

while IFS= read -r service; do
    name=$(echo "$service" | jq -r '.name')
    host=$(echo "$service" | jq -r '.host')
    port=$(echo "$service" | jq -r '.port')
    protocol=$(echo "$service" | jq -r '.protocol')

    # 检查服务是否已存在
    if curl -s -f $KONG_ADMIN/services/$name > /dev/null 2>&1; then
        # 更新现有服务
        curl -s -X PATCH $KONG_ADMIN/services/$name \
            --data "host=$host" \
            --data "port=$port" \
            --data "protocol=$protocol" \
            > /dev/null
    else
        # 创建新服务
        curl -s -X POST $KONG_ADMIN/services \
            --data "name=$name" \
            --data "host=$host" \
            --data "port=$port" \
            --data "protocol=$protocol" \
            > /dev/null
    fi

    service_restored=$((service_restored + 1))
done <<< "$services"

log_success "恢复 $service_restored 个 Services"

# 2. 恢复 Routes
log_info "恢复 Routes..."
routes=$(echo "$backup_data" | jq -c '.routes[]')
route_restored=0

while IFS= read -r route; do
    name=$(echo "$route" | jq -r '.name')
    service_id=$(echo "$route" | jq -r '.service.id')
    paths=$(echo "$route" | jq -r '.paths[]')

    # 通过 service id 获取 service name
    service_name=$(curl -s $KONG_ADMIN/services/$service_id | jq -r '.name')

    # 检查路由是否已存在
    if curl -s -f $KONG_ADMIN/routes/$name > /dev/null 2>&1; then
        log_warning "Route $name 已存在,跳过"
        continue
    fi

    # 构建路径参数
    path_params=""
    for path in $paths; do
        path_params="${path_params} --data paths[]=$path"
    done

    # 创建路由
    curl -s -X POST $KONG_ADMIN/services/$service_name/routes \
        --data "name=$name" \
        $path_params \
        --data "strip_path=false" \
        > /dev/null

    route_restored=$((route_restored + 1))
done <<< "$routes"

log_success "恢复 $route_restored 个 Routes"

# 3. 恢复 Consumers
log_info "恢复 Consumers..."
consumers=$(echo "$backup_data" | jq -c '.consumers[]')
consumer_restored=0

while IFS= read -r consumer; do
    username=$(echo "$consumer" | jq -r '.username')
    custom_id=$(echo "$consumer" | jq -r '.custom_id // ""')

    # 检查 Consumer 是否已存在
    if curl -s -f $KONG_ADMIN/consumers/$username > /dev/null 2>&1; then
        log_warning "Consumer $username 已存在,跳过"
        continue
    fi

    # 创建 Consumer
    if [ -n "$custom_id" ]; then
        curl -s -X POST $KONG_ADMIN/consumers \
            --data "username=$username" \
            --data "custom_id=$custom_id" \
            > /dev/null
    else
        curl -s -X POST $KONG_ADMIN/consumers \
            --data "username=$username" \
            > /dev/null
    fi

    consumer_restored=$((consumer_restored + 1))
done <<< "$consumers"

log_success "恢复 $consumer_restored 个 Consumers"

# 4. 恢复 JWT Credentials
log_info "恢复 JWT Credentials..."
jwt_credentials=$(echo "$backup_data" | jq -c '.jwt_credentials[]')
jwt_restored=0

while IFS= read -r jwt; do
    consumer=$(echo "$jwt" | jq -r '.consumer')
    key=$(echo "$jwt" | jq -r '.key')
    secret=$(echo "$jwt" | jq -r '.secret')
    algorithm=$(echo "$jwt" | jq -r '.algorithm')

    # 检查 JWT Credential 是否已存在
    existing_jwt=$(curl -s $KONG_ADMIN/consumers/$consumer/jwt | jq --arg key "$key" '.data[] | select(.key == $key)')
    if [ -n "$existing_jwt" ]; then
        log_warning "JWT Credential for $consumer (key: $key) 已存在,跳过"
        continue
    fi

    # 创建 JWT Credential
    curl -s -X POST $KONG_ADMIN/consumers/$consumer/jwt \
        --data "key=$key" \
        --data "secret=$secret" \
        --data "algorithm=$algorithm" \
        > /dev/null

    jwt_restored=$((jwt_restored + 1))
done <<< "$jwt_credentials"

log_success "恢复 $jwt_restored 个 JWT Credentials"

# 5. 恢复 API Keys
log_info "恢复 API Keys..."
api_keys=$(echo "$backup_data" | jq -c '.api_keys[]')
api_key_restored=0

while IFS= read -r api_key; do
    consumer=$(echo "$api_key" | jq -r '.consumer')
    key=$(echo "$api_key" | jq -r '.key')

    # 检查 API Key 是否已存在
    existing_key=$(curl -s $KONG_ADMIN/consumers/$consumer/key-auth | jq --arg key "$key" '.data[] | select(.key == $key)')
    if [ -n "$existing_key" ]; then
        log_warning "API Key for $consumer 已存在,跳过"
        continue
    fi

    # 创建 API Key
    curl -s -X POST $KONG_ADMIN/consumers/$consumer/key-auth \
        --data "key=$key" \
        > /dev/null

    api_key_restored=$((api_key_restored + 1))
done <<< "$api_keys"

log_success "恢复 $api_key_restored 个 API Keys"

# 6. 恢复 Plugins (跳过,因为 kong-setup.sh 会重新创建)
log_warning "插件配置需要手动运行 kong-setup.sh 重新创建"

echo ""
log_success "恢复完成!"
echo ""

echo "=========================================="
echo "  恢复摘要"
echo "=========================================="
echo ""
echo "  - Services: $service_restored / $services_count"
echo "  - Routes: $route_restored / $routes_count"
echo "  - Consumers: $consumer_restored / $consumers_count"
echo "  - JWT Credentials: $jwt_restored / $jwt_count"
echo "  - API Keys: $api_key_restored / $api_keys_count"
echo ""

echo "=========================================="
echo "  后续步骤"
echo "=========================================="
echo ""
echo "1. 重新配置插件:"
echo "   bash backend/scripts/kong-setup.sh"
echo ""
echo "2. 验证配置:"
echo "   bash backend/scripts/test-kong.sh"
echo ""
echo "=========================================="
echo ""
