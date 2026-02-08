package gorm_model

import (
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

type CV struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	UserID    string         `gorm:"column:user_id;type:uuid;not null"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	Path      string         `gorm:"column:path;type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var CVAllowedSort = []string{"name", "created_at", "updated_at"}

type CVFilter struct {
	DefaultFilter
	Name *string
	Path *string
}

func (f *CVFilter) Query(q *gorm.DB) {

	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where("name = ?", *f.Name)
	}

}

type CVResp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (row *CV) ToCVResp() CVResp {
	path := row.Path
	// If path is relative (doesn't start with http), prepend public url
	if len(path) > 0 && path[0] != 'h' {
		path = fmt.Sprintf("%s/%s", os.Getenv("S3_PUBLIC_URL"), row.Path)
	}
	return CVResp{
		ID:        row.ID,
		Name:      row.Name,
		Path:      path,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
