package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	ID                          string          `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	SubrequestID                string          `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest                  *Subrequest     `gorm:"foreignKey:SubrequestID"`
	CandidateUserID             string          `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser               *User           `gorm:"foreignKey:CandidateUserID"`
	EmployeeUserID              string          `gorm:"column:employee_user_id;type:uuid;not null"`
	EmployeeUser                *User           `gorm:"foreignKey:EmployeeUserID"`
	OnboardHistoryID            string          `gorm:"column:onboard_history_id;type:uuid;not null"`
	OnboardHistory              *OnboardHistory `gorm:"foreignKey:OnboardHistoryID"`
	WorkQuality                 int             `gorm:"column:work_quality;type:int"`
	Timeliness                  int             `gorm:"column:timeliness;type:int"`
	CommunicationCollaboration  int             `gorm:"column:communication_collaboration;type:int"`
	ProblemSolvingInitiative    int             `gorm:"column:problem_solving_initiative;type:int"`
	FinalRecommendation         string          `gorm:"column:final_recommendation;type:review_final_recommendation"`
	Notes                       *string         `gorm:"column:notes;type:text"`
	CreatedAt                   time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                   time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt                   gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}
