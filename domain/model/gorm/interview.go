package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Interview struct {
	ID              string          `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID     *string         `gorm:"column:admin_user_id;type:uuid"`
	AdminUser       *User           `gorm:"foreignKey:AdminUserID"`
	CandidateUserID string          `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User           `gorm:"foreignKey:CandidateUserID"`
	SubrequestID    string          `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest      *Subrequest     `gorm:"foreignKey:SubrequestID"`
	StageID         string          `gorm:"column:stage_id;type:uuid;not null"`
	Stage           *InterviewStage `gorm:"foreignKey:StageID"`
	Title           string          `gorm:"column:title;type:varchar(255)"`
	Description     *string         `gorm:"column:description;type:text"`
	ScheduledAt     *time.Time      `gorm:"column:scheduled_at;type:timestamp"`
	Method          *string         `gorm:"column:method;type:varchar(50)"`
	Status          string          `gorm:"column:status;type:interview_status"`
	MeetingLink     *string         `gorm:"column:meeting_link;type:varchar(255)"`
	MeetingLocation *string         `gorm:"column:meeting_location;type:text"`
	IsEmailSent       bool            `gorm:"column:is_email_sent;type:boolean;default:false"`
	Is24hReminderSent bool            `gorm:"column:is_24h_reminder_sent;type:boolean;default:false"`
	Is1hReminderSent  bool            `gorm:"column:is_1h_reminder_sent;type:boolean;default:false"`
	CreatedAt         time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt         gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}


var InterviewAllowedSort = []string{"scheduled_at", "created_at", "updated_at"}

type InterviewFilter struct {
	DefaultFilter
	AdminUserID     *string
	CandidateUserID *string
	SubrequestID    *string
	StageID         *string
	Status          *string
	ScheduledOnly   bool
}

func (f *InterviewFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.AdminUserID != nil {
		q.Where("admin_user_id = ?", *f.AdminUserID)
	}
	if f.CandidateUserID != nil {
		q.Where("candidate_user_id = ?", *f.CandidateUserID)
	}
	if f.SubrequestID != nil {
		q.Where("subrequest_id = ?", *f.SubrequestID)
	}
	if f.StageID != nil {
		q.Where("stage_id = ?", *f.StageID)
	}
	if f.Status != nil {
		q.Where("status = ?", *f.Status)
	}
	if f.ScheduledOnly {
		q.Where("scheduled_at IS NOT NULL")
	}
}

type InterviewResp struct {
	ID                          string             `json:"id"`
	AdminUserID                 *string            `json:"admin_user_id,omitempty"`
	CandidateUserID             string             `json:"candidate_user_id"`
	CandidateUserName           *string            `json:"candidate_user_name,omitempty"`
	CandidateUserProfilePicture *string            `json:"candidate_user_profile_picture,omitempty"`
	SubrequestID                string             `json:"subrequest_id"`
	Subrequest                  *SubrequestResp    `json:"subrequest,omitempty"`
	StageID                     string             `json:"stage_id"`
	Stage                       *InterviewStageResp `json:"stage,omitempty"`
	Title                       string             `json:"title"`
	Description                 *string            `json:"description"`
	ScheduledAt                 *string            `json:"scheduled_at"`
	Method                      *string            `json:"method"`
	Status                      string             `json:"status"`
	MeetingLink                 *string            `json:"meeting_link"`
	MeetingLocation             *string            `json:"meeting_location"`
	IsEmailSent                 bool               `json:"is_email_sent"`
	CreatedAt                   time.Time          `json:"created_at"`
	UpdatedAt                   time.Time          `json:"updated_at"`
}

func (row *Interview) ToInterviewResp() InterviewResp {
	var stageResp *InterviewStageResp
	if row.Stage != nil {
		s := row.Stage.ToInterviewStageResp()
		stageResp = &s
	}

	var subResp *SubrequestResp
	if row.Subrequest != nil {
		s := row.Subrequest.ToSubrequestResp()
		subResp = &s
	}

	var scheduledStr *string
	if row.ScheduledAt != nil {
		s := row.ScheduledAt.Format(time.RFC3339)
		scheduledStr = &s
	}

	var candidateName *string
	var candidateProfile *string
	if row.CandidateUser != nil {
		candidateName = &row.CandidateUser.Name
		candidateProfile = buildProfilePictureURL(row.CandidateUser.ProfilePicture)
	}

	return InterviewResp{
		ID:                          row.ID,
		AdminUserID:                 row.AdminUserID,
		CandidateUserID:             row.CandidateUserID,
		CandidateUserName:           candidateName,
		CandidateUserProfilePicture: candidateProfile,
		SubrequestID:                row.SubrequestID,
		Subrequest:                  subResp,
		StageID:                     row.StageID,
		Stage:                       stageResp,
		Title:                       row.Title,
		Description:                 row.Description,
		ScheduledAt:                 scheduledStr,
		Method:                      row.Method,
		Status:                      row.Status,
		MeetingLink:                 row.MeetingLink,
		MeetingLocation:             row.MeetingLocation,
		IsEmailSent:                 row.IsEmailSent,
		CreatedAt:                   row.CreatedAt,
		UpdatedAt:                   row.UpdatedAt,
	}
}
