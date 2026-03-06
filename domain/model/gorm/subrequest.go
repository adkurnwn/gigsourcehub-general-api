package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Subrequest struct {
	ID                 string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	RequestID          string         `gorm:"column:request_id;type:uuid;not null"`
	Request            *Request       `gorm:"foreignKey:RequestID"`
	MinYearsExperience int            `gorm:"column:min_years_experience;type:int"`
	JobTitleID         *string        `gorm:"column:job_title_id;type:uuid"`
	JobTitle           *JobTitle      `gorm:"foreignKey:JobTitleID"`
	TechStack          *string        `gorm:"column:tech_stack;type:jsonb"`
	Notes              *string        `gorm:"column:notes;type:text"`
	IsFilled           bool           `gorm:"column:is_filled;type:boolean;default:false"`
	CreatedAt          time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
