package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Request struct {
	ID                string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	ProjectName       string         `gorm:"column:project_name;type:varchar(255);not null"`
	DueDate           *time.Time     `gorm:"column:due_date;type:date"`
	AdminUserID       *string        `gorm:"column:admin_user_id;type:uuid"`
	AdminUser         *User          `gorm:"foreignKey:AdminUserID"`
	EmployeeUserID    string         `gorm:"column:employee_user_id;type:uuid;not null"`
	EmployeeUser      *User          `gorm:"foreignKey:EmployeeUserID"`
	RequiredHeadcount int            `gorm:"column:required_headcount;type:int;not null"`
	Status            string         `gorm:"column:status;type:varchar(50)"`
	Urgency           string         `gorm:"column:urgency;type:request_urgency"`
	FulfillmentDate   *time.Time     `gorm:"column:fulfillment_date;type:date"`
	Subrequests       []Subrequest   `gorm:"foreignKey:RequestID"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
