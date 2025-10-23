package errors

// 业务错误码定义
const (
	// 通用错误码 (1000-1999)
	ErrCodeSuccess          = "SUCCESS"
	ErrCodeInvalidRequest   = "INVALID_REQUEST"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeConflict         = "CONFLICT"

	// 认证相关错误 (2000-2999)
	ErrCodeAuthFailed        = "AUTH_FAILED"
	ErrCodeTokenExpired      = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid      = "TOKEN_INVALID"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired    = "SESSION_EXPIRED"
	ErrCode2FARequired       = "2FA_REQUIRED"
	ErrCode2FAFailed         = "2FA_FAILED"

	// 商户相关错误 (3000-3999)
	ErrCodeMerchantNotFound     = "MERCHANT_NOT_FOUND"
	ErrCodeMerchantNotActive    = "MERCHANT_NOT_ACTIVE"
	ErrCodeMerchantFrozen       = "MERCHANT_FROZEN"
	ErrCodeMerchantDuplicate    = "MERCHANT_DUPLICATE"
	ErrCodeInvalidAPIKey        = "INVALID_API_KEY"
	ErrCodeInvalidSignature     = "INVALID_SIGNATURE"
	ErrCodeMerchantNotApproved  = "MERCHANT_NOT_APPROVED"

	// 支付相关错误 (4000-4999)
	ErrCodePaymentFailed        = "PAYMENT_FAILED"
	ErrCodeDuplicateOrder       = "DUPLICATE_ORDER"
	ErrCodeInvalidAmount        = "INVALID_AMOUNT"
	ErrCodeInvalidCurrency      = "INVALID_CURRENCY"
	ErrCodePaymentExpired       = "PAYMENT_EXPIRED"
	ErrCodePaymentCancelled     = "PAYMENT_CANCELLED"
	ErrCodeInsufficientBalance  = "INSUFFICIENT_BALANCE"
	ErrCodeAmountExceedsLimit   = "AMOUNT_EXCEEDS_LIMIT"
	ErrCodeChannelUnavailable   = "CHANNEL_UNAVAILABLE"
	ErrCodeChannelNotSupported  = "CHANNEL_NOT_SUPPORTED"

	// 退款相关错误 (5000-5999)
	ErrCodeRefundFailed         = "REFUND_FAILED"
	ErrCodeRefundNotAllowed     = "REFUND_NOT_ALLOWED"
	ErrCodeRefundAmountExceeds  = "REFUND_AMOUNT_EXCEEDS"
	ErrCodeInvalidRefundStatus  = "INVALID_REFUND_STATUS"
	ErrCodePartialRefundFailed  = "PARTIAL_REFUND_FAILED"

	// 风控相关错误 (6000-6999)
	ErrCodeRiskRejected         = "RISK_REJECTED"
	ErrCodeRiskScoreTooHigh     = "RISK_SCORE_TOO_HIGH"
	ErrCodeBlacklistMatch       = "BLACKLIST_MATCH"
	ErrCodeManualReviewRequired = "MANUAL_REVIEW_REQUIRED"
	ErrCodeFraudDetected        = "FRAUD_DETECTED"

	// 订单相关错误 (7000-7999)
	ErrCodeOrderNotFound        = "ORDER_NOT_FOUND"
	ErrCodeOrderStatusInvalid   = "ORDER_STATUS_INVALID"
	ErrCodeOrderCancelled       = "ORDER_CANCELLED"
	ErrCodeOrderExpired         = "ORDER_EXPIRED"
	ErrCodeOrderAlreadyPaid     = "ORDER_ALREADY_PAID"

	// 结算相关错误 (8000-8999)
	ErrCodeSettlementFailed     = "SETTLEMENT_FAILED"
	ErrCodeSettlementNotReady   = "SETTLEMENT_NOT_READY"
	ErrCodeInvalidSettlementPeriod = "INVALID_SETTLEMENT_PERIOD"
	ErrCodeSettlementAlreadyProcessed = "SETTLEMENT_ALREADY_PROCESSED"

	// 提现相关错误 (9000-9999)
	ErrCodeWithdrawalFailed     = "WITHDRAWAL_FAILED"
	ErrCodeWithdrawalNotAllowed = "WITHDRAWAL_NOT_ALLOWED"
	ErrCodeWithdrawalAmountInvalid = "WITHDRAWAL_AMOUNT_INVALID"
	ErrCodeInsufficientWithdrawableBalance = "INSUFFICIENT_WITHDRAWABLE_BALANCE"
	ErrCodeWithdrawalLimitExceeded = "WITHDRAWAL_LIMIT_EXCEEDED"

	// KYC 相关错误 (10000-10999)
	ErrCodeKYCNotSubmitted      = "KYC_NOT_SUBMITTED"
	ErrCodeKYCPending           = "KYC_PENDING"
	ErrCodeKYCRejected          = "KYC_REJECTED"
	ErrCodeKYCDocumentInvalid   = "KYC_DOCUMENT_INVALID"
	ErrCodeKYCExpired           = "KYC_EXPIRED"

	// 通知相关错误 (11000-11999)
	ErrCodeNotificationFailed   = "NOTIFICATION_FAILED"
	ErrCodeEmailSendFailed      = "EMAIL_SEND_FAILED"
	ErrCodeSMSSendFailed        = "SMS_SEND_FAILED"
	ErrCodeWebhookCallbackFailed = "WEBHOOK_CALLBACK_FAILED"

	// 配置相关错误 (12000-12999)
	ErrCodeConfigNotFound       = "CONFIG_NOT_FOUND"
	ErrCodeConfigInvalid        = "CONFIG_INVALID"
	ErrCodeConfigUpdateFailed   = "CONFIG_UPDATE_FAILED"
)

