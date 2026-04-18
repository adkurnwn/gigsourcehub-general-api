package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchSystemRole(ctx context.Context, options gorm_model.SystemRoleFilter) (cur *sql.Rows, err error) {
	// generate query
	q := r.db.Model(&gorm_model.SystemRole{})
	options.Query(q)

	cur, err = q.WithContext(ctx).Rows()
	if err != nil {
		logrus.Error("FetchSystemRole Find:", err)
		return
	}

	return
}
