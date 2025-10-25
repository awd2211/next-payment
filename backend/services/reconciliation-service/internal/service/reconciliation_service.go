package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payment-platform/reconciliation-service/internal/model"
	"payment-platform/reconciliation-service/internal/repository"
)

// ReconciliationService 对账服务接口
type ReconciliationService interface {
	// Task management
	CreateTask(ctx context.Context, input *CreateTaskInput) (*model.ReconciliationTask, error)
	ExecuteTask(ctx context.Context, taskID uuid.UUID) error
	GetTaskDetails(ctx context.Context, taskID uuid.UUID) (*TaskDetails, error)
	ListTasks(ctx context.Context, filters *TaskFilters, page, pageSize int) (*TaskListResult, error)
	RetryTask(ctx context.Context, taskID uuid.UUID) error

	// Record management
	ResolveRecord(ctx context.Context, recordID, resolvedBy uuid.UUID, note string) error
	ListRecords(ctx context.Context, filters *RecordFilters, page, pageSize int) (*RecordListResult, error)
	GetRecordDetails(ctx context.Context, recordID uuid.UUID) (*model.ReconciliationRecord, error)

	// File management
	DownloadSettlementFile(ctx context.Context, channel string, settlementDate time.Time) (*model.ChannelSettlementFile, error)
	GetFileDetails(ctx context.Context, fileNo string) (*model.ChannelSettlementFile, error)
	ListFiles(ctx context.Context, filters *FileFilters, page, pageSize int) (*FileListResult, error)

	// Report generation
	GenerateReport(ctx context.Context, taskID uuid.UUID) (string, error)
}

// Input/Output DTOs
type CreateTaskInput struct {
	TaskDate time.Time `json:"task_date" binding:"required"`
	Channel  string    `json:"channel" binding:"required"`
	TaskType string    `json:"task_type" binding:"required"` // daily, manual, reconcile
}

type TaskFilters struct {
	TaskDate  *time.Time
	Channel   string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
}

type RecordFilters struct {
	TaskID     *uuid.UUID
	DiffType   string
	IsResolved *bool
	MerchantID *uuid.UUID
}

type FileFilters struct {
	Channel        string
	SettlementDate *time.Time
	Status         string
	StartDate      *time.Time
	EndDate        *time.Time
}

type TaskDetails struct {
	Task    *model.ReconciliationTask      `json:"task"`
	Records []*model.ReconciliationRecord  `json:"records,omitempty"`
	File    *model.ChannelSettlementFile   `json:"file,omitempty"`
	Summary *TaskSummary                   `json:"summary"`
}

type TaskSummary struct {
	TotalCount       int     `json:"total_count"`
	MatchedCount     int     `json:"matched_count"`
	DiffCount        int     `json:"diff_count"`
	MatchRate        float64 `json:"match_rate"`
	UnresolvedCount  int     `json:"unresolved_count"`
	PlatformOnlyCount int    `json:"platform_only_count"`
	ChannelOnlyCount  int    `json:"channel_only_count"`
	AmountDiffCount   int    `json:"amount_diff_count"`
	StatusDiffCount   int    `json:"status_diff_count"`
}

