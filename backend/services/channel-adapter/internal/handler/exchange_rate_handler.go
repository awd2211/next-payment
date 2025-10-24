package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"payment-platform/channel-adapter/internal/repository"
)

// ExchangeRateHandler 汇率HTTP处理器
type ExchangeRateHandler struct {
	exchangeRateRepo repository.ExchangeRateRepository
}

// NewExchangeRateHandler 创建汇率处理器
func NewExchangeRateHandler(exchangeRateRepo repository.ExchangeRateRepository) *ExchangeRateHandler {
	return &ExchangeRateHandler{
		exchangeRateRepo: exchangeRateRepo,
	}
}

// RegisterRoutes 注册汇率路由
func (h *ExchangeRateHandler) RegisterRoutes(r *gin.Engine) {
	exchange := r.Group("/api/v1/exchange-rates")
	{
		exchange.GET("/latest", h.GetLatestRate)        // 获取最新汇率
		exchange.GET("/snapshot", h.GetLatestSnapshot)  // 获取最新快照
		exchange.GET("/history", h.GetRateHistory)      // 获取历史汇率
	}
}

// GetLatestRateRequest 获取最新汇率请求
type GetLatestRateRequest struct {
	From string `form:"from" binding:"required"` // 基础货币
	To   string `form:"to" binding:"required"`   // 目标货币
}

// GetLatestRateResponse 获取最新汇率响应
type GetLatestRateResponse struct {
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	Source         string    `json:"source"`
	ValidFrom      time.Time `json:"valid_from"`
}

// GetLatestRate 获取最新汇率
//
//	@Summary		获取最新汇率
//	@Description	获取两种货币之间的最新汇率
//	@Tags			ExchangeRate
//	@Accept			json
//	@Produce		json
//	@Param			from	query		string	true	"基础货币代码 (如: USD)"
//	@Param			to		query		string	true	"目标货币代码 (如: CNY)"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/exchange-rates/latest [get]
func (h *ExchangeRateHandler) GetLatestRate(c *gin.Context) {
	var req GetLatestRateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	rate, err := h.exchangeRateRepo.GetLatestRate(c.Request.Context(), req.From, req.To)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "未找到汇率数据: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": GetLatestRateResponse{
			BaseCurrency:   rate.BaseCurrency,
			TargetCurrency: rate.TargetCurrency,
			Rate:           rate.Rate,
			Source:         rate.Source,
			ValidFrom:      rate.ValidFrom,
		},
	})
}

// GetLatestSnapshotRequest 获取最新快照请求
type GetLatestSnapshotRequest struct {
	Base string `form:"base" binding:"required"` // 基础货币
}

// GetLatestSnapshot 获取最新汇率快照（一次获取多个货币对）
//
//	@Summary		获取最新汇率快照
//	@Description	获取基础货币对多个目标货币的最新汇率
//	@Tags			ExchangeRate
//	@Accept			json
//	@Produce		json
//	@Param			base	query		string	true	"基础货币代码 (如: USD)"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/exchange-rates/snapshot [get]
func (h *ExchangeRateHandler) GetLatestSnapshot(c *gin.Context) {
	var req GetLatestSnapshotRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	snapshot, err := h.exchangeRateRepo.GetLatestSnapshot(c.Request.Context(), req.Base)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "未找到快照数据: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"base_currency":  snapshot.BaseCurrency,
			"rates":          snapshot.Rates,
			"source":         snapshot.Source,
			"snapshot_time":  snapshot.SnapshotTime,
		},
	})
}

// GetRateHistoryRequest 获取历史汇率请求
type GetRateHistoryRequest struct {
	From      string `form:"from" binding:"required"`       // 基础货币
	To        string `form:"to" binding:"required"`         // 目标货币
	StartTime string `form:"start_time" binding:"required"` // 开始时间 RFC3339格式
	EndTime   string `form:"end_time" binding:"required"`   // 结束时间 RFC3339格式
}

// GetRateHistory 获取历史汇率
//
//	@Summary		获取历史汇率
//	@Description	获取指定时间范围内的历史汇率数据
//	@Tags			ExchangeRate
//	@Accept			json
//	@Produce		json
//	@Param			from		query		string	true	"基础货币代码"
//	@Param			to			query		string	true	"目标货币代码"
//	@Param			start_time	query		string	true	"开始时间 (RFC3339格式)"
//	@Param			end_time	query		string	true	"结束时间 (RFC3339格式)"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v1/exchange-rates/history [get]
func (h *ExchangeRateHandler) GetRateHistory(c *gin.Context) {
	var req GetRateHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误: " + err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误: " + err.Error(),
		})
		return
	}

	rates, err := h.exchangeRateRepo.GetRateHistory(c.Request.Context(), req.From, req.To, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取历史汇率失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"base_currency":   req.From,
			"target_currency": req.To,
			"start_time":      startTime,
			"end_time":        endTime,
			"rates":           rates,
			"count":           len(rates),
		},
	})
}
