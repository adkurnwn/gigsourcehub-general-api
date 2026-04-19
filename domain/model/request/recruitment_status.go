package request_model

type CreateRecruitmentStatusRequest struct {
	Name    string `json:"name" binding:"required"`
	HexCode string `json:"hex_code"`
	IsActive *bool  `json:"is_active"`
}

type UpdateRecruitmentStatusRequest struct {
	Name    string `json:"name" binding:"required"`
	HexCode string `json:"hex_code"`
	IsActive *bool  `json:"is_active"`
}