type TaskListResult struct {
	Tasks      []*model.ReconciliationTask `json:"tasks"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

type RecordListResult struct {
	Records    []*model.ReconciliationRecord `json:"records"`
	Total      int64                         `json:"total"`
	Page       int                           `json:"page"`
	PageSize   int                           `json:"page_size"`
	TotalPages int                           `json:"total_pages"`
}

type FileListResult struct {
	Files      []*model.ChannelSettlementFile `json:"files"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

// reconciliationService 对账服务实现
type reconciliationService struct {
	repo              repository.ReconciliationRepository
	db                *gorm.DB
	channelDownloader ChannelDownloader
	platformFetcher   PlatformDataFetcher
	reportGenerator   ReportGenerator
}

// NewReconciliationService 创建对账服务实例
func NewReconciliationService(
	repo repository.ReconciliationRepository,
	db *gorm.DB,
	channelDownloader ChannelDownloader,
	platformFetcher PlatformDataFetcher,
	reportGenerator ReportGenerator,
) ReconciliationService {
	return &reconciliationService{
		repo:              repo,
		db:                db,
		channelDownloader: channelDownloader,
		platformFetcher:   platformFetcher,
		reportGenerator:   reportGenerator,
	}
}

// CreateTask 创建对账任务
func (s *reconciliationService) CreateTask(ctx context.Context, input *CreateTaskInput) (*model.ReconciliationTask, error) {
	// Check if task already exists for this date and channel
	existing, err := s.repo.GetTaskByDateAndChannel(ctx, input.TaskDate, input.Channel)
	if err != nil {
		return nil, fmt.Errorf("check existing task failed: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("task already exists for date %s and channel %s", input.TaskDate.Format("2006-01-02"), input.Channel)
	}

	// Generate task number
	taskNo := generateTaskNo(input.Channel, input.TaskDate)

	task := &model.ReconciliationTask{
		TaskNo:   taskNo,
		TaskDate: input.TaskDate,
		Channel:  input.Channel,
		TaskType: input.TaskType,
		Status:   model.TaskStatusPending,
		Progress: 0,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	return task, nil
}

// ExecuteTask 执行对账任务
func (s *reconciliationService) ExecuteTask(ctx context.Context, taskID uuid.UUID) error {
	// Get task
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task failed: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	// Check status
	if task.Status == model.TaskStatusCompleted {
		return fmt.Errorf("task already completed")
	}
	if task.Status == model.TaskStatusProcessing {
		return fmt.Errorf("task is already processing")
	}

	// Update status to processing
	now := time.Now()
	task.Status = model.TaskStatusProcessing
	task.StartedAt = &now
	task.Progress = 0
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("update task status failed: %w", err)
	}

	// Execute reconciliation in background (in real implementation, use goroutine or queue)
	if err := s.executeReconciliation(ctx, task); err != nil {
		// Update task status to failed
		task.Status = model.TaskStatusFailed
		task.ErrorMessage = err.Error()
		s.repo.UpdateTask(ctx, task)
		return fmt.Errorf("execute reconciliation failed: %w", err)
	}

	return nil
}

// executeReconciliation 执行对账核心逻辑
func (s *reconciliationService) executeReconciliation(ctx context.Context, task *model.ReconciliationTask) error {
	// Step 1: Download channel settlement file (10% progress)
	task.Progress = 10
	s.repo.UpdateTask(ctx, task)

	channelFile, err := s.channelDownloader.Download(ctx, task.Channel, task.TaskDate)
	if err != nil {
		return fmt.Errorf("download channel file failed: %w", err)
	}
	task.ChannelFileURL = channelFile.FileURL

	// Step 2: Fetch platform payment data (30% progress)
	task.Progress = 30
	s.repo.UpdateTask(ctx, task)

	platformRecords, err := s.platformFetcher.FetchPayments(ctx, task.TaskDate, task.Channel)
	if err != nil {
		return fmt.Errorf("fetch platform data failed: %w", err)
	}

	// Step 3: Parse channel file (50% progress)
	task.Progress = 50
	s.repo.UpdateTask(ctx, task)

	channelRecords, err := s.channelDownloader.Parse(ctx, channelFile.FileURL)
	if err != nil {
		return fmt.Errorf("parse channel file failed: %w", err)
	}

	// Step 4: Three-way matching (70% progress)
	task.Progress = 70
	s.repo.UpdateTask(ctx, task)

	diffRecords := s.performMatching(task, platformRecords, channelRecords)

	// Step 5: Save diff records (90% progress)
	task.Progress = 90
	s.repo.UpdateTask(ctx, task)

	if len(diffRecords) > 0 {
		if err := s.repo.BatchCreateRecords(ctx, diffRecords); err != nil {
			return fmt.Errorf("save diff records failed: %w", err)
		}
	}

	// Step 6: Update task statistics (100% progress)
	task.PlatformCount = len(platformRecords)
	task.ChannelCount = len(channelRecords)
	task.PlatformAmount = calculateTotalAmount(platformRecords)
	task.ChannelAmount = calculateChannelTotalAmount(channelRecords)

	matchedCount := 0
	matchedAmount := int64(0)
	diffCount := 0
	diffAmount := int64(0)

	for _, record := range diffRecords {
		if record.DiffType == model.DiffTypeMatched {
			matchedCount++
			matchedAmount += record.PlatformAmount
		} else {
			diffCount++
			diffAmount += record.DiffAmount
		}
	}

	task.MatchedCount = matchedCount
	task.MatchedAmount = matchedAmount
	task.DiffCount = diffCount
	task.DiffAmount = diffAmount
	task.Status = model.TaskStatusCompleted
	task.Progress = 100
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("update task statistics failed: %w", err)
	}

	return nil
}

// performMatching 三方匹配算法
func (s *reconciliationService) performMatching(
	task *model.ReconciliationTask,
	platformRecords []*PlatformPayment,
	channelRecords []*ChannelPayment,
) []*model.ReconciliationRecord {
	var diffRecords []*model.ReconciliationRecord

	// Build channel record map for fast lookup
	channelMap := make(map[string]*ChannelPayment)
	for _, cr := range channelRecords {
		channelMap[cr.ChannelTradeNo] = cr
	}

	// Match platform records with channel records
	matchedChannelTradeNos := make(map[string]bool)

	for _, pr := range platformRecords {
		cr, exists := channelMap[pr.ChannelTradeNo]

		if !exists {
			// Platform only
			diffRecords = append(diffRecords, &model.ReconciliationRecord{
				TaskID:         task.ID,
				TaskNo:         task.TaskNo,
				PaymentNo:      pr.PaymentNo,
				ChannelTradeNo: pr.ChannelTradeNo,
				OrderNo:        pr.OrderNo,
				MerchantID:     pr.MerchantID,
				PlatformAmount: pr.Amount,
				ChannelAmount:  0,
				DiffAmount:     pr.Amount,
				Currency:       pr.Currency,
				PlatformStatus: pr.Status,
				ChannelStatus:  "",
				DiffType:       model.DiffTypePlatformOnly,
				DiffReason:     "Platform record not found in channel settlement",
				IsResolved:     false,
			})
			continue
		}

		matchedChannelTradeNos[cr.ChannelTradeNo] = true

		// Check for differences
		diffType := model.DiffTypeMatched
		diffReason := ""
		diffAmount := int64(0)

		if pr.Amount != cr.Amount {
			diffType = model.DiffTypeAmountDiff
			diffReason = fmt.Sprintf("Amount mismatch: platform=%d, channel=%d", pr.Amount, cr.Amount)
			diffAmount = pr.Amount - cr.Amount
		} else if pr.Status != cr.Status {
			diffType = model.DiffTypeStatusDiff
			diffReason = fmt.Sprintf("Status mismatch: platform=%s, channel=%s", pr.Status, cr.Status)
		}

		diffRecords = append(diffRecords, &model.ReconciliationRecord{
			TaskID:         task.ID,
			TaskNo:         task.TaskNo,
			PaymentNo:      pr.PaymentNo,
			ChannelTradeNo: pr.ChannelTradeNo,
			OrderNo:        pr.OrderNo,
			MerchantID:     pr.MerchantID,
			PlatformAmount: pr.Amount,
			ChannelAmount:  cr.Amount,
			DiffAmount:     diffAmount,
			Currency:       pr.Currency,
			PlatformStatus: pr.Status,
			ChannelStatus:  cr.Status,
			DiffType:       diffType,
			DiffReason:     diffReason,
			IsResolved:     diffType == model.DiffTypeMatched,
		})
	}

	// Find channel-only records
	for _, cr := range channelRecords {
		if !matchedChannelTradeNos[cr.ChannelTradeNo] {
			diffRecords = append(diffRecords, &model.ReconciliationRecord{
				TaskID:         task.ID,
				TaskNo:         task.TaskNo,
				ChannelTradeNo: cr.ChannelTradeNo,
				PlatformAmount: 0,
				ChannelAmount:  cr.Amount,
				DiffAmount:     -cr.Amount,
				Currency:       cr.Currency,
				PlatformStatus: "",
				ChannelStatus:  cr.Status,
				DiffType:       model.DiffTypeChannelOnly,
				DiffReason:     "Channel record not found in platform database",
				IsResolved:     false,
			})
		}
	}

	return diffRecords
}

// GetTaskDetails 获取任务详情
func (s *reconciliationService) GetTaskDetails(ctx context.Context, taskID uuid.UUID) (*TaskDetails, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task failed: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Get diff records summary
	filters := repository.RecordFilters{TaskID: &taskID}
	records, _, err := s.repo.ListRecords(ctx, filters, 1, 10) // Get first 10 records
	if err != nil {
		return nil, fmt.Errorf("get records failed: %w", err)
	}

	// Calculate summary
	unresolvedCount := 0
	platformOnlyCount := 0
	channelOnlyCount := 0
	amountDiffCount := 0
	statusDiffCount := 0

	for _, record := range records {
		if !record.IsResolved {
			unresolvedCount++
		}
		switch record.DiffType {
		case model.DiffTypePlatformOnly:
			platformOnlyCount++
		case model.DiffTypeChannelOnly:
			channelOnlyCount++
		case model.DiffTypeAmountDiff:
			amountDiffCount++
		case model.DiffTypeStatusDiff:
			statusDiffCount++
		}
	}

	totalCount := task.PlatformCount + task.ChannelCount
	matchRate := 0.0
	if totalCount > 0 {
		matchRate = float64(task.MatchedCount) / float64(totalCount) * 100
	}

	summary := &TaskSummary{
		TotalCount:        totalCount,
		MatchedCount:      task.MatchedCount,
		DiffCount:         task.DiffCount,
		MatchRate:         matchRate,
		UnresolvedCount:   unresolvedCount,
		PlatformOnlyCount: platformOnlyCount,
		ChannelOnlyCount:  channelOnlyCount,
		AmountDiffCount:   amountDiffCount,
		StatusDiffCount:   statusDiffCount,
	}

	return &TaskDetails{
		Task:    task,
		Records: records,
		Summary: summary,
	}, nil
}

// ListTasks 查询任务列表
func (s *reconciliationService) ListTasks(ctx context.Context, filters *TaskFilters, page, pageSize int) (*TaskListResult, error) {
	repoFilters := repository.TaskFilters{
		TaskDate:  filters.TaskDate,
		Channel:   filters.Channel,
		Status:    filters.Status,
		StartDate: filters.StartDate,
		EndDate:   filters.EndDate,
	}

	tasks, total, err := s.repo.ListTasks(ctx, repoFilters, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list tasks failed: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &TaskListResult{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// RetryTask 重试失败的任务
func (s *reconciliationService) RetryTask(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task failed: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	if task.Status != model.TaskStatusFailed {
		return fmt.Errorf("only failed tasks can be retried")
	}

	// Reset task status
	task.Status = model.TaskStatusPending
	task.Progress = 0
	task.ErrorMessage = ""
	task.StartedAt = nil
	task.CompletedAt = nil

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("reset task status failed: %w", err)
	}

	// Execute task
	return s.ExecuteTask(ctx, taskID)
}

// ResolveRecord 标记差异已解决
func (s *reconciliationService) ResolveRecord(ctx context.Context, recordID, resolvedBy uuid.UUID, note string) error {
	record, err := s.repo.GetRecordByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("get record failed: %w", err)
	}
	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.IsResolved {
		return fmt.Errorf("record already resolved")
	}

	if err := s.repo.ResolveRecord(ctx, recordID, resolvedBy, note); err != nil {
		return fmt.Errorf("resolve record failed: %w", err)
	}

	return nil
}

// ListRecords 查询差异记录列表
func (s *reconciliationService) ListRecords(ctx context.Context, filters *RecordFilters, page, pageSize int) (*RecordListResult, error) {
	repoFilters := repository.RecordFilters{
		TaskID:     filters.TaskID,
		DiffType:   filters.DiffType,
		IsResolved: filters.IsResolved,
		MerchantID: filters.MerchantID,
	}

	records, total, err := s.repo.ListRecords(ctx, repoFilters, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list records failed: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &RecordListResult{
		Records:    records,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetRecordDetails 获取差异记录详情
func (s *reconciliationService) GetRecordDetails(ctx context.Context, recordID uuid.UUID) (*model.ReconciliationRecord, error) {
	record, err := s.repo.GetRecordByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("get record failed: %w", err)
	}
	if record == nil {
		return nil, fmt.Errorf("record not found")
	}
	return record, nil
}

// DownloadSettlementFile 下载渠道账单文件
func (s *reconciliationService) DownloadSettlementFile(ctx context.Context, channel string, settlementDate time.Time) (*model.ChannelSettlementFile, error) {
	// Check if file already exists
	existing, err := s.repo.GetFileByDateAndChannel(ctx, settlementDate, channel)
	if err != nil {
		return nil, fmt.Errorf("check existing file failed: %w", err)
	}
	if existing != nil && existing.Status == model.FileStatusImported {
		return existing, nil
	}

	// Download file
	file, err := s.channelDownloader.Download(ctx, channel, settlementDate)
	if err != nil {
		return nil, fmt.Errorf("download file failed: %w", err)
	}

	return file, nil
}

// GetFileDetails 获取文件详情
func (s *reconciliationService) GetFileDetails(ctx context.Context, fileNo string) (*model.ChannelSettlementFile, error) {
	file, err := s.repo.GetFileByNo(ctx, fileNo)
	if err != nil {
		return nil, fmt.Errorf("get file failed: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}
	return file, nil
}

// ListFiles 查询文件列表
func (s *reconciliationService) ListFiles(ctx context.Context, filters *FileFilters, page, pageSize int) (*FileListResult, error) {
	repoFilters := repository.FileFilters{
		Channel:        filters.Channel,
		SettlementDate: filters.SettlementDate,
		Status:         filters.Status,
		StartDate:      filters.StartDate,
		EndDate:        filters.EndDate,
	}

	files, total, err := s.repo.ListFiles(ctx, repoFilters, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list files failed: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &FileListResult{
		Files:      files,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GenerateReport 生成对账报告
func (s *reconciliationService) GenerateReport(ctx context.Context, taskID uuid.UUID) (string, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("get task failed: %w", err)
	}
	if task == nil {
		return "", fmt.Errorf("task not found")
	}

	if task.Status != model.TaskStatusCompleted {
		return "", fmt.Errorf("only completed tasks can generate reports")
	}

	// Generate report (delegated to report generator)
	reportURL, err := s.reportGenerator.Generate(ctx, task)
	if err != nil {
		return "", fmt.Errorf("generate report failed: %w", err)
	}

	// Update task with report URL
	task.ReportFileURL = reportURL
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return "", fmt.Errorf("update task report URL failed: %w", err)
	}

	return reportURL, nil
}

// Helper functions

func generateTaskNo(channel string, taskDate time.Time) string {
	return fmt.Sprintf("RECON-%s-%s-%d",
		channel,
		taskDate.Format("20060102"),
		time.Now().Unix()%10000,
	)
}

func calculateTotalAmount(records []*PlatformPayment) int64 {
	total := int64(0)
	for _, r := range records {
		total += r.Amount
	}
	return total
}

func calculateChannelTotalAmount(records []*ChannelPayment) int64 {
	total := int64(0)
	for _, r := range records {
		total += r.Amount
	}
	return total
}
