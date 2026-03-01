package usecase_member

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

func NewAppUsecase(r RepoInjection, timeout time.Duration) domain.MemberAppUsecase {
	return &appUsecase{
		gormDbRepo:     r.GormDbRepo,
		storageRepo:    r.StorageRepo,
		contextTimeout: timeout,
	}
}
