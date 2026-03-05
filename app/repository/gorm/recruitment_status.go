package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchRecruitmentStatus(ctx context.Context, options gorm_model.RecruitmentStatusFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.RecruitmentStatus{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchRecruitmentStatus Find:", err)
		return
	}

	return
}

func (r *gormRepo) CreateRecruitmentStatus(ctx context.Context, model *gorm_model.RecruitmentStatus) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateRecruitmentStatus(ctx context.Context, model *gorm_model.RecruitmentStatus) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteRecruitmentStatus(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.RecruitmentStatus{}, "id = ?", id).Error
}
