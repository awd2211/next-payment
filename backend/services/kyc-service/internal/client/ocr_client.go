package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/payment-platform/pkg/httpclient"
)

// OCRClient OCR服务客户端（用于文档识别）
type OCRClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewOCRClient 创建OCR客户端实例（带熔断器和降级）
func NewOCRClient(baseURL string) *OCRClient {
	config := &httpclient.Config{
		Timeout:       60 * time.Second, // OCR处理时间较长
		MaxRetries:    2,
		RetryDelay:    2 * time.Second,
		EnableLogging: false,
	}
	breakerConfig := httpclient.DefaultBreakerConfig("ocr-service")

	return &OCRClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// OCRExtractRequest OCR提取请求
type OCRExtractRequest struct {
	ImageURL     string `json:"image_url"`
	DocumentType string `json:"document_type"` // passport, id_card, driving_license, business_license
	Language     string `json:"language,omitempty"` // en, zh, auto
}

// OCRExtractResponse OCR提取响应
type OCRExtractResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		// 通用字段
		DocumentType   string                 `json:"document_type"`
		DocumentNumber string                 `json:"document_number"`
		FullName       string                 `json:"full_name"`
		DateOfBirth    string                 `json:"date_of_birth,omitempty"`
		Gender         string                 `json:"gender,omitempty"`
		Nationality    string                 `json:"nationality,omitempty"`
		IssueDate      string                 `json:"issue_date,omitempty"`
		ExpiryDate     string                 `json:"expiry_date,omitempty"`
		IssuingCountry string                 `json:"issuing_country,omitempty"`

		// 地址信息
		Address        string `json:"address,omitempty"`

		// 商业执照特有字段
		CompanyName    string `json:"company_name,omitempty"`
		RegistrationNo string `json:"registration_no,omitempty"`
		BusinessScope  string `json:"business_scope,omitempty"`

		// OCR置信度和原始数据
		Confidence     float64                `json:"confidence"`      // 0.0 - 1.0
		RawText        string                 `json:"raw_text"`        // OCR原始识别文本
		ExtractedData  map[string]interface{} `json:"extracted_data"`  // 其他提取字段

		// 质量检测
		QualityScore   float64 `json:"quality_score,omitempty"` // 图像质量分数
		IsBlurry       bool    `json:"is_blurry,omitempty"`
		IsTampered     bool    `json:"is_tampered,omitempty"`   // 是否疑似篡改

		// 处理时间
		ProcessTime    int64   `json:"process_time_ms,omitempty"` // 毫秒
	} `json:"data"`
}

// OCRData OCR识别数据（用于存储到数据库）
type OCRData struct {
	DocumentNumber string                 `json:"document_number,omitempty"`
	FullName       string                 `json:"full_name,omitempty"`
	DateOfBirth    string                 `json:"date_of_birth,omitempty"`
	Gender         string                 `json:"gender,omitempty"`
	Nationality    string                 `json:"nationality,omitempty"`
	IssueDate      string                 `json:"issue_date,omitempty"`
	ExpiryDate     string                 `json:"expiry_date,omitempty"`
	IssuingCountry string                 `json:"issuing_country,omitempty"`
	Address        string                 `json:"address,omitempty"`
	CompanyName    string                 `json:"company_name,omitempty"`
	RegistrationNo string                 `json:"registration_no,omitempty"`
	BusinessScope  string                 `json:"business_scope,omitempty"`
	Confidence     float64                `json:"confidence"`
	QualityScore   float64                `json:"quality_score,omitempty"`
	IsBlurry       bool                   `json:"is_blurry,omitempty"`
	IsTampered     bool                   `json:"is_tampered,omitempty"`
	RawText        string                 `json:"raw_text,omitempty"`
	ExtractedData  map[string]interface{} `json:"extracted_data,omitempty"`
}

