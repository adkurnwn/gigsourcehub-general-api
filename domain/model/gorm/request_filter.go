package gorm_model

// RequestFilter holds optional filters for admin request listings.
type RequestFilter struct {
	Status      *string
	Urgency     *string
	Search      *string
	AdminUserID *string
	ProposedBy  *string
	AdminName   *string
}

func (f RequestFilter) GetDescription() string {
	var parts []string
	if f.Status != nil && *f.Status != "" {
		parts = append(parts, "Status: "+*f.Status)
	}
	if f.Urgency != nil && *f.Urgency != "" {
		parts = append(parts, "Urgency: "+*f.Urgency)
	}
	if f.Search != nil && *f.Search != "" {
		parts = append(parts, "Search: "+*f.Search)
	}
	if f.AdminUserID != nil && *f.AdminUserID != "" {
		parts = append(parts, "Admin User ID: "+*f.AdminUserID)
	}
	if f.ProposedBy != nil && *f.ProposedBy != "" {
		parts = append(parts, "Proposed By: "+*f.ProposedBy)
	}
	if f.AdminName != nil && *f.AdminName != "" {
		parts = append(parts, "Admin Name: "+*f.AdminName)
	}
	if len(parts) == 0 {
		return "None"
	}
	res := ""
	for i, p := range parts {
		if i > 0 {
			res += ", "
		}
		res += p
	}
	return res
}
