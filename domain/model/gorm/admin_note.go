package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminNote struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID     string         `gorm:"column:admin_user_id;type:uuid;not null"`
	AdminUser       *User          `gorm:"foreignKey:AdminUserID"`
	CandidateUserID string         `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User          `gorm:"foreignKey:CandidateUserID"`
	Content         string         `gorm:"column:content;type:text;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type AdminNoteResp struct {
	ID              string     `json:"id"`
	AdminUserID     string     `json:"admin_user_id"`
	AdminUserName   *string    `json:"admin_user_name,omitempty"`
	CandidateUserID string     `json:"candidate_user_id"`
	Content         string     `json:"content"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

func (m *AdminNote) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

func (m *AdminNote) ToAdminNoteResp() AdminNoteResp {
	var adminName *string
	if m.AdminUser != nil {
		adminName = &m.AdminUser.Name
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return AdminNoteResp{
		ID:              m.ID,
		AdminUserID:     m.AdminUserID,
		AdminUserName:   adminName,
		CandidateUserID: m.CandidateUserID,
		Content:         m.Content,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}
