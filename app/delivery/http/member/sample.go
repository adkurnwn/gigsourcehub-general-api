package http_member

import (
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/gin-gonic/gin"
)

func (h *routeHandler) handleSampleRoute(prefixPath string) {
	// (optional). add prefix api version
	api := h.Route.Group(prefixPath)

	api.GET("/user/list", h.Middleware.Auth(), h.UserList)
	api.GET("/user/detail/:id", h.Middleware.Auth(), h.UserDetail)
	api.GET("/user/export", h.Middleware.Auth(), h.UserExport)
}

// List Member
//
//	@Summary		List member
//	@Description	Get list all member
//	@Tags			sample
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.List
//	@Router			/sample/user/list [get]
//
//	@Security		BearerAuth
func (r *routeHandler) UserList(c *gin.Context) {
	ctx := c.Request.Context()

	response := r.Usecase.SampleUserList(ctx, c.MustGet("token_data").(domain.JWTClaimUser), c.Request.URL.Query())
	c.JSON(response.Status, response)
}

// Detail Member
//
//	@Summary		Detail member
//	@Description	Get Detail member by id
//	@Tags			sample
//	@Accept			json
//	@Produce		json
//
//	@Param			id	path		string	true	"User ID"
//
//	@Success		200	{object}	gorm_model.UserResp
//	@Router			/sample/user/detail/{id} [get]
//
//	@Security		BearerAuth
func (r *routeHandler) UserDetail(c *gin.Context) {
	ctx := c.Request.Context()

	response := r.Usecase.SampleUserDetail(ctx, c.MustGet("token_data").(domain.JWTClaimUser), c.Param("id"))
	c.JSON(response.Status, response)
}

// Export Member
//
//	@Summary		Export member
//	@Description	Export data member
//	@Tags			sample
//	@Accept			json
//	@Produce		json
//
//	@Success		200	{object}	map[string]any	"base64 encoded"
//	@Router			/sample/user/export [get]
//
//	@Security		BearerAuth
func (r *routeHandler) UserExport(c *gin.Context) {
	ctx := c.Request.Context()

	response := r.Usecase.SampleUserExport(ctx, c.MustGet("token_data").(domain.JWTClaimUser), c.Request.URL.Query())
	c.JSON(response.Status, response)
}
