package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchSector(ctx context.Context, options gorm_model.SectorFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.Sector{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchSector Find:", err)
		return
	}

	return
}

func (r *gormRepo) GetActiveSectors(ctx context.Context) ([]gorm_model.Sector, error) {
	var sectors []gorm_model.Sector
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("name asc").Find(&sectors).Error
	return sectors, err
}

func (r *gormRepo) CreateSector(ctx context.Context, model *gorm_model.Sector) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateSector(ctx context.Context, model *gorm_model.Sector) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteSector(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.Sector{}, "id = ?", id).Error
}
