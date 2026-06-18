package gorm_model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OnboardHistory struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	CandidateUserID string         `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User          `gorm:"foreignKey:CandidateUserID"`
	OfferingID      *string        `gorm:"column:offering_id;type:uuid"`
	Offering        *Offering      `gorm:"foreignKey:OfferingID"`
	IsStopped                bool           `gorm:"column:is_stopped;type:boolean;default:false;not null"`
	IsExpiryNotificationSent bool           `gorm:"column:is_expiry_notification_sent;type:boolean;default:false;not null"`
	StartDate       *time.Time     `gorm:"column:start_date;type:date"`
	EndDate         *time.Time     `gorm:"column:end_date;type:date"`
	Snapshot        *string        `gorm:"column:snapshot;type:jsonb"`
	CancelledReason *string        `gorm:"column:cancelled_reason;type:text"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Review          *Review        `gorm:"foreignKey:OnboardHistoryID"`
}

type OnboardHistoryResp struct {
	ID              string     `json:"id"`
	CandidateUserID string     `json:"candidate_user_id"`
	CandidateUser   *UserResp  `json:"candidate_user,omitempty"`
	OfferingID      *string    `json:"offering_id,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	ProjectName     *string    `json:"project_name,omitempty"`
	JobRoleName     *string    `json:"job_role_name,omitempty"`
	Snapshot        *string    `json:"snapshot,omitempty"`
	CancelledReason *string    `json:"cancelled_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ReviewID        *string    `json:"review_id"`
}

type onboardHistorySnapshot struct {
	ProjectName *string `json:"project_name"`
	JobRoleName *string `json:"job_role_name"`
}

func (m *OnboardHistory) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

func (m *OnboardHistory) ToOnboardHistoryResp() OnboardHistoryResp {
	var candidateResp *UserResp
	if m.CandidateUser != nil {
		resp := m.CandidateUser.ToUserResp()
		candidateResp = &resp
	}

	var reviewID *string
	if m.Review != nil {
		reviewID = &m.Review.ID
	}

	var projectName *string
	var jobRoleName *string
	if m.Snapshot != nil && *m.Snapshot != "" {
		var snapshot onboardHistorySnapshot
		if err := json.Unmarshal([]byte(*m.Snapshot), &snapshot); err == nil {
			projectName = snapshot.ProjectName
			jobRoleName = snapshot.JobRoleName
		}
	}

	return OnboardHistoryResp{
		ID:              m.ID,
		CandidateUserID: m.CandidateUserID,
		CandidateUser:   candidateResp,
		OfferingID:      m.OfferingID,
		StartDate:       m.StartDate,
		EndDate:         m.EndDate,
		ProjectName:     projectName,
		JobRoleName:     jobRoleName,
		Snapshot:        m.Snapshot,
		CancelledReason: m.CancelledReason,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		ReviewID:        reviewID,
	}
}
