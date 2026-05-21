package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type InterviewStage struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var InterviewStageAllowedSort = []string{"name", "created_at", "updated_at"}

type InterviewStageFilter struct {
	DefaultFilter
	Name   *string
	Search *string
}

func (f *InterviewStageFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where("name = ?", *f.Name)
	}
	if f.Search != nil && *f.Search != "" {
		q.Where("name ILIKE ?", "%"+*f.Search+"%")
	}
}

type InterviewStageResp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (row *InterviewStage) ToInterviewStageResp() InterviewStageResp {
	return InterviewStageResp{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
