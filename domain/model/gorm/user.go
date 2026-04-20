package gorm_model

import (
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Name                string         `gorm:"column:name;type:varchar(255);not null"`
	Email               string         `gorm:"column:email;type:varchar(255);unique;not null"`
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
	Summary             *string        `gorm:"column:summary;type:text"`
	ProfilePicture      *string        `gorm:"column:profile_picture;type:varchar(255)"`
	CandidateLevel      *string        `gorm:"column:candidate_level;type:varchar(50)"`
	RecruitmentStatusId *string        `gorm:"column:recruitment_status_id;type:uuid"`
	UnavailableUntil    *time.Time     `gorm:"column:unavailable_until;type:date"`
	SystemRoleId        *string        `gorm:"column:system_role_id;type:uuid"`
	SystemRole          *SystemRole    `gorm:"foreignKey:SystemRoleId"`
	AssignedRoleId      *string        `gorm:"column:assigned_role_id;type:uuid"`
	AssignedRole        *JobRole       `gorm:"foreignKey:AssignedRoleId"`
	JobRoles            []JobRole      `gorm:"many2many:user_has_job_roles;"`
	AccountStatus       *string        `gorm:"column:account_status;type:user_account_status"`
	JobTitleId          *string        `gorm:"column:job_title_id;type:uuid"`
	JobTitle            *JobTitle      `gorm:"foreignKey:JobTitleId"`
	VerifiedAt          *time.Time     `gorm:"column:verified_at;type:timestamptz"`
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
	Summary             *string    `json:"summary"`
	ProfilePicture      *string    `json:"profile_picture"`
	CandidateLevel      *string    `json:"candidate_level,omitempty"`
	RecruitmentStatusId *string    `json:"recruitment_status_id"`
	UnavailableUntil    *time.Time `json:"unavailable_until"`
	SystemRoleName      *string    `json:"system_role_name"`
	AssignedRoleId      *string    `json:"assigned_role_id"`
	AccountStatus       *string    `json:"account_status"`
	IsBookmark          *bool         `json:"is_bookmark,omitempty"`
	JobTitleId          *string       `json:"job_title_id,omitempty"`
	Bidang              *string       `json:"bidang,omitempty"`
	JobRoles            []JobRoleResp `json:"job_roles"`
}

type AuthMeResp struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	ProfilePicture *string `json:"profile_picture"`
	SystemRoleName *string `json:"system_role_name"`
}

func (row *User) ToUserResp() UserResp {
	var roleName *string
	if row.SystemRole != nil {
		roleName = &row.SystemRole.Name
	}

	candidateLevel := row.CandidateLevel

	if roleName != nil && *roleName == "Candidate" {
		candidateLevel = nil
	}

	// Construct full public URL for profile picture
	var profilePicture *string
	if row.ProfilePicture != nil && *row.ProfilePicture != "" {
		pp := *row.ProfilePicture
		if len(pp) > 0 && pp[0] != 'h' {
			pp = fmt.Sprintf("%s/%s", os.Getenv("S3_PUBLIC_URL"), pp)
		}
		profilePicture = &pp
	}

	roles := []JobRoleResp{}
	for _, r := range row.JobRoles {
		roles = append(roles, r.ToJobRoleResp())
	}

	bidangStr := "-"
	if row.JobTitle != nil && row.JobTitle.Sector != nil {
		bidangStr = row.JobTitle.Sector.Name
	} else if row.AssignedRole != nil && row.AssignedRole.Sector != nil {
		bidangStr = row.AssignedRole.Sector.Name
	} else if len(row.JobRoles) > 0 && row.JobRoles[0].Sector != nil {
		bidangStr = row.JobRoles[0].Sector.Name
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
		Summary:             row.Summary,
		ProfilePicture:      profilePicture,
		CandidateLevel:      candidateLevel,
		RecruitmentStatusId: row.RecruitmentStatusId,
		UnavailableUntil:    row.UnavailableUntil,
		SystemRoleName:      roleName,
		AssignedRoleId:      row.AssignedRoleId,
		AccountStatus:       row.AccountStatus,
		JobTitleId:          row.JobTitleId,
		Bidang:              &bidangStr,
		JobRoles:            roles,
	}
}

func (row *User) ToAuthMeResp() AuthMeResp {
	var roleName *string
	if row.SystemRole != nil {
		roleName = &row.SystemRole.Name
	}
	return AuthMeResp{
		ID:             row.ID,
		Name:           row.Name,
		Email:          row.Email,
		ProfilePicture: row.ProfilePicture,
		SystemRoleName: roleName,
	}
}

// UserDetailResp bundles the standard user output along with a restricted CV link
type UserDetailResp struct {
	UserResp
	CV *CVPrivateResp `json:"cv"`
}
