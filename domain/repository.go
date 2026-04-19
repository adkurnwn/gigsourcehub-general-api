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

	FetchKabupatenKota(ctx context.Context, options gorm_model.KabupatenKotaFilter) (*sql.Rows, error)
	FetchProvinsi(ctx context.Context, options gorm_model.ProvinsiFilter) (*sql.Rows, error)
	FetchSystemRole(ctx context.Context, options gorm_model.SystemRoleFilter) (*sql.Rows, error)

	CreateBookmark(ctx context.Context, model *gorm_model.Bookmark) error
	GetBookmark(ctx context.Context, adminID, candidateID string) (*gorm_model.Bookmark, error)
	DeleteBookmark(ctx context.Context, adminID, candidateID string) (int64, error)
	FetchBookmarksByAdmin(ctx context.Context, adminID string, limit, offset int64) (*sql.Rows, error)
	CountBookmarksByAdmin(ctx context.Context, adminID string) (int64, error)

	CreateRequest(ctx context.Context, model *gorm_model.Request) error
	FetchRequestsByEmployee(ctx context.Context, employeeID string, limit, offset int64) (*sql.Rows, error)
	CountRequestsByEmployee(ctx context.Context, employeeID string) (int64, error)
	GetRequestByID(ctx context.Context, id string) (*gorm_model.Request, error)
	UpdateRequestByEmployee(ctx context.Context, model *gorm_model.Request) error
	GetSubrequestByID(ctx context.Context, id string) (*gorm_model.Subrequest, error)
	CreateSubrequestByEmployee(ctx context.Context, model *gorm_model.Subrequest) error
	UpdateSubrequestByEmployee(ctx context.Context, model *gorm_model.Subrequest) error

	GetDB() *gorm.DB

	// GetReviewScoresByUserIDs fetches aggregated (AVG) review scores for a list of candidate user IDs.
	// Returns a map of userID -> ReviewAggregateScore. Users without any review are omitted from the map.
	GetReviewScoresByUserIDs(ctx context.Context, userIDs []string) (map[string]ReviewAggregateScore, error)
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
}
