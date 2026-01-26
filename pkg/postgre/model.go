package postgre_pkg

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/auth"
	"github.com/adkurnwn/gigsourcehub-general-api/app/user"
)

func GetAllModels() []interface{} {
	return []interface{}{
		&user.User{},
		&auth.PasswordResetToken{},
	}
}
