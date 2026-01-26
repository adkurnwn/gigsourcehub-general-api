package auth

import "github.com/adkurnwn/gigsourcehub-general-api/app/user"

type LoginResponse struct {
	Token string           `json:"token"`
	User  user.UserProfile `json:"user"`
}
