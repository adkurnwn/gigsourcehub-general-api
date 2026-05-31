package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CandidateStatusHistory struct {
	ID                  string             `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	CandidateUserID     string             `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser       *User              `gorm:"foreignKey:CandidateUserID"`
	RecruitmentStatusID *string            `gorm:"column:recruitment_status_id;type:uuid"`
	RecruitmentStatus   *RecruitmentStatus `gorm:"foreignKey:RecruitmentStatusID"`
	SubrequestID        *string            `gorm:"column:subrequest_id;type:uuid"`
	Subrequest          *Subrequest        `gorm:"foreignKey:SubrequestID"`
	ChangedByUserID     string             `gorm:"column:changed_by_user_id;type:uuid;not null"`
	ChangedByUser       *User              `gorm:"foreignKey:ChangedByUserID"`
	ChangedAt           time.Time          `gorm:"column:changed_at;type:timestamp;default:now()"`
}

func (m *CandidateStatusHistory) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.ChangedAt.IsZero() {
		m.ChangedAt = time.Now()
	}
	return
}
