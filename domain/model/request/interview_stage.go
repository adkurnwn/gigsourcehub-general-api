package request_model

type CreateInterviewStageRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateInterviewStageRequest struct {
	Name string `json:"name" binding:"required"`
}