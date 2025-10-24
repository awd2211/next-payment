#!/bin/bash
set -e

# Kong JWT 认证配置脚本
# 用途: 为 Admin Portal 和 Merchant Portal 配置 JWT Consumer

KONG_ADMIN="http://localhost:40081"

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

# 读取 JWT Secret (从环境变量或提示输入)
read_jwt_secret() {
    if [ -n "$JWT_SECRET" ]; then
        echo "$JWT_SECRET"
    else
        log_warning "未找到 JWT_SECRET 环境变量"
        echo -n "请输入 JWT Secret (按回车使用默认值 'your-secret-key'): "
        read input_secret
        echo "${input_secret:-your-secret-key}"
    fi
}

echo ""
echo "=========================================="
echo "  Kong JWT 认证配置工具"
echo "=========================================="
echo ""

log_info "读取 JWT Secret..."
JWT_SECRET=$(read_jwt_secret)
log_success "JWT Secret: ${JWT_SECRET:0:10}... (已加载)"

echo ""
log_info "开始配置 JWT Consumers..."
echo ""

# 1. 创建 Admin Portal Consumer
log_info "创建 Admin Portal Consumer..."
admin_consumer_result=$(curl -s -X POST $KONG_ADMIN/consumers \
    --data "username=admin-portal" \
    --data "custom_id=admin-portal-app" 2>&1)

if echo "$admin_consumer_result" | grep -q '"username":"admin-portal"'; then
    log_success "Admin Portal Consumer 已创建"
elif echo "$admin_consumer_result" | grep -q "already exists"; then
    log_warning "Admin Portal Consumer 已存在,跳过创建"
else
    log_error "创建失败: $admin_consumer_result"
fi

# 2. 为 Admin Portal 创建 JWT Credential
log_info "创建 Admin Portal JWT Credential..."
admin_jwt_result=$(curl -s -X POST $KONG_ADMIN/consumers/admin-portal/jwt \
    --data "key=admin-portal" \
    --data "algorithm=HS256" \
    --data "secret=$JWT_SECRET" 2>&1)

if echo "$admin_jwt_result" | grep -q '"key":"admin-portal"'; then
    log_success "Admin Portal JWT Credential 已创建"
elif echo "$admin_jwt_result" | grep -q "already exists"; then
    log_warning "Admin Portal JWT Credential 已存在,跳过创建"
else
    log_error "创建失败: $admin_jwt_result"
fi

echo ""

# 3. 创建 Merchant Portal Consumer
log_info "创建 Merchant Portal Consumer..."
merchant_consumer_result=$(curl -s -X POST $KONG_ADMIN/consumers \
    --data "username=merchant-portal" \
    --data "custom_id=merchant-portal-app" 2>&1)

if echo "$merchant_consumer_result" | grep -q '"username":"merchant-portal"'; then
    log_success "Merchant Portal Consumer 已创建"
elif echo "$merchant_consumer_result" | grep -q "already exists"; then
    log_warning "Merchant Portal Consumer 已存在,跳过创建"
else
    log_error "创建失败: $merchant_consumer_result"
fi

# 4. 为 Merchant Portal 创建 JWT Credential
log_info "创建 Merchant Portal JWT Credential..."
merchant_jwt_result=$(curl -s -X POST $KONG_ADMIN/consumers/merchant-portal/jwt \
    --data "key=merchant-portal" \
    --data "algorithm=HS256" \
    --data "secret=$JWT_SECRET" 2>&1)

if echo "$merchant_jwt_result" | grep -q '"key":"merchant-portal"'; then
    log_success "Merchant Portal JWT Credential 已创建"
elif echo "$merchant_jwt_result" | grep -q "already exists"; then
    log_warning "Merchant Portal JWT Credential 已存在,跳过创建"
else
    log_error "创建失败: $merchant_jwt_result"
fi

echo ""
log_success "JWT Consumer 配置完成!"
echo ""

# 5. 显示配置摘要
echo "=========================================="
echo "  JWT Consumer 配置摘要"
echo "=========================================="
echo ""

admin_jwt=$(curl -s $KONG_ADMIN/consumers/admin-portal/jwt 2>/dev/null)
merchant_jwt=$(curl -s $KONG_ADMIN/consumers/merchant-portal/jwt 2>/dev/null)

echo "Admin Portal:"
echo "  Consumer: admin-portal"
echo "  JWT Key (iss): admin-portal"
echo "  Algorithm: HS256"
echo "  Secret: ${JWT_SECRET:0:10}..."
echo ""

echo "Merchant Portal:"
echo "  Consumer: merchant-portal"
echo "  JWT Key (iss): merchant-portal"
echo "  Algorithm: HS256"
echo "  Secret: ${JWT_SECRET:0:10}..."
echo ""

# 6. 后端服务需要的代码变更提示
echo "=========================================="
echo "  后端服务需要的变更"
echo "=========================================="
echo ""

cat << 'EOF'
后端服务在签发 JWT 时,必须在 iss (issuer) 声明中指定 Consumer key:

1. Admin Service (admin-service/internal/handler/admin_handler.go):

   claims := auth.Claims{
       UserID:     admin.ID.String(),
       UserType:   "admin",
       TenantID:   admin.TenantID.String(),
       RegisteredClaims: jwt.RegisteredClaims{
           Issuer:    "admin-portal",  // 必须与 Kong Consumer key 一致
           ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
           IssuedAt:  jwt.NewNumericDate(time.Now()),
       },
   }

2. Merchant Service (merchant-service/internal/handler/merchant_handler.go):

   claims := auth.Claims{
       UserID:     merchant.ID.String(),
       UserType:   "merchant",
       TenantID:   merchant.ID.String(),
       RegisteredClaims: jwt.RegisteredClaims{
           Issuer:    "merchant-portal",  // 必须与 Kong Consumer key 一致
           ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
           IssuedAt:  jwt.NewNumericDate(time.Now()),
       },
   }

3. 环境变量:
   确保后端服务的 JWT_SECRET 与此脚本使用的值一致:

   export JWT_SECRET="your-secret-key"

EOF

echo ""
log_info "测试 JWT 认证:"
echo ""
echo "  1. 启动 admin-service 和 merchant-service"
echo "  2. 通过 Kong 调用登录接口:"
echo ""
echo "     curl -X POST http://localhost:40080/api/v1/admin/login \\"
echo "       -H \"Content-Type: application/json\" \\"
echo "       -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"
echo ""
echo "  3. 使用返回的 JWT token 访问受保护资源:"
echo ""
echo "     curl -X GET http://localhost:40080/api/v1/admin \\"
echo "       -H \"Authorization: Bearer <your-jwt-token>\""
echo ""
echo "=========================================="
echo ""
