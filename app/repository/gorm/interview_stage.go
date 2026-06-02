package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchInterviewStage(ctx context.Context, options gorm_model.InterviewStageFilter) (cur *sql.Rows, err error) {
	q := r.db.Model(&gorm_model.InterviewStage{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchInterviewStage Find:", err)
		return
	}

	return
}

func (r *gormRepo) CreateInterviewStage(ctx context.Context, model *gorm_model.InterviewStage) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateInterviewStage(ctx context.Context, model *gorm_model.InterviewStage) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteInterviewStage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.InterviewStage{}, "id = ?", id).Error
}