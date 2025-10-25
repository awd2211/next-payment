# 全球化支付平台 API 接口设计

本文档定义全球化方案所需的所有 RESTful API 接口，包括对账系统、拒付管理、商户额度管理等模块。

## 概览

**设计原则**:
- RESTful 风格
- 统一响应格式
- JWT 认证（管理后台）
- API Key 认证（商户 API）
- 分页查询标准化
- 完整的错误码定义

**新增服务 API**:
1. **reconciliation-service** (Port: 40016)
2. **dispute-service** (Port: 40017)
3. **merchant-limit-service** (Port: 40018)

**扩展现有服务 API**:
- merchant-service - 新增商户等级管理接口
- payment-gateway - 新增超时配置接口

---

## 通用约定

### 响应格式

**成功响应**:
```json
{
  "code": "SUCCESS",
  "message": "操作成功",
  "data": { ... },
  "trace_id": "trace-abc123"
}
```

**错误响应**:
```json
{
  "code": "ERROR_CODE",
  "message": "错误描述",
  "details": "详细错误信息",
  "trace_id": "trace-abc123"
}
```

### 分页参数

**请求参数**:
```json
{
  "page": 1,
  "page_size": 20
}
```

**分页响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

### 认证方式

| 服务 | 认证方式 | Header |
|------|---------|--------|
| Admin API | JWT Token | `Authorization: Bearer {token}` |
| Merchant API | API Key | `X-Api-Key: {api_key}`, `X-Signature: {signature}` |
| Internal API | mTLS (可选) | 证书认证 |

---

## 一、对账系统 API (reconciliation-service)

**Base URL**: `http://localhost:40016/api/v1`

### 1.1 对账任务管理

#### POST /reconciliation/tasks

创建对账任务（手动触发）

**认证**: JWT (Admin)

**请求**:
```json
{
  "task_date": "2025-01-24",
  "channel": "stripe",
  "task_type": "manual"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "对账任务创建成功",
  "data": {
    "task_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "task_no": "REC20250124001",
    "task_date": "2025-01-24",
    "channel": "stripe",
    "status": "pending"
  }
}
```

---

#### GET /reconciliation/tasks

查询对账任务列表

**认证**: JWT (Admin)

**查询参数**:
```
?task_date=2025-01-24
&channel=stripe
&status=completed
&page=1
&page_size=20
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [
      {
        "task_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "task_no": "REC20250124001",
        "task_date": "2025-01-24",
        "channel": "stripe",
        "status": "completed",
        "platform_count": 1234,
        "platform_amount": 567890000,
        "channel_count": 1230,
        "channel_amount": 567800000,
        "matched_count": 1228,
        "matched_amount": 567500000,
        "diff_count": 6,
        "diff_amount": 390000,
        "progress": 100,
        "started_at": "2025-01-24T02:00:00Z",
        "completed_at": "2025-01-24T02:15:30Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 45,
      "total_pages": 3
    }
  }
}
```

---

#### GET /reconciliation/tasks/:task_id

查询对账任务详情

**认证**: JWT (Admin)

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "task_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "task_no": "REC20250124001",
    "task_date": "2025-01-24",
    "channel": "stripe",
    "task_type": "daily",
    "status": "completed",
    "progress": 100,

    "statistics": {
      "platform_count": 1234,
      "platform_amount": 567890000,
      "channel_count": 1230,
      "channel_amount": 567800000,
      "matched_count": 1228,
      "matched_amount": 567500000,
      "diff_count": 6,
      "diff_amount": 390000
    },

    "files": {
      "channel_file_url": "https://storage.example.com/settlements/stripe/2025-01-24.csv",
      "report_file_url": "https://storage.example.com/reports/REC20250124001.pdf"
    },

    "started_at": "2025-01-24T02:00:00Z",
    "completed_at": "2025-01-24T02:15:30Z",
    "created_at": "2025-01-24T02:00:00Z"
  }
}
```

---

#### POST /reconciliation/tasks/:task_id/retry

重试失败的对账任务

**认证**: JWT (Admin)

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "对账任务已重新启动",
  "data": {
    "task_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "status": "processing"
  }
}
```

---

### 1.2 对账差异管理

