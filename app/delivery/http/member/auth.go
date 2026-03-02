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

	api.GET("/me", h.Middleware.Auth(), h.GetMe)
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
