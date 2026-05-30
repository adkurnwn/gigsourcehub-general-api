package request_model

type CreateJobTitleRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

type UpdateJobTitleRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
	IsActive *bool  `json:"is_active"`
}
