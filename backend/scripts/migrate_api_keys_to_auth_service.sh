#!/bin/bash
# =============================================================================
# 脚本名称: migrate_api_keys_to_auth_service.sh
# 描述: 将 API Keys 从 payment_gateway 迁移到 payment_merchant_auth 数据库
# 用途: Phase 1 - APIKey 迁移到 merchant-auth-service
# =============================================================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 数据库配置
PG_HOST="${DB_HOST:-localhost}"
PG_PORT="${DB_PORT:-40432}"
PG_USER="${DB_USER:-postgres}"
PG_PASSWORD="${DB_PASSWORD:-postgres}"
SOURCE_DB="payment_gateway"
TARGET_DB="payment_merchant_auth"

# 临时文件
TEMP_DIR="/tmp"
EXPORT_FILE="$TEMP_DIR/api_keys_export.csv"
BACKUP_FILE="$TEMP_DIR/api_keys_backup_$(date +%Y%m%d_%H%M%S).sql"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}API Keys 数据迁移工具${NC}"
echo -e "${GREEN}从 $SOURCE_DB 迁移到 $TARGET_DB${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 步骤 1: 检查数据库连接
echo -e "${YELLOW}[1/7] 检查数据库连接...${NC}"
if docker exec payment-postgres psql -U "$PG_USER" -d "$SOURCE_DB" -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 源数据库连接成功 ($SOURCE_DB)${NC}"
else
    echo -e "${RED}✗ 无法连接源数据库 ($SOURCE_DB)${NC}"
    exit 1
fi

if docker exec payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 目标数据库连接成功 ($TARGET_DB)${NC}"
else
    echo -e "${RED}✗ 无法连接目标数据库 ($TARGET_DB)${NC}"
    exit 1
fi
echo ""

