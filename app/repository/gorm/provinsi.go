package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchProvinsi(ctx context.Context, options gorm_model.ProvinsiFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.Provinsi{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchProvinsi Find:", err)	
		return
	}

	return
}
