package request_model

type UpdateCompanyProfileRequest struct {
	Address      *string `json:"address"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	FacebookURL  *string `json:"facebook_url"`
	InstagramURL *string `json:"instagram_url"`
	LinkedinURL  *string `json:"linkedin_url"`
	TwitterURL   *string `json:"twitter_url"`
}
