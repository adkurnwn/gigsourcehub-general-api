package gorm_model

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CareerDepartment struct {
	ID          string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Name        string         `gorm:"column:name;type:text;not null"`
	Description string         `gorm:"column:description;type:text;not null"`
	ImagePath   *string        `gorm:"column:image_path;type:varchar(255)"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *CareerDepartment) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

// --- Filter ---

var CareerDepartmentAllowedSort = []string{"name", "created_at", "updated_at"}

type CareerDepartmentFilter struct {
	DefaultFilter
	Search *string
}

func (f *CareerDepartmentFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.Search != nil && *f.Search != "" {
		q.Where("name ILIKE ?", "%"+*f.Search+"%")
	}
}

// --- Response ---

type CareerDepartmentResp struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ImageURL       *string    `json:"image_url"`
	Author         *string    `json:"author"`
	Status         *string    `json:"status"`
	RejectedReason *string    `json:"rejected_reason,omitempty"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (row *CareerDepartment) ToCareerDepartmentResp(status string) CareerDepartmentResp {
	var publishedAt *time.Time = &row.CreatedAt

	// Resolve image URL from S3 public URL
	var imageURL *string
	if row.ImagePath != nil && *row.ImagePath != "" {
		ip := *row.ImagePath
		if len(ip) > 0 && ip[0] != 'h' {
			ip = fmt.Sprintf("%s/%s", os.Getenv("S3_PUBLIC_URL"), ip)
		}
		imageURL = &ip
	}

	return CareerDepartmentResp{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		ImageURL:       imageURL,
		Status:         &status,
		RejectedReason: nil,
		PublishedAt:    publishedAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
