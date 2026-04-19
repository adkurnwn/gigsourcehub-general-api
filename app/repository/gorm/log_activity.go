package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
)

func (r *gormRepo) CreateLogActivity(ctx context.Context, model *gorm_model.LogActivity) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) FetchLogActivity(ctx context.Context, options gorm_model.LogActivityFilter) ([]gorm_model.LogActivity, error) {
	var results []gorm_model.LogActivity
	db := r.db.WithContext(ctx).Model(&gorm_model.LogActivity{}).Preload("Actor.SystemRole")
	options.Query(db)

	err := db.Order("created_at DESC").Find(&results).Error
	return results, err
}

func (r *gormRepo) FetchCountLogActivity(ctx context.Context, options gorm_model.LogActivityFilter) (int64, error) {
	var total int64
	db := r.db.WithContext(ctx).Model(&gorm_model.LogActivity{})
	options.Query(db)

	err := db.Count(&total).Error
	return total, err
}
