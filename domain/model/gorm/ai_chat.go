package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIChat struct {
	ID          string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID string         `gorm:"column:admin_user_id;type:uuid;not null"`
	AdminUser   *User          `gorm:"foreignKey:AdminUserID"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *AIChat) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
