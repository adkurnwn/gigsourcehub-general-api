package request_model

type CreateUserBySuperadminRequest struct {
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Password       string  `json:"password"`
	AssignedRoleId *string `json:"assigned_role_id"`
	SystemRoleId   *string `json:"system_role_id"`
	JobTitleId     *string `json:"job_title_id"`
	AccountStatus  *string `json:"account_status"`
}

type EditUserBySuperadminRequest struct {
	Name           *string `json:"name"`
	AssignedRoleId *string `json:"assigned_role_id"`
	SystemRoleId   *string `json:"system_role_id"`
	JobTitleId     *string `json:"job_title_id"`
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
	Summary            *string  `json:"summary"`
	JobRoleIds         []string `json:"job_role_ids"`
	AvailabilityStatus *string  `json:"availability_status"`
	UnavailableUntil   *string  `json:"unavailable_until"` // Expecting YYYY-MM-DD
}

type PatchUserRecruitmentStatusRequest struct {
	CandidateLevel      *string `json:"candidate_level"`
	RecruitmentStatusId *string `json:"recruitment_status_id"`
}

type FinalizeRecruitmentRequest struct {
	CandidateUserID string  `json:"candidate_user_id"`
	SubrequestID    string  `json:"subrequest_id"`
	StartDate       *string `json:"start_date"` // format: "2006-01-02"
	EndDate         *string `json:"end_date"`   // format: "2006-01-02"
}

type DeleteAccountRequest struct {
	Password string `json:"password"`
}
