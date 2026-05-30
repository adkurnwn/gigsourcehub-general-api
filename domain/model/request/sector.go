package request_model

type CreateSectorRequest struct {
	Name string `json:"name"`
}

type UpdateSectorRequest struct {
	Name string `json:"name"`
	IsActive bool `json:"is_active"`
}
