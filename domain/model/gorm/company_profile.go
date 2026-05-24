package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyProfile struct {
	ID           string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Address      *string        `gorm:"column:address;type:text"`
	Phone        *string        `gorm:"column:phone;type:varchar(20)"`
	Email        *string        `gorm:"column:email;type:varchar(255)"`
	FacebookURL  *string        `gorm:"column:facebook_url;type:text"`
	InstagramURL *string        `gorm:"column:instagram_url;type:text"`
	LinkedinURL  *string        `gorm:"column:linkedin_url;type:text"`
	TwitterURL   *string        `gorm:"column:twitter_url;type:text"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *CompanyProfile) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
