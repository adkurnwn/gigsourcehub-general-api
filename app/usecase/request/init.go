package usecase_request

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	contextTimeout time.Duration
}

func NewRequestAppUsecase(gormDbRepo domain.GormRepo, timeout time.Duration) domain.RequestAppUsecase {
	return &appUsecase{
		gormDbRepo:     gormDbRepo,
		contextTimeout: timeout,
	}
}
