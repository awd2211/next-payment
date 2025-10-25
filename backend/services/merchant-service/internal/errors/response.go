package errors

// 错误代码常量
const (
	ErrCodeSuccess         = 0
	ErrCodeUnauthorized    = 401
	ErrCodeForbidden       = 403
	ErrCodeNotFound        = 404
	ErrCodeInternalError   = 500
	ErrCodeInvalidParam    = 400
	ErrCodeBusinessError   = 1000
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"trace_id,omitempty"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message, detail string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Code: ErrCodeSuccess,
		Data: data,
	}
}

// WithTraceID 添加 Trace ID
func (e *ErrorResponse) WithTraceID(traceID string) *ErrorResponse {
	e.TraceID = traceID
	return e
}

// WithTraceID 添加 Trace ID
func (s *SuccessResponse) WithTraceID(traceID string) *SuccessResponse {
	s.TraceID = traceID
	return s
}
