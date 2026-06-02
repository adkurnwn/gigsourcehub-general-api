package gorm_model

// RequestFilter holds optional filters for admin request listings.
type RequestFilter struct {
	Status  *string
	Urgency *string
	Search  *string
	AdminUserID *string
}
