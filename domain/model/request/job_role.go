package request_model

type CreateJobRoleRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
}

type UpdateJobRoleRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
}
