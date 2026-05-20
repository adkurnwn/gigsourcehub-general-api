package request_model

type CreateSubrequestRequest struct {
	JobRoleID *string  `json:"job_role_id"`
	Level     string   `json:"level" validate:"required,oneof=Junior Middle Senior"`
	TechStack []string `json:"tech_stack"`
	Notes     *string  `json:"notes"`
	Overview  *string  `json:"overview"`
}

type CreateRequestRequest struct {
	ProjectName     string                    `json:"project_name" validate:"required"`
	ProjectDuration *string                   `json:"project_duration"`
	DueDate         *string                   `json:"due_date"`
	Urgency         string                    `json:"urgency" validate:"required,oneof=Low Medium High"`
	Subrequests     []CreateSubrequestRequest `json:"subrequests" validate:"required,min=1,dive"`
}

type UpdateSubrequestRequest struct {
	JobRoleID *string  `json:"job_role_id"`
	Level     string   `json:"level" validate:"required,oneof=Junior Middle Senior"`
	TechStack []string `json:"tech_stack"`
	Notes     *string  `json:"notes"`
	Overview  *string  `json:"overview"`
}

type AssignCandidateToSubrequestRequest struct {
	CandidateUserID string `json:"candidate_user_id" validate:"required"`
}

type UpdateRequestRequest struct {
	ProjectName     string  `json:"project_name" validate:"required"`
	ProjectDuration *string `json:"project_duration"`
	DueDate         *string `json:"due_date"`
	Urgency         string  `json:"urgency" validate:"required,oneof=Low Medium High"`
}

type RejectRequestRequest struct {
	RejectedReason string `json:"rejected_reason" validate:"required"`
}
