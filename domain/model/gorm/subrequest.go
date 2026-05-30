package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subrequest struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	RequestID string         `gorm:"column:request_id;type:uuid;not null"`
	Request   *Request       `gorm:"foreignKey:RequestID"`
	Level     *string        `gorm:"column:level;type:subrequest_level"`
	JobRoleID *string        `gorm:"column:job_role_id;type:uuid"`
	JobRole   *JobRole       `gorm:"foreignKey:JobRoleID"`
	TechStack *string        `gorm:"column:tech_stack;type:jsonb"`
	Notes     *string        `gorm:"column:notes;type:text"`
	IsFilled  bool           `gorm:"column:is_filled;type:boolean;default:false"`
	Overview  *string        `gorm:"column:overview;type:text"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *Subrequest) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

type SubrequestResp struct {
	ID              string  `json:"id"`
	RequestID       string  `json:"request_id"`
	ProjectName     string  `json:"project_name,omitempty"`
	ProjectDuration *string `json:"project_duration,omitempty"`
	Level           *string `json:"level"`
	JobRoleID       *string `json:"job_role_id"`
	JobRole         *string `json:"job_role"`
	Bidang          *string `json:"bidang,omitempty"`
	TechStack       *string `json:"tech_stack"`
	Notes           *string `json:"notes"`
	IsFilled        bool    `json:"is_filled"`
	Overview        *string `json:"overview"`
}

func (row *Subrequest) ToSubrequestResp() SubrequestResp {
	var jobRoleName *string
	var bidangName *string
	if row.JobRole != nil {
		jobRoleName = &row.JobRole.Name
		if row.JobRole.Sector != nil {
			bidangName = &row.JobRole.Sector.Name
		}
	}

	var projectName *string
	var projectDuration *string
	if row.Request != nil {
		projectName = &row.Request.ProjectName
		projectDuration = row.Request.ProjectDuration
	}

	return SubrequestResp{
		ID:              row.ID,
		RequestID:       row.RequestID,
		ProjectName:     func() string { if projectName != nil { return *projectName }; return "" }(),
		ProjectDuration: projectDuration,
		Level:           row.Level,
		JobRoleID:       row.JobRoleID,
		JobRole:         jobRoleName,
		Bidang:          bidangName,
		TechStack:       row.TechStack,
		Notes:           row.Notes,
		IsFilled:        row.IsFilled,
		Overview:        row.Overview,
	}
}