# 步骤 2: 检查源表是否存在
echo -e "${YELLOW}[2/7] 检查源表是否存在...${NC}"
SOURCE_EXISTS=$(docker exec payment-postgres psql -U "$PG_USER" -d "$SOURCE_DB" -t -c "
    SELECT COUNT(*) FROM information_schema.tables
    WHERE table_name = 'api_keys' AND table_schema = 'public'
")

if [ "$SOURCE_EXISTS" -eq 0 ]; then
    echo -e "${RED}✗ 源表 api_keys 不存在于 $SOURCE_DB${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 源表存在${NC}"
echo ""

# 步骤 3: 统计源数据
echo -e "${YELLOW}[3/7] 统计源数据...${NC}"
SOURCE_COUNT=$(docker exec payment-postgres psql -U "$PG_USER" -d "$SOURCE_DB" -t -c "SELECT COUNT(*) FROM api_keys" | xargs)
echo -e "${GREEN}源表记录数: $SOURCE_COUNT${NC}"

if [ "$SOURCE_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}⚠ 源表为空，无需迁移${NC}"
    exit 0
fi
echo ""

# 步骤 4: 备份源数据
echo -e "${YELLOW}[4/7] 备份源数据...${NC}"
docker exec payment-postgres pg_dump -U "$PG_USER" -d "$SOURCE_DB" -t api_keys > "$BACKUP_FILE"
echo -e "${GREEN}✓ 备份文件: $BACKUP_FILE${NC}"
echo ""

# 步骤 5: 导出数据到 CSV
echo -e "${YELLOW}[5/7] 导出数据到 CSV...${NC}"
docker exec payment-postgres psql -U "$PG_USER" -d "$SOURCE_DB" -c "
COPY (
    SELECT
        id,
        merchant_id,
        api_key,
        api_secret,
        name,
        environment,
        is_active,
        last_used_at,
        expires_at,
        created_at,
        updated_at
    FROM api_keys
    ORDER BY created_at
) TO STDOUT CSV HEADER
" > "$EXPORT_FILE"

EXPORT_LINES=$(wc -l < "$EXPORT_FILE")
echo -e "${GREEN}✓ 导出 $EXPORT_LINES 行到 $EXPORT_FILE${NC}"
echo ""

# 步骤 6: 导入数据到目标数据库
echo -e "${YELLOW}[6/7] 导入数据到目标数据库...${NC}"

# 确保目标表存在（merchant-auth-service 应该已经创建）
TARGET_TABLE_EXISTS=$(docker exec payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -t -c "
    SELECT COUNT(*) FROM information_schema.tables
    WHERE table_name = 'api_keys' AND table_schema = 'public'
")

if [ "$TARGET_TABLE_EXISTS" -eq 0 ]; then
    echo -e "${RED}✗ 目标表 api_keys 不存在于 $TARGET_DB${NC}"
    echo -e "${YELLOW}请先启动 merchant-auth-service 创建表结构${NC}"
    exit 1
fi

# 检查目标表是否已有数据
TARGET_COUNT=$(docker exec payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -t -c "SELECT COUNT(*) FROM api_keys" | xargs)
if [ "$TARGET_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠ 目标表已有 $TARGET_COUNT 条记录${NC}"
    read -p "是否清空目标表再导入? (y/N): " confirm
    if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
        docker exec payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -c "TRUNCATE TABLE api_keys CASCADE"
        echo -e "${GREEN}✓ 目标表已清空${NC}"
    else
        echo -e "${YELLOW}保留现有数据，追加导入${NC}"
    fi
fi

# 导入数据
cat "$EXPORT_FILE" | docker exec -i payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -c "
COPY api_keys(
    id,
    merchant_id,
    api_key,
    api_secret,
    name,
    environment,
    is_active,
    last_used_at,
    expires_at,
    created_at,
    updated_at
) FROM STDIN CSV HEADER
"
echo -e "${GREEN}✓ 数据导入完成${NC}"
echo ""

# 步骤 7: 验证数据一致性
echo -e "${YELLOW}[7/7] 验证数据一致性...${NC}"
TARGET_COUNT_AFTER=$(docker exec payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -t -c "SELECT COUNT(*) FROM api_keys" | xargs)

echo -e "${GREEN}源数据库 ($SOURCE_DB): $SOURCE_COUNT 条${NC}"
echo -e "${GREEN}目标数据库 ($TARGET_DB): $TARGET_COUNT_AFTER 条${NC}"

if [ "$SOURCE_COUNT" -eq "$TARGET_COUNT_AFTER" ]; then
    echo -e "${GREEN}✓ 数据迁移成功！行数一致${NC}"
else
    echo -e "${RED}✗ 数据迁移失败！行数不一致${NC}"
    echo -e "${YELLOW}请检查日志并从备份恢复: $BACKUP_FILE${NC}"
    exit 1
fi
echo ""

# 步骤 8: 数据抽样验证
echo -e "${YELLOW}抽样验证数据完整性...${NC}"
SAMPLE_ID=$(docker exec payment-postgres psql -U "$PG_USER" -d "$SOURCE_DB" -t -c "SELECT id FROM api_keys LIMIT 1" | xargs)
if [ ! -z "$SAMPLE_ID" ]; then
    SOURCE_SAMPLE=$(docker exec payment-postgres psql -U "$PG_USER" -d "$SOURCE_DB" -t -c "SELECT api_key, merchant_id FROM api_keys WHERE id = '$SAMPLE_ID'" | xargs)
    TARGET_SAMPLE=$(docker exec payment-postgres psql -U "$PG_USER" -d "$TARGET_DB" -t -c "SELECT api_key, merchant_id FROM api_keys WHERE id = '$SAMPLE_ID'" | xargs)

    if [ "$SOURCE_SAMPLE" == "$TARGET_SAMPLE" ]; then
        echo -e "${GREEN}✓ 抽样数据验证通过${NC}"
    else
        echo -e "${RED}✗ 抽样数据不一致${NC}"
        echo "Source: $SOURCE_SAMPLE"
        echo "Target: $TARGET_SAMPLE"
    fi
fi
echo ""

# 清理临时文件（保留备份）
rm -f "$EXPORT_FILE"

# 总结
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}迁移完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}下一步操作：${NC}"
echo "1. 测试 merchant-auth-service 的 API Key 验证功能"
echo "2. 在测试环境启用新方案: export USE_AUTH_SERVICE=true"
echo "3. 验证 payment-gateway 可以正常调用 merchant-auth-service"
echo "4. 运行集成测试确认无问题"
echo ""
echo -e "${YELLOW}回滚方案：${NC}"
echo "如果需要回滚，使用备份文件恢复："
echo "  docker exec -i payment-postgres psql -U $PG_USER -d $SOURCE_DB < $BACKUP_FILE"
echo ""
echo -e "${YELLOW}备份文件位置：${NC}"
echo "  $BACKUP_FILE"
echo ""