// ExtractDocument 提取文档信息（主方法）
func (c *OCRClient) ExtractDocument(ctx context.Context, req *OCRExtractRequest) (*OCRData, error) {
	url := fmt.Sprintf("%s/api/v1/ocr/extract", c.baseURL)

	httpReq := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Body:   req,
		Ctx:    ctx,
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		// 降级策略: OCR服务不可用时返回空数据，不阻塞KYC流程
		return &OCRData{
			Confidence: 0.0,
			RawText:    "OCR服务暂时不可用",
		}, fmt.Errorf("OCR服务调用失败（已降级）: %w", err)
	}

	var result OCRExtractResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析OCR响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("OCR服务错误: %s", result.Message)
	}

	if result.Data == nil {
		return nil, fmt.Errorf("OCR响应数据为空")
	}

	// 转换为内部OCRData结构
	ocrData := &OCRData{
		DocumentNumber: result.Data.DocumentNumber,
		FullName:       result.Data.FullName,
		DateOfBirth:    result.Data.DateOfBirth,
		Gender:         result.Data.Gender,
		Nationality:    result.Data.Nationality,
		IssueDate:      result.Data.IssueDate,
		ExpiryDate:     result.Data.ExpiryDate,
		IssuingCountry: result.Data.IssuingCountry,
		Address:        result.Data.Address,
		CompanyName:    result.Data.CompanyName,
		RegistrationNo: result.Data.RegistrationNo,
		BusinessScope:  result.Data.BusinessScope,
		Confidence:     result.Data.Confidence,
		QualityScore:   result.Data.QualityScore,
		IsBlurry:       result.Data.IsBlurry,
		IsTampered:     result.Data.IsTampered,
		RawText:        result.Data.RawText,
		ExtractedData:  result.Data.ExtractedData,
	}

	return ocrData, nil
}

// ExtractPassport 提取护照信息（便捷方法）
func (c *OCRClient) ExtractPassport(ctx context.Context, imageURL string) (*OCRData, error) {
	return c.ExtractDocument(ctx, &OCRExtractRequest{
		ImageURL:     imageURL,
		DocumentType: "passport",
		Language:     "auto",
	})
}

// ExtractIDCard 提取身份证信息（便捷方法）
func (c *OCRClient) ExtractIDCard(ctx context.Context, imageURL string) (*OCRData, error) {
	return c.ExtractDocument(ctx, &OCRExtractRequest{
		ImageURL:     imageURL,
		DocumentType: "id_card",
		Language:     "auto",
	})
}

// ExtractBusinessLicense 提取营业执照信息（便捷方法）
func (c *OCRClient) ExtractBusinessLicense(ctx context.Context, imageURL string) (*OCRData, error) {
	return c.ExtractDocument(ctx, &OCRExtractRequest{
		ImageURL:     imageURL,
		DocumentType: "business_license",
		Language:     "auto",
	})
}

// ValidateOCRQuality 验证OCR质量（用于决定是否需要人工审核）
func ValidateOCRQuality(ocrData *OCRData) (bool, string) {
	// 置信度检查
	if ocrData.Confidence < 0.7 {
		return false, fmt.Sprintf("OCR置信度过低 (%.2f < 0.70)", ocrData.Confidence)
	}

	// 图像质量检查
	if ocrData.IsBlurry {
		return false, "图像模糊，建议重新上传"
	}

	// 篡改检测
	if ocrData.IsTampered {
		return false, "检测到文档可能被篡改"
	}

	// 质量分数检查
	if ocrData.QualityScore > 0 && ocrData.QualityScore < 0.6 {
		return false, fmt.Sprintf("图像质量分数过低 (%.2f < 0.60)", ocrData.QualityScore)
	}

	// 必要字段检查（根据文档类型）
	if ocrData.DocumentNumber == "" {
		return false, "未能识别证件号码"
	}

	if ocrData.FullName == "" && ocrData.CompanyName == "" {
		return false, "未能识别姓名/公司名称"
	}

	return true, "OCR质量验证通过"
}
