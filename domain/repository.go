package domain

import (
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	sql_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/sql"
	storage_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/storage"
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/jmoiron/sqlx"
)


type SqlRepo interface {
	FetchUser(ctx context.Context, options sql_model.UserFilter) (*sqlx.Rows, error)
	FetchOneUser(ctx context.Context, options sql_model.UserFilter) (*sql_model.User, error)
	CountUser(ctx context.Context, options sql_model.UserFilter) int64
	CreateUser(ctx context.Context, model *sql_model.User) (err error)
}

type GormRepo interface {
	StructScan(rows *sql.Rows, dest any) error
	FetchUser(ctx context.Context, options gorm_model.UserFilter) (*sql.Rows, error)
	FetchOneUser(ctx context.Context, options gorm_model.UserFilter) (*gorm_model.User, error)
	CountUser(ctx context.Context, options gorm_model.UserFilter) int64
	CreateUser(ctx context.Context, model *gorm_model.User) (err error)
}

type CacheRepo interface {
	Enabled() bool
	GetTTL() time.Duration
	Get(ctx context.Context, key string) (value []byte, err error)
	Set(ctx context.Context, key string, value []byte, expiration *time.Duration) (err error)
}

type StorageRepo interface {
	GetPresignedLink(objectKey string, expires *time.Duration) string
	GetPublicLink(objectKey string) string
	UploadFilePublic(objectKey string, body io.Reader, contentType string) (uploadData *storage_model.UploadResponse, err error)
	UploadFilePrivate(objectKey string, body io.Reader, contentType string, expires *time.Duration) (uploadData *storage_model.UploadResponse, err error)
}
