package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobVacancy struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	SubrequestID    string         `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest      *Subrequest    `gorm:"foreignKey:SubrequestID"`
	Name            string         `gorm:"column:name;type:varchar(255);not null"`
	TakedownDate    *time.Time     `gorm:"column:takedown_date;type:date"`
	FulfillmentDate *time.Time     `gorm:"column:fulfillment_date;type:date"`
	Schema          *string        `gorm:"column:schema;type:job_vacancy_schema"`
	Status          *string        `gorm:"column:status;type:job_vacancy_status"`
	Description     *string        `gorm:"column:description;type:varchar(50)"`
	Overview        *string        `gorm:"column:overview;type:text"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *JobVacancy) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
