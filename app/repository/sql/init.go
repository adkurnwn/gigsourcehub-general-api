package sqlrepo

import (
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/jmoiron/sqlx"
)

type sqlRepo struct {
	db        *sqlx.DB
	userTable string
}

func NewSqlRepo(db *sqlx.DB) domain.SqlRepo {
	return &sqlRepo{
		db:        db,
		userTable: "users",
	}
}
