package domain

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

type MemberAppUsecase interface {
	Login(ctx context.Context, payload request_model.LoginRequest) response.Base
	Register(ctx context.Context, payload request_model.RegisterRequest) response.Base
	GetMe(ctx context.Context, claim JWTClaimUser) response.Base
	GetProfile(ctx context.Context, claim JWTClaimUser) response.Base
	FetchUsers(ctx context.Context, page, limit int64, cursor string, roleName *string) response.Base
	FetchUserDetail(ctx context.Context, id string) response.Base
	CreateBySuperadmin(ctx context.Context, req request_model.CreateUserBySuperadminRequest) response.Base
	EditUserBySuperadmin(ctx context.Context, id string, req request_model.EditUserBySuperadminRequest) response.Base
	BlockUserBySuperadmin(ctx context.Context, id string) response.Base
}

type RoleAppliedAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.RoleAppliedFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, req request_model.CreateRoleAppliedRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateRoleAppliedRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
}

type SectorAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.SectorFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, req request_model.CreateSectorRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateSectorRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
}

type RecruitmentStatusAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.RecruitmentStatusFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, req request_model.CreateRecruitmentStatusRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateRecruitmentStatusRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
}

type KabupatenKotaAppUsecase interface {
	FetchAll(ctx context.Context, filter gorm_model.KabupatenKotaFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
}

type ProvinsiAppUsecase interface {
	FetchAll(ctx context.Context, filter gorm_model.ProvinsiFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
}
