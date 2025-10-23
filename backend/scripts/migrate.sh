#!/bin/bash

# 数据库迁移管理脚本
# 使用 golang-migrate

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-40432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_SSL_MODE="${DB_SSL_MODE:-disable}"

# 服务和数据库映射
declare -A SERVICE_DB=(
    ["admin-service"]="payment_admin"
    ["merchant-service"]="payment_merchant"
    ["payment-gateway"]="payment_gateway"
    ["order-service"]="payment_order"
    ["channel-adapter"]="payment_channel"
    ["risk-service"]="payment_risk"
    ["accounting-service"]="payment_accounting"
    ["notification-service"]="payment_notification"
    ["analytics-service"]="payment_analytics"
    ["config-service"]="payment_config"
)

BACKEND_DIR="/home/eric/payment/backend/services"

# 显示帮助
show_help() {
    cat << HELP
数据库迁移管理脚本

用法: $0 <command> [service] [options]

命令:
  up [service]       - 执行所有待迁移（或指定服务）
  down [service] [N] - 回滚N步迁移（默认1步）
  reset [service]    - 重置数据库（删除所有表）
  version [service]  - 显示当前迁移版本
  force [service] N  - 强制设置版本（用于修复dirty状态）
  create <name>      - 创建新的迁移文件
  status             - 显示所有服务的迁移状态

服务列表:
  admin-service, merchant-service, payment-gateway, order-service,
  channel-adapter, risk-service, accounting-service, notification-service,
  analytics-service, config-service
  
  或使用 'all' 应用到所有服务

示例:
  $0 up all                    # 迁移所有服务
  $0 up admin-service          # 只迁移 admin-service
  $0 down admin-service 1      # 回滚 admin-service 1步
  $0 reset admin-service       # 重置 admin-service 数据库
  $0 version all               # 显示所有服务的版本
  
环境变量:
  DB_HOST     - 数据库主机 (默认: localhost)
  DB_PORT     - 数据库端口 (默认: 40432)
  DB_USER     - 数据库用户 (默认: postgres)
  DB_PASSWORD - 数据库密码 (默认: postgres)
  DB_SSL_MODE - SSL模式 (默认: disable)

HELP
}

# 构建数据库URL
get_db_url() {
    local db_name=$1
    echo "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${db_name}?sslmode=${DB_SSL_MODE}"
}

# 执行迁移命令
run_migrate() {
    local service=$1
    local command=$2
    shift 2
    local args="$@"
    
    if [ ! -d "$BACKEND_DIR/$service" ]; then
        echo -e "${RED}✗ 服务不存在: $service${NC}"
        return 1
    fi
    
    local db_name="${SERVICE_DB[$service]}"
    if [ -z "$db_name" ]; then
        echo -e "${RED}✗ 未找到服务对应的数据库: $service${NC}"
        return 1
    fi
    
    local migrations_dir="$BACKEND_DIR/$service/migrations"
    if [ ! -d "$migrations_dir" ]; then
        echo -e "${RED}✗ 迁移目录不存在: $migrations_dir${NC}"
        return 1
    fi
    
    local db_url=$(get_db_url "$db_name")
    
    echo -e "${BLUE}=== $service ($db_name) ===${NC}"
    
    # 使用 golang-migrate CLI
    migrate -path "$migrations_dir" -database "$db_url" $command $args
    
    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓ 完成${NC}"
    else
        echo -e "${RED}✗ 失败${NC}"
    fi
    echo ""
    
    return $exit_code
}

# 处理 all 参数
process_all_services() {
    local command=$1
    shift
    local args="$@"
    
    local failed=0
    for service in "${!SERVICE_DB[@]}"; do
        if ! run_migrate "$service" "$command" $args; then
            failed=$((failed + 1))
        fi
    done
    
    echo -e "${BLUE}==============================${NC}"
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}✓ 所有服务处理完成${NC}"
    else
        echo -e "${YELLOW}⚠ $failed 个服务处理失败${NC}"
    fi
}

# 显示迁移版本
show_version() {
    local service=$1
    
    if [ "$service" = "all" ]; then
        echo -e "${BLUE}所有服务的迁移版本:${NC}"
        echo ""
        for svc in "${!SERVICE_DB[@]}"; do
            run_migrate "$svc" "version" 2>&1 | grep -E "version|dirty" || echo "无版本信息"
        done
    else
        run_migrate "$service" "version"
    fi
}

# 显示状态
show_status() {
    echo -e "${BLUE}==============================${NC}"
    echo -e "${BLUE}迁移状态总览${NC}"
    echo -e "${BLUE}==============================${NC}"
    echo ""
    
    printf "%-25s %-20s %s\n" "服务" "数据库" "版本"
    printf "%-25s %-20s %s\n" "-------------------------" "--------------------" "----------"
    
    for service in $(echo "${!SERVICE_DB[@]}" | tr ' ' '\n' | sort); do
        local db_name="${SERVICE_DB[$service]}"
        local migrations_dir="$BACKEND_DIR/$service/migrations"
        local db_url=$(get_db_url "$db_name")
        
        # 获取版本
        local version_info=$(migrate -path "$migrations_dir" -database "$db_url" version 2>&1 || echo "error")
        
        if echo "$version_info" | grep -q "dirty"; then
            printf "%-25s %-20s ${RED}%s${NC}\n" "$service" "$db_name" "DIRTY"
        elif echo "$version_info" | grep -q "no migration"; then
            printf "%-25s %-20s ${YELLOW}%s${NC}\n" "$service" "$db_name" "未迁移"
        else
            local version=$(echo "$version_info" | grep -oP '\d+' | head -1)
            printf "%-25s %-20s ${GREEN}v%s${NC}\n" "$service" "$db_name" "${version:-0}"
        fi
    done
    
    echo ""
}

# 主逻辑
main() {
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi
    
    local command=$1
    shift
    
    case "$command" in
        up)
            local service=${1:-all}
            if [ "$service" = "all" ]; then
                process_all_services "up"
            else
                run_migrate "$service" "up"
            fi
            ;;
        down)
            local service=${1}
            local steps=${2:-1}
            if [ -z "$service" ]; then
                echo -e "${RED}错误: 必须指定服务${NC}"
                show_help
                exit 1
            fi
            if [ "$service" = "all" ]; then
                process_all_services "down" "$steps"
            else
                run_migrate "$service" "down" "$steps"
            fi
            ;;
        reset)
            local service=${1}
            if [ -z "$service" ]; then
                echo -e "${RED}错误: 必须指定服务${NC}"
                show_help
                exit 1
            fi
            echo -e "${YELLOW}⚠ 警告: 这将删除所有数据！${NC}"
            read -p "确认重置 $service? (输入 'yes' 确认): " confirm
            if [ "$confirm" != "yes" ]; then
                echo "已取消"
                exit 0
            fi
            if [ "$service" = "all" ]; then
                process_all_services "down" "-all"
            else
                run_migrate "$service" "down" "-all"
            fi
            ;;
        version)
            local service=${1:-all}
            show_version "$service"
            ;;
        force)
            local service=${1}
            local version=${2}
            if [ -z "$service" ] || [ -z "$version" ]; then
                echo -e "${RED}错误: 必须指定服务和版本${NC}"
                show_help
                exit 1
            fi
            run_migrate "$service" "force" "$version"
            ;;
        status)
            show_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}错误: 未知命令 '$command'${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
