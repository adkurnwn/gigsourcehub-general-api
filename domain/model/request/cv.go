package request_model

type ConfirmCVRequest struct {
	EditedData map[string]interface{} `json:"edited_data" binding:"required"`
}
