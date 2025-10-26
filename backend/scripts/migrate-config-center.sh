#!/bin/bash

# 配置中心集成批量迁移脚本
# 将所有服务迁移为使用配置中心客户端

set -e

BACKEND_DIR="/home/eric/payment/backend"
SERVICES_DIR="$BACKEND_DIR/services"

# 待迁移服务列表（已完成：payment-gateway, order-service, channel-adapter）
SERVICES=(
    "risk-service"
    "accounting-service"
    "notification-service"
    "analytics-service"
    "merchant-auth-service"
    "settlement-service"
    "withdrawal-service"
    "kyc-service"
    "cashier-service"
    "admin-service"
    "merchant-service"
    "admin-bff-service"
    "merchant-bff-service"
    "reconciliation-service"
    "dispute-service"
    "merchant-limit-service"
)

# 统计变量
SUCCESS_COUNT=0
FAIL_COUNT=0
SKIPPED_COUNT=0

echo "========================================="
echo "配置中心集成批量迁移"
echo "========================================="
echo ""
echo "待迁移服务总数: ${#SERVICES[@]}"
echo "已手动完成: payment-gateway, order-service, channel-adapter"
echo ""

# 为每个服务添加配置中心客户端
for service in "${SERVICES[@]}"; do
    echo ""
    echo "=== 处理服务: $service ==="

    SERVICE_DIR="$SERVICES_DIR/$service"
    MAIN_GO="$SERVICE_DIR/cmd/main.go"

    # 检查服务目录和main.go是否存在
    if [ ! -f "$MAIN_GO" ]; then
        echo "  ⚠️  跳过: $MAIN_GO 不存在"
        ((SKIPPED_COUNT++))
        continue
    fi

    # 检查是否已经包含configclient导入
    if grep -q "github.com/payment-platform/pkg/configclient" "$MAIN_GO"; then
        echo "  ✅ 已集成配置中心客户端，跳过"
        ((SKIPPED_COUNT++))
        continue
    fi

    # 备份原文件
    cp "$MAIN_GO" "$MAIN_GO.bak"
    echo "  📋 已备份: $MAIN_GO.bak"

    # 创建临时文件
    TMP_FILE=$(mktemp)

    # 处理文件：添加import和配置客户端初始化代码
    awk '
    /^import \(/ {
        print
        in_import = 1
        next
    }

    in_import && /^[[:space:]]*"github.com\/payment-platform\/pkg\/config"/ {
        print
        print "\t\"github.com/payment-platform/pkg/configclient\""
        config_imported = 1
        next
    }

    in_import && /^[[:space:]]*"github.com\/payment-platform\/pkg\/logger"/ {
        print
        if (!zap_imported) {
            print "\t\"go.uber.org/zap\""
            zap_imported = 1
        }
        next
    }

    in_import && /^\)/ {
        if (!config_imported) {
            print "\t\"github.com/payment-platform/pkg/configclient\""
        }
        if (!zap_imported) {
            print "\t\"go.uber.org/zap\""
        }
        print
        in_import = 0
        next
    }

    /^func main\(\) {/ {
        print
        print "\t// 1. 初始化配置客户端（可选，失败不影响启动）"
        print "\tvar configClient *configclient.Client"
        print "\tenableConfigClient := config.GetEnv(\"ENABLE_CONFIG_CLIENT\", \"false\") == \"true\""
        print ""
        print "\tif enableConfigClient {"
        print "\t\t// 检查是否启用 mTLS"
        print "\t\tenableConfigMTLS := config.GetEnvBool(\"CONFIG_CLIENT_MTLS\", false)"
        print ""
        print "\t\tclientCfg := configclient.ClientConfig{"
        print "\t\t\tServiceName: \"'$service'\","
        print "\t\t\tEnvironment: config.GetEnv(\"ENV\", \"production\"),"
        print "\t\t\tConfigURL:   config.GetEnv(\"CONFIG_SERVICE_URL\", \"http://localhost:40010\"),"
        print "\t\t\tRefreshRate: 30 * time.Second,"
        print "\t\t}"
        print ""
        print "\t\t// 如果启用 mTLS,添加证书配置"
        print "\t\tif enableConfigMTLS {"
        print "\t\t\tclientCfg.EnableMTLS = true"
        print "\t\t\tclientCfg.TLSCertFile = config.GetEnv(\"TLS_CERT_FILE\", \"\")"
        print "\t\t\tclientCfg.TLSKeyFile = config.GetEnv(\"TLS_KEY_FILE\", \"\")"
        print "\t\t\tclientCfg.TLSCAFile = config.GetEnv(\"TLS_CA_FILE\", \"\")"
        print "\t\t}"
        print ""
        print "\t\tclient, err := configclient.NewClient(clientCfg)"
        print "\t\tif err != nil {"
        print "\t\t\tlogger.Warn(\"配置客户端初始化失败，将使用环境变量\", zap.Error(err))"
        print "\t\t} else {"
        print "\t\t\tconfigClient = client"
        print "\t\t\tdefer configClient.Stop()"
        print "\t\t\tlogger.Info(\"配置中心客户端初始化成功\")"
        print "\t\t}"
        print "\t}"
        print ""
        print "\t// 定义配置获取函数：优先从配置中心获取，失败则使用环境变量"
        print "\tgetConfig := func(key, defaultValue string) string {"
        print "\t\tif configClient != nil {"
        print "\t\t\tif val := configClient.Get(key); val != \"\" {"
        print "\t\t\t\treturn val"
        print "\t\t\t}"
        print "\t\t}"
        print "\t\treturn config.GetEnv(key, defaultValue)"
        print "\t}"
        print ""
        next
    }

    { print }
    ' "$MAIN_GO" > "$TMP_FILE"

    # 替换原文件
    mv "$TMP_FILE" "$MAIN_GO"
    echo "  ✅ 已添加配置中心客户端初始化代码"

    # 测试编译
    echo "  🔨 测试编译..."
    cd "$SERVICE_DIR"
    if GOWORK="$BACKEND_DIR/go.work" timeout 30 go build -o "/tmp/$service-test" ./cmd/main.go 2>/dev/null; then
        echo "  ✅ 编译成功"
        rm -f "/tmp/$service-test"
        ((SUCCESS_COUNT++))
    else
        echo "  ❌ 编译失败，恢复备份"
        mv "$MAIN_GO.bak" "$MAIN_GO"
        ((FAIL_COUNT++))
    fi

    # 清理备份（编译成功时）
    if [ $? -eq 0 ]; then
        rm -f "$MAIN_GO.bak"
    fi
done

echo ""
echo "========================================="
echo "迁移完成统计"
echo "========================================="
echo "✅ 成功: $SUCCESS_COUNT"
echo "❌ 失败: $FAIL_COUNT"
echo "⏭️  跳过: $SKIPPED_COUNT"
echo "📊 总计: ${#SERVICES[@]}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo "🎉 所有服务迁移成功！"
    exit 0
else
    echo "⚠️  部分服务迁移失败，请手动检查"
    exit 1
fi
