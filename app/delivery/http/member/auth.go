package http_member

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

func (h *routeHandler) handleAuthRoute(prefixPath string) {
	// (optional). add prefix api version
	api := h.Route.Group(prefixPath)

	api.POST("/login", h.Login)
	api.POST("/register", h.Register)

	api.GET("/verify", h.VerifyAccount)
	api.POST("/forgot-password", h.ForgotPassword)
	api.POST("/reset-password", h.ResetPassword)

	api.GET("/me", h.Middleware.Auth(), h.GetMe)
}

// Verify Account
//
//	@Summary		Verify account
//	@Description	Verify account use token from email
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			token	query		string	true	"Verification Token"
//	@Success		200		{object}	response.Base
//	@Failure		400		{object}	response.Base
//	@Failure		500		{object}	response.Base
//	@Router			/auth/verify [get]
func (r *routeHandler) VerifyAccount(c *gin.Context) {
	ctx := c.Request.Context()
	token := c.Query("token")

	if token == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "token is required"))
		return
	}

	response := r.Usecase.VerifyAccount(ctx, token)
	c.JSON(response.Status, response)
}

// Forgot Password
//
//	@Summary		Forgot password
//	@Description	Request password reset link
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request_model.ForgotPasswordRequest	true	"Forgot Password Request"
//	@Success		200		{object}	response.Base
//	@Failure		400		{object}	response.Base
//	@Failure		500		{object}	response.Base
//	@Router			/auth/forgot-password [post]
func (r *routeHandler) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()

	payload := request_model.ForgotPasswordRequest{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid json data"))
		return
	}

	response := r.Usecase.ForgotPassword(ctx, payload)
	c.JSON(response.Status, response)
}

// Reset Password
//
//	@Summary		Reset password
//	@Description	Reset password use token from email
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request_model.ResetPasswordRequest	true	"Reset Password Request"
//	@Success		200		{object}	response.Base
//	@Failure		400		{object}	response.Base
//	@Failure		500		{object}	response.Base
//	@Router			/auth/reset-password [post]
func (r *routeHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	payload := request_model.ResetPasswordRequest{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid json data"))
		return
	}

	response := r.Usecase.ResetPassword(ctx, payload)
	c.JSON(response.Status, response)
}

// Login User
//
//	@Summary		Login as user
//	@Description	Login use email and password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request_model.LoginRequest	true	"Login Request"
//	@Success		200		{object}	gorm_model.UserResp
//	@Failure		400		{object}	response.Base
//	@Failure		404		{object}	response.Base
//	@Failure		500		{object}	response.Base
//	@Router			/auth/login [post]
func (r *routeHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	payload := request_model.LoginRequest{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid json data"))
		return
	}

	response := r.Usecase.Login(ctx, payload)
	c.JSON(response.Status, response)
}

// Register User
//
//	@Summary		Register user
//	@Description	Create a new user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request_model.RegisterRequest	true	"Register Request"
//	@Success		200		{object}	gorm_model.UserResp
//	@Failure		400		{object}	response.Base
//	@Failure		404		{object}	response.Base
//	@Failure		500		{object}	response.Base
//	@Router			/auth/register [post]
func (r *routeHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	payload := request_model.RegisterRequest{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid json data"))
		return
	}

	response := r.Usecase.Register(ctx, payload)
	c.JSON(response.Status, response)
}

// Detail User
//
//	@Summary		Detail user
//	@Description	Get detail current user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	gorm_model.AuthMeResp
//	@Failure		400	{object}	response.Base
//	@Failure		404	{object}	response.Base
//	@Failure		500	{object}	response.Base
//	@Router			/auth/me [get]
//
//	@Security		BearerAuth
func (r *routeHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	response := r.Usecase.GetMe(ctx, c.MustGet("token_data").(domain.JWTClaimUser))
	c.JSON(response.Status, response)
}
