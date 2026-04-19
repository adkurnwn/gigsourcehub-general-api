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
		AccountStatus: &activeStatus,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = u.gormDbRepo.CreateUser(ctx, &newUser)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	// generate token
	tokenString, err := jwt_helper.GenerateJWTToken(
		jwt_helper.GetJwtCredential().Member,
		domain.JWTClaimUser{
			UserID: newUser.ID,
		},
	)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "failed to generate token")
	}

	return response.Success(map[string]interface{}{
		"name":  newUser.Name,
		"email": newUser.Email,
		"token": tokenString,
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
