package usecase_member

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	jwt_helper "github.com/adkurnwn/gigsourcehub-general-api/helpers/jsonwebtoken"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// add minimal implementations so appUsecase implements domain.MemberAppUsecase

func (u *appUsecase) SampleUserList(ctx context.Context, claim domain.JWTClaimUser, query url.Values) response.Base {
	return response.Error(http.StatusNotImplemented, "not implemented")
}

func (u *appUsecase) SampleUserDetail(ctx context.Context, claim domain.JWTClaimUser, id string) response.Base {
	return response.Error(http.StatusNotImplemented, "not implemented")
}

func (u *appUsecase) SampleUserExport(ctx context.Context, claim domain.JWTClaimUser, query url.Values) response.Base {
	return response.Error(http.StatusNotImplemented, "not implemented")
}

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
		return response.Error(http.StatusBadRequest, "user not found")
	}

	// check password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
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

	newUser := gorm_model.User{
		ID:        uuid.New().String(),
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = u.gormDbRepo.CreateUser(ctx, &newUser)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	return response.Success(newUser.ToUserResp())
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

	return response.Success(user.ToUserResp())
}
