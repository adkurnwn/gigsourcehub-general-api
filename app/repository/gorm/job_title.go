package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchJobTitle(ctx context.Context, options gorm_model.JobTitleFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.JobTitle{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchJobTitle Find:", err)
		return
	}

	return
}

func (r *gormRepo) CreateJobTitle(ctx context.Context, model *gorm_model.JobTitle) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateJobTitle(ctx context.Context, model *gorm_model.JobTitle) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteJobTitle(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.JobTitle{}, "id = ?", id).Error
}