#### GET /reconciliation/records

查询对账差异记录

**认证**: JWT (Admin)

**查询参数**:
```
?task_id=f47ac10b-58cc-4372-a567-0e02b2c3d479
&diff_type=amount_diff
&is_resolved=false
&merchant_id=e55feb66-16f9-41be-a68b-a8961df898b6
&page=1
&page_size=20
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [
      {
        "record_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "task_no": "REC20250124001",
        "payment_no": "PY20250124123456abcdefgh",
        "channel_trade_no": "pi_3QdVg42eZvKYlo2C0H7jSI9x",
        "order_no": "ORDER-12345",
        "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",

        "amounts": {
          "platform_amount": 50000,
          "channel_amount": 49500,
          "diff_amount": 500,
          "currency": "USD"
        },

        "status": {
          "platform_status": "success",
          "channel_status": "succeeded"
        },

        "diff_info": {
          "diff_type": "amount_diff",
          "diff_reason": "平台金额与渠道金额不一致"
        },

        "is_resolved": false,
        "created_at": "2025-01-24T02:05:30Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 6,
      "total_pages": 1
    }
  }
}
```

---

#### POST /reconciliation/records/:record_id/resolve

标记差异已解决

**认证**: JWT (Admin)

**请求**:
```json
{
  "resolution_note": "已确认为 Stripe 手续费扣除，差异正常"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "差异已标记为已解决",
  "data": {
    "record_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "is_resolved": true,
    "resolved_by": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "resolved_at": "2025-01-24T10:30:00Z"
  }
}
```

---

### 1.3 渠道账单管理

#### GET /reconciliation/settlement-files

查询渠道账单文件列表

**认证**: JWT (Admin)

**查询参数**:
```
?channel=stripe
&settlement_date=2025-01-24
&status=imported
&page=1
&page_size=20
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [
      {
        "file_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
        "file_no": "SF-STRIPE-20250124",
        "channel": "stripe",
        "settlement_date": "2025-01-24",
        "file_url": "https://storage.example.com/settlements/stripe/2025-01-24.csv",
        "file_size": 1048576,
        "file_hash": "a1b2c3d4e5f6...",
        "record_count": 1230,
        "total_amount": 567800000,
        "currency": "USD",
        "status": "imported",
        "downloaded_at": "2025-01-24T01:00:00Z",
        "parsed_at": "2025-01-24T01:05:00Z",
        "imported_at": "2025-01-24T01:10:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 30,
      "total_pages": 2
    }
  }
}
```

---

#### POST /reconciliation/settlement-files/download

手动触发账单下载

**认证**: JWT (Admin)

**请求**:
```json
{
  "channel": "stripe",
  "settlement_date": "2025-01-24"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "账单下载任务已创建",
  "data": {
    "file_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "status": "pending"
  }
}
```

---

### 1.4 对账报表

#### GET /reconciliation/reports/:task_id

下载对账报表

**认证**: JWT (Admin)

**响应**: PDF 文件流

**报表内容**:
- 对账任务概览
- 匹配统计（成功率、差异率）
- 差异明细列表（按类型分组）
- 建议处理措施

---

## 二、拒付管理 API (dispute-service)

**Base URL**: `http://localhost:40017/api/v1`

### 2.1 拒付工单管理

#### GET /disputes

查询拒付工单列表

**认证**: JWT (Admin) 或 Merchant API Key

**查询参数**:
```
?merchant_id=e55feb66-16f9-41be-a68b-a8961df898b6  # Admin 可选，Merchant 自动填充
&status=needs_response
&channel=stripe
&start_date=2025-01-01
&end_date=2025-01-31
&page=1
&page_size=20
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [
      {
        "dispute_id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
        "dispute_no": "DIS20250124001",
        "channel": "stripe",
        "channel_dispute_id": "dp_1QdVg42eZvKYlo2C",

        "payment_info": {
          "payment_no": "PY20250124123456abcdefgh",
          "order_no": "ORDER-12345",
          "amount": 50000,
          "currency": "USD"
        },

        "dispute_info": {
          "reason": "fraudulent",
          "status": "needs_response",
          "evidence_due_by": "2025-01-31T23:59:59Z"
        },

        "evidence": {
          "has_evidence": false,
          "evidence_count": 0
        },

        "created_at": "2025-01-24T10:00:00Z",
        "updated_at": "2025-01-24T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 3,
      "total_pages": 1
    },
    "summary": {
      "total_disputes": 3,
      "needs_response": 2,
      "under_review": 1,
      "total_amount": 125000
    }
  }
}
```

---

#### GET /disputes/:dispute_id

查询拒付工单详情

**认证**: JWT (Admin) 或 Merchant API Key

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "dispute_id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
    "dispute_no": "DIS20250124001",
    "channel": "stripe",
    "channel_dispute_id": "dp_1QdVg42eZvKYlo2C",

    "payment_info": {
      "payment_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "payment_no": "PY20250124123456abcdefgh",
      "order_no": "ORDER-12345",
      "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
      "amount": 50000,
      "currency": "USD"
    },

    "dispute_info": {
      "reason": "fraudulent",
      "reason_description": "客户声称交易未授权",
      "status": "needs_response",
      "evidence_due_by": "2025-01-31T23:59:59Z"
    },

    "evidence": {
      "has_evidence": true,
      "evidence_count": 3,
      "evidence_submitted_at": "2025-01-25T14:30:00Z",
      "evidence_list": [
        {
          "evidence_id": "d4e5f6a7-b8c9-0123-def0-234567890123",
          "evidence_type": "invoice",
          "file_name": "invoice-ORDER-12345.pdf",
          "file_url": "https://storage.example.com/disputes/evidence/invoice-ORDER-12345.pdf",
          "description": "订单发票",
          "uploaded_at": "2025-01-25T14:30:00Z"
        }
      ]
    },

    "assignment": {
      "assigned_to": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "assigned_to_name": "张三"
    },

    "notes": {
      "internal_notes": "客户已确认收货，有签收记录",
      "merchant_notes": "订单已正常配送"
    },

    "timeline": [
      {
        "event_type": "created",
        "event_description": "拒付工单创建",
        "operator_type": "system",
        "created_at": "2025-01-24T10:00:00Z"
      },
      {
        "event_type": "evidence_submitted",
        "event_description": "商户提交证据",
        "operator_type": "merchant",
        "created_at": "2025-01-25T14:30:00Z"
      }
    ],

    "created_at": "2025-01-24T10:00:00Z",
    "updated_at": "2025-01-25T14:30:00Z"
  }
}
```

---

#### POST /disputes/:dispute_id/evidence

上传拒付证据

**认证**: JWT (Admin) 或 Merchant API Key

**请求** (multipart/form-data):
```
evidence_type: invoice
description: 订单发票
file: (binary)
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "证据上传成功",
  "data": {
    "evidence_id": "d4e5f6a7-b8c9-0123-def0-234567890123",
    "evidence_type": "invoice",
    "file_name": "invoice-ORDER-12345.pdf",
    "file_url": "https://storage.example.com/disputes/evidence/invoice-ORDER-12345.pdf",
    "file_size": 524288,
    "uploaded_at": "2025-01-25T14:30:00Z"
  }
}
```

---

#### POST /disputes/:dispute_id/submit

提交证据到渠道（Stripe/PayPal）

**认证**: JWT (Admin)

**请求**:
```json
{
  "evidence_details": {
    "customer_name": "John Doe",
    "customer_email_address": "customer@example.com",
    "billing_address": "123 Main St, New York, NY 10001",
    "receipt": "https://storage.example.com/disputes/evidence/receipt.pdf",
    "customer_signature": "https://storage.example.com/disputes/evidence/signature.jpg",
    "uncategorized_text": "Customer confirmed receipt via email on 2025-01-20"
  }
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "证据已提交到 Stripe",
  "data": {
    "dispute_id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
    "status": "under_review",
    "evidence_submitted_at": "2025-01-25T15:00:00Z"
  }
}
```

---

#### POST /disputes/:dispute_id/assign

分配拒付工单

**认证**: JWT (Admin)

**请求**:
```json
{
  "assigned_to": "f47ac10b-58cc-4372-a567-0e02b2c3d479"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "工单已分配",
  "data": {
    "dispute_id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
    "assigned_to": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "assigned_to_name": "张三"
  }
}
```

---

### 2.2 拒付统计

#### GET /disputes/statistics

查询拒付统计数据

**认证**: JWT (Admin)

**查询参数**:
```
?merchant_id=e55feb66-16f9-41be-a68b-a8961df898b6  # 可选
&start_date=2025-01-01
&end_date=2025-01-31
&channel=stripe
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "overview": {
      "total_disputes": 45,
      "total_amount": 2250000,
      "won_count": 20,
      "won_amount": 1000000,
      "lost_count": 15,
      "lost_amount": 750000,
      "pending_count": 10,
      "pending_amount": 500000,
      "win_rate": 0.5714
    },

    "by_reason": [
      {
        "reason": "fraudulent",
        "count": 20,
        "amount": 1000000
      },
      {
        "reason": "product_not_received",
        "count": 15,
        "amount": 750000
      }
    ],

    "by_channel": [
      {
        "channel": "stripe",
        "count": 30,
        "amount": 1500000
      },
      {
        "channel": "paypal",
        "count": 15,
        "amount": 750000
      }
    ],

    "trend": [
      {
        "date": "2025-01-01",
        "count": 2,
        "amount": 100000
      }
    ]
  }
}
```

---

## 三、商户额度管理 API (merchant-limit-service)

**Base URL**: `http://localhost:40018/api/v1`

