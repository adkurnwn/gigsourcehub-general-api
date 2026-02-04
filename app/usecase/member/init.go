package usecase_member

import (
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"time"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	contextTimeout time.Duration
}

type RepoInjection struct {
	GormDbRepo domain.GormRepo
}

func NewAppUsecase(r RepoInjection, timeout time.Duration) domain.MemberAppUsecase {
	return &appUsecase{
		gormDbRepo:     r.GormDbRepo,
		contextTimeout: timeout,
	}
}
