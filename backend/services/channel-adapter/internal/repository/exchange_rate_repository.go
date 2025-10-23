package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"payment-platform/channel-adapter/internal/model"
)

// ExchangeRateRepository 汇率仓库接口
type ExchangeRateRepository interface {
	// 保存单个汇率记录
	SaveRate(ctx context.Context, rate *model.ExchangeRate) error

	// 保存汇率快照（批量）
	SaveSnapshot(ctx context.Context, snapshot *model.ExchangeRateSnapshot) error

	// 获取最新汇率
	GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*model.ExchangeRate, error)

	// 获取指定时间的汇率
	GetRateAtTime(ctx context.Context, baseCurrency, targetCurrency string, timestamp time.Time) (*model.ExchangeRate, error)

	// 查询时间范围内的汇率历史
	GetRateHistory(ctx context.Context, baseCurrency, targetCurrency string, startTime, endTime time.Time) ([]*model.ExchangeRate, error)

	// 获取最新的汇率快照
	GetLatestSnapshot(ctx context.Context, baseCurrency string) (*model.ExchangeRateSnapshot, error)

	// 查询时间范围内的快照
	GetSnapshotHistory(ctx context.Context, baseCurrency string, startTime, endTime time.Time) ([]*model.ExchangeRateSnapshot, error)
}

// exchangeRateRepository 汇率仓库实现
type exchangeRateRepository struct {
	db *gorm.DB
}

// NewExchangeRateRepository 创建汇率仓库实例
func NewExchangeRateRepository(db *gorm.DB) ExchangeRateRepository {
	return &exchangeRateRepository{db: db}
}

// SaveRate 保存单个汇率记录
func (r *exchangeRateRepository) SaveRate(ctx context.Context, rate *model.ExchangeRate) error {
	// 设置生效时间为当前时间（如果未设置）
	if rate.ValidFrom.IsZero() {
		rate.ValidFrom = time.Now()
	}

	return r.db.WithContext(ctx).Create(rate).Error
}

// SaveSnapshot 保存汇率快照
func (r *exchangeRateRepository) SaveSnapshot(ctx context.Context, snapshot *model.ExchangeRateSnapshot) error {
	// 设置快照时间为当前时间（如果未设置）
	if snapshot.SnapshotTime.IsZero() {
		snapshot.SnapshotTime = time.Now()
	}

	return r.db.WithContext(ctx).Create(snapshot).Error
}

// GetLatestRate 获取最新汇率
func (r *exchangeRateRepository) GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*model.ExchangeRate, error) {
	var rate model.ExchangeRate
	err := r.db.WithContext(ctx).
		Where("base_currency = ? AND target_currency = ?", baseCurrency, targetCurrency).
		Order("valid_from DESC").
		First(&rate).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rate, err
}

// GetRateAtTime 获取指定时间的汇率
func (r *exchangeRateRepository) GetRateAtTime(ctx context.Context, baseCurrency, targetCurrency string, timestamp time.Time) (*model.ExchangeRate, error) {
	var rate model.ExchangeRate
	err := r.db.WithContext(ctx).
		Where("base_currency = ? AND target_currency = ?", baseCurrency, targetCurrency).
		Where("valid_from <= ?", timestamp).
		Where("(valid_to IS NULL OR valid_to > ?)", timestamp).
		Order("valid_from DESC").
		First(&rate).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rate, err
}

// GetRateHistory 查询时间范围内的汇率历史
func (r *exchangeRateRepository) GetRateHistory(ctx context.Context, baseCurrency, targetCurrency string, startTime, endTime time.Time) ([]*model.ExchangeRate, error) {
	var rates []*model.ExchangeRate
	err := r.db.WithContext(ctx).
		Where("base_currency = ? AND target_currency = ?", baseCurrency, targetCurrency).
		Where("valid_from >= ? AND valid_from <= ?", startTime, endTime).
		Order("valid_from DESC").
		Find(&rates).Error

	return rates, err
}

// GetLatestSnapshot 获取最新的汇率快照
func (r *exchangeRateRepository) GetLatestSnapshot(ctx context.Context, baseCurrency string) (*model.ExchangeRateSnapshot, error) {
	var snapshot model.ExchangeRateSnapshot
	err := r.db.WithContext(ctx).
		Where("base_currency = ?", baseCurrency).
		Order("snapshot_time DESC").
		First(&snapshot).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &snapshot, err
}

// GetSnapshotHistory 查询时间范围内的快照
func (r *exchangeRateRepository) GetSnapshotHistory(ctx context.Context, baseCurrency string, startTime, endTime time.Time) ([]*model.ExchangeRateSnapshot, error) {
	var snapshots []*model.ExchangeRateSnapshot
	err := r.db.WithContext(ctx).
		Where("base_currency = ?", baseCurrency).
		Where("snapshot_time >= ? AND snapshot_time <= ?", startTime, endTime).
		Order("snapshot_time DESC").
		Find(&snapshots).Error

	return snapshots, err
}