### 3.1 商户等级管理

#### GET /tiers

查询商户等级列表

**认证**: JWT (Admin) 或 Public

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [
      {
        "tier_id": "e4f5a6b7-c8d9-0123-ef01-345678901234",
        "tier_code": "starter",
        "tier_name": "入门版",
        "tier_level": 1,

        "limits": {
          "daily_limit": 5000000,
          "monthly_limit": 100000000,
          "single_transaction_limit": 100000
        },

        "fees": {
          "transaction_fee_rate": 0.0320,
          "fixed_fee": 30
        },

        "features": {
          "pre_auth": false,
          "batch_payment": false,
          "api_access": true
        },

        "upgrade_conditions": {
          "min_volume": 10000000,
          "min_transactions": 100
        },

        "is_active": true,
        "description": "适合初创企业和个人商户"
      }
    ]
  }
}
```

---

#### POST /tiers

创建商户等级（Admin）

**认证**: JWT (Admin)

**请求**:
```json
{
  "tier_code": "vip",
  "tier_name": "VIP 定制版",
  "tier_level": 6,
  "daily_limit": 999999999999,
  "monthly_limit": 999999999999,
  "single_transaction_limit": 999999999999,
  "transaction_fee_rate": 0.0150,
  "fixed_fee": 30,
  "features": {
    "pre_auth": true,
    "batch_payment": true,
    "api_access": true,
    "dedicated_support": true
  },
  "description": "VIP 客户专属等级"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "商户等级创建成功",
  "data": {
    "tier_id": "f5a6b7c8-d9e0-1234-f012-456789012345",
    "tier_code": "vip",
    "tier_level": 6
  }
}
```

---

### 3.2 商户额度管理

#### GET /limits/:merchant_id

查询商户额度信息

**认证**: JWT (Admin) 或 Merchant API Key

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",

    "tier_info": {
      "tier_id": "e4f5a6b7-c8d9-0123-ef01-345678901234",
      "tier_code": "starter",
      "tier_name": "入门版",
      "tier_level": 1
    },

    "limits": {
      "daily_limit": 5000000,
      "monthly_limit": 100000000,
      "single_transaction_limit": 100000,
      "is_custom": false
    },

    "usage": {
      "daily_used": 1250000,
      "daily_remaining": 3750000,
      "daily_usage_rate": 0.25,

      "monthly_used": 35000000,
      "monthly_remaining": 65000000,
      "monthly_usage_rate": 0.35,

      "last_reset_date": "2025-01-24"
    },

    "status": {
      "is_suspended": false,
      "suspend_reason": null
    }
  }
}
```

---

#### PUT /limits/:merchant_id

更新商户额度（Admin）

**认证**: JWT (Admin)

