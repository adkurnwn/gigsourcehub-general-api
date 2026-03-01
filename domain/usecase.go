package domain

import (
	"context"

	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

type MemberAppUsecase interface {
	Login(ctx context.Context, payload request_model.LoginRequest) response.Base
	Register(ctx context.Context, payload request_model.RegisterRequest) response.Base
	GetMe(ctx context.Context, claim JWTClaimUser) response.Base
	FetchUsers(ctx context.Context, page, limit int64, cursor string, roleName *string) response.Base
	FetchUserDetail(ctx context.Context, id string) response.Base
}
