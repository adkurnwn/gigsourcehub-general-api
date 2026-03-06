package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Conversation struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID     string         `gorm:"column:admin_user_id;type:uuid;not null"`
	AdminUser       *User          `gorm:"foreignKey:AdminUserID"`
	CandidateUserID string         `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User          `gorm:"foreignKey:CandidateUserID"`
	Messages        []Message      `gorm:"foreignKey:ConversationID"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
