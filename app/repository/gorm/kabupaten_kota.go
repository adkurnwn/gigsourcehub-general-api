package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchKabupatenKota(ctx context.Context, options gorm_model.KabupatenKotaFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.KabupatenKota{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchKabupatenKota Find:", err)	
		return
	}

	return
}
