# 订单服务设计文档

## 概述

Order Service是订单管理的核心服务，负责订单创建、状态管理、统计分析等功能，与Payment Gateway和Merchant Service紧密配合。

## 核心功能

### 1. 订单管理

#### 创建订单

**API：** `POST /api/v1/orders`

**请求示例：**
```json
{
  "merchant_id": "uuid",
  "customer_email": "customer@example.com",
  "customer_name": "John Doe",
  "customer_phone": "+1234567890",
  "customer_ip": "192.168.1.100",
  "currency": "USD",
  "language": "en",
  "items": [
    {
      "product_id": "PROD001",
      "product_name": "Product A",
      "product_sku": "SKU-A-001",
      "product_image": "https://cdn.example.com/product-a.jpg",
      "unit_price": 5000,
      "quantity": 2,
      "attributes": {
        "color": "red",
        "size": "L"
      }
    }
  ],
  "shipping_method": "express",
  "shipping_fee": 500,
  "shipping_address": {
    "country": "US",
    "province": "CA",
    "city": "San Francisco",
    "street": "123 Market St",
    "postal_code": "94102",
    "phone": "+1234567890",
    "name": "John Doe"
  },
  "discount_amount": 1000,
  "remark": "Please deliver before 5pm",
  "expire_minutes": 30
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "order_no": "OD202401151234567890ABC",
    "merchant_id": "uuid",
    "total_amount": 10500,
    "pay_amount": 9500,
    "discount_amount": 1000,
    "shipping_fee": 500,
    "currency": "USD",
    "status": "pending",
    "pay_status": "pending",
    "shipping_status": "pending",
    "customer_email": "customer@example.com",
    "customer_name": "John Doe",
    "expired_at": "2024-01-15T12:30:00Z",
    "created_at": "2024-01-15T12:00:00Z",
    "items": [
      {
        "product_id": "PROD001",
        "product_name": "Product A",
        "unit_price": 5000,
        "quantity": 2,
        "total_price": 10000
      }
    ]
  }
}
```

#### 查询订单

**API：** `GET /api/v1/orders/{order_no}`

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "order_no": "OD202401151234567890ABC",
    "payment_no": "PY202401151234567890XYZ",
    "total_amount": 10500,
    "pay_amount": 9500,
    "currency": "USD",
    "status": "paid",
    "pay_status": "paid",
    "shipping_status": "pending",
    "customer_email": "customer@example.com",
    "customer_name": "John Doe",
    "paid_at": "2024-01-15T12:05:00Z",
    "created_at": "2024-01-15T12:00:00Z",
    "items": [...]
  }
}
```

---

### 2. 订单状态管理

#### 订单状态流转

```
pending → paid → processing → shipped → completed
            ↓
        cancelled
            ↓
        refunded
```

#### 订单状态说明

| 状态 | 说明 | 可执行操作 |
|-----|------|-----------|
| pending | 待支付 | 取消、支付 |
| paid | 已支付 | 发货、退款 |
| processing | 处理中 | 发货 |
| shipped | 已发货 | 确认收货 |
| completed | 已完成 | 售后 |
| cancelled | 已取消 | 无 |
| refunded | 已退款 | 无 |
| expired | 已过期 | 重新下单 |

#### 支付状态

| 状态 | 说明 |
|-----|------|
| pending | 待支付 |
| paid | 已支付 |
| failed | 支付失败 |
| refunded | 已退款 |
| partial_refunded | 部分退款 |

#### 配送状态

| 状态 | 说明 |
|-----|------|
| pending | 待发货 |
| preparing | 备货中 |
| shipped | 已发货 |
| in_transit | 运输中 |
| delivered | 已送达 |
| returned | 已退货 |

---

### 3. 订单操作

#### 取消订单

**API：** `POST /api/v1/orders/{order_no}/cancel`

```json
{
  "reason": "Customer requested cancellation"
}
```

**限制：**
- 只有pending或paid（未发货）状态可以取消
- 已支付订单取消需要先退款

#### 支付订单（内部调用）

**API：** `POST /api/v1/orders/{order_no}/pay`

```json
{
  "payment_no": "PY202401151234567890XYZ"
}
```

**说明：**
- 由Payment Gateway在支付成功后调用
- 更新订单状态为paid
- 更新支付状态为paid
- 记录支付流水号

#### 退款订单

**API：** `POST /api/v1/orders/{order_no}/refund`

```json
{
  "amount": 9500,
  "reason": "Product quality issue"
}
```

**支持：**
- 全额退款
- 部分退款
- 多次退款（累计不超过支付金额）

#### 发货

**API：** `POST /api/v1/orders/{order_no}/ship`

```json
{
  "tracking_no": "SF1234567890",
  "carrier": "SF Express",
  "estimated_delivery": "2024-01-20"
}
```

**限制：**
- 只有paid状态的订单可以发货

#### 完成订单

**API：** `POST /api/v1/orders/{order_no}/complete`

**限制：**
- 只有shipped状态的订单可以完成

---

### 4. 订单查询

#### 订单列表

**API：** `GET /api/v1/orders`

**查询参数：**
```
merchant_id: 商户ID（必填）
customer_email: 客户邮箱
status: 订单状态
pay_status: 支付状态
shipping_status: 配送状态
currency: 货币类型
start_time: 开始时间
end_time: 结束时间
min_amount: 最小金额
max_amount: 最大金额
keyword: 关键词（订单号、客户姓名、邮箱）
page: 页码（默认1）
page_size: 每页数量（默认20，最大100）
```

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "order_no": "OD202401151234567890ABC",
        "total_amount": 10500,
        "currency": "USD",
        "status": "paid",
        "customer_email": "customer@example.com",
        "created_at": "2024-01-15T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 5. 订单统计

#### 每日汇总

**API：** `GET /api/v1/orders/statistics/daily-summary`

**查询参数：**
```
merchant_id: 商户ID
date: 日期（YYYY-MM-DD）
```

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "total_orders": 150,
    "paid_orders": 120,
    "cancelled_orders": 20,
    "total_amount": 1500000,
    "paid_amount": 1200000
  }
}
```

