package request_model

type CreateRecruitmentStatusRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateRecruitmentStatusRequest struct {
	Name string `json:"name" binding:"required"`
}
