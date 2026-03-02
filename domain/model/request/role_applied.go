package request_model

type CreateRoleAppliedRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
}

type UpdateRoleAppliedRequest struct {
	Name     string `json:"name" binding:"required"`
	SectorID string `json:"sector_id" binding:"required"`
}
