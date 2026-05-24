package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FAQ struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Question  string         `gorm:"column:question;type:text;not null"`
	Answer    string         `gorm:"column:answer;type:text;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *FAQ) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

// --- Filter ---

var FAQAllowedSort = []string{"question", "created_at", "updated_at"}

type FAQFilter struct {
	DefaultFilter
	Search *string
}

func (f *FAQFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.Search != nil && *f.Search != "" {
		q.Where("question ILIKE ?", "%"+*f.Search+"%")
	}
}

// --- Response ---

type FAQResp struct {
	ID          string     `json:"id"`
	Question    string     `json:"question"`
	Answer      string     `json:"answer"`
	Author      *string    `json:"author"`
	Status      *string    `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (row *FAQ) ToFAQResp(status string) FAQResp {
	var publishedAt *time.Time = &row.CreatedAt
	return FAQResp{
		ID:          row.ID,
		Question:    row.Question,
		Answer:      row.Answer,
		Status:      &status,
		PublishedAt: publishedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
