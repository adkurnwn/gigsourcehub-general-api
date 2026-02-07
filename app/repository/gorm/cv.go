package gormrepo

import (
    "context"
    gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
)

func (r *gormRepo) CreateCV(ctx context.Context, cv *gorm_model.CV) error {
    return r.db.WithContext(ctx).Create(cv).Error
}

func (r *gormRepo) FetchCVs(ctx context.Context, userID string) ([]gorm_model.CV, error) {
    var cvs []gorm_model.CV
    err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&cvs).Error
    return cvs, err
}
