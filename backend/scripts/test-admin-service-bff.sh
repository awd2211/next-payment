#!/bin/bash

# Admin Service BFF 测试脚本
# 用途: 测试所有 BFF 路由是否正常工作

echo "=================================================="
echo "  Admin Service BFF 接口测试"
echo "=================================================="

# Admin Service 地址
ADMIN_SERVICE="http://localhost:40001"

echo ""
echo "步骤 1: 管理员登录获取 Token"
echo "=================================================="

# 登录获取 token
LOGIN_RESPONSE=$(curl -s -X POST "${ADMIN_SERVICE}/api/v1/admin/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo "登录响应:"
echo "$LOGIN_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$LOGIN_RESPONSE"

# 提取 token
TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token', ''))" 2>/dev/null)

if [ -z "$TOKEN" ]; then
    echo ""
    echo "❌ 登录失败,无法获取 token"
    echo "请确保:"
    echo "  1. Admin Service 已启动 (端口 40001)"
    echo "  2. 数据库中已有 admin 用户"
    echo "  3. 密码正确"
    exit 1
fi

echo ""
echo "✅ 登录成功,Token: ${TOKEN:0:50}..."
echo ""

# 等待用户确认
echo "按 Enter 继续测试 BFF 接口..."
read

# ========================================
# 测试 Config Service BFF
# ========================================
echo ""
echo "步骤 2: 测试 Config Service BFF"
echo "=================================================="

echo "2.1 获取配置列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/configs?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "2.2 获取功能开关列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/feature-flags?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "2.3 获取服务注册列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/services?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

# ========================================
# 测试 Risk Service BFF
# ========================================
echo ""
echo "步骤 3: 测试 Risk Service BFF"
echo "=================================================="

echo "3.1 获取风控规则列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/risk/rules?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "3.2 获取黑名单列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/risk/blacklist?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "3.3 获取风控检查记录"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/risk/checks?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

# ========================================
# 测试 KYC Service BFF
# ========================================
echo ""
echo "步骤 4: 测试 KYC Service BFF"
echo "=================================================="

echo "4.1 获取所有 KYC 文档"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/kyc/documents?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "4.2 获取待审核文档"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/kyc/documents/pending?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "4.3 获取资质列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/kyc/qualifications?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "4.4 获取等级统计"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/kyc/levels/statistics" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

# ========================================
# 测试 Merchant Limit Service BFF
# ========================================
echo ""
echo "步骤 5: 测试 Merchant Limit Service BFF"
echo "=================================================="

echo "5.1 获取 Tier 列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/merchant-tiers?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "5.2 获取所有商户限额"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/merchant-limits?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

# ========================================
# 测试 Merchant Service BFF
# ========================================
echo ""
echo "步骤 6: 测试 Merchant Service BFF"
echo "=================================================="

echo "6.1 获取商户列表"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/merchants?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "6.2 获取商户统计"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/merchants/statistics?period=today" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

# ========================================
# 测试 Analytics Service BFF
# ========================================
echo ""
echo "步骤 7: 测试 Analytics Service BFF"
echo "=================================================="

echo "7.1 获取 Dashboard 数据"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/analytics/dashboard?period=today" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "7.2 获取平台总览"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/analytics/platform/overview?period=7days" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "7.3 获取支付统计"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/analytics/payments/statistics?period=today" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

echo ""
echo "7.4 获取商户统计"
curl -s -X GET "${ADMIN_SERVICE}/api/v1/admin/analytics/merchants/statistics?period=today" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null || echo "请求失败"

# ========================================
# 测试总结
# ========================================
echo ""
echo "=================================================="
echo "  测试完成!"
echo "=================================================="
echo ""
echo "已测试的 BFF Handler:"
echo "  ✅ ConfigBFF       - /api/v1/admin/configs, /feature-flags, /services"
echo "  ✅ RiskBFF         - /api/v1/admin/risk/rules, /blacklist, /checks"
echo "  ✅ KYCBFF          - /api/v1/admin/kyc/documents, /qualifications, /levels"
echo "  ✅ LimitBFF        - /api/v1/admin/merchant-tiers, /merchant-limits"
echo "  ✅ MerchantBFF     - /api/v1/admin/merchants"
echo "  ✅ AnalyticsBFF    - /api/v1/admin/analytics/*"
echo ""
echo "如果所有接口都返回了 JSON 响应(而不是错误),说明 BFF 工作正常!"
echo ""
