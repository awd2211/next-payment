#!/bin/bash

# 幂等性测试脚本
# 测试相同的 Idempotency-Key 是否返回缓存的响应

set -e

echo "========================================="
echo "幂等性测试脚本"
echo "========================================="
echo ""

# 生成唯一的幂等性Key
IDEMPOTENCY_KEY="test-$(date +%s)-$(uuidgen)"
echo "生成幂等性Key: $IDEMPOTENCY_KEY"
echo ""

# 测试数据
ORDER_NO="ORDER-$(date +%s)"
AMOUNT=10000  # 100.00 USD

echo "测试数据:"
echo "  订单号: $ORDER_NO"
echo "  金额: $AMOUNT (100.00 USD)"
echo ""

# 第一次请求 - 应该处理并创建支付
echo "========================================="
echo "第一次请求 (应该创建新支付)"
echo "========================================="

RESPONSE1=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:40003/api/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -H "X-Signature: test-signature" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d "{
    \"order_no\": \"$ORDER_NO\",
    \"amount\": $AMOUNT,
    \"currency\": \"USD\",
    \"channel\": \"stripe\",
    \"subject\": \"测试支付\",
    \"body\": \"幂等性测试\",
    \"callback_url\": \"http://localhost:8080/callback\",
    \"return_url\": \"http://localhost:8080/return\"
  }")

HTTP_CODE1=$(echo "$RESPONSE1" | tail -n 1)
BODY1=$(echo "$RESPONSE1" | sed '$d')

echo "HTTP状态码: $HTTP_CODE1"
echo "响应体:"
echo "$BODY1" | jq . 2>/dev/null || echo "$BODY1"
echo ""

# 等待2秒确保第一次请求完成
echo "等待2秒..."
sleep 2

# 第二次请求 - 应该返回缓存的响应
echo "========================================="
echo "第二次请求 (应该返回缓存响应)"
echo "========================================="

RESPONSE2=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:40003/api/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -H "X-Signature: test-signature" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d "{
    \"order_no\": \"$ORDER_NO\",
    \"amount\": $AMOUNT,
    \"currency\": \"USD\",
    \"channel\": \"stripe\",
    \"subject\": \"测试支付\",
    \"body\": \"幂等性测试\",
    \"callback_url\": \"http://localhost:8080/callback\",
    \"return_url\": \"http://localhost:8080/return\"
  }")

HTTP_CODE2=$(echo "$RESPONSE2" | tail -n 1)
BODY2=$(echo "$RESPONSE2" | sed '$d')

echo "HTTP状态码: $HTTP_CODE2"
echo "响应体:"
echo "$BODY2" | jq . 2>/dev/null || echo "$BODY2"
echo ""

# 比较两次响应
echo "========================================="
echo "结果验证"
echo "========================================="

if [ "$BODY1" == "$BODY2" ]; then
    echo "✅ 幂等性测试通过: 两次请求返回相同的响应"
else
    echo "❌ 幂等性测试失败: 两次请求返回不同的响应"
    echo ""
    echo "第一次响应:"
    echo "$BODY1"
    echo ""
    echo "第二次响应:"
    echo "$BODY2"
fi

echo ""
echo "========================================="
echo "测试完成"
echo "========================================="
