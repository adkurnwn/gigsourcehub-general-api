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
	DeclineRecruitment(ctx context.Context, id string) response.Base
	ConfirmDeclineRecruitment(ctx context.Context, id string) response.Base
	StopOnboarding(ctx context.Context, id string) response.Base
	PatchUserRecruitmentStatus(ctx context.Context, id string, req request_model.PatchUserRecruitmentStatusRequest) response.Base
	CancelRecruitment(ctx context.Context, id string) response.Base
	GetActiveSubrequest(ctx context.Context, id string) response.Base
	FinalizeRecruitment(ctx context.Context, adminID string, req request_model.FinalizeRecruitmentRequest) response.Base
	FetchOnboardingByCandidate(ctx context.Context, candidateID string, page, limit int64, cursor string) response.Base
	FetchOnboardingActiveTeam(ctx context.Context, employeeID string, page, limit int64, cursor string) response.Base
	FetchOnboardingHistory(ctx context.Context, employeeID string, page, limit int64, cursor string) response.Base
	FetchCandidateRecruitment(ctx context.Context, page, limit int64, cursor string) response.Base
	FetchCandidateBookmarked(ctx context.Context, adminID string, page, limit int64, cursor string) response.Base
	FetchOnboardingActive(ctx context.Context, page, limit int64, cursor string) response.Base
	FetchOnboardingArchive(ctx context.Context, page, limit int64, cursor string) response.Base

	VerifyAccount(ctx context.Context, token string) response.Base
	ResendVerification(ctx context.Context, req request_model.ResendVerificationRequest) response.Base
	ForgotPassword(ctx context.Context, req request_model.ForgotPasswordRequest) response.Base
	ResetPassword(ctx context.Context, req request_model.ResetPasswordRequest) response.Base
	UpdatePassword(ctx context.Context, userID string, req request_model.UpdatePasswordRequest) response.Base
	Logout(ctx context.Context, claim JWTClaimUser, refreshToken string) response.Base
	RefreshToken(ctx context.Context, payload request_model.RefreshTokenRequest) response.Base
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

type InterviewStageAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.InterviewStageFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, req request_model.CreateInterviewStageRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateInterviewStageRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
}

type InterviewAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.InterviewFilter) response.Base
	FetchScheduled(ctx context.Context, adminID string, page, limit int64, cursor string) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, adminID string, req request_model.CreateInterviewRequest) response.Base
	Update(ctx context.Context, adminID string, id string, req request_model.UpdateInterviewRequest) response.Base
	PatchStage(ctx context.Context, adminID string, req request_model.PatchInterviewStageRequest) response.Base
	PatchStatus(ctx context.Context, adminID string, req request_model.PatchInterviewStatusRequest) response.Base
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

type AdminNoteAppUsecase interface {
	Create(ctx context.Context, actorID string, candidateID string, req request_model.CreateAdminNoteRequest) response.Base
	FetchByCandidate(ctx context.Context, candidateID string) response.Base
}

type ReviewAppUsecase interface {
	CreateOrUpdate(ctx context.Context, employeeID string, req request_model.CreateReviewRequest) response.Base
	FetchByID(ctx context.Context, reviewID string) response.Base
	FetchQuestions(ctx context.Context) response.Base
}

