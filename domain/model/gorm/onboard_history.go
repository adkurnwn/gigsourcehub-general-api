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

func (m *OnboardHistory) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
