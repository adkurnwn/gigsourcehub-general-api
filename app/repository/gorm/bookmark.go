package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) CreateBookmark(ctx context.Context, model *gorm_model.Bookmark) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) GetBookmark(ctx context.Context, adminID, candidateID string) (*gorm_model.Bookmark, error) {
	var bookmark gorm_model.Bookmark
	err := r.db.WithContext(ctx).
		Where("admin_id = ? AND candidate_id = ?", adminID, candidateID).
		First(&bookmark).Error
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

func (r *gormRepo) DeleteBookmark(ctx context.Context, adminID, candidateID string) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("admin_id = ? AND candidate_id = ?", adminID, candidateID).
		Delete(&gorm_model.Bookmark{})
	return result.RowsAffected, result.Error
}

func (r *gormRepo) FetchBookmarksByAdmin(ctx context.Context, adminID string, limit, offset int64) (cur *sql.Rows, err error) {
	q := r.db.WithContext(ctx).Model(&gorm_model.Bookmark{}).Where("admin_id = ?", adminID).
		Order("created_at DESC").
		Limit(int(limit)).Offset(int(offset))

	cur, err = q.Rows()
	if err != nil {
		logrus.Error("FetchBookmarksByAdmin error: ", err)
		return
	}

	return
}

func (r *gormRepo) CountBookmarksByAdmin(ctx context.Context, adminID string) (total int64, err error) {
	err = r.db.WithContext(ctx).Model(&gorm_model.Bookmark{}).Where("admin_id = ?", adminID).Count(&total).Error
	if err != nil {
		logrus.Error("CountBookmarksByAdmin error: ", err)
	}
	return
}
