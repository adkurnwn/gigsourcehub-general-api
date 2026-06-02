package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Offering struct {
	ID           string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Filename     string         `gorm:"column:filename;type:varchar(255);not null"`
	UserID       string         `gorm:"column:user_id_kandidat;type:uuid;not null"`
	User         *User          `gorm:"foreignKey:UserID;references:ID"`
	Path         string         `gorm:"column:path;type:varchar(255);not null"`
	SubrequestID string         `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest   *Subrequest    `gorm:"foreignKey:SubrequestID"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *Offering) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
