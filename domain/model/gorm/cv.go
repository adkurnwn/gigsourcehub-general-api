package gorm_model

import (
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

type CV struct {
	ID         string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	UserID     string         `gorm:"column:user_id;type:uuid;not null"`
	Filename   string         `gorm:"column:filename;type:varchar(255);not null"`
	Path       string         `gorm:"column:path;type:varchar(255);not null"`
	ParsedData *string        `gorm:"column:parsed_data;type:jsonb"`
	Status     string         `gorm:"column:status;type:varchar(50);default:'UPLOADED'"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var CVAllowedSort = []string{"filename", "created_at", "updated_at"}

type CVFilter struct {
	DefaultFilter
	Filename *string
	Path     *string
}

func (f *CVFilter) Query(q *gorm.DB) {

	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.Filename != nil {
		q.Where("filename = ?", *f.Filename)
	}

}

type CVResp struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	Path       string    `json:"path"`
	ParsedData string    `json:"parsed_data,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (row *CV) ToCVResp() CVResp {
	path := row.Path
	// If path is relative (doesn't start with http), prepend public url
	if len(path) > 0 && path[0] != 'h' {
		path = fmt.Sprintf("%s/%s", os.Getenv("S3_PUBLIC_URL"), row.Path)
	}
	var parsedData string
	if row.ParsedData != nil {
		parsedData = *row.ParsedData
	}
	return CVResp{
		ID:         row.ID,
		Filename:   row.Filename,
		Path:       path,
		ParsedData: parsedData,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