#### 时间范围统计

**API：** `GET /api/v1/orders/statistics/range`

**查询参数：**
```
merchant_id: 商户ID
start_date: 开始日期
end_date: 结束日期
currency: 货币类型（可选）
```

**响应示例：**
```json
{
  "code": 0,
  "data": [
    {
      "stat_date": "2024-01-15",
      "currency": "USD",
      "total_orders": 50,
      "paid_orders": 40,
      "cancelled_orders": 5,
      "total_amount": 500000,
      "paid_amount": 400000,
      "refund_amount": 10000,
      "avg_order_amount": 10000
    }
  ]
}
```

---

### 6. 订单日志

每个订单操作都会自动记录日志，包括：
- 订单创建
- 状态变更
- 支付成功
- 发货
- 退款
- 取消

**API：** `GET /api/v1/orders/{order_no}/logs`

**响应示例：**
```json
{
  "code": 0,
  "data": [
    {
      "action": "pay",
      "old_status": "pending",
      "new_status": "paid",
      "operator_type": "system",
      "remark": "支付成功，支付流水号: PY202401151234567890XYZ",
      "created_at": "2024-01-15T12:05:00Z"
    },
    {
      "action": "create",
      "old_status": "",
      "new_status": "pending",
      "operator_type": "system",
      "remark": "订单创建",
      "created_at": "2024-01-15T12:00:00Z"
    }
  ]
}
```

---

## 数据模型

### orders（订单表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| merchant_id | UUID | 商户ID |
| order_no | VARCHAR(64) | 订单号（唯一） |
| payment_no | VARCHAR(64) | 支付流水号 |
| total_amount | BIGINT | 订单总金额（分） |
| pay_amount | BIGINT | 实付金额（分） |
| discount_amount | BIGINT | 优惠金额（分） |
| shipping_fee | BIGINT | 运费（分） |
| currency | VARCHAR(10) | 货币类型 |
| status | VARCHAR(20) | 订单状态 |
| pay_status | VARCHAR(20) | 支付状态 |
| shipping_status | VARCHAR(20) | 配送状态 |
| customer_id | UUID | 客户ID |
| customer_email | VARCHAR(255) | 客户邮箱 |
| customer_name | VARCHAR(100) | 客户姓名 |
| customer_phone | VARCHAR(20) | 客户手机 |
| customer_ip | VARCHAR(50) | 客户IP |
| shipping_method | VARCHAR(50) | 配送方式 |
| shipping_address | JSONB | 配送地址 |
| billing_address | JSONB | 账单地址 |
| remark | TEXT | 备注 |
| extra | JSONB | 扩展信息 |
| language | VARCHAR(10) | 语言 |
| paid_at | TIMESTAMPTZ | 支付时间 |
| shipped_at | TIMESTAMPTZ | 发货时间 |
| completed_at | TIMESTAMPTZ | 完成时间 |
| cancelled_at | TIMESTAMPTZ | 取消时间 |
| expired_at | TIMESTAMPTZ | 过期时间 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

**索引：**
- order_no（唯一）
- merchant_id
- customer_email
- payment_no
- status
- pay_status
- created_at

