package system_setting

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
)

type Usecase interface {
	Fetch(ctx context.Context) (*gorm_model.SystemSetting, error)
	Update(ctx context.Context, req request_model.UpdateSystemSettingRequest) error
	UploadCVTemplate(ctx context.Context, fileHeader *multipart.FileHeader) (string, error)
}

type appUsecase struct {
	repo           domain.GormRepo
	storageRepo    domain.StorageRepo
	contextTimeout time.Duration
}

type RepoInjection struct {
	GormDbRepo  domain.GormRepo
	StorageRepo domain.StorageRepo
}

func NewAppUsecase(repo RepoInjection, timeout time.Duration) Usecase {
	return &appUsecase{
		repo:           repo.GormDbRepo,
		storageRepo:    repo.StorageRepo,
		contextTimeout: timeout,
	}
}
