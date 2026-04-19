package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type JobTitle struct {
	ID        string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	SectorID  string         `gorm:"column:sector_id;type:uuid;not null"`
	Sector    *Sector        `gorm:"foreignKey:SectorID"`
	Name      string         `gorm:"column:name;type:varchar(150);not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var JobTitleAllowedSort = []string{"name", "created_at", "updated_at"}

type JobTitleFilter struct {
	DefaultFilter
	SectorID *string
	Name     *string
	Search   *string
}

func (f *JobTitleFilter) Query(q *gorm.DB) {
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
}

type JobTitleResp struct {
	ID       string      `json:"id"`
	SectorID string      `json:"sector_id"`
	Name     string      `json:"name"`
	Sector   *SectorResp `json:"sector,omitempty"`
}

func (row *JobTitle) ToJobTitleResp() JobTitleResp {
	var sectorResp *SectorResp
	if row.Sector != nil {
		s := row.Sector.ToSectorResp()
		sectorResp = &s
	}
	return JobTitleResp{
		ID:       row.ID,
		SectorID: row.SectorID,
		Name:     row.Name,
		Sector:   sectorResp,
	}
}