**请求**:
```json
{
  "daily_limit": 10000000,
  "monthly_limit": 200000000,
  "single_transaction_limit": 200000,
  "is_custom": true,
  "reason": "商户业务量增长，申请提额"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "商户额度更新成功",
  "data": {
    "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
    "daily_limit": 10000000,
    "monthly_limit": 200000000,
    "is_custom": true,
    "updated_at": "2025-01-24T16:00:00Z"
  }
}
```

---

#### POST /limits/:merchant_id/suspend

暂停商户额度（风控冻结）

**认证**: JWT (Admin)

**请求**:
```json
{
  "suspend_reason": "风控检测到异常交易，暂停交易额度"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "商户额度已暂停",
  "data": {
    "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
    "is_suspended": true,
    "suspended_at": "2025-01-24T16:30:00Z"
  }
}
```

---

#### POST /limits/:merchant_id/resume

恢复商户额度

**认证**: JWT (Admin)

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "商户额度已恢复",
  "data": {
    "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
    "is_suspended": false
  }
}
```

---

### 3.3 额度检查（内部 API）

#### POST /limits/check

检查商户额度是否充足

**认证**: Internal (mTLS)

**请求**:
```json
{
  "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
  "amount": 50000,
  "currency": "USD"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "is_allowed": true,
    "reason": null,
    "limits": {
      "daily_remaining": 3750000,
      "monthly_remaining": 65000000
    }
  }
}
```

**或（超限）**:
```json
{
  "code": "LIMIT_EXCEEDED",
  "message": "超过日限额",
  "data": {
    "is_allowed": false,
    "reason": "daily_limit_exceeded",
    "limits": {
      "daily_limit": 5000000,
      "daily_used": 4980000,
      "daily_remaining": 20000,
      "requested_amount": 50000
    }
  }
}
```

---

#### POST /limits/consume

消费商户额度（创建支付时调用）

**认证**: Internal (mTLS)

**请求**:
```json
{
  "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
  "payment_no": "PY20250124123456abcdefgh",
  "amount": 50000,
  "currency": "USD",
  "operation": "increase"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "额度已消费",
  "data": {
    "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
    "before_daily_used": 1250000,
    "after_daily_used": 1300000,
    "daily_remaining": 3700000
  }
}
```

---

#### POST /limits/release

释放商户额度（退款/取消时调用）

**认证**: Internal (mTLS)

**请求**:
```json
{
  "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
  "payment_no": "PY20250124123456abcdefgh",
  "amount": 50000,
  "currency": "USD",
  "operation": "decrease"
}
```

**响应**:
```json
{
  "code": "SUCCESS",
  "message": "额度已释放",
  "data": {
    "merchant_id": "e55feb66-16f9-41be-a68b-a8961df898b6",
    "before_daily_used": 1300000,
    "after_daily_used": 1250000,
    "daily_remaining": 3750000
  }
}
```

---

### 3.4 额度使用统计

#### GET /limits/:merchant_id/usage-history

查询额度使用历史

**认证**: JWT (Admin) 或 Merchant API Key

**查询参数**:
```
?start_date=2025-01-01
&end_date=2025-01-31
&page=1
&page_size=20
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "list": [
      {
        "log_id": "a6b7c8d9-e0f1-2345-a012-567890123456",
        "payment_no": "PY20250124123456abcdefgh",
        "amount": 50000,
        "currency": "USD",
        "operation": "increase",

        "usage": {
          "before_daily_used": 1250000,
          "after_daily_used": 1300000,
          "before_monthly_used": 35000000,
          "after_monthly_used": 35050000
        },

        "limits": {
          "daily_limit": 5000000,
          "monthly_limit": 100000000
        },

        "created_at": "2025-01-24T14:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 245,
      "total_pages": 13
    }
  }
}
```

---

#### GET /limits/:merchant_id/statistics

查询商户额度统计

**认证**: JWT (Admin) 或 Merchant API Key

**查询参数**:
```
?period=month  # day, week, month
```

**响应**:
```json
{
  "code": "SUCCESS",
  "data": {
    "current_period": {
      "period": "2025-01",
      "total_transactions": 245,
      "total_amount": 35050000,
      "average_amount": 143061,
      "peak_daily_usage": 2500000,
      "peak_daily_usage_date": "2025-01-15"
    },

    "limit_info": {
      "daily_limit": 5000000,
      "monthly_limit": 100000000,
      "daily_usage_rate": 0.26,
      "monthly_usage_rate": 0.35
    },

    "trend": [
      {
        "date": "2025-01-01",
        "daily_used": 1200000,
        "transaction_count": 24
      }
    ]
  }
}
```

---

## 四、错误码定义

### 4.1 通用错误码

| 错误码 | HTTP状态码 | 说明 |
|--------|-----------|------|
| SUCCESS | 200 | 成功 |
| INVALID_REQUEST | 400 | 无效的请求参数 |
| UNAUTHORIZED | 401 | 未授权 |
| FORBIDDEN | 403 | 禁止访问 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 资源冲突 |
| INTERNAL_ERROR | 500 | 内部服务器错误 |

### 4.2 对账系统错误码

| 错误码 | 说明 |
|--------|------|
| TASK_ALREADY_EXISTS | 对账任务已存在 |
| TASK_NOT_FOUND | 对账任务不存在 |
| TASK_IN_PROGRESS | 对账任务正在执行 |
| FILE_DOWNLOAD_FAILED | 账单文件下载失败 |
| FILE_PARSE_FAILED | 账单文件解析失败 |

### 4.3 拒付管理错误码

| 错误码 | 说明 |
|--------|------|
| DISPUTE_NOT_FOUND | 拒付工单不存在 |
| EVIDENCE_UPLOAD_FAILED | 证据上传失败 |
| EVIDENCE_OVERDUE | 证据提交已过期 |
| INVALID_EVIDENCE_TYPE | 无效的证据类型 |
| CHANNEL_API_ERROR | 渠道 API 调用失败 |

### 4.4 商户额度错误码

| 错误码 | 说明 |
|--------|------|
| LIMIT_EXCEEDED | 超过额度限制 |
| DAILY_LIMIT_EXCEEDED | 超过日限额 |
| MONTHLY_LIMIT_EXCEEDED | 超过月限额 |
| SINGLE_TRANSACTION_LIMIT_EXCEEDED | 超过单笔限额 |
| MERCHANT_SUSPENDED | 商户已暂停 |
| TIER_NOT_FOUND | 商户等级不存在 |
| INVALID_TIER_LEVEL | 无效的等级级别 |

---

## 五、Webhook 回调

### 5.1 Stripe Dispute Webhook

**URL**: `POST /api/v1/webhooks/stripe/dispute`

**Event Types**:
- `charge.dispute.created` - 拒付创建
- `charge.dispute.updated` - 拒付更新
- `charge.dispute.closed` - 拒付关闭

**Payload 示例**:
```json
{
  "id": "evt_1QdVg42eZvKYlo2C",
  "type": "charge.dispute.created",
  "data": {
    "object": {
      "id": "dp_1QdVg42eZvKYlo2C",
      "amount": 50000,
      "currency": "usd",
      "charge": "ch_3QdVg42eZvKYlo2C",
      "reason": "fraudulent",
      "status": "needs_response",
      "evidence_due_by": 1738368000
    }
  }
}
```

---

## 六、总结

### 6.1 API 端点统计

| 服务 | 端点数量 | 核心功能 |
|------|---------|---------|
| reconciliation-service | 10 | 对账任务、差异管理、报表 |
| dispute-service | 8 | 拒付工单、证据管理、统计 |
| merchant-limit-service | 12 | 等级管理、额度检查、统计 |
| **总计** | **30** | **30** |

### 6.2 认证方式

| API 类型 | 认证方式 | 使用场景 |
|---------|---------|---------|
| Admin API | JWT Token | 管理后台 |
| Merchant API | API Key + Signature | 商户调用 |
| Internal API | mTLS (可选) | 服务间调用 |

### 6.3 下一步

- [ ] Swagger/OpenAPI 规范生成
- [ ] Postman Collection 导出
- [ ] API 自动化测试
- [ ] 速率限制配置
- [ ] API 文档网站部署

---

**文档版本**: 1.0.0
**最后更新**: 2025-01-24
**状态**: ✅ API 设计完成，待评审
