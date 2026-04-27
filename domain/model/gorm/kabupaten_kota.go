package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type KabupatenKota struct {
	ID         string         `gorm:"column:id;primarykey;type:varchar(5); not null"`
	Name       string         `gorm:"column:name;type:varchar(255);not null"`
	ProvinsiID string         `gorm:"column:provinsi_id;type:varchar(2);not null"`
	Provinsi   *Provinsi      `gorm:"foreignKey:ProvinsiID;references:ID"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (KabupatenKota) TableName() string {
	return "kabupaten_kota"
}

var KabupatenKotaAllowedSort = []string{"id"}

type KabupatenKotaFilter struct {
	DefaultFilter
	ID         *string
	Name       *string
	ProvinsiID *string
}

func (k *KabupatenKotaFilter) Query(q *gorm.DB) {
	k.DefaultFilter.DefaultQuery(q)

	if k.ID != nil {
		q.Where("id = ?", *k.ID)
	}
	if k.Name != nil {
		q.Where("name ILIKE ?", "%"+*k.Name+"%")
	}
	if k.ProvinsiID != nil {
		q.Where("provinsi_id = ?", *k.ProvinsiID)
	}
}

type KabupatenKotaResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (row *KabupatenKota) ToKabupatenKotaResp() KabupatenKotaResp {
	return KabupatenKotaResp{
		ID:   row.ID,
		Name: row.Name,
	}
}
