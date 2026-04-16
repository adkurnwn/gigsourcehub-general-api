package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID               string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserIDOwner string         `gorm:"column:admin_user_id_owner;type:uuid;not null"`
	AdminUserOwner   *User          `gorm:"foreignKey:AdminUserIDOwner"`
	Title            string         `gorm:"column:title;type:varchar(255);not null"`
	Description      *string        `gorm:"column:description;type:text"`
	IsRead           bool           `gorm:"column:is_read;type:boolean;default:false;not null"`
	IsAdminBroadcast bool           `gorm:"column:is_admin_broadcast;type:boolean;default:false;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
