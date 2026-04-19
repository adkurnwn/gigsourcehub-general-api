package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/google/uuid"
)

func (r *gormRepo) GetSystemSetting(ctx context.Context) (*gorm_model.SystemSetting, error) {
	var model gorm_model.SystemSetting
	err := r.db.WithContext(ctx).First(&model).Error
	if err != nil {
		// If not found, create a default one
		model.ID = uuid.New().String()
		model.IsAIModeEnabled = true
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return nil, err
		}
	}
	return &model, nil
}

func (r *gormRepo) UpdateSystemSetting(ctx context.Context, model *gorm_model.SystemSetting) error {
	return r.db.WithContext(ctx).Save(model).Error
}
