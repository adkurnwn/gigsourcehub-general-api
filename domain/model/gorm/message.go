package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID             string              `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	ConversationID string              `gorm:"column:conversation_id;type:uuid;not null"`
	Conversation   *Conversation       `gorm:"foreignKey:ConversationID"`
	Content        string              `gorm:"column:content;type:text;not null"`
	SenderUserID   string              `gorm:"column:sender_user_id;type:uuid;not null"`
	SenderUser     *User               `gorm:"foreignKey:SenderUserID"`
	Attachments    []MessageAttachment `gorm:"foreignKey:MessageID"`
	CreatedAt      time.Time           `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time           `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt      `gorm:"column:deleted_at;index"`
}
