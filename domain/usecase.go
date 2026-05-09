package domain

import (
	"context"
	"mime/multipart"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

type MemberAppUsecase interface {
	Login(ctx context.Context, payload request_model.LoginRequest) response.Base
	Register(ctx context.Context, payload request_model.RegisterRequest) response.Base
	GetMe(ctx context.Context, claim JWTClaimUser) response.Base
	GetProfile(ctx context.Context, claim JWTClaimUser) response.Base
	FetchUsers(ctx context.Context, page, limit int64, cursor string, search *string, roleName *string, adminID *string) response.Base
	FetchUserDetail(ctx context.Context, id string) response.Base
	CreateBySuperadmin(ctx context.Context, req request_model.CreateUserBySuperadminRequest) response.Base
	EditUserBySuperadmin(ctx context.Context, id string, req request_model.EditUserBySuperadminRequest) response.Base
	BlockUserBySuperadmin(ctx context.Context, id string) response.Base
	DisableUserBySuperadmin(ctx context.Context, id string) response.Base
	ActivateUserBySuperadmin(ctx context.Context, id string) response.Base
	UploadProfilePicture(ctx context.Context, userID string, file *multipart.FileHeader) response.Base
	FetchUserThumb(ctx context.Context, id string) response.Base
	UpdateProfile(ctx context.Context, userID string, req request_model.UpdateProfileRequest) response.Base

	VerifyAccount(ctx context.Context, token string) response.Base
	ForgotPassword(ctx context.Context, req request_model.ForgotPasswordRequest) response.Base
	ResetPassword(ctx context.Context, req request_model.ResetPasswordRequest) response.Base
	UpdatePassword(ctx context.Context, userID string, req request_model.UpdatePasswordRequest) response.Base
}

type JobRoleAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.JobRoleFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	FetchSystemRoles(ctx context.Context) response.Base
	Create(ctx context.Context, req request_model.CreateJobRoleRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateJobRoleRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
}

type JobTitleAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.JobTitleFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, req request_model.CreateJobTitleRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateJobTitleRequest) response.Base
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

type BookmarkAppUsecase interface {
	Create(ctx context.Context, adminID string, req request_model.CreateBookmarkRequest) response.Base
	Delete(ctx context.Context, adminID string, candidateID string) response.Base
	FetchByAdmin(ctx context.Context, adminID string, page, limit int64) response.Base
}

type RequestAppUsecase interface {
	CreateByEmployee(ctx context.Context, employeeID string, req request_model.CreateRequestRequest) response.Base
	FetchByEmployee(ctx context.Context, employeeID string, page, limit int64) response.Base
	GetByID(ctx context.Context, employeeID, requestID string) response.Base
	UpdateByEmployee(ctx context.Context, employeeID string, requestID string, req request_model.UpdateRequestRequest) response.Base
	UpdateSubrequestByEmployee(ctx context.Context, employeeID string, requestID string, subrequestID string, req request_model.UpdateSubrequestRequest) response.Base
	AddSubrequestByEmployee(ctx context.Context, employeeID string, requestID string, req request_model.CreateSubrequestRequest) response.Base
}

type AIChatAppUsecase interface {
	FetchMyChats(ctx context.Context, adminID string) response.Base
	CreateChat(ctx context.Context, adminID string, firstQuery string) response.Base
	DeleteChat(ctx context.Context, adminID string, chatID string) response.Base
	FetchChatMessages(ctx context.Context, adminID string, chatID string) response.Base
	StoreChatMessage(ctx context.Context, adminID string, chatID string, role string, content string, isLast bool) response.Base
}

type ActivityLogAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.LogActivityFilter) response.Base
	ExportData(ctx context.Context, filter gorm_model.LogActivityFilter, format string, actorID string) ([]byte, string, string, error)
}

type ChatAppUsecase interface {
	CreateConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest) response.Base
	FetchMyConversations(ctx context.Context, userID string, page, limit int64) response.Base
	GetConversation(ctx context.Context, userID string, conversationID string) response.Base
	SendMessage(ctx context.Context, userID string, conversationID string, req request_model.SendMessageRequest) response.Base
	FetchMessages(ctx context.Context, userID string, conversationID string, page, limit int64) response.Base
	MarkAsRead(ctx context.Context, userID string, conversationID string) response.Base
}
