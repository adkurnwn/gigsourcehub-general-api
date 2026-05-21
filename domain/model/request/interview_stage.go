package request_model

type CreateInterviewStageRequest struct {
	Name     string `json:"name" binding:"required"`
	HexCode  string `json:"hex_code"`
	IsActive *bool  `json:"is_active"`
}

type UpdateInterviewStageRequest struct {
	Name     string `json:"name" binding:"required"`
	HexCode  string `json:"hex_code"`
	IsActive *bool  `json:"is_active"`
}