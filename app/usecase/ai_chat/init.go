package usecase_aichat

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	contextTimeout time.Duration
}

func NewAIChatUsecase(gormDbRepo domain.GormRepo, timeout time.Duration) domain.AIChatAppUsecase {
	return &appUsecase{
		gormDbRepo:     gormDbRepo,
		contextTimeout: timeout,
	}
}
