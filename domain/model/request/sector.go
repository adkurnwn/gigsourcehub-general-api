package request_model

type CreateSectorRequest struct {
	Name string `json:"name"`
	HexCode string `json:"hex_code"`
}

type UpdateSectorRequest struct {
	Name string `json:"name"`
	HexCode string `json:"hex_code"`
	IsActive bool `json:"is_active"`
}
