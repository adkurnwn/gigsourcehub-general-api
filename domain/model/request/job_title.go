package request_model

type CreateJobTitleRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
}

type UpdateJobTitleRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
}
