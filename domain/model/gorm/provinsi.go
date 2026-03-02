package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type Provinsi struct {
	ID        string         `gorm:"column:id;primarykey;type:varchar(5); not null"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Provinsi) TableName() string {
	return "provinsi"
}

var ProvinsiAllowedSort = []string{"id"}

type ProvinsiFilter struct {
	DefaultFilter
	ID   *string
	Name *string
}

func (p *ProvinsiFilter) Query(q *gorm.DB) {
	p.DefaultFilter.DefaultQuery(q)

	if p.ID != nil {
		q.Where("id = ?", *p.ID)
	}
	if p.Name != nil {
		q.Where("name ILIKE ?", "%"+*p.Name+"%")
	}
}

type ProvinsiResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (row *Provinsi) ToProvinsiResp() ProvinsiResp {
	return ProvinsiResp{
		ID:   row.ID,
		Name: row.Name,
	}
}
