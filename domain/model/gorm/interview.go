package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Interview struct {
	ID              string          `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	CandidateUserID string          `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User           `gorm:"foreignKey:CandidateUserID"`
	SubrequestID    string          `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest      *Subrequest     `gorm:"foreignKey:SubrequestID"`
	StageID         string          `gorm:"column:stage_id;type:uuid;not null"`
	Stage           *InterviewStage `gorm:"foreignKey:StageID"`
	ScheduledAt     *time.Time      `gorm:"column:scheduled_at;type:timestamp"`
	Method          *string         `gorm:"column:method;type:varchar(50)"`
	Status          string          `gorm:"column:status;type:interview_status"`
	MeetingLink     *string         `gorm:"column:meeting_link;type:varchar(255)"`
	IsEmailSent     bool            `gorm:"column:is_email_sent;type:boolean;default:false"`
	CreatedAt       time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}
