package gormrepo

import (
	"database/sql"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type gormRepo struct {
	db *gorm.DB
}

func NewGormRepo(db *sqlx.DB, log logger.Interface) domain.GormRepo {
	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn: db,
		}),
		&gorm.Config{
			Logger: log,
		},
	)

	if err != nil {
		panic("error connecting to gorm: " + err.Error())
	}

	return &gormRepo{
		db: gormDB,
	}
}

func (r *gormRepo) StructScan(rows *sql.Rows, dest any) error {
	return r.db.ScanRows(rows, dest)
}
