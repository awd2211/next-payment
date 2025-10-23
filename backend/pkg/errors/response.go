package errors

import (
	"net/http"
)

// Response 统一响应结构
type Response struct {
	Code    string      `json:"code"`              // 业务状态码
	Message string      `json:"message"`           // 响应消息
	Data    interface{} `json:"data,omitempty"`    // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    string `json:"code"`              // 错误码
	Message string `json:"message"`           // 错误消息
	Details string `json:"details,omitempty"` // 错误详情
	TraceID string `json:"trace_id,omitempty"` // 追踪ID
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Code    string      `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
	Total   int64       `json:"total"`   // 总记录数
	Page    int         `json:"page"`    // 当前页码
	PageSize int        `json:"page_size"` // 每页大小
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    ErrCodeSuccess,
		Message: GetMessage(ErrCodeSuccess),
		Data:    data,
	}
}

// NewSuccessResponseWithMessage 创建带自定义消息的成功响应
func NewSuccessResponseWithMessage(message string, data interface{}) *Response {
	return &Response{
		Code:    ErrCodeSuccess,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code, message, details string) *ErrorResponse {
	if message == "" {
		message = GetMessage(code)
	}
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewErrorResponseFromBusinessError 从业务错误创建错误响应
func NewErrorResponseFromBusinessError(err *BusinessError) *ErrorResponse {
	return &ErrorResponse{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	}
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse(data interface{}, total int64, page, pageSize int) *PaginatedResponse {
	return &PaginatedResponse{
		Code:     ErrCodeSuccess,
		Message:  GetMessage(ErrCodeSuccess),
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// WithTraceID 添加追踪ID
func (r *Response) WithTraceID(traceID string) *Response {
	r.TraceID = traceID
	return r
}

// WithTraceID 添加追踪ID到错误响应
func (r *ErrorResponse) WithTraceID(traceID string) *ErrorResponse {
	r.TraceID = traceID
	return r
}

// WithTraceID 添加追踪ID到分页响应
func (r *PaginatedResponse) WithTraceID(traceID string) *PaginatedResponse {
	r.TraceID = traceID
	return r
}

// GetHTTPStatus 根据错误码获取HTTP状态码
func GetHTTPStatus(code string) int {
	switch code {
	case ErrCodeSuccess:
		return http.StatusOK
	case ErrCodeInvalidRequest, ErrCodeInvalidAmount, ErrCodeInvalidCurrency,
		ErrCodeInvalidAPIKey, ErrCodeInvalidSignature:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeAuthFailed, ErrCodeTokenExpired,
		ErrCodeTokenInvalid, ErrCodeInvalidCredentials, ErrCodeSessionExpired:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeMerchantFrozen, ErrCodeMerchantNotApproved,
		ErrCodeRefundNotAllowed, ErrCodeWithdrawalNotAllowed:
		return http.StatusForbidden
	case ErrCodeResourceNotFound, ErrCodeMerchantNotFound, ErrCodeOrderNotFound,
		ErrCodeConfigNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeDuplicateOrder, ErrCodeMerchantDuplicate,
		ErrCodeOrderAlreadyPaid:
		return http.StatusConflict
	case ErrCodeInternalError, ErrCodePaymentFailed, ErrCodeRefundFailed,
		ErrCodeSettlementFailed, ErrCodeWithdrawalFailed, ErrCodeNotificationFailed:
		return http.StatusInternalServerError
	case ErrCodeRiskRejected, ErrCodeRiskScoreTooHigh, ErrCodeBlacklistMatch:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// GetHTTPStatusFromError 从业务错误获取HTTP状态码
func GetHTTPStatusFromError(err error) int {
	if bizErr, ok := GetBusinessError(err); ok {
		return GetHTTPStatus(bizErr.Code)
	}
	return http.StatusInternalServerError
}
