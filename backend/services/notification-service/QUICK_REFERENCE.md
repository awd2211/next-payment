# Notification Service 优化功能快速参考

## 🚀 快速开始

### 启动优化后的服务
```bash
cd /home/eric/payment/backend/services/notification-service

# 运行优化版本
DB_HOST=localhost DB_PORT=40432 DB_USER=postgres DB_PASSWORD=postgres \
DB_NAME=payment_notify REDIS_HOST=localhost REDIS_PORT=40379 \
JWT_SECRET=your-secret-key PORT=8007 \
/tmp/notification-service-optimized
```

---

## ✨ 新功能1：高级模板语法

### 使用条件判断
```html
<!-- 模板内容 -->
{{if .is_vip}}
    <div class="vip-section">
        <h3>尊贵的VIP用户</h3>
        <p>专享{{.discount}}%折扣</p>
    </div>
{{else}}
    <p>普通用户</p>
{{end}}
```

### 使用循环
```html
<!-- 遍历订单商品 -->
<table>
{{range .items}}
    <tr>
        <td>{{.name}}</td>
        <td>数量：{{.quantity}}</td>
        <td>{{formatMoney .price "USD"}}</td>
    </tr>
{{end}}
</table>
```

### 使用自定义函数
```html
<!-- 格式化金额（分 -> 美元） -->
<p>总价：{{formatMoney .total_amount "USD"}}</p>
<!-- 输入：10000，输出：$100.00 -->

<!-- 格式化日期 -->
<p>交易时间：{{formatDate .created_at}}</p>
<!-- 输出：2024-10-23 12:30:45 -->

<!-- 字符串大小写转换 -->
<p>商户：{{upper .merchant_name}}</p>
<p>状态：{{lower .status}}</p>
```

### 支持的货币符号
| 货币代码 | 符号 |
|---------|------|
| USD | $ |
| EUR | € |
| GBP | £ |
| JPY | ¥ |
| CNY | ¥ |
| HKD | HK$ |

---

## 🔔 新功能2：通知偏好管理

### 常用操作

#### 1. 关闭某类型的邮件通知
```bash
curl -X POST http://localhost:8007/api/v1/preferences \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "email",
    "event_type": "payment.success",
    "is_enabled": false,
    "description": "不想收到支付成功邮件"
  }'
```

#### 2. 查看我的所有偏好
```bash
curl -X GET http://localhost:8007/api/v1/preferences \
  -H "Authorization: Bearer $TOKEN"
```

#### 3. 开启某类型通知
```bash
# 先查询偏好ID
curl -X GET http://localhost:8007/api/v1/preferences \
  -H "Authorization: Bearer $TOKEN"

# 更新偏好
curl -X PUT http://localhost:8007/api/v1/preferences/{preference_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_enabled": true
  }'
```

#### 4. 删除偏好（恢复默认）
```bash
curl -X DELETE http://localhost:8007/api/v1/preferences/{preference_id} \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📋 支持的事件类型

### 账户相关
- `merchant.registered` - 商户注册
- `kyc.approved` - KYC审核通过
- `kyc.rejected` - KYC审核拒绝
- `merchant.frozen` - 商户冻结
- `password.reset` - 密码重置

### 交易相关
- `payment.success` - 支付成功
- `payment.failed` - 支付失败
- `refund.completed` - 退款完成
- `order.created` - 订单创建
- `order.cancelled` - 订单取消

### 财务相关
- `settlement.completed` - 结算完成

### 系统相关
- `system.maintenance` - 系统维护

---

## 💡 实用示例

### 示例1：创建VIP用户专属模板
```bash
curl -X POST http://localhost:8007/api/v1/templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "vip_welcome",
    "name": "VIP欢迎邮件",
    "type": "account",
    "channel": "email",
    "subject": "欢迎尊贵的VIP用户 - {{.username}}",
    "content": "<html><body><h1>{{upper .username}}</h1>{{if .is_vip}}<p>您的VIP等级：{{.vip_level}}</p><p>专属折扣：{{.discount}}%</p>{{end}}<p>账户余额：{{formatMoney .balance \"USD\"}}</p></body></html>",
    "variables": "[\"username\", \"is_vip\", \"vip_level\", \"discount\", \"balance\"]",
    "is_enabled": true
  }'
