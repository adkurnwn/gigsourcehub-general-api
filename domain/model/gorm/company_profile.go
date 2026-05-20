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

// --- Response ---

type CompanyProfileResp struct {
	ID           string    `json:"id"`
	Address      *string   `json:"address"`
	Phone        *string   `json:"phone"`
	Email        *string   `json:"email"`
	FacebookURL  *string   `json:"facebook_url"`
	InstagramURL *string   `json:"instagram_url"`
	LinkedinURL  *string   `json:"linkedin_url"`
	TwitterURL   *string   `json:"twitter_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (row *CompanyProfile) ToCompanyProfileResp() CompanyProfileResp {
	return CompanyProfileResp{
		ID:           row.ID,
		Address:      row.Address,
		Phone:        row.Phone,
		Email:        row.Email,
		FacebookURL:  row.FacebookURL,
		InstagramURL: row.InstagramURL,
		LinkedinURL:  row.LinkedinURL,
		TwitterURL:   row.TwitterURL,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
