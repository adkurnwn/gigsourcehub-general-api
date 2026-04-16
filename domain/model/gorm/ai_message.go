package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIMessage struct {
	ID            string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AIChatID      string         `gorm:"column:ai_chat_id;type:uuid;not null"`
	AIChat        *AIChat        `gorm:"foreignKey:AIChatID"`
	Content       string         `gorm:"column:content;type:json;not null"`
	IsLastMessage bool           `gorm:"column:is_last_message;type:boolean;default:false;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *AIMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
