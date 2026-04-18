package request_model

type CreateUserBySuperadminRequest struct {
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Password       string  `json:"password"`
	AssignedRoleId *string `json:"assigned_role_id"`
	SystemRoleId   *string `json:"system_role_id"`
	AccountStatus  *string `json:"account_status"`
}

type EditUserBySuperadminRequest struct {
	Name           *string `json:"name"`
	AssignedRoleId *string `json:"assigned_role_id"`
	AccountStatus  *string `json:"account_status"`
}

type UpdateProfileRequest struct {
	Name             *string  `json:"name"`
	Birthdate        *string  `json:"birthdate"` // Expecting YYYY-MM-DD
	SchoolUniversity *string  `json:"school_university"`
	Major            *string  `json:"major"`
	Gpa              *float64 `json:"gpa"`
	PhoneNumber      *string  `json:"phone_number"`
	PortofolioLink   *string  `json:"portofolio_link"`
	KabupatenKotaId  *string  `json:"kabupaten_kota_id"`
	YearsExperience  *int     `json:"years_experience"`
	TechStack        []string `json:"tech_stack"`
	Summary          *string  `json:"summary"`
	JobRoleIds       []string `json:"job_role_ids"`
}