```

### 示例2：发送使用高级语法的邮件
```bash
curl -X POST http://localhost:8007/api/v1/notifications/email/template \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["vip@example.com"],
    "template_code": "vip_welcome",
    "template_data": {
      "username": "John Doe",
      "is_vip": true,
      "vip_level": "Gold",
      "discount": 20,
      "balance": 50000
    },
    "provider": "smtp",
    "event_type": "merchant.registered"
  }'
```

### 示例3：设置营销通知偏好
```bash
# 商户想关闭所有营销类通知
curl -X POST http://localhost:8007/api/v1/preferences \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "email",
    "event_type": "marketing.*",
    "is_enabled": false,
    "description": "不接收任何营销邮件"
  }'

curl -X POST http://localhost:8007/api/v1/preferences \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "sms",
    "event_type": "marketing.*",
    "is_enabled": false,
    "description": "不接收任何营销短信"
  }'
```

---

## 🔍 故障排查

### 问题1：模板解析失败
**症状**：邮件显示`{{variable}}`而不是实际值

**原因**：模板语法错误

**解决**：
1. 检查模板语法是否正确
2. 确保变量名与`template_data`中的键匹配
3. 查看日志确认是否有解析错误

**注意**：模板解析失败会自动降级到简单替换，不会导致发送失败

### 问题2：偏好设置不生效
**症状**：设置了禁用但仍然收到通知

**原因**：发送请求中未指定`event_type`

**解决**：
在发送通知时必须指定`event_type`参数：
```json
{
  "to": ["user@example.com"],
  "subject": "测试",
  "event_type": "payment.success"  // 必须指定
}
```

### 问题3：自定义函数不work
**症状**：`formatMoney`等函数不起作用

**原因**：可能是模板解析失败降级了

**解决**：
1. 确保使用正确的函数名
2. 检查参数类型（amount必须是int64，currency必须是string）
3. 查看服务日志

---

## 📖 最佳实践

### 1. 模板设计
- ✅ **DO**：使用条件判断区分用户类型
- ✅ **DO**：使用formatMoney显示金额
- ✅ **DO**：在模板中添加降级处理（如：`{{.name | default "用户"}}`）
- ❌ **DON'T**：不要在模板中包含敏感信息（如密码、token）

### 2. 偏好管理
- ✅ **DO**：为每个事件类型单独设置偏好
- ✅ **DO**：提供用户界面让用户自行管理
- ✅ **DO**：发送通知时始终指定event_type
- ❌ **DON'T**：不要全局禁用所有通知（可能错过重要安全通知）

### 3. 性能优化
- ✅ **DO**：为常用模板启用缓存（计划中）
- ✅ **DO**：使用批量API发送大量通知（计划中）
- ✅ **DO**：监控发送成功率
- ❌ **DON'T**：不要同步发送大量邮件（会阻塞API）

---

## 🆕 API变更总结

### 新增API（5个）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/preferences | 创建偏好 |
| GET | /api/v1/preferences/:id | 获取偏好详情 |
| GET | /api/v1/preferences | 列出偏好 |
| PUT | /api/v1/preferences/:id | 更新偏好 |
| DELETE | /api/v1/preferences/:id | 删除偏好 |

### 修改的API
发送邮件/短信API新增可选字段：
- `user_id`: UUID（用于偏好检查）
- `event_type`: string（用于偏好检查）

**向后兼容**：不提供这些字段时，偏好检查被跳过

---

## 📞 获取帮助

- 完整文档：`NOTIFICATION_SERVICE_GUIDE.md`
- 优化总结：`OPTIMIZATION_SUMMARY.md`
- Swagger文档：http://localhost:8007/swagger/index.html
- 健康检查：http://localhost:8007/health

---

**优化版本**：v2.0.0
**编译时间**：2024-10-23
**编译文件**：`/tmp/notification-service-optimized`
