package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchCareerDepartment(ctx context.Context, options gorm_model.CareerDepartmentFilter) ([]gorm_model.CareerDepartment, error) {
	var items []gorm_model.CareerDepartment
	q := r.db.WithContext(ctx).Model(&gorm_model.CareerDepartment{})
	options.Query(q)

	if err := q.Find(&items).Error; err != nil {
		logrus.Error("FetchCareerDepartment error:", err)
		return nil, err
	}
	return items, nil
}

func (r *gormRepo) CountCareerDepartment(ctx context.Context, options gorm_model.CareerDepartmentFilter) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&gorm_model.CareerDepartment{})

	// Count without limit/offset
	countFilter := options
	countFilter.Limit = nil
	countFilter.Offset = nil
	countFilter.Query(q)

	if err := q.Count(&count).Error; err != nil {
		logrus.Error("CountCareerDepartment error:", err)
		return 0, err
	}
	return count, nil
}

func (r *gormRepo) GetCareerDepartmentByID(ctx context.Context, id string) (*gorm_model.CareerDepartment, error) {
	var item gorm_model.CareerDepartment
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormRepo) CreateCareerDepartment(ctx context.Context, model *gorm_model.CareerDepartment) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateCareerDepartment(ctx context.Context, model *gorm_model.CareerDepartment) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteCareerDepartment(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.CareerDepartment{}, "id = ?", id).Error
}
