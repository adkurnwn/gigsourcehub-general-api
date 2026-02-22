package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Name                string         `gorm:"column:name;type:varchar(255);not null"`
	Email               string         `gorm:"column:email;type:varchar(255);not null"`
	Password            string         `gorm:"column:password;type:varchar(255);not null"`
	CreatedAt           time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
	PendidikanTerakhir  *string        `gorm:"column:pendidikan_terakhir;type:varchar(255)"`
	InstansiPendidikan  *string        `gorm:"column:instansi_pendidikan;type:varchar(255)"`
	Jurusan             *string        `gorm:"column:jurusan;type:varchar(255)"`
	Ipk                 *string        `gorm:"column:ipk;type:varchar(50)"`
	KabupatenId         *string        `gorm:"column:kabupaten_id;type:varchar(5)"`
	ProvinsiId          *string        `gorm:"column:provinsi_id;type:varchar(2)"`
	LamaPengalamanKerja *string        `gorm:"column:lama_pengalaman_kerja;type:varchar(255)"`
	BidangMinat         *string        `gorm:"column:bidang_minat;type:varchar(255)"`
	AppliedRole         *string        `gorm:"column:applied_role;type:varchar(255)"`
	Skills              *string        `gorm:"column:skills;type:jsonb"`
	LinkPortofolio      *string        `gorm:"column:link_portofolio;type:varchar(255)"`
}

var UserAllowedSort = []string{"name", "email", "created_at", "updated_at"}

type UserFilter struct {
	DefaultFilter
	Name  *string
	Email *string
}

func (f *UserFilter) Query(q *gorm.DB) {

	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where("name = ?", *f.Name)
	}
	if f.Email != nil {
		q.Where("email = ?", *f.Email)
	}
}

type UserResp struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	PendidikanTerakhir  *string   `json:"pendidikan_terakhir"`
	InstansiPendidikan  *string   `json:"instansi_pendidikan"`
	Jurusan             *string   `json:"jurusan"`
	Ipk                 *string   `json:"ipk"`
	KabupatenId         *string   `json:"kabupaten_id"`
	ProvinsiId          *string   `json:"provinsi_id"`
	LamaPengalamanKerja *string   `json:"lama_pengalaman_kerja"`
	BidangMinat         *string   `json:"bidang_minat"`
	AppliedRole         *string   `json:"applied_role"`
	Skills              *string   `json:"skills"`
	LinkPortofolio      *string   `json:"link_portofolio"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (row *User) ToUserResp() UserResp {
	return UserResp{
		ID:                  row.ID,
		Name:                row.Name,
		Email:               row.Email,
		PendidikanTerakhir:  row.PendidikanTerakhir,
		InstansiPendidikan:  row.InstansiPendidikan,
		Jurusan:             row.Jurusan,
		Ipk:                 row.Ipk,
		KabupatenId:         row.KabupatenId,
		ProvinsiId:          row.ProvinsiId,
		LamaPengalamanKerja: row.LamaPengalamanKerja,
		BidangMinat:         row.BidangMinat,
		AppliedRole:         row.AppliedRole,
		Skills:              row.Skills,
		LinkPortofolio:      row.LinkPortofolio,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