### order_items（订单项表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| order_id | UUID | 订单ID |
| product_id | VARCHAR(64) | 商品ID |
| product_name | VARCHAR(200) | 商品名称 |
| product_sku | VARCHAR(100) | 商品SKU |
| product_image | VARCHAR(500) | 商品图片 |
| unit_price | BIGINT | 单价（分） |
| quantity | INTEGER | 数量 |
| total_price | BIGINT | 小计（分） |
| discount_price | BIGINT | 优惠金额（分） |
| attributes | JSONB | 商品属性 |
| extra | JSONB | 扩展信息 |
| created_at | TIMESTAMPTZ | 创建时间 |

### order_logs（订单日志表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| order_id | UUID | 订单ID |
| action | VARCHAR(50) | 操作类型 |
| old_status | VARCHAR(20) | 旧状态 |
| new_status | VARCHAR(20) | 新状态 |
| operator_id | UUID | 操作人ID |
| operator_type | VARCHAR(20) | 操作人类型 |
| remark | TEXT | 备注 |
| extra | JSONB | 扩展信息 |
| created_at | TIMESTAMPTZ | 创建时间 |

### order_statistics（订单统计表）

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| merchant_id | UUID | 商户ID |
| stat_date | DATE | 统计日期 |
| currency | VARCHAR(10) | 货币类型 |
| total_orders | INTEGER | 订单总数 |
| paid_orders | INTEGER | 已支付订单数 |
| cancelled_orders | INTEGER | 已取消订单数 |
| total_amount | BIGINT | 订单总金额 |
| paid_amount | BIGINT | 已支付金额 |
| refund_amount | BIGINT | 退款金额 |
| avg_order_amount | BIGINT | 平均订单金额 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

**唯一索引：** (merchant_id, stat_date)

---

## 业务流程

### 标准下单流程

```
1. 商户调用创建订单API
2. Order Service生成订单号，保存订单
3. 返回订单信息给商户
4. 商户引导用户到Payment Gateway支付
5. Payment Gateway支付成功后回调Order Service
6. Order Service更新订单状态为paid
7. 商户发货，调用发货API
8. Order Service更新配送状态
9. 用户确认收货，订单完成
```

### 订单取消流程

```
1. 未支付订单：
   - 直接取消，更新状态为cancelled

2. 已支付订单：
   - 先申请退款
   - 退款成功后，更新订单状态为refunded
```

### 订单退款流程

```
1. 商户/用户发起退款请求
2. Order Service验证订单状态
3. 调用Payment Gateway退款API
4. Payment Gateway执行退款
5. 退款成功后回调Order Service
6. Order Service更新支付状态
7. 全额退款：订单状态变为refunded
8. 部分退款：支付状态变为partial_refunded
```

---

## 与其他服务的交互

### 与Payment Gateway的交互

**Order → Payment：**
- 支付前查询订单信息
- 验证订单金额

**Payment → Order：**
- 支付成功回调，更新订单状态
- 退款成功回调，更新退款状态

### 与Merchant Service的交互

**Order → Merchant：**
- 验证商户状态
- 获取商户配置

### 与Notification Service的交互

**Order → Notification：**
- 订单创建通知
- 支付成功通知
- 发货通知
- 退款通知

---

## API完整列表

### 订单管理
- `POST /api/v1/orders` - 创建订单
- `GET /api/v1/orders/{order_no}` - 查询订单
- `GET /api/v1/orders` - 订单列表
- `POST /api/v1/orders/{order_no}/cancel` - 取消订单

### 订单操作
- `POST /api/v1/orders/{order_no}/pay` - 支付订单（内部）
- `POST /api/v1/orders/{order_no}/refund` - 退款订单
- `POST /api/v1/orders/{order_no}/ship` - 发货
- `POST /api/v1/orders/{order_no}/complete` - 完成订单

### 订单日志
- `GET /api/v1/orders/{order_no}/logs` - 订单日志

### 订单统计
- `GET /api/v1/orders/statistics/daily-summary` - 每日汇总
- `GET /api/v1/orders/statistics/range` - 时间范围统计

---

## 技术栈

- **语言**：Go 1.21+
- **框架**：Gin（HTTP）
- **数据库**：PostgreSQL
- **缓存**：Redis
- **消息队列**：Kafka

---

## 扩展功能

### 计划中
- [ ] 订单导出（Excel, CSV）
- [ ] 订单搜索优化（ElasticSearch）
- [ ] 订单标签系统
- [ ] 订单备注和附件
- [ ] 订单分配和调度

### 未来
- [ ] 订单预测和推荐
- [ ] 智能库存管理
- [ ] 物流追踪集成
- [ ] 订单评价系统
