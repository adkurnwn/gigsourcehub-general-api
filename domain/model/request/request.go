package request_model

type CreateSubrequestRequest struct {
	JobTitleID         *string  `json:"job_title_id"`
	MinYearsExperience int      `json:"min_years_experience" validate:"required,min=0"`
	TechStack          []string `json:"tech_stack"`
	Notes              *string  `json:"notes"`
}

type CreateRequestRequest struct {
	ProjectName string                    `json:"project_name" validate:"required"`
	DueDate     *string                   `json:"due_date"`
	Urgency     string                    `json:"urgency" validate:"required,oneof=Low Medium High"`
	Subrequests []CreateSubrequestRequest `json:"subrequests" validate:"required,min=1,dive"`
}
