package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type RoleSystem struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Name      string         `gorm:"column:name;type:varchar(150);not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var RoleSystemAllowedSort = []string{"name", "created_at", "updated_at"}

type RoleSystemFilter struct {
	DefaultFilter
	Name *string
}

func (f *RoleSystemFilter) Query(q *gorm.DB) {
	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where("name = ?", *f.Name)
	}
}

type RoleSystemResp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (row *RoleSystem) ToRoleSystemResp() RoleSystemResp {
	return RoleSystemResp{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
