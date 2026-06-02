package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchJobRole(ctx context.Context, options gorm_model.JobRoleFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.JobRole{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchJobRole Find:", err)
		return
	}

	return
}

func (r *gormRepo) CreateJobRole(ctx context.Context, model *gorm_model.JobRole) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateJobRole(ctx context.Context, model *gorm_model.JobRole) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteJobRole(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.JobRole{}, "id = ?", id).Error
}

func (r *gormRepo) GetActiveJobRoles(ctx context.Context) ([]gorm_model.JobRole, error) {
	var roles []gorm_model.JobRole
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("name asc").Find(&roles).Error
	return roles, err
}
