package request_model

type CreateSectorRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateSectorRequest struct {
	Name string `json:"name" binding:"required"`
}
