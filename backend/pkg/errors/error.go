package errors

import (
	"fmt"
)

// BusinessError 业务错误类型
type BusinessError struct {
	Code    string // 错误码
	Message string // 错误消息
	Details string // 错误详情（可选）
	Err     error  // 原始错误（可选）
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *BusinessError) Unwrap() error {
	return e.Err
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// NewBusinessErrorWithDetails 创建带详情的业务错误
func NewBusinessErrorWithDetails(code, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewBusinessErrorWrap 包装原始错误
func NewBusinessErrorWrap(code, message string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails 添加错误详情
func (e *BusinessError) WithDetails(details string) *BusinessError {
	e.Details = details
	return e
}

// WithError 添加原始错误
func (e *BusinessError) WithError(err error) *BusinessError {
	e.Err = err
	return e
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// GetBusinessError 获取业务错误
func GetBusinessError(err error) (*BusinessError, bool) {
	if err == nil {
		return nil, false
	}
	bizErr, ok := err.(*BusinessError)
	return bizErr, ok
}

// 快捷方法：创建常见业务错误

// NewInvalidRequestError 无效请求错误
func NewInvalidRequestError(message string) *BusinessError {
	if message == "" {
		message = GetMessage(ErrCodeInvalidRequest)
	}
	return NewBusinessError(ErrCodeInvalidRequest, message)
}

// NewUnauthorizedError 未授权错误
func NewUnauthorizedError(message string) *BusinessError {
	if message == "" {
		message = GetMessage(ErrCodeUnauthorized)
	}
	return NewBusinessError(ErrCodeUnauthorized, message)
}

// NewNotFoundError 资源不存在错误
func NewNotFoundError(message string) *BusinessError {
	if message == "" {
		message = GetMessage(ErrCodeResourceNotFound)
	}
	return NewBusinessError(ErrCodeResourceNotFound, message)
}

// NewInternalError 内部错误
func NewInternalError(message string) *BusinessError {
	if message == "" {
		message = GetMessage(ErrCodeInternalError)
	}
	return NewBusinessError(ErrCodeInternalError, message)
}

// NewConflictError 资源冲突错误
func NewConflictError(message string) *BusinessError {
	if message == "" {
		message = GetMessage(ErrCodeConflict)
	}
	return NewBusinessError(ErrCodeConflict, message)
}

// NewForbiddenError 禁止访问错误
func NewForbiddenError(message string) *BusinessError {
	if message == "" {
		message = GetMessage(ErrCodeForbidden)
	}
	return NewBusinessError(ErrCodeForbidden, message)
}
