# KYC Service

> KYC和资质审核服务

## 职责

从 `merchant-service` 拆分出来的KYC审核模块，负责商户的身份验证和业务资质审核。

## 核心功能

- 📄 KYC文档上传
- 🔍 人工审核
- 🤖 OCR识别（身份证、营业执照）
- ✅ 资质审核
- 📅 证件到期提醒

## 数据库

**Database**: `payment_kyc`

**Tables**:
- kyc_documents（KYC文档）
- business_qualifications（业务资质）
- kyc_reviews（审核记录）

## 端口

**Port**: `8014`

## 状态

📋 **预留中** - 待从 merchant-service 拆分

## 依赖服务

- merchant-service（更新商户KYC状态）
- document-service（文档存储，未来）
- notification-service（审核结果通知）

## API端点

```
POST   /api/v1/kyc/documents           # 上传KYC文档
GET    /api/v1/kyc/documents           # KYC文档列表
POST   /api/v1/kyc/documents/:id/review # 审核文档
GET    /api/v1/kyc/qualifications      # 资质列表
POST   /api/v1/kyc/qualifications      # 上传资质
```

## 启动命令

```bash
PORT=8014 \
DB_NAME=payment_kyc \
go run ./cmd/main.go
```

## 拆分计划

- 预计工作量：3周
- 优先级：P2（第四批拆分）
- 开始时间：待定
