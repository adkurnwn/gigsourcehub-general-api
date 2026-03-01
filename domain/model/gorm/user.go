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
	Birthdate           *time.Time     `gorm:"column:birthdate;type:date"`
	SchoolUniversity    *string        `gorm:"column:school_university;type:varchar(255)"`
	Major               *string        `gorm:"column:major;type:varchar(255)"`
	Gpa                 *float64       `gorm:"column:gpa;type:decimal(3,2)"`
	PhoneNumber         *string        `gorm:"column:phone_number;type:varchar(20)"`
	PortofolioLink      *string        `gorm:"column:portofolio_link;type:text"`
	KabupatenKotaId     *string        `gorm:"column:kabupaten_kota_id;type:varchar(5)"`
	YearsExperience     *int           `gorm:"column:years_experience;type:int"`
	TechStack           *string        `gorm:"column:tech_stack;type:jsonb"`
	ProfilePicture      *string        `gorm:"column:profile_picture;type:varchar(255)"`
	CandidateLevel      *string        `gorm:"column:candidate_level;type:varchar(50)"`
	RecruitmentStatusId *string        `gorm:"column:recruitment_status_id;type:uuid"`
	UnavailableUntil    *time.Time     `gorm:"column:unavailable_until;type:date"`
	RoleSystemId        *string        `gorm:"column:role_system_id;type:uuid"`
	RoleSystem          *RoleSystem    `gorm:"foreignKey:RoleSystemId"`
	RoleAppliedId       *string        `gorm:"column:role_applied_id;type:uuid"`
	AccountStatus       *string        `gorm:"column:account_status;type:varchar(10)"`
	Jabatan             *string        `gorm:"column:jabatan;type:varchar(100)"`
	CreatedAt           time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
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
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Email               string     `json:"email"`
	Birthdate           *time.Time `json:"birthdate"`
	SchoolUniversity    *string    `json:"school_university"`
	Major               *string    `json:"major"`
	Gpa                 *float64   `json:"gpa"`
	PhoneNumber         *string    `json:"phone_number"`
	PortofolioLink      *string    `json:"portofolio_link"`
	KabupatenKotaId     *string    `json:"kabupaten_kota_id"`
	YearsExperience     *int       `json:"years_experience"`
	TechStack           *string    `json:"tech_stack"`
	ProfilePicture      *string    `json:"profile_picture"`
	CandidateLevel      *string    `json:"candidate_level,omitempty"`
	RecruitmentStatusId *string    `json:"recruitment_status_id"`
	UnavailableUntil    *time.Time `json:"unavailable_until"`
	RoleSystemName      *string    `json:"role_system_name"`
	RoleAppliedId       *string    `json:"role_applied_id"`
	AccountStatus       *string    `json:"account_status"`
	Jabatan             *string    `json:"jabatan,omitempty"`
}

func (row *User) ToUserResp() UserResp {
	var roleName *string
	if row.RoleSystem != nil {
		roleName = &row.RoleSystem.Name
	}

	candidateLevel := row.CandidateLevel
	jabatan := row.Jabatan

	if roleName != nil && *roleName == "Candidate" {
		candidateLevel = nil
		jabatan = nil
	}

	return UserResp{
		ID:                  row.ID,
		Name:                row.Name,
		Email:               row.Email,
		Birthdate:           row.Birthdate,
		SchoolUniversity:    row.SchoolUniversity,
		Major:               row.Major,
		Gpa:                 row.Gpa,
		PhoneNumber:         row.PhoneNumber,
		PortofolioLink:      row.PortofolioLink,
		KabupatenKotaId:     row.KabupatenKotaId,
		YearsExperience:     row.YearsExperience,
		TechStack:           row.TechStack,
		ProfilePicture:      row.ProfilePicture,
		CandidateLevel:      candidateLevel,
		RecruitmentStatusId: row.RecruitmentStatusId,
		UnavailableUntil:    row.UnavailableUntil,
		RoleSystemName:      roleName,
		RoleAppliedId:       row.RoleAppliedId,
		AccountStatus:       row.AccountStatus,
		Jabatan:             jabatan,
	}
}
