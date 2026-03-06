package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Contract struct {
	ID           string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Filename     string         `gorm:"column:filename;type:varchar(255);not null"`
	UserID       string         `gorm:"column:user_id_kandidat;type:uuid;not null"`
	User         *User          `gorm:"foreignKey:UserID;references:ID"`
	Path         string         `gorm:"column:path;type:varchar(255);not null"`
	SubrequestID string         `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest   *Subrequest    `gorm:"foreignKey:SubrequestID"`
	StartDate    *time.Time     `gorm:"column:start_date;type:date"`
	EndDate      *time.Time     `gorm:"column:end_date;type:date"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
