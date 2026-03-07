package gorm_model

import (
	"time"

	"github.com/google/uuid"
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
	Status            string         `gorm:"column:status;type:request_status"`
	Urgency           string         `gorm:"column:urgency;type:request_urgency"`
	FulfillmentDate   *time.Time     `gorm:"column:fulfillment_date;type:date"`
	Subrequests       []Subrequest   `gorm:"foreignKey:RequestID"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *Request) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

type RequestResp struct {
	ID                string           `json:"id"`
	ProjectName       string           `json:"project_name"`
	DueDate           *string          `json:"due_date"`
	AdminUserID       *string          `json:"admin_user_id"`
	EmployeeUserID    string           `json:"employee_user_id"`
	RequiredHeadcount int              `json:"required_headcount"`
	Status            string           `json:"status"`
	Urgency           string           `json:"urgency"`
	FulfillmentDate   *string          `json:"fulfillment_date"`
	Subrequests       []SubrequestResp `json:"subrequests,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

func (row *Request) ToRequestResp() RequestResp {
	var dueDateStr *string
	if row.DueDate != nil {
		str := row.DueDate.Format(time.RFC3339)
		dueDateStr = &str
	}
	var fulfillmentStr *string
	if row.FulfillmentDate != nil {
		str := row.FulfillmentDate.Format(time.RFC3339)
		fulfillmentStr = &str
	}

	var subResponses []SubrequestResp
	if row.Subrequests != nil {
		for _, sub := range row.Subrequests {
			subResponses = append(subResponses, sub.ToSubrequestResp())
		}
	}

	return RequestResp{
		ID:                row.ID,
		ProjectName:       row.ProjectName,
		DueDate:           dueDateStr,
		AdminUserID:       row.AdminUserID,
		EmployeeUserID:    row.EmployeeUserID,
		RequiredHeadcount: row.RequiredHeadcount,
		Status:            row.Status,
		Urgency:           row.Urgency,
		FulfillmentDate:   fulfillmentStr,
		Subrequests:       subResponses,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
