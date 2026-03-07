package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) CreateRequest(ctx context.Context, model *gorm_model.Request) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		logrus.Errorf("CreateRequest DB Error: %v\n", err)
		return err
	}
	return nil
}
