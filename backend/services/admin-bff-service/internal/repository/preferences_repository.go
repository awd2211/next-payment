package repository

import (
	"context"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"gorm.io/gorm"
)

// PreferencesRepository 用户偏好设置仓储接口
type PreferencesRepository interface {
	Create(ctx context.Context, prefs *model.UserPreferences) error
	GetByUserID(ctx context.Context, userID uuid.UUID, userType string) (*model.UserPreferences, error)
	Update(ctx context.Context, prefs *model.UserPreferences) error
	Delete(ctx context.Context, userID uuid.UUID, userType string) error
}

type preferencesRepository struct {
	db *gorm.DB
}

// NewPreferencesRepository 创建用户偏好设置仓储实例
func NewPreferencesRepository(db *gorm.DB) PreferencesRepository {
	return &preferencesRepository{db: db}
}

// Create 创建用户偏好设置
func (r *preferencesRepository) Create(ctx context.Context, prefs *model.UserPreferences) error {
	return r.db.WithContext(ctx).Create(prefs).Error
}

// GetByUserID 根据用户ID获取偏好设置
func (r *preferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID, userType string) (*model.UserPreferences, error) {
	var prefs model.UserPreferences
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		First(&prefs).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &prefs, nil
}

// Update 更新用户偏好设置
func (r *preferencesRepository) Update(ctx context.Context, prefs *model.UserPreferences) error {
	return r.db.WithContext(ctx).Save(prefs).Error
}

// Delete 删除用户偏好设置
func (r *preferencesRepository) Delete(ctx context.Context, userID uuid.UUID, userType string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		Delete(&model.UserPreferences{}).Error
}
