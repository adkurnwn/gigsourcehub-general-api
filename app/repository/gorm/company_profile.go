package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/google/uuid"
)

func (r *gormRepo) GetCompanyProfile(ctx context.Context) (*gorm_model.CompanyProfile, error) {
	var model gorm_model.CompanyProfile
	err := r.db.WithContext(ctx).First(&model).Error
	if err != nil {
		// If not found, create a default one
		model.ID = uuid.New().String()
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return nil, err
		}
	}
	return &model, nil
}

func (r *gormRepo) UpdateCompanyProfile(ctx context.Context, model *gorm_model.CompanyProfile) error {
	return r.db.WithContext(ctx).Save(model).Error
}
