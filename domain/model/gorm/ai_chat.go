package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIChat struct {
	ID          string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()" json:"id"`
	AdminUserID string         `gorm:"column:admin_user_id;type:uuid;not null" json:"admin_user_id"`
	AdminUser   *User          `gorm:"foreignKey:AdminUserID" json:"admin_user,omitempty"`
	Title       string         `gorm:"column:title;type:varchar(255)" json:"title"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *AIChat) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
