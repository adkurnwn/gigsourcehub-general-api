package usecase_chat

import (
	"time"

	http_chat "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/chat"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	storageRepo    domain.StorageRepo
	hub            *http_chat.Hub
	contextTimeout time.Duration
}

type RepoInjection struct {
	GormDbRepo domain.GormRepo
	StorageRepo domain.StorageRepo
}

func NewAppUsecase(r RepoInjection, hub *http_chat.Hub, timeout time.Duration) domain.ChatAppUsecase {
	return &appUsecase{
		gormDbRepo:     r.GormDbRepo,
		storageRepo:    r.StorageRepo,
		hub:            hub,
		contextTimeout: timeout,
	}
}