// 错误码对应的默认消息
var errorMessages = map[string]string{
	ErrCodeSuccess:          "成功",
	ErrCodeInvalidRequest:   "请求参数错误",
	ErrCodeInternalError:    "内部服务错误",
	ErrCodeResourceNotFound: "资源不存在",
	ErrCodeUnauthorized:     "未授权访问",
	ErrCodeForbidden:        "禁止访问",
	ErrCodeBadRequest:       "错误的请求",
	ErrCodeConflict:         "资源冲突",

	ErrCodeAuthFailed:        "认证失败",
	ErrCodeTokenExpired:      "Token已过期",
	ErrCodeTokenInvalid:      "Token无效",
	ErrCodeInvalidCredentials: "用户名或密码错误",
	ErrCodeSessionExpired:    "会话已过期",
	ErrCode2FARequired:       "需要二次认证",
	ErrCode2FAFailed:         "二次认证失败",

	ErrCodeMerchantNotFound:     "商户不存在",
	ErrCodeMerchantNotActive:    "商户未激活",
	ErrCodeMerchantFrozen:       "商户已冻结",
	ErrCodeMerchantDuplicate:    "商户已存在",
	ErrCodeInvalidAPIKey:        "API密钥无效",
	ErrCodeInvalidSignature:     "签名验证失败",
	ErrCodeMerchantNotApproved:  "商户未审核通过",

	ErrCodePaymentFailed:        "支付失败",
	ErrCodeDuplicateOrder:       "订单号重复",
	ErrCodeInvalidAmount:        "金额无效",
	ErrCodeInvalidCurrency:      "货币类型不支持",
	ErrCodePaymentExpired:       "支付已过期",
	ErrCodePaymentCancelled:     "支付已取消",
	ErrCodeInsufficientBalance:  "余额不足",
	ErrCodeAmountExceedsLimit:   "金额超过限制",
	ErrCodeChannelUnavailable:   "支付渠道不可用",
	ErrCodeChannelNotSupported:  "不支持该支付渠道",

	ErrCodeRefundFailed:         "退款失败",
	ErrCodeRefundNotAllowed:     "不允许退款",
	ErrCodeRefundAmountExceeds:  "退款金额超过原支付金额",
	ErrCodeInvalidRefundStatus:  "支付状态不允许退款",
	ErrCodePartialRefundFailed:  "部分退款失败",

	ErrCodeRiskRejected:         "风控拒绝",
	ErrCodeRiskScoreTooHigh:     "风险评分过高",
	ErrCodeBlacklistMatch:       "命中黑名单",
	ErrCodeManualReviewRequired: "需要人工审核",
	ErrCodeFraudDetected:        "检测到欺诈行为",

	ErrCodeOrderNotFound:        "订单不存在",
	ErrCodeOrderStatusInvalid:   "订单状态无效",
	ErrCodeOrderCancelled:       "订单已取消",
	ErrCodeOrderExpired:         "订单已过期",
	ErrCodeOrderAlreadyPaid:     "订单已支付",

	ErrCodeSettlementFailed:     "结算失败",
	ErrCodeSettlementNotReady:   "结算未就绪",
	ErrCodeInvalidSettlementPeriod: "无效的结算周期",
	ErrCodeSettlementAlreadyProcessed: "结算已处理",

	ErrCodeWithdrawalFailed:     "提现失败",
	ErrCodeWithdrawalNotAllowed: "不允许提现",
	ErrCodeWithdrawalAmountInvalid: "提现金额无效",
	ErrCodeInsufficientWithdrawableBalance: "可提现余额不足",
	ErrCodeWithdrawalLimitExceeded: "超过提现限额",

	ErrCodeKYCNotSubmitted:      "未提交KYC",
	ErrCodeKYCPending:           "KYC审核中",
	ErrCodeKYCRejected:          "KYC审核被拒绝",
	ErrCodeKYCDocumentInvalid:   "KYC文档无效",
	ErrCodeKYCExpired:           "KYC已过期",

	ErrCodeNotificationFailed:   "通知发送失败",
	ErrCodeEmailSendFailed:      "邮件发送失败",
	ErrCodeSMSSendFailed:        "短信发送失败",
	ErrCodeWebhookCallbackFailed: "Webhook回调失败",

	ErrCodeConfigNotFound:       "配置不存在",
	ErrCodeConfigInvalid:        "配置无效",
	ErrCodeConfigUpdateFailed:   "配置更新失败",
}

// GetMessage 获取错误码对应的默认消息
func GetMessage(code string) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
