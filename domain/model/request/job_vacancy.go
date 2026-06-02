package request_model

type CreateJobVacancyRequest struct {
	SubrequestID    string  `json:"subrequest_id"    binding:"required"`
	Name            string  `json:"name"             binding:"required"`
	TakedownDate    *string `json:"takedown_date"`    // format: "2006-01-02"
	FulfillmentDate *string `json:"fulfillment_date"` // format: "2006-01-02"
	Schema          *string `json:"schema"`           // ONSITE | REMOTE | HYBRID
	Status          *string `json:"status"`           // DRAFT | ARCHIVED | PUBLISHED
	Description     *string `json:"description"`
	Overview        *string `json:"overview"`
}

type UpdateJobVacancyRequest struct {
	SubrequestID    string  `json:"subrequest_id"    binding:"required"`
	Name            string  `json:"name"             binding:"required"`
	TakedownDate    *string `json:"takedown_date"`    // format: "2006-01-02"
	FulfillmentDate *string `json:"fulfillment_date"` // format: "2006-01-02"
	Schema          *string `json:"schema"`           // ONSITE | REMOTE | HYBRID
	Status          *string `json:"status"`           // DRAFT | ARCHIVED | PUBLISHED
	Description     *string `json:"description"`
	Overview        *string `json:"overview"`
}
