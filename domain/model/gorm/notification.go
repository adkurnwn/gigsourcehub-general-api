package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID               string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserIDOwner string         `gorm:"column:admin_user_id_owner;type:uuid;not null"` // Used as generic user_id_owner for all roles
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

type NotificationResp struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

func (n *Notification) ToNotificationResp() NotificationResp {
	return NotificationResp{
		ID:          n.ID,
		Title:       n.Title,
		Description: n.Description,
		IsRead:      n.IsRead,
		CreatedAt:   n.CreatedAt,
	}
}
