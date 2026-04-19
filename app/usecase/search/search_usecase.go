package usecase_search

import (
	"context"
	"net/http"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

type SearchUsecase interface {
	Search(ctx context.Context, query string) response.Base
}

type searchUsecase struct {
	aiSearchRepo domain.AISearchRepository
	gormRepo     domain.GormRepo
	timeout      time.Duration
}

func NewSearchUsecase(aiSearchRepo domain.AISearchRepository, gormRepo domain.GormRepo, timeout time.Duration) SearchUsecase {
	return &searchUsecase{
		aiSearchRepo: aiSearchRepo,
		gormRepo:     gormRepo,
		timeout:      timeout,
	}
}

func (u *searchUsecase) Search(ctx context.Context, query string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	// Guard: Check if AI module is enabled
	settings, err := u.gormRepo.GetSystemSetting(ctx)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "failed to fetch system settings")
	}

	if !settings.IsAIModeEnabled {
		return response.Error(http.StatusForbidden, "AI Search is currently disabled by administrator")
	}

	results, err := u.aiSearchRepo.Search(ctx, query)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "search failed: "+err.Error())
	}

	return response.Success(results)
}
