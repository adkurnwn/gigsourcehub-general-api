package request_model

type SearchRequest struct {
	Query string `json:"query" binding:"required"`
}
