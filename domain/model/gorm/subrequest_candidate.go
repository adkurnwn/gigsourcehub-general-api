package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubrequestCandidate struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	SubrequestID    string         `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest      *Subrequest    `gorm:"foreignKey:SubrequestID"`
	CandidateUserID string         `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User          `gorm:"foreignKey:CandidateUserID"`
	Name            string         `gorm:"column:name;type:varchar(150);not null"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *SubrequestCandidate) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
