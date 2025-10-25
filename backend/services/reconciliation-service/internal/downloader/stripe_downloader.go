package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/reporting/reportrun"

	"payment-platform/reconciliation-service/internal/model"
	"payment-platform/reconciliation-service/internal/repository"
	"payment-platform/reconciliation-service/internal/service"
)

// StripeDownloader Stripe账单下载器
type StripeDownloader struct {
	apiKey   string
	repo     repository.ReconciliationRepository
	basePath string // 文件存储路径
}

// NewStripeDownloader 创建Stripe下载器
func NewStripeDownloader(apiKey string, repo repository.ReconciliationRepository, basePath string) *StripeDownloader {
	stripe.Key = apiKey
	return &StripeDownloader{
		apiKey:   apiKey,
		repo:     repo,
		basePath: basePath,
	}
}

// Download 下载Stripe账单文件
func (d *StripeDownloader) Download(ctx context.Context, channel string, settlementDate time.Time) (*model.ChannelSettlementFile, error) {
	if channel != model.ChannelStripe {
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}

	// Check if file already exists
	existing, err := d.repo.GetFileByDateAndChannel(ctx, settlementDate, channel)
	if err != nil {
		return nil, fmt.Errorf("check existing file failed: %w", err)
	}
	if existing != nil && existing.Status == model.FileStatusImported {
		return existing, nil
	}

	// Generate file record
	fileNo := generateFileNo(channel, settlementDate)
	file := &model.ChannelSettlementFile{
		FileNo:         fileNo,
		Channel:        channel,
		SettlementDate: settlementDate,
		Status:         model.FileStatusPending,
	}

	if err := d.repo.CreateFile(ctx, file); err != nil {
		return nil, fmt.Errorf("create file record failed: %w", err)
	}

	// Create report run in Stripe
	params := &stripe.ReportingReportRunParams{
		ReportType: stripe.String("balance.summary.1"),
		Parameters: &stripe.ReportingReportRunParametersParams{
			IntervalStart: stripe.Int64(settlementDate.Unix()),
			IntervalEnd:   stripe.Int64(settlementDate.Add(24 * time.Hour).Unix()),
		},
	}

	reportRun, err := reportrun.New(params)
	if err != nil {
		file.Status = model.FileStatusPending
		d.repo.UpdateFile(ctx, file)
		return nil, fmt.Errorf("create stripe report run failed: %w", err)
	}

	// Wait for report to be ready (polling with timeout)
	maxWaitTime := 5 * time.Minute
	pollInterval := 10 * time.Second
	deadline := time.Now().Add(maxWaitTime)

	for time.Now().Before(deadline) {
		reportRun, err = reportrun.Get(reportRun.ID, nil)
		if err != nil {
			return nil, fmt.Errorf("get report run status failed: %w", err)
		}

		if reportRun.Status == "succeeded" {
			break
		} else if reportRun.Status == "failed" {
			file.Status = model.FileStatusPending
			d.repo.UpdateFile(ctx, file)
			return nil, fmt.Errorf("stripe report generation failed")
		}

		time.Sleep(pollInterval)
	}

	if reportRun.Status != "succeeded" {
		return nil, fmt.Errorf("stripe report generation timeout")
	}

	// Download file from result URL
	downloadURL := reportRun.Result.URL
	if downloadURL == "" {
		return nil, fmt.Errorf("stripe report URL is empty")
	}

	localPath, fileSize, fileHash, err := d.downloadFile(downloadURL, fileNo)
	if err != nil {
		return nil, fmt.Errorf("download file failed: %w", err)
	}

	// Update file record
	now := time.Now()
	file.FileURL = localPath
	file.FileSize = fileSize
	file.FileHash = fileHash
	file.Status = model.FileStatusDownloaded
	file.DownloadedAt = &now

	if err := d.repo.UpdateFile(ctx, file); err != nil {
		return nil, fmt.Errorf("update file record failed: %w", err)
	}

	return file, nil
}

// Parse 解析Stripe账单文件
func (d *StripeDownloader) Parse(ctx context.Context, fileURL string) ([]*service.ChannelPayment, error) {
	// Open CSV file
	file, err := os.Open(fileURL)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header failed: %w", err)
	}

	// Map column indices
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	// Parse records
	var payments []*service.ChannelPayment
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read line %d failed: %w", lineNum, err)
		}
		lineNum++

		// Extract fields (adjust based on actual Stripe CSV format)
		payment := &service.ChannelPayment{
			ChannelTradeNo: getField(record, columnMap, "id"),
			Status:         mapStripeStatus(getField(record, columnMap, "status")),
			Currency:       strings.ToUpper(getField(record, columnMap, "currency")),
		}

		// Parse amount (Stripe amounts are in cents)
		amountStr := getField(record, columnMap, "amount")
		if amountStr != "" {
			amount, err := strconv.ParseInt(amountStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse amount on line %d failed: %w", lineNum, err)
			}
			payment.Amount = amount
		}

		// Parse settlement time
		timeStr := getField(record, columnMap, "created")
		if timeStr != "" {
			timestamp, err := strconv.ParseInt(timeStr, 10, 64)
			if err == nil {
				payment.SettlementTime = time.Unix(timestamp, 0)
			}
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// downloadFile 下载文件到本地
func (d *StripeDownloader) downloadFile(url, fileNo string) (string, int64, string, error) {
	// Create directory if not exists
	if err := os.MkdirAll(d.basePath, 0755); err != nil {
		return "", 0, "", fmt.Errorf("create directory failed: %w", err)
	}

	// Local file path
	fileName := fmt.Sprintf("%s.csv", fileNo)
	localPath := filepath.Join(d.basePath, fileName)

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, "", fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create local file
	out, err := os.Create(localPath)
	if err != nil {
		return "", 0, "", fmt.Errorf("create local file failed: %w", err)
	}
	defer out.Close()

	// Calculate hash while copying
	hash := sha256.New()
	multiWriter := io.MultiWriter(out, hash)

	written, err := io.Copy(multiWriter, resp.Body)
	if err != nil {
		return "", 0, "", fmt.Errorf("copy file failed: %w", err)
	}

	fileHash := hex.EncodeToString(hash.Sum(nil))

	return localPath, written, fileHash, nil
}

// Helper functions

func generateFileNo(channel string, settlementDate time.Time) string {
	return fmt.Sprintf("FILE-%s-%s-%s",
		strings.ToUpper(channel),
		settlementDate.Format("20060102"),
		uuid.New().String()[:8],
	)
}

func getField(record []string, columnMap map[string]int, fieldName string) string {
	if idx, exists := columnMap[fieldName]; exists && idx < len(record) {
		return record[idx]
	}
	return ""
}

func mapStripeStatus(stripeStatus string) string {
	// Map Stripe charge status to platform status
	switch strings.ToLower(stripeStatus) {
	case "succeeded":
		return "success"
	case "pending":
		return "pending"
	case "failed":
		return "failed"
	case "refunded":
		return "refunded"
	default:
		return stripeStatus
	}
}
