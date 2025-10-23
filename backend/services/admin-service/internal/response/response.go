package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误代码
type ErrorCode string

const (
	// 通用错误
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeInternalServer  ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"

	// 业务错误
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeAccountLocked      ErrorCode = "ACCOUNT_LOCKED"
	ErrCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ErrCodeResourceExists     ErrorCode = "RESOURCE_EXISTS"
	ErrCodeResourceNotFound   ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeOperationFailed    ErrorCode = "OPERATION_FAILED"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`          // 是否成功
	Code    ErrorCode   `json:"code,omitempty"`   // 错误代码
	Message string      `json:"message"`          // 消息
	Data    interface{} `json:"data,omitempty"`   // 数据
	Error   *ErrorInfo  `json:"error,omitempty"`  // 错误详情
}

// ErrorInfo 错误详情
type ErrorInfo struct {
	Code    ErrorCode `json:"code"`              // 错误代码
	Message string    `json:"message"`           // 错误消息
	Details string    `json:"details,omitempty"` // 详细信息（用于调试）
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page      int   `json:"page"`       // 当前页
	PageSize  int   `json:"page_size"`  // 每页数量
	Total     int64 `json:"total"`      // 总数
	TotalPage int64 `json:"total_page"` // 总页数
}

// ListResponse 列表响应
type ListResponse struct {
	Success    bool            `json:"success"`    // 是否成功
	Message    string          `json:"message"`    // 消息
	Data       interface{}     `json:"data"`       // 数据列表
	Pagination *PaginationInfo `json:"pagination"` // 分页信息
}

// Success 返回成功响应
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithData 返回成功响应（仅数据）
func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created 返回创建成功响应
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessList 返回列表成功响应
func SuccessList(c *gin.Context, data interface{}, pagination *PaginationInfo) {
	c.JSON(http.StatusOK, ListResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, statusCode int, code ErrorCode, message string, details string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequest 返回400错误
func BadRequest(c *gin.Context, message string, details string) {
	Error(c, http.StatusBadRequest, ErrCodeBadRequest, message, details)
}

// ValidationError 返回参数验证错误
func ValidationError(c *gin.Context, message string, details string) {
	Error(c, http.StatusBadRequest, ErrCodeValidation, message, details)
}

// Unauthorized 返回401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, message, "")
}

// Forbidden 返回403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, ErrCodeForbidden, message, "")
}

// NotFound 返回404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, ErrCodeNotFound, message, "")
}

// Conflict 返回409错误
func Conflict(c *gin.Context, message string, details string) {
	Error(c, http.StatusConflict, ErrCodeConflict, message, details)
}

// InternalServerError 返回500错误
func InternalServerError(c *gin.Context, message string, details string) {
	Error(c, http.StatusInternalServerError, ErrCodeInternalServer, message, details)
}

// HandleServiceError 处理服务层错误
func HandleServiceError(c *gin.Context, err error, customMessage string) {
	if err == nil {
		return
	}

	// 根据错误类型返回不同的响应
	// 这里可以根据具体的service error进行判断
	switch err.Error() {
	case "配置不存在", "角色不存在", "管理员不存在", "权限不存在":
		NotFound(c, err.Error())
	case "配置键已存在", "角色代码已存在", "用户名已存在":
		Conflict(c, err.Error(), "")
	case "用户名或密码错误":
		Error(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, err.Error(), "")
	case "账户已被禁用":
		Error(c, http.StatusForbidden, ErrCodeAccountLocked, err.Error(), "")
	case "无权限操作":
		Error(c, http.StatusForbidden, ErrCodePermissionDenied, err.Error(), "")
	default:
		InternalServerError(c, customMessage, err.Error())
	}
}

// NewPagination 创建分页信息
func NewPagination(page, pageSize int, total int64) *PaginationInfo {
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)
	return &PaginationInfo{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPage,
	}
}
