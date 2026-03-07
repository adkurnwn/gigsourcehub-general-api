package request_model

type CreateBookmarkRequest struct {
	CandidateID string `json:"candidate_id" binding:"required"`
}
