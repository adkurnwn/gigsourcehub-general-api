package usecase_chat

import (
	"time"

	http_chat "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/chat"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
)

type appUsecase struct {
	gormDbRepo     domain.GormRepo
	hub            *http_chat.Hub
	contextTimeout time.Duration
}

type RepoInjection struct {
	GormDbRepo domain.GormRepo
}

func NewAppUsecase(r RepoInjection, hub *http_chat.Hub, timeout time.Duration) domain.ChatAppUsecase {
	return &appUsecase{
		gormDbRepo:     r.GormDbRepo,
		hub:            hub,
		contextTimeout: timeout,
	}
}
