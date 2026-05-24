package usecase_faq

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	contextTimeout time.Duration
}

type RepoInjection struct {
	GormDbRepo domain.GormRepo
}

func NewAppUsecase(r RepoInjection, timeout time.Duration) domain.FAQAppUsecase {
	return &appUsecase{
		gormDbRepo:     r.GormDbRepo,
		contextTimeout: timeout,
	}
}
