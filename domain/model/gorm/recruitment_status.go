package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type RecruitmentStatus struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Name      string         `gorm:"column:name;type:varchar(150);not null"`
	HexCode   string         `gorm:"column:hex_code;type:varchar(10)"`
	IsActive  bool           `gorm:"column:is_active;type:boolean;default:true"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var RecruitmentStatusAllowedSort = []string{"name", "created_at", "updated_at"}

type RecruitmentStatusFilter struct {
	DefaultFilter
	Name     *string
	IsActive *bool
	RecruitmentStatusID *string
}

func (f *RecruitmentStatusFilter) Query(q *gorm.DB) {
	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where("name = ?", *f.Name)
	}
	if f.IsActive != nil {
		q.Where("is_active = ?", *f.IsActive)
	}
}

type RecruitmentStatusResp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	HexCode   string    `json:"hex_code"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (row *RecruitmentStatus) ToRecruitmentStatusResp() RecruitmentStatusResp {
	return RecruitmentStatusResp{
		ID:        row.ID,
		Name:      row.Name,
		HexCode:   row.HexCode,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
