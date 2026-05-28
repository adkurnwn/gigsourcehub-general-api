package gorm_model

import (
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
	StartDate       *time.Time     `gorm:"column:start_date;type:date"`
	EndDate         *time.Time     `gorm:"column:end_date;type:date"`
	Snapshot        *string        `gorm:"column:snapshot;type:jsonb"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type OnboardHistoryResp struct {
	ID              string     `json:"id"`
	CandidateUserID string     `json:"candidate_user_id"`
	CandidateUser   *UserResp  `json:"candidate_user,omitempty"`
	OfferingID      *string    `json:"offering_id,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	Snapshot        *string    `json:"snapshot,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
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

	return OnboardHistoryResp{
		ID:              m.ID,
		CandidateUserID: m.CandidateUserID,
		CandidateUser:   candidateResp,
		OfferingID:      m.OfferingID,
		StartDate:       m.StartDate,
		EndDate:         m.EndDate,
		Snapshot:        m.Snapshot,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
