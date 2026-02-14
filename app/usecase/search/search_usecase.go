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
	timeout      time.Duration
}

func NewSearchUsecase(aiSearchRepo domain.AISearchRepository, timeout time.Duration) SearchUsecase {
	return &searchUsecase{
		aiSearchRepo: aiSearchRepo,
		timeout:      timeout,
	}
}

func (u *searchUsecase) Search(ctx context.Context, query string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	results, err := u.aiSearchRepo.Search(ctx, query)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "search failed: "+err.Error())
	}

	return response.Success(results)
}
