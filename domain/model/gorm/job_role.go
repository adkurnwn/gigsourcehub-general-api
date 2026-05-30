package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type JobRole struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	SectorID  string         `gorm:"column:sector_id;type:uuid;not null"`
	Sector    *Sector        `gorm:"foreignKey:SectorID"`
	Name      string         `gorm:"column:name;type:varchar(150);not null"`
	IsActive  bool           `gorm:"column:is_active;type:boolean;default:true"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var JobRoleAllowedSort = []string{"name", "created_at", "updated_at"}

type JobRoleFilter struct {
	DefaultFilter
	SectorID *string
	Name     *string
	Search   *string
	IsActive *bool
}

func (f *JobRoleFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where("name = ?", *f.Name)
	}
	if f.Search != nil && *f.Search != "" {
		q.Where("name ILIKE ?", "%"+*f.Search+"%")
	}
	if f.SectorID != nil {
		q.Where("sector_id = ?", *f.SectorID)
	}
	if f.IsActive != nil {
		q.Where("is_active = ?", *f.IsActive)
	}
}

type JobRoleResp struct {
	ID       string      `json:"id"`
	SectorID string      `json:"sector_id"`
	Name     string      `json:"name"`
	IsActive bool        `json:"is_active"`
	Sector   *SectorResp `json:"sector,omitempty"`
}

func (row *JobRole) ToJobRoleResp() JobRoleResp {
	var sectorResp *SectorResp
	if row.Sector != nil {
		s := row.Sector.ToSectorResp()
		sectorResp = &s
	}
	return JobRoleResp{
		ID:       row.ID,
		SectorID: row.SectorID,
		Name:     row.Name,
		IsActive: row.IsActive,
		Sector:   sectorResp,
	}
}
