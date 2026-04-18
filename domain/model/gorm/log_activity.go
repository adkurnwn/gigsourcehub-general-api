package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogActivity struct {
	ID          string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	ActorID     string         `gorm:"column:actor_id;type:uuid;not null"`
	Actor       *User          `gorm:"foreignKey:ActorID"`
	ActionType  string         `gorm:"column:action_type;type:varchar(100);not null"`
	Module      string         `gorm:"column:module;type:varchar(100);not null"`
	Description *string        `gorm:"column:description;type:text"`
	Metadata    *string        `gorm:"column:metadata;type:jsonb"`
	IPAddress   *string        `gorm:"column:ip_address;type:varchar(45)"`
	IsSuccess   bool           `gorm:"column:is_success;type:boolean;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *LogActivity) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
