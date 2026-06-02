package gorm_model

type FinalizeRecruitmentSnapshot struct {
	SubrequestID   string  `gorm:"column:subrequest_id"`
	JobRoleID      *string `gorm:"column:job_role_id"`
	JobRoleName    *string `gorm:"column:job_role_name"`
	ProjectName    string  `gorm:"column:project_name"`
	EmployeeUserID string  `gorm:"column:employee_user_id"`
	EmployeeName   string  `gorm:"column:employee_name"`
}
