package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FAQ struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Question  string         `gorm:"column:question;type:text;not null"`
	Answer    string         `gorm:"column:answer;type:text;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *FAQ) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
