package domain

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	storage_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/storage"
	"gorm.io/gorm"
)

type GormRepo interface {
	StructScan(rows *sql.Rows, dest any) error
	FetchUser(ctx context.Context, options gorm_model.UserFilter) (*sql.Rows, error)
	FetchOneUser(ctx context.Context, options gorm_model.UserFilter) (*gorm_model.User, error)
	CountUser(ctx context.Context, options gorm_model.UserFilter) int64
	CreateUser(ctx context.Context, model *gorm_model.User) (err error)
	UpdateUser(ctx context.Context, model *gorm_model.User) (err error)
	GetRoleNameByUserID(ctx context.Context, userID string) (string, error)

	CreateCV(ctx context.Context, cv *gorm_model.CV) error
	GetCVByUserID(ctx context.Context, userID string) (*gorm_model.CV, error)
	GetCVByID(ctx context.Context, id string) (*gorm_model.CV, error)
	UpdateCV(ctx context.Context, cv *gorm_model.CV) error

	GetProvinsiName(ctx context.Context, id string) (string, error)
	GetKabupatenName(ctx context.Context, id string) (string, error)

	FetchRoleApplied(ctx context.Context, options gorm_model.RoleAppliedFilter) (*sql.Rows, error)
	CreateRoleApplied(ctx context.Context, model *gorm_model.RoleApplied) error
	UpdateRoleApplied(ctx context.Context, model *gorm_model.RoleApplied) error
	DeleteRoleApplied(ctx context.Context, id string) error

	FetchKabupatenKota(ctx context.Context, options gorm_model.KabupatenKotaFilter) (*sql.Rows, error)
	FetchProvinsi(ctx context.Context, options gorm_model.ProvinsiFilter) (*sql.Rows, error)

	GetDB() *gorm.DB
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

type MessageBroker interface {
	Publish(ctx context.Context, queueName string, message interface{}) error
	Consume(ctx context.Context, queueName string, handler func(msg []byte) error) error
	Close() error
}

type AISearchRepository interface {
	Search(ctx context.Context, query string) ([]model.SearchResult, error)
}
