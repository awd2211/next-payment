#!/bin/bash

# 数据库迁移脚本
# 用法: ./migrate.sh [up|down|steps|version|force|status] [options]

set -e

# 默认配置
DATABASE_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5432/payment_platform?sslmode=disable"}
MIGRATIONS_PATH=${MIGRATIONS_PATH:-"file://./migrations"}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查是否安装了 migrate 工具
check_migrate_cli() {
    if ! command -v migrate &> /dev/null; then
        print_error "migrate CLI 未安装"
        echo ""
        echo "安装方法："
        echo "  macOS:   brew install golang-migrate"
        echo "  Linux:   curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz && mv migrate /usr/local/bin/"
        echo "  Windows: scoop install migrate"
        echo ""
        echo "或者使用 Go 安装："
        echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
数据库迁移脚本

用法:
  ./migrate.sh <command> [options]

命令:
  up              执行所有待执行的 up 迁移
  down            回滚所有迁移
  steps N         执行 N 步迁移（正数向上，负数向下）
  goto V          迁移到指定版本 V
  version         显示当前数据库版本
  force V         强制设置数据库版本为 V（用于修复脏状态）
  status          显示迁移状态
  create NAME     创建新的迁移文件

环境变量:
  DATABASE_URL       数据库连接字符串
  MIGRATIONS_PATH    迁移文件路径

示例:
  ./migrate.sh up
  ./migrate.sh down
  ./migrate.sh steps -1        # 回滚最后一个迁移
  ./migrate.sh steps 2         # 向上执行 2 个迁移
  ./migrate.sh goto 3          # 迁移到版本 3
  ./migrate.sh version
  ./migrate.sh create add_users_table
  ./migrate.sh force 1         # 修复脏状态

EOF
}

# 主命令处理
case "${1}" in
    up)
        check_migrate_cli
        print_info "执行所有待执行的 up 迁移..."
        migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" up
        print_info "✅ 迁移完成"
        ;;

    down)
        check_migrate_cli
        print_warning "⚠️  警告：这将回滚所有迁移！"
        read -p "确认继续？ (yes/no): " confirm
        if [ "$confirm" = "yes" ]; then
            print_info "回滚所有迁移..."
            migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" down
            print_info "✅ 回滚完成"
        else
            print_info "已取消"
        fi
        ;;

    steps)
        check_migrate_cli
        if [ -z "$2" ]; then
            print_error "错误：steps 命令需要提供步骤数"
            echo "示例: ./migrate.sh steps -1"
            exit 1
        fi
        print_info "执行 $2 步迁移..."
        migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" steps "$2"
        print_info "✅ 迁移完成"
        ;;

    goto)
        check_migrate_cli
        if [ -z "$2" ]; then
            print_error "错误：goto 命令需要提供目标版本"
            echo "示例: ./migrate.sh goto 3"
            exit 1
        fi
        print_info "迁移到版本 $2..."
        migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" goto "$2"
        print_info "✅ 迁移完成"
        ;;

    version)
        check_migrate_cli
        migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" version
        ;;

    force)
        check_migrate_cli
        if [ -z "$2" ]; then
            print_error "错误：force 命令需要提供版本号"
            echo "示例: ./migrate.sh force 1"
            exit 1
        fi
        print_warning "⚠️  强制设置版本为 $2"
        migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" force "$2"
        print_info "✅ 版本设置完成"
        ;;

    status)
        check_migrate_cli
        current_version=$(migrate -database "${DATABASE_URL}" -path "${MIGRATIONS_PATH}" version 2>&1 || echo "error")
        echo "当前状态: $current_version"
        ;;

    create)
        check_migrate_cli
        if [ -z "$2" ]; then
            print_error "错误：create 命令需要提供迁移名称"
            echo "示例: ./migrate.sh create add_users_table"
            exit 1
        fi

        # 移除 file:// 前缀
        path="${MIGRATIONS_PATH#file://}"
        migrate create -ext sql -dir "${path}" -seq "$2"
        print_info "✅ 迁移文件创建完成"
        ;;

    help|--help|-h)
        show_help
        ;;

    *)
        print_error "未知命令: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
