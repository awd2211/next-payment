package handler

// SuccessResponse 成功响应
type SuccessResponse struct {
	Code    int         `json:"code" example:"0"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error" example:"invalid parameter"`
}

// ListResponse 列表响应
type ListResponse struct {
	Code       int         `json:"code" example:"0"`
	Message    string      `json:"message" example:"success"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total" example:"100"`
	Page       int         `json:"page" example:"1"`
	PageSize   int         `json:"page_size" example:"20"`
	TotalPages int         `json:"total_pages" example:"5"`
}