type RequestAppUsecase interface {
	FetchAll(ctx context.Context, page, limit int64) response.Base
	CreateByEmployee(ctx context.Context, employeeID string, req request_model.CreateRequestRequest) response.Base
	FetchByEmployee(ctx context.Context, employeeID string, page, limit int64) response.Base
	FetchByAdmin(ctx context.Context, page, limit int64, filter gorm_model.RequestFilter) response.Base
	FetchPendingForAdmin(ctx context.Context, page, limit int64) response.Base
	FetchMyRequestsForAdmin(ctx context.Context, adminID string, page, limit int64) response.Base
	GetByID(ctx context.Context, employeeID, requestID string) response.Base
	UpdateByEmployee(ctx context.Context, employeeID string, requestID string, req request_model.UpdateRequestRequest) response.Base
	UpdateSubrequestByEmployee(ctx context.Context, employeeID string, requestID string, subrequestID string, req request_model.UpdateSubrequestRequest) response.Base
	AddSubrequestByEmployee(ctx context.Context, employeeID string, requestID string, req request_model.CreateSubrequestRequest) response.Base
	AssignCandidateToSubrequest(ctx context.Context, adminID string, requestID string, subrequestID string, req request_model.AssignCandidateToSubrequestRequest) response.Base
	AssignPIC(ctx context.Context, adminID, requestID string) response.Base
	RejectRequest(ctx context.Context, adminID, requestID string, rejectedReason string) response.Base
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

type JobVacancyAppUsecase interface {
	// CMS — Admin & Superadmin
	FetchAll(ctx context.Context, page, limit int64, filter gorm_model.JobVacancyFilter) response.Base
	FetchData(ctx context.Context, id string) response.Base
	Create(ctx context.Context, req request_model.CreateJobVacancyRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateJobVacancyRequest) response.Base
	Delete(ctx context.Context, id string) response.Base

	// Public — tanpa auth, hanya PUBLISHED & belum takedown
	FetchPublic(ctx context.Context, page, limit int64, filter gorm_model.JobVacancyFilter) response.Base
	FetchPublicByID(ctx context.Context, id string) response.Base
}

type ChatAppUsecase interface {
	CreateConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest) response.Base
	StartConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest) response.Base
	UploadOffering(ctx context.Context, adminID string, conversationID string, file *multipart.FileHeader) response.Base
	FetchMyConversations(ctx context.Context, userID string, page, limit int64) response.Base
	GetConversation(ctx context.Context, userID string, conversationID string) response.Base
	SendMessage(ctx context.Context, userID string, conversationID string, req request_model.SendMessageRequest) response.Base
	FetchMessages(ctx context.Context, userID string, conversationID string, page, limit int64) response.Base
	MarkAsRead(ctx context.Context, userID string, conversationID string) response.Base
}

type FAQAppUsecase interface {
	// Admin & Superadmin — CMS
	FetchAll(ctx context.Context, page, limit int64, search *string) response.Base
	FetchData(ctx context.Context, id string) response.Base

	// Admin only
	Create(ctx context.Context, adminID string, req request_model.CreateFAQRequest) response.Base
	Update(ctx context.Context, adminID string, id string, req request_model.UpdateFAQRequest) response.Base
	Delete(ctx context.Context, id string) response.Base

	// Superadmin — Approvals
	FetchApprovals(ctx context.Context, page, limit int64) response.Base
	ApproveRequest(ctx context.Context, superadminID string, approvalID string) response.Base
	RejectRequest(ctx context.Context, superadminID string, approvalID string, req request_model.ReviewApprovalRequest) response.Base
	TakedownRequest(ctx context.Context, superadminID string, approvalID string) response.Base

	// Public — no auth
	FetchPublic(ctx context.Context, page, limit int64, search *string) response.Base
	FetchPublicByID(ctx context.Context, id string) response.Base
}

type CompanyProfileAppUsecase interface {
	// Admin & Superadmin — CMS
	Get(ctx context.Context) response.Base

	// Admin only
	AdminUpdate(ctx context.Context, adminID string, req request_model.UpdateCompanyProfileRequest) response.Base
	CancelPendingApproval(ctx context.Context, approvalID string) response.Base

	// Superadmin — Approvals
	FetchPendingApprovals(ctx context.Context, page, limit int64) response.Base
	ApproveRequest(ctx context.Context, superadminID string, approvalID string) response.Base
	RejectRequest(ctx context.Context, superadminID string, approvalID string, req request_model.ReviewApprovalRequest) response.Base

	// Public — no auth
	GetPublic(ctx context.Context) response.Base
}

type CareerDepartmentAppUsecase interface {
	// Admin & Superadmin — CMS
	FetchAll(ctx context.Context, page, limit int64, search *string) response.Base
	FetchData(ctx context.Context, id string) response.Base

	// Admin only
	Create(ctx context.Context, adminID string, req request_model.CreateCareerDepartmentRequest) response.Base
	Update(ctx context.Context, adminID string, id string, req request_model.UpdateCareerDepartmentRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
	UploadImage(ctx context.Context, id string, file *multipart.FileHeader) response.Base

	// Superadmin — Approvals
	FetchApprovals(ctx context.Context, page, limit int64) response.Base
	ApproveRequest(ctx context.Context, superadminID string, approvalID string) response.Base
	RejectRequest(ctx context.Context, superadminID string, approvalID string, req request_model.ReviewApprovalRequest) response.Base
	TakedownRequest(ctx context.Context, superadminID string, approvalID string) response.Base

	// Public — no auth
	FetchPublic(ctx context.Context, page, limit int64, search *string) response.Base
	FetchPublicByID(ctx context.Context, id string) response.Base
}
type DashboardAppUsecase interface {
	GetAdminDashboardSummary(ctx context.Context, adminID string) response.Base
	GetAdminDashboardAnalytics(ctx context.Context, period string) response.Base
	GetSuperadminDashboardSummary(ctx context.Context) response.Base
}
