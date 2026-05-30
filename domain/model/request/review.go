package request_model

type CreateReviewRequest struct {
	OnboardHistoryID           string  `json:"onboard_history_id" binding:"required"`
	FinalRecommendation        *string `json:"final_recommendation"`
	Notes                      *string `json:"notes"`
	WorkQuality                []int   `json:"work_quality" binding:"required"`
	Timeliness                 []int   `json:"timeliness" binding:"required"`
	CommunicationCollaboration []int   `json:"communication_collaboration" binding:"required"`
	ProblemSolvingInitiative   []int   `json:"problem_solving_initiative" binding:"required"`
}
