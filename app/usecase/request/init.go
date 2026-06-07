package usecase_request

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	mailerRepo     domain.Mailer
	contextTimeout time.Duration
}

func NewRequestAppUsecase(gormDbRepo domain.GormRepo, mailerRepo domain.Mailer, timeout time.Duration) domain.RequestAppUsecase {
	return &appUsecase{
		gormDbRepo:     gormDbRepo,
		mailerRepo:     mailerRepo,
		contextTimeout: timeout,
	}
}
