package usecase_member

import (
	"context"
	"net/http"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	jwt_helper "github.com/adkurnwn/gigsourcehub-general-api/helpers/jsonwebtoken"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// add minimal implementations so appUsecase implements domain.MemberAppUsecase

func (u *appUsecase) Login(ctx context.Context, payload request_model.LoginRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	// validating request
	if payload.Email == "" {
		errValidation["email"] = "email field is required"
	}

	if payload.Password == "" {
		errValidation["password"] = "password field is required"
	}

	if len(errValidation) > 0 {
		return response.ErrorValidation(errValidation, "error validation")
	}

	// check the db
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		Email: &payload.Email,
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	if user == nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Login", "Authentication", payload.Email, nil, false)
		return response.Error(http.StatusBadRequest, "user not found")
	}

	// check password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		helpers.LogActivity(ctx, u.gormDbRepo, "Login", "Authentication", payload.Email, nil, false)
		return response.Error(http.StatusBadRequest, "Wrong password")
	}

	// check account status
	if user.AccountStatus != nil {
		if *user.AccountStatus == "Blocked" {
			return response.Error(http.StatusForbidden, "Your account has been blocked.")
		}
		if *user.AccountStatus == "Inactive" {
			return response.Error(http.StatusForbidden, "Your account is inactive.")
		}
	}

	// check verification status
	if user.VerifiedAt == nil {
		return response.Error(http.StatusForbidden, "Account not verified. Please check your email to verify your account.")
	}

	// generate token
	tokenString, err := jwt_helper.GenerateJWTToken(
		jwt_helper.GetJwtCredential().Member,
		domain.JWTClaimUser{
			UserID: user.ID,
		},
	)
	if err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Login", "Authentication", user.Email, nil, true)

	// Manually ensure the Role Name is fetched so ToUserResp can properly suppress Candidate fields
	roleName, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, user.ID)
	if errRole == nil && roleName != "" {
		user.SystemRole = &gorm_model.SystemRole{
			Name: roleName,
		}
	}

	return response.Success(map[string]interface{}{
		"user":  user.ToUserResp(),
		"token": tokenString,
	})
}

func (u *appUsecase) Register(ctx context.Context, payload request_model.RegisterRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	// validating request
	if payload.Name == "" {
		errValidation["name"] = "name field is required"
	}

	if payload.Email == "" {
		errValidation["email"] = "email field is required"
	} else if !helpers.IsValidEmail(payload.Email) {
		errValidation["email"] = "invalid email format"
	}

	if payload.Password == "" {
		errValidation["password"] = "password field is required"
	}

	if len(errValidation) > 0 {
		return response.ErrorValidation(errValidation, "error validation")
	}

	// check the db
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		Email: &payload.Email,
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user != nil {
		return response.Error(http.StatusBadRequest, "email already taken")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	// Fetch default role "Candidate"
	var candidateRole gorm_model.SystemRole
	var systemRoleID *string
	if err := u.gormDbRepo.GetDB().Where("name = ?", "Candidate").First(&candidateRole).Error; err == nil {
		systemRoleID = &candidateRole.ID
	}

	activeStatus := "Active"

	newUser := gorm_model.User{
		ID:            uuid.New().String(),
		Name:          payload.Name,
		Email:         payload.Email,
		Password:      string(hashedPassword),
		SystemRoleId:  systemRoleID,
		SystemRole:    &candidateRole,
		AccountStatus: &activeStatus,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = u.gormDbRepo.CreateUser(ctx, &newUser)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	// Create Verification Token
	verifyToken := uuid.New().String()
	userToken := gorm_model.NewUserToken(
		newUser.ID,
		gorm_model.TokenTypeVerification,
		verifyToken,
		time.Now().Add(24*time.Hour),
	)
	_ = u.gormDbRepo.CreateUserToken(ctx, &userToken)

	// Send Verification Email
	_ = u.mailerRepo.SendVerificationEmail(newUser.Email, newUser.Name, verifyToken)

	return response.Success(map[string]interface{}{
		"user":    newUser.ToUserResp(),
		"message": "Registration successful! Please check your email to verify your account.",
	})
}

func (u *appUsecase) GetMe(ctx context.Context, claim domain.JWTClaimUser) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userID := claim.UserID

	// check the db
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{
			ID: userID,
		},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	if user == nil {
		return response.Error(http.StatusBadRequest, "user not found")
	}

	// Manually ensure the Role Name is fetched so ToUserResp can properly suppress Candidate fields
	roleName, err := u.gormDbRepo.GetRoleNameByUserID(ctx, userID)
	if err == nil && roleName != "" {
		user.SystemRole = &gorm_model.SystemRole{
			Name: roleName,
		}
	}

	return response.Success(user.ToAuthMeResp())
}

func (u *appUsecase) VerifyAccount(ctx context.Context, token string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Get Token
	userToken, err := u.gormDbRepo.GetUserToken(ctx, token, gorm_model.TokenTypeVerification)
	if err != nil {
		return response.Error(http.StatusBadRequest, "Invalid or expired verification token")
	}

	// 2. Check Expiration
	if time.Now().After(userToken.ExpiresAt) {
		_ = u.gormDbRepo.DeleteUserToken(ctx, userToken.ID)
		return response.Error(http.StatusBadRequest, "Verification token has expired")
	}

	// 3. Update User Status
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{
			ID: userToken.UserID,
		},
	})
	if err != nil || user == nil {
		return response.Error(http.StatusInternalServerError, "User not found during verification")
	}

	activeStatus := "Active"
	now := time.Now()
	user.AccountStatus = &activeStatus
	user.VerifiedAt = &now

	err = u.gormDbRepo.UpdateUser(ctx, user)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to update user status")
	}

	// 4. Delete Token
	_ = u.gormDbRepo.DeleteUserToken(ctx, userToken.ID)

	return response.Success(map[string]string{
		"message": "Account verified successfully! You can now log in.",
	})
}

