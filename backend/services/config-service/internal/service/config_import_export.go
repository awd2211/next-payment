package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
	"payment-platform/config-service/internal/repository"
)

// ConfigExport 配置导出结构
type ConfigExport struct {
	Version     string          `json:"version" yaml:"version"`
	ServiceName string          `json:"service_name" yaml:"service_name"`
	Environment string          `json:"environment" yaml:"environment"`
	ExportedAt  string          `json:"exported_at" yaml:"exported_at"`
	Configs     []ConfigItem    `json:"configs" yaml:"configs"`
	Flags       []FeatureFlagItem `json:"feature_flags,omitempty" yaml:"feature_flags,omitempty"`
}

type ConfigItem struct {
	ConfigKey   string `json:"config_key" yaml:"config_key"`
	ConfigValue string `json:"config_value" yaml:"config_value"`
	ValueType   string `json:"value_type" yaml:"value_type"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	IsEncrypted bool   `json:"is_encrypted" yaml:"is_encrypted"`
}

type FeatureFlagItem struct {
	FlagKey     string                 `json:"flag_key" yaml:"flag_key"`
	FlagName    string                 `json:"flag_name" yaml:"flag_name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool                   `json:"enabled" yaml:"enabled"`
	Conditions  map[string]interface{} `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Percentage  int                    `json:"percentage" yaml:"percentage"`
}

// ExportConfigs 导出配置为 JSON 或 YAML
func (s *configService) ExportConfigs(ctx context.Context, serviceName, environment, format string) ([]byte, error) {
	// 1. 查询所有配置
	configs, _, err := s.ListConfigs(ctx, &repository.ConfigQuery{
		ServiceName: serviceName,
		Environment: environment,
		Page:        1,
		PageSize:    1000,
	})
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}

	// 2. 查询功能开关
	flags, _, err := s.ListFeatureFlags(ctx, &repository.FeatureFlagQuery{
		Environment: environment,
		Page:        1,
		PageSize:    1000,
	})
	if err != nil {
		return nil, fmt.Errorf("查询功能开关失败: %w", err)
	}

	// 3. 构建导出数据
	export := &ConfigExport{
		Version:     "1.0",
		ServiceName: serviceName,
		Environment: environment,
		ExportedAt:  time.Now().Format(time.RFC3339),
		Configs:     make([]ConfigItem, 0, len(configs)),
		Flags:       make([]FeatureFlagItem, 0, len(flags)),
	}

	for _, cfg := range configs {
		export.Configs = append(export.Configs, ConfigItem{
			ConfigKey:   cfg.ConfigKey,
			ConfigValue: cfg.ConfigValue,
			ValueType:   cfg.ValueType,
			Description: cfg.Description,
			IsEncrypted: cfg.IsEncrypted,
		})
	}

	for _, flag := range flags {
		export.Flags = append(export.Flags, FeatureFlagItem{
			FlagKey:     flag.FlagKey,
			FlagName:    flag.FlagName,
			Description: flag.Description,
			Enabled:     flag.Enabled,
			Conditions:  flag.Conditions,
			Percentage:  flag.Percentage,
		})
	}

	// 4. 序列化为指定格式
	switch format {
	case "yaml", "yml":
		return yaml.Marshal(export)
	case "json":
		return json.MarshalIndent(export, "", "  ")
	default:
		return nil, fmt.Errorf("不支持的格式: %s (支持 json 或 yaml)", format)
	}
}

// ImportConfigs 从 JSON 或 YAML 导入配置
func (s *configService) ImportConfigs(ctx context.Context, data []byte, format, importedBy string, override bool) (*ImportResult, error) {
	// 1. 反序列化
	var export ConfigExport
	var err error

	switch format {
	case "yaml", "yml":
		err = yaml.Unmarshal(data, &export)
	case "json":
		err = json.Unmarshal(data, &export)
	default:
		return nil, fmt.Errorf("不支持的格式: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("解析导入数据失败: %w", err)
	}

	result := &ImportResult{
		TotalConfigs:   len(export.Configs),
		TotalFlags:     len(export.Flags),
		CreatedConfigs: 0,
		UpdatedConfigs: 0,
		SkippedConfigs: 0,
		CreatedFlags:   0,
		UpdatedFlags:   0,
		SkippedFlags:   0,
		Errors:         []string{},
	}

	// 2. 导入配置
	for _, item := range export.Configs {
		// 检查是否已存在
		existing, _ := s.GetConfig(ctx, export.ServiceName, item.ConfigKey, export.Environment)

		if existing != nil {
			if !override {
				result.SkippedConfigs++
				continue
			}
			// 更新现有配置
			_, err := s.UpdateConfig(ctx, existing.ID, &UpdateConfigInput{
				ConfigValue:  item.ConfigValue,
				Description:  item.Description,
				UpdatedBy:    importedBy,
				ChangeReason: "从导入文件更新",
			})
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("更新配置 %s 失败: %v", item.ConfigKey, err))
			} else {
				result.UpdatedConfigs++
			}
		} else {
			// 创建新配置
			_, err := s.CreateConfig(ctx, &CreateConfigInput{
				ServiceName: export.ServiceName,
				ConfigKey:   item.ConfigKey,
				ConfigValue: item.ConfigValue,
				ValueType:   item.ValueType,
				Environment: export.Environment,
				Description: item.Description,
				IsEncrypted: item.IsEncrypted,
				CreatedBy:   importedBy,
			})
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建配置 %s 失败: %v", item.ConfigKey, err))
			} else {
				result.CreatedConfigs++
			}
		}
	}

	// 3. 导入功能开关
	for _, item := range export.Flags {
		existing, _ := s.GetFeatureFlag(ctx, item.FlagKey)

		if existing != nil {
			if !override {
				result.SkippedFlags++
				continue
			}
			// 更新
			enabled := item.Enabled
			_, err := s.UpdateFeatureFlag(ctx, existing.ID, &UpdateFeatureFlagInput{
				FlagName:    item.FlagName,
				Description: item.Description,
				Enabled:     &enabled,
				Conditions:  item.Conditions,
				Percentage:  item.Percentage,
				UpdatedBy:   importedBy,
			})
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("更新功能开关 %s 失败: %v", item.FlagKey, err))
			} else {
				result.UpdatedFlags++
			}
		} else {
			// 创建
			_, err := s.CreateFeatureFlag(ctx, &CreateFeatureFlagInput{
				FlagKey:     item.FlagKey,
				FlagName:    item.FlagName,
				Description: item.Description,
				Enabled:     item.Enabled,
				Environment: export.Environment,
				Conditions:  item.Conditions,
				Percentage:  item.Percentage,
				CreatedBy:   importedBy,
			})
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建功能开关 %s 失败: %v", item.FlagKey, err))
			} else {
				result.CreatedFlags++
			}
		}
	}

	return result, nil
}

// ImportResult 导入结果统计
type ImportResult struct {
	TotalConfigs   int      `json:"total_configs"`
	CreatedConfigs int      `json:"created_configs"`
	UpdatedConfigs int      `json:"updated_configs"`
	SkippedConfigs int      `json:"skipped_configs"`
	TotalFlags     int      `json:"total_flags"`
	CreatedFlags   int      `json:"created_flags"`
	UpdatedFlags   int      `json:"updated_flags"`
	SkippedFlags   int      `json:"skipped_flags"`
	Errors         []string `json:"errors,omitempty"`
}
