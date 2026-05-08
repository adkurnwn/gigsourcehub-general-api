package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchJobVacancy(ctx context.Context, options gorm_model.JobVacancyFilter) (cur *sql.Rows, err error) {
	q := r.db.Model(&gorm_model.JobVacancy{}).Preload("Subrequest")
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchJobVacancy error:", err)
		return
	}
	return
}

func (r *gormRepo) GetJobVacancyByID(ctx context.Context, id string) (*gorm_model.JobVacancy, error) {
	var vacancy gorm_model.JobVacancy
	err := r.db.WithContext(ctx).
		Preload("Subrequest").
		Where("id = ?", id).
		First(&vacancy).Error
	if err != nil {
		return nil, err
	}
	return &vacancy, nil
}

func (r *gormRepo) CreateJobVacancy(ctx context.Context, model *gorm_model.JobVacancy) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateJobVacancy(ctx context.Context, model *gorm_model.JobVacancy) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteJobVacancy(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.JobVacancy{}, "id = ?", id).Error
}
