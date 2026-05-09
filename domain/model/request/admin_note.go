package request_model

type CreateAdminNoteRequest struct {
	Content string `json:"content" binding:"required"`
}
