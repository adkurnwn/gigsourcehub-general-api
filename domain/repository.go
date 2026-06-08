package domain

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	storage_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/storage"
	pb "github.com/adkurnwn/gigsourcehub-general-api/proto"
	"gorm.io/gorm"
)

// ReviewAggregateScore holds averaged Likert-scale scores (1–5) for a candidate.
// All fields are pointers so we can detect SQL NULL (no reviews exist).
type ReviewAggregateScore struct {
	UserID                      string
	AvgWorkQuality              *float64
	AvgTimeliness               *float64
	AvgCommunicationCollab      *float64
	AvgProblemSolvingInitiative *float64
}

type GormRepo interface {
	StructScan(rows *sql.Rows, dest any) error
	FetchUser(ctx context.Context, options gorm_model.UserFilter) (*sql.Rows, error)
	FetchOneUser(ctx context.Context, options gorm_model.UserFilter) (*gorm_model.User, error)
	CountUser(ctx context.Context, options gorm_model.UserFilter) int64
	CreateUser(ctx context.Context, model *gorm_model.User) (err error)
	CreateUserBySuperadmin(ctx context.Context, model *gorm_model.User) (err error)
	UpdateUser(ctx context.Context, model *gorm_model.User) (err error)
	SoftDeleteUser(ctx context.Context, userID string) error
	GetRoleNameByUserID(ctx context.Context, userID string) (string, error)
	GetUserAccountStatus(ctx context.Context, userID string) (string, error)

	CreateCV(ctx context.Context, cv *gorm_model.CV) error
	GetCVByUserID(ctx context.Context, userID string) (*gorm_model.CV, error)
	GetCVByID(ctx context.Context, id string) (*gorm_model.CV, error)
	UpdateCV(ctx context.Context, cv *gorm_model.CV) error

	GetProvinsiName(ctx context.Context, id string) (string, error)
	GetKabupatenName(ctx context.Context, id string) (string, error)

	FetchJobRole(ctx context.Context, options gorm_model.JobRoleFilter) (*sql.Rows, error)
	CreateJobRole(ctx context.Context, model *gorm_model.JobRole) error
	UpdateJobRole(ctx context.Context, model *gorm_model.JobRole) error
	DeleteJobRole(ctx context.Context, id string) error
	GetActiveJobRoles(ctx context.Context) ([]gorm_model.JobRole, error)

	FetchJobTitle(ctx context.Context, options gorm_model.JobTitleFilter) (*sql.Rows, error)
	CreateJobTitle(ctx context.Context, model *gorm_model.JobTitle) error
	UpdateJobTitle(ctx context.Context, model *gorm_model.JobTitle) error
	DeleteJobTitle(ctx context.Context, id string) error

	FetchSector(ctx context.Context, options gorm_model.SectorFilter) (*sql.Rows, error)
	GetActiveSectors(ctx context.Context) ([]gorm_model.Sector, error)
	CreateSector(ctx context.Context, model *gorm_model.Sector) error
	UpdateSector(ctx context.Context, model *gorm_model.Sector) error
	DeleteSector(ctx context.Context, id string) error

	FetchRecruitmentStatus(ctx context.Context, options gorm_model.RecruitmentStatusFilter) (*sql.Rows, error)
	CreateRecruitmentStatus(ctx context.Context, model *gorm_model.RecruitmentStatus) error
	UpdateRecruitmentStatus(ctx context.Context, model *gorm_model.RecruitmentStatus) error
	DeleteRecruitmentStatus(ctx context.Context, id string) error

	FetchInterviewStage(ctx context.Context, options gorm_model.InterviewStageFilter) (*sql.Rows, error)
	CreateInterviewStage(ctx context.Context, model *gorm_model.InterviewStage) error
	UpdateInterviewStage(ctx context.Context, model *gorm_model.InterviewStage) error
	DeleteInterviewStage(ctx context.Context, id string) error

	GetInterviewByID(ctx context.Context, id string) (*gorm_model.Interview, error)
	CreateInterview(ctx context.Context, model *gorm_model.Interview) error
	UpdateInterview(ctx context.Context, model *gorm_model.Interview) error
	PatchInterviewStage(ctx context.Context, interviewID string, stageID string, status *string) error
	PatchInterviewStatus(ctx context.Context, interviewID string, status string) error

	FetchKabupatenKota(ctx context.Context, options gorm_model.KabupatenKotaFilter) (*sql.Rows, error)
	FetchProvinsi(ctx context.Context, options gorm_model.ProvinsiFilter) (*sql.Rows, error)
	FetchSystemRole(ctx context.Context, options gorm_model.SystemRoleFilter) (*sql.Rows, error)

	CreateBookmark(ctx context.Context, model *gorm_model.Bookmark) error
	GetBookmark(ctx context.Context, adminID, candidateID string) (*gorm_model.Bookmark, error)
	DeleteBookmark(ctx context.Context, adminID, candidateID string) (int64, error)
	FetchBookmarksByAdmin(ctx context.Context, adminID string, limit, offset int64) (*sql.Rows, error)
	CountBookmarksByAdmin(ctx context.Context, adminID string) (int64, error)

	CreateAdminNote(ctx context.Context, model *gorm_model.AdminNote) error
	FetchAdminNotesByCandidate(ctx context.Context, candidateID string) ([]gorm_model.AdminNote, error)

	CreateRequest(ctx context.Context, model *gorm_model.Request) error
	FetchRequestsByEmployee(ctx context.Context, employeeID string, limit, offset int64) (*sql.Rows, error)
	CountRequestsByEmployee(ctx context.Context, employeeID string) (int64, error)
	FetchRequestsByAdmin(ctx context.Context, filter gorm_model.RequestFilter, limit, offset int64) (*sql.Rows, error)
	CountRequestsByAdmin(ctx context.Context, filter gorm_model.RequestFilter) (int64, error)
	GetRequestByID(ctx context.Context, id string) (*gorm_model.Request, error)
	UpdateRequestByEmployee(ctx context.Context, model *gorm_model.Request) error
	UpdateRequestByAdmin(ctx context.Context, model *gorm_model.Request) error
	UpdateRequestAdminUser(ctx context.Context, requestID string, adminUserID string) error
	GetSubrequestByID(ctx context.Context, id string) (*gorm_model.Subrequest, error)
	CreateSubrequestByEmployee(ctx context.Context, model *gorm_model.Subrequest) error
	UpdateSubrequestByEmployee(ctx context.Context, model *gorm_model.Subrequest) error
	AssignCandidateToSubrequest(ctx context.Context, model *gorm_model.SubrequestCandidate, recruitmentStatusID string, requestID string, markRequestProcessing bool) error
	CountActiveSubrequestCandidatesByCandidateID(ctx context.Context, candidateID string) (int64, error)
	SoftDeleteSubrequestCandidatesByCandidateID(ctx context.Context, candidateID string) error
	StopOnboardingByCandidateID(ctx context.Context, candidateID string) error
	CancelRecruitmentByCandidateID(ctx context.Context, candidateID string) error
	GetActiveSubrequestByCandidateID(ctx context.Context, candidateID string) (*gorm_model.ActiveSubrequestInfo, error)
	GetFinalizeSnapshotData(ctx context.Context, subrequestID string) (*gorm_model.FinalizeRecruitmentSnapshot, error)
	FinalizeRecruitment(ctx context.Context, candidateID, subrequestID, requestID, acceptedStatusID string, startDate, endDate *time.Time, offeringID *string, snapshotJSON string) error
	CreateOffering(ctx context.Context, model *gorm_model.Offering) error
	FetchOnboardHistoriesByCandidate(ctx context.Context, candidateID string, limit, offset int64) ([]gorm_model.OnboardHistory, error)
	CountOnboardHistoriesByCandidate(ctx context.Context, candidateID string) (int64, error)
	FetchOnboardHistoriesByEmployee(ctx context.Context, employeeID string, limit, offset int64) ([]gorm_model.OnboardHistory, error)
	CountOnboardHistoriesByEmployee(ctx context.Context, employeeID string) (int64, error)
	FetchOnboardHistoriesByEmployeeHistory(ctx context.Context, employeeID string, beforeDate time.Time, limit, offset int64) ([]gorm_model.OnboardHistory, error)
	CountOnboardHistoriesByEmployeeHistory(ctx context.Context, employeeID string, beforeDate time.Time) (int64, error)
	GetOnboardHistoryByID(ctx context.Context, id string) (*gorm_model.OnboardHistory, error)

	GetDB() *gorm.DB

	// GetReviewScoresByUserIDs fetches aggregated (AVG) review scores for a list of candidate user IDs.
	// Returns a map of userID -> ReviewAggregateScore. Users without any review are omitted from the map.
	GetReviewScoresByUserIDs(ctx context.Context, userIDs []string) (map[string]ReviewAggregateScore, error)
	GetReviewByID(ctx context.Context, id string) (*gorm_model.Review, error)
	GetReviewByOnboardEmployee(ctx context.Context, onboardHistoryID, employeeUserID string) (*gorm_model.Review, error)
	CreateReview(ctx context.Context, model *gorm_model.Review) error
	UpdateReview(ctx context.Context, model *gorm_model.Review) error
	FetchReviewQuestions(ctx context.Context) ([]gorm_model.ReviewQuestion, error)
	ReplaceReviewAnswers(ctx context.Context, reviewID string, answers []gorm_model.ReviewAnswer) error
	// GetJobRolesByUserIDs fetches names of job roles for a list of user IDs.
	// Returns a map of userID -> list of job role names.
	GetJobRolesByUserIDs(ctx context.Context, userIDs []string) (map[string][]string, error)
	// GetCandidateLevelsByUserIDs fetches candidate levels for a list of user IDs.
	// Returns a map of userID -> candidate level.
	GetCandidateLevelsByUserIDs(ctx context.Context, userIDs []string) (map[string]string, error)
	// AI Chat & Messages persistence
	CreateAIChat(ctx context.Context, chat *gorm_model.AIChat) error
	GetAIChatByID(ctx context.Context, id string) (*gorm_model.AIChat, error)
	FetchAIChatsByAdmin(ctx context.Context, adminID string) ([]gorm_model.AIChat, error)
	DeleteAIChat(ctx context.Context, id string) error
	CreateLogActivity(ctx context.Context, model *gorm_model.LogActivity) error
	FetchLogActivity(ctx context.Context, options gorm_model.LogActivityFilter) ([]gorm_model.LogActivity, error)
	FetchCountLogActivity(ctx context.Context, options gorm_model.LogActivityFilter) (int64, error)

	CreateAIMessage(ctx context.Context, msg *gorm_model.AIMessage) error
	FetchAIMessagesByChat(ctx context.Context, chatID string) ([]gorm_model.AIMessage, error)

	GetSystemSetting(ctx context.Context) (*gorm_model.SystemSetting, error)
	UpdateSystemSetting(ctx context.Context, model *gorm_model.SystemSetting) error

	CreateUserToken(ctx context.Context, model *gorm_model.UserToken) error
	GetUserToken(ctx context.Context, token string, tokenType string) (*gorm_model.UserToken, error)
	DeleteUserToken(ctx context.Context, id string) error
	DeleteUserTokensByUserID(ctx context.Context, userID string, tokenType string) error

	GetUserVerifiedAt(ctx context.Context, userID string) (*time.Time, error)
	GetUserMustResetPassword(ctx context.Context, userID string) (bool, error)

	FetchJobVacancy(ctx context.Context, options gorm_model.JobVacancyFilter) (*sql.Rows, error)
	GetJobVacancyByID(ctx context.Context, id string) (*gorm_model.JobVacancy, error)
	CreateJobVacancy(ctx context.Context, model *gorm_model.JobVacancy) error
	UpdateJobVacancy(ctx context.Context, model *gorm_model.JobVacancy) error
	DeleteJobVacancy(ctx context.Context, id string) error

	// Chat / Conversation
	CreateConversation(ctx context.Context, conv *gorm_model.Conversation) error
	StartConversation(ctx context.Context, conv *gorm_model.Conversation, contactedStatusID string) error
	GetConversationByID(ctx context.Context, id string) (*gorm_model.Conversation, error)
	GetConversationBySubrequestAndCandidate(ctx context.Context, subrequestID, candidateID string) (*gorm_model.Conversation, error)
	GetActiveConversationByCandidateID(ctx context.Context, candidateID string) (*gorm_model.Conversation, error)
	FetchConversationsByUser(ctx context.Context, userID string, limit, offset int64) ([]gorm_model.Conversation, error)
	CountConversationsByUser(ctx context.Context, userID string) (int64, error)
	DeleteConversationsByCandidateID(ctx context.Context, candidateID string) error

	// Messages
	CreateMessage(ctx context.Context, msg *gorm_model.Message) error
	GetMessageByID(ctx context.Context, id string) (*gorm_model.Message, error)
	FetchMessagesByConversation(ctx context.Context, conversationID string, limit, offset int64) ([]gorm_model.Message, error)
	CountMessagesByConversation(ctx context.Context, conversationID string) (int64, error)
	MarkMessagesAsRead(ctx context.Context, conversationID, readerUserID string) error
	CountUnreadMessagesByUser(ctx context.Context, userID string) (int64, error)
	CountUnreadMessagesByConversation(ctx context.Context, conversationID, userID string) (int64, error)

	// Chat Authorization helpers
	IsAdminOfSubrequest(ctx context.Context, adminID, subrequestID string) (bool, error)
	IsCandidateOnSubrequest(ctx context.Context, candidateID, subrequestID string) (bool, error)
	GetCandidateRecruitmentStatusName(ctx context.Context, candidateID string) (string, error)

	// FAQ
	FetchFAQ(ctx context.Context, options gorm_model.FAQFilter) ([]gorm_model.FAQ, error)
	CountFAQ(ctx context.Context, options gorm_model.FAQFilter) (int64, error)
	GetFAQByID(ctx context.Context, id string) (*gorm_model.FAQ, error)
	CreateFAQ(ctx context.Context, model *gorm_model.FAQ) error
	UpdateFAQ(ctx context.Context, model *gorm_model.FAQ) error
	DeleteFAQ(ctx context.Context, id string) error

	// Career Department
	FetchCareerDepartment(ctx context.Context, options gorm_model.CareerDepartmentFilter) ([]gorm_model.CareerDepartment, error)
	CountCareerDepartment(ctx context.Context, options gorm_model.CareerDepartmentFilter) (int64, error)
	GetCareerDepartmentByID(ctx context.Context, id string) (*gorm_model.CareerDepartment, error)
	CreateCareerDepartment(ctx context.Context, model *gorm_model.CareerDepartment) error
	UpdateCareerDepartment(ctx context.Context, model *gorm_model.CareerDepartment) error
	DeleteCareerDepartment(ctx context.Context, id string) error

	// Company Profile
	GetCompanyProfile(ctx context.Context) (*gorm_model.CompanyProfile, error)
	UpdateCompanyProfile(ctx context.Context, model *gorm_model.CompanyProfile) error

	// Approval Request
	CreateApprovalRequest(ctx context.Context, model *gorm_model.ApprovalRequest) error
	GetApprovalRequestByID(ctx context.Context, id string) (*gorm_model.ApprovalRequest, error)
	FetchApprovalRequests(ctx context.Context, options gorm_model.ApprovalRequestFilter) ([]gorm_model.ApprovalRequest, error)
	CountApprovalRequests(ctx context.Context, options gorm_model.ApprovalRequestFilter) (int64, error)
	UpdateApprovalRequest(ctx context.Context, model *gorm_model.ApprovalRequest) error
	GetPendingApprovalByRecord(ctx context.Context, tableName, recordID string) (*gorm_model.ApprovalRequest, error)
	DeleteApprovalRequest(ctx context.Context, id string) error

	// Dashboard & Analytics
	CreateCandidateStatusHistory(ctx context.Context, model *gorm_model.CandidateStatusHistory) error
	FetchUpcomingInterviewsForAdmin(ctx context.Context, adminID string, limit int) ([]gorm_model.Interview, error)
	FetchRequestsSummaryStats(ctx context.Context) (gorm_model.RequestsSummaryResp, error)
	FetchDashboardAlertsForAdmin(ctx context.Context, adminID string) (gorm_model.DashboardAlertResp, error)
	FetchRecentActivitiesForAdmin(ctx context.Context, limit int) ([]gorm_model.LogActivity, error)
	FetchDashboardAnalytics(ctx context.Context, period string) (gorm_model.DashboardAnalyticsResp, error)
	FetchSuperadminDashboardStats(ctx context.Context) (gorm_model.SuperadminDashboardResp, error)

	// Notifications
	CreateNotification(ctx context.Context, model *gorm_model.Notification) error
	FetchNotificationsByUser(ctx context.Context, userID string, limit, offset int64) ([]gorm_model.Notification, error)
	CountNotificationsByUser(ctx context.Context, userID string) (int64, error)
	CountUnreadNotificationsByUser(ctx context.Context, userID string) (int64, error)
	MarkNotificationAsRead(ctx context.Context, notificationID, userID string) error
	MarkAllNotificationsAsRead(ctx context.Context, userID string) error
	GetSuperadminUserIDs(ctx context.Context) ([]string, error)
	GetUpcomingInterviewsForNotification(ctx context.Context, from, to time.Time) ([]gorm_model.Interview, error)
	GetUpcomingInterviewsFor24hReminder(ctx context.Context, from, to time.Time) ([]gorm_model.Interview, error)
	GetUpcomingInterviewsFor1hReminder(ctx context.Context, from, to time.Time) ([]gorm_model.Interview, error)
	MarkInterview24hReminderSent(ctx context.Context, interviewID string) error
	MarkInterview1hReminderSent(ctx context.Context, interviewID string) error
	GetExpiringContractsForNotification(ctx context.Context, targetDate time.Time) ([]gorm_model.OnboardHistory, error)
	EndExpiredContract(ctx context.Context, onboardHistoryID string, candidateID string) error
}

type CacheRepo interface {
	Enabled() bool
	GetTTL() time.Duration
	Get(ctx context.Context, key string) (value []byte, err error)
	Set(ctx context.Context, key string, value []byte, expiration *time.Duration) (err error)
}

type StorageRepo interface {
	GetPresignedLink(objectKey string, expires *time.Duration) string
	GetPublicLink(objectKey string) string
	UploadFilePublic(objectKey string, body io.Reader, contentType string) (uploadData *storage_model.UploadResponse, err error)
	UploadFilePrivate(objectKey string, body io.Reader, contentType string, expires *time.Duration) (uploadData *storage_model.UploadResponse, err error)
	DeleteFile(objectKey string) error
}

type MessageBroker interface {
	Publish(ctx context.Context, queueName string, message interface{}) error
	Consume(ctx context.Context, queueName string, handler func(msg []byte) error) error
	Close() error
}

type AISearchRepository interface {
	Search(ctx context.Context, query string) ([]model.SearchResult, error)
	UpdateCandidate(ctx context.Context, req *pb.UpdateCandidateRequest) error
	DeleteCandidate(ctx context.Context, userID string) error
}
