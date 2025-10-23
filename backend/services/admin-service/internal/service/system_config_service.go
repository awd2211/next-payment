package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

var (
	ErrConfigNotFound = errors.New("配置不存在")
	ErrConfigExists   = errors.New("配置键已存在")
)

// SystemConfigService 系统配置服务接口
type SystemConfigService interface {
	CreateConfig(ctx context.Context, req *CreateConfigRequest) (*model.SystemConfig, error)
	GetConfig(ctx context.Context, id uuid.UUID) (*model.SystemConfig, error)
	GetConfigByKey(ctx context.Context, key string) (*model.SystemConfig, error)
	ListConfigs(ctx context.Context, category string, page, pageSize int) ([]*model.SystemConfig, int64, error)
	ListConfigsByCategory(ctx context.Context) (map[string][]*model.SystemConfig, error)
	UpdateConfig(ctx context.Context, req *UpdateConfigRequest) (*model.SystemConfig, error)
	DeleteConfig(ctx context.Context, id uuid.UUID) error
	BatchUpdateConfigs(ctx context.Context, configs []UpdateConfigRequest) error
}

type systemConfigService struct {
	configRepo repository.SystemConfigRepository
}

// NewSystemConfigService 创建系统配置服务实例
func NewSystemConfigService(
	configRepo repository.SystemConfigRepository,
) SystemConfigService {
	return &systemConfigService{
		configRepo: configRepo,
	}
}

// CreateConfigRequest 创建配置请求
type CreateConfigRequest struct {
	Key         string
	Value       string
	Type        string // string, number, boolean, json
	Category    string
	Description string
	IsPublic    bool
	UpdatedBy   uuid.UUID
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	ID          uuid.UUID
	Key         string
	Value       string
	Type        string
	Category    string
	Description string
	IsPublic    bool
	UpdatedBy   uuid.UUID
}

// CreateConfig 创建系统配置
func (s *systemConfigService) CreateConfig(ctx context.Context, req *CreateConfigRequest) (*model.SystemConfig, error) {
	// 检查键是否已存在
	existingConfig, err := s.configRepo.GetByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	if existingConfig != nil {
		return nil, ErrConfigExists
	}

	config := &model.SystemConfig{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UpdatedBy:   req.UpdatedBy,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetConfig 获取配置详情
func (s *systemConfigService) GetConfig(ctx context.Context, id uuid.UUID) (*model.SystemConfig, error) {
	config, err := s.configRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}
	return config, nil
}

// GetConfigByKey 根据键获取配置
func (s *systemConfigService) GetConfigByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	config, err := s.configRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}
	return config, nil
}

// ListConfigs 获取配置列表
func (s *systemConfigService) ListConfigs(ctx context.Context, category string, page, pageSize int) ([]*model.SystemConfig, int64, error) {
	return s.configRepo.List(ctx, category, page, pageSize)
}

// ListConfigsByCategory 按类别分组获取配置
func (s *systemConfigService) ListConfigsByCategory(ctx context.Context) (map[string][]*model.SystemConfig, error) {
	return s.configRepo.ListByCategory(ctx)
}

// UpdateConfig 更新配置
func (s *systemConfigService) UpdateConfig(ctx context.Context, req *UpdateConfigRequest) (*model.SystemConfig, error) {
	// 获取现有配置
	config, err := s.configRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}

	// 如果修改了键，检查新键是否已存在
	if req.Key != "" && req.Key != config.Key {
		existingConfig, err := s.configRepo.GetByKey(ctx, req.Key)
		if err != nil {
			return nil, err
		}
		if existingConfig != nil {
			return nil, ErrConfigExists
		}
		config.Key = req.Key
	}

	// 更新字段
	if req.Value != "" {
		config.Value = req.Value
	}
	if req.Type != "" {
		config.Type = req.Type
	}
	if req.Category != "" {
		config.Category = req.Category
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	config.IsPublic = req.IsPublic
	config.UpdatedBy = req.UpdatedBy

	if err := s.configRepo.Update(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// DeleteConfig 删除配置
func (s *systemConfigService) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	config, err := s.configRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if config == nil {
		return ErrConfigNotFound
	}

	return s.configRepo.Delete(ctx, id)
}

// BatchUpdateConfigs 批量更新配置
func (s *systemConfigService) BatchUpdateConfigs(ctx context.Context, configs []UpdateConfigRequest) error {
	for _, req := range configs {
		if _, err := s.UpdateConfig(ctx, &req); err != nil {
			return err
		}
	}
	return nil
}