func (u *appUsecase) ForgotPassword(ctx context.Context, req request_model.ForgotPasswordRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Find User
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		Email: &req.Email,
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	if user == nil {
		// Return success even if not found for security (prevent email enumeration)
		return response.Success(map[string]string{
			"message": "If that email matches an account, we have sent a password reset link.",
		})
	}

	// 2. Create Reset Token (Delete old ones first)
	_ = u.gormDbRepo.DeleteUserTokensByUserID(ctx, user.ID, gorm_model.TokenTypePasswordReset)

	resetToken := uuid.New().String()
	userToken := gorm_model.NewUserToken(
		user.ID,
		gorm_model.TokenTypePasswordReset,
		resetToken,
		time.Now().Add(1*time.Hour),
	)
	err = u.gormDbRepo.CreateUserToken(ctx, &userToken)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to create reset token")
	}

	// 3. Send Email
	_ = u.mailerRepo.SendResetPasswordEmail(user.Email, user.Name, resetToken)

	return response.Success(map[string]string{
		"message": "If that email matches an account, we have sent a password reset link.",
	})
}

func (u *appUsecase) ResetPassword(ctx context.Context, req request_model.ResetPasswordRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Get Token
	userToken, err := u.gormDbRepo.GetUserToken(ctx, req.Token, gorm_model.TokenTypePasswordReset)
	if err != nil {
		return response.Error(http.StatusBadRequest, "Invalid or expired reset token")
	}

	// 2. Check Expiration
	if time.Now().After(userToken.ExpiresAt) {
		_ = u.gormDbRepo.DeleteUserToken(ctx, userToken.ID)
		return response.Error(http.StatusBadRequest, "Reset token has expired")
	}

	// 3. Update Password
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{
			ID: userToken.UserID,
		},
	})
	if err != nil || user == nil {
		return response.Error(http.StatusInternalServerError, "User not found")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	err = u.gormDbRepo.UpdateUser(ctx, user)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to update password")
	}

	// 4. Delete Token
	_ = u.gormDbRepo.DeleteUserTokensByUserID(ctx, user.ID, gorm_model.TokenTypePasswordReset)

	return response.Success(map[string]string{
		"message": "Password reset successfully! You can now log in with your new password.",
	})
}

func (u *appUsecase) Logout(ctx context.Context, claim domain.JWTClaimUser) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userID := claim.UserID

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{
			ID: userID,
		},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	if user == nil {
		return response.Error(http.StatusBadRequest, "user not found")
	}

	if claim.ID != "" {
		expiresAt := time.Now().Add(24 * time.Hour) // Fallback expires at
		if claim.ExpiresAt != nil {
			expiresAt = claim.ExpiresAt.Time
		}

		blacklistedToken := gorm_model.NewUserToken(
			userID,
			gorm_model.TokenTypeBlacklist,
			claim.ID,
			expiresAt,
		)
		err = u.gormDbRepo.CreateUserToken(ctx, &blacklistedToken)
		if err != nil {
			return response.Error(http.StatusInternalServerError, "Failed to logout session: "+err.Error())
		}
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Logout", "Authentication", user.Email, nil, true)

	return response.Success(map[string]string{
		"message": "Logout successful",
	})
}
