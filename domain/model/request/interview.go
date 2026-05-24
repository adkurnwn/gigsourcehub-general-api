package request_model

type CreateInterviewRequest struct {
	CandidateUserID *string `json:"candidate_user_id"`
	SubrequestID    *string `json:"subrequest_id"`
	StageID         string  `json:"stage_id" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Description     *string `json:"description"`
	ScheduledAt     *string `json:"scheduled_at" binding:"required"` // RFC3339
	Method          string  `json:"method" binding:"required"`       // Online | Offline
	MeetingLocation *string `json:"meeting_location"`
	MeetingLink     *string `json:"meeting_link"`
}

type UpdateInterviewRequest struct {
	StageID         string  `json:"stage_id" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Description     *string `json:"description"`
	ScheduledAt     *string `json:"scheduled_at" binding:"required"` // RFC3339
	Method          string  `json:"method" binding:"required"`       // Online | Offline
	MeetingLocation *string `json:"meeting_location"`
	MeetingLink     *string `json:"meeting_link"`
}

type PatchInterviewStageRequest struct {
	InterviewID string  `json:"interview_id" binding:"required"`
	StageID     string  `json:"stage_id" binding:"required"`
	Status      *string `json:"status"`
}

type PatchInterviewStatusRequest struct {
	InterviewID string `json:"interview_id" binding:"required"`
	Status      string `json:"status" binding:"required"`
}