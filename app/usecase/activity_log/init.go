package usecase_activity_log

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

func NewAppUsecase(repo RepoInjection, timeout time.Duration) domain.ActivityLogAppUsecase {
	return &appUsecase{
		gormDbRepo:     repo.GormDbRepo,
		contextTimeout: timeout,
	}
}
