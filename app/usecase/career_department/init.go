package usecase_career_department

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	storageRepo    domain.StorageRepo
	contextTimeout time.Duration
}

type RepoInjection struct {
	GormDbRepo  domain.GormRepo
	StorageRepo domain.StorageRepo
}

func NewAppUsecase(r RepoInjection, timeout time.Duration) domain.CareerDepartmentAppUsecase {
	return &appUsecase{
		gormDbRepo:     r.GormDbRepo,
		storageRepo:    r.StorageRepo,
		contextTimeout: timeout,
	}
}
