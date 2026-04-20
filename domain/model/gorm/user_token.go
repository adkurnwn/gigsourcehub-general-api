package gorm_model

import (
	"time"

	"github.com/google/uuid"
)

const (
	TokenTypeVerification  = "VERIFICATION"
	TokenTypePasswordReset = "PASSWORD_RESET"
)

type UserToken struct {
	ID        string    `gorm:"column:id;primarykey;type:uuid"`
	UserID    string    `gorm:"column:user_id;type:uuid;not null"`
	Type      string    `gorm:"column:type;type:varchar(50);not null"`
	Token     string    `gorm:"column:token;type:varchar(255);not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UserToken) TableName() string {
	return "user_tokens"
}

func NewUserToken(userID, tokenType, token string, expiresAt time.Time) UserToken {
	return UserToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      tokenType,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}
