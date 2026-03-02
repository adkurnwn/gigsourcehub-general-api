package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchRoleApplied(ctx context.Context, options gorm_model.RoleAppliedFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.RoleApplied{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchRoleApplied Find:", err)
		return
	}

	return
}

func (r *gormRepo) CreateRoleApplied(ctx context.Context, model *gorm_model.RoleApplied) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateRoleApplied(ctx context.Context, model *gorm_model.RoleApplied) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteRoleApplied(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.RoleApplied{}, "id = ?", id).Error
}
