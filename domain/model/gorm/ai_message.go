package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIMessage struct {
	ID            string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()" json:"id"`
	AIChatID      string         `gorm:"column:ai_chat_id;type:uuid;not null" json:"ai_chat_id"`
	AIChat        *AIChat        `gorm:"foreignKey:AIChatID" json:"ai_chat,omitempty"`
	Role          string         `gorm:"column:role;type:varchar(20);not null" json:"role"`
	Content       string         `gorm:"column:content;type:text;not null" json:"content"`
	IsLastMessage bool           `gorm:"column:is_last_message;type:boolean;default:false;not null" json:"is_last_message"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *AIMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
