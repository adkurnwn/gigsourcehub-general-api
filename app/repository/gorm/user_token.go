package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
)

func (r *gormRepo) CreateUserToken(ctx context.Context, model *gorm_model.UserToken) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) GetUserToken(ctx context.Context, token string, tokenType string) (*gorm_model.UserToken, error) {
	var model gorm_model.UserToken
	err := r.db.WithContext(ctx).Where("token = ? AND type = ?", token, tokenType).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *gormRepo) DeleteUserToken(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.UserToken{}, "id = ?", id).Error
}

func (r *gormRepo) DeleteUserTokensByUserID(ctx context.Context, userID string, tokenType string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.UserToken{}, "user_id = ? AND type = ?", userID, tokenType).Error
}
