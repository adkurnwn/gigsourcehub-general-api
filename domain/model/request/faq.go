package request_model

type CreateFAQRequest struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer"   binding:"required"`
}

type UpdateFAQRequest struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer"   binding:"required"`
}

type ReviewApprovalRequest struct {
	RejectedReason *string `json:"rejected_reason"`
}
