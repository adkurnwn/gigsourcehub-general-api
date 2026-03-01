package http_member

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

func (h *routeHandler) handleUserRoute(path string) {
	userGroup := h.Route.Group(path)

	// Candidates list: Only for Admin & Superadmin
	userGroup.GET("/candidates", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin"), h.FetchCandidates)

	// Standard users list: Only Superadmin
	userGroup.GET("", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.FetchAllUsers)

	// User detail: For Admin & Superadmin
	userGroup.GET("/:id", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin"), h.FetchUserDetail)
}

// FetchCandidates
// @Summary Fetch Candidates
// @Description Fetch a paginated list of Candidate users
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/candidates [get]
// @Security BearerAuth
func (h *routeHandler) FetchCandidates(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	roleName := "Candidate"

	res := h.Usecase.FetchUsers(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, &roleName)
	c.JSON(res.Status, res)
}

// FetchAllUsers
// @Summary Fetch All Users
// @Description Fetch a paginated list of all users
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users [get]
// @Security BearerAuth
func (h *routeHandler) FetchAllUsers(c *gin.Context) {
	pagination := helpers.GetPagination(c)

	res := h.Usecase.FetchUsers(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, nil)
	c.JSON(res.Status, res)
}

// FetchUserDetail
// @Summary Fetch User Detail
// @Description Fetch detail of a specific user including their CV URL
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/{id} [get]
// @Security BearerAuth
func (h *routeHandler) FetchUserDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.FetchUserDetail(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
