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

// --- Filter ---

var JobVacancyAllowedSort = []string{"name", "created_at", "updated_at", "takedown_date"}

type JobVacancyFilter struct {
	DefaultFilter
	SubrequestID  *string
	Status        *string
	Schema        *string
	Search        *string
	// OnlyPublicValid filters to PUBLISHED vacancies whose takedown_date is NULL or >= today
	OnlyPublicValid bool
}

func (f *JobVacancyFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.SubrequestID != nil {
		q.Where("subrequest_id = ?", *f.SubrequestID)
	}
	if f.Status != nil {
		q.Where("status = ?", *f.Status)
	}
	if f.Schema != nil {
		q.Where("schema = ?", *f.Schema)
	}
	if f.Search != nil && *f.Search != "" {
		q.Where("name ILIKE ?", "%"+*f.Search+"%")
	}
	if f.OnlyPublicValid {
		q.Where("status = 'PUBLISHED' AND (takedown_date IS NULL OR takedown_date >= CURRENT_DATE)")
	}
}

// --- Response ---

type JobVacancyResp struct {
	ID              string          `json:"id"`
	SubrequestID    string          `json:"subrequest_id"`
	Subrequest      *SubrequestResp `json:"subrequest,omitempty"`
	Name            string          `json:"name"`
	TakedownDate    *string         `json:"takedown_date"`
	FulfillmentDate *string         `json:"fulfillment_date"`
	Schema          *string         `json:"schema"`
	Status          *string         `json:"status"`
	Description     *string         `json:"description"`
	Overview        *string         `json:"overview"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (row *JobVacancy) ToJobVacancyResp() JobVacancyResp {
	var subrequestResp *SubrequestResp
	if row.Subrequest != nil {
		s := row.Subrequest.ToSubrequestResp()
		subrequestResp = &s
	}

	var takedownStr *string
	if row.TakedownDate != nil {
		s := row.TakedownDate.Format("2006-01-02")
		takedownStr = &s
	}

	var fulfillmentStr *string
	if row.FulfillmentDate != nil {
		s := row.FulfillmentDate.Format("2006-01-02")
		fulfillmentStr = &s
	}

	return JobVacancyResp{
		ID:              row.ID,
		SubrequestID:    row.SubrequestID,
		Subrequest:      subrequestResp,
		Name:            row.Name,
		TakedownDate:    takedownStr,
		FulfillmentDate: fulfillmentStr,
		Schema:          row.Schema,
		Status:          row.Status,
		Description:     row.Description,
		Overview:        row.Overview,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
