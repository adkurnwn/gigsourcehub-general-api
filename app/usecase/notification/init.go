package usecase_notification

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	contextTimeout time.Duration
}

func NewAppUsecase(gormDbRepo domain.GormRepo, timeout time.Duration) domain.NotificationAppUsecase {
	return &appUsecase{
		gormDbRepo:     gormDbRepo,
		contextTimeout: timeout,
	}
}
