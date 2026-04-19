package http_member

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

func (h *routeHandler) handleUserRoute(path string) {
	userGroup := h.Route.Group(path)

	// candidates list: Admin & Superadmin
	userGroup.GET("/candidates", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin"), h.FetchCandidates)

	// users list: Superadmin
	userGroup.GET("", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.FetchAllUsers)

	// user detail: Admin & Superadmin
	userGroup.GET("/:id", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin"), h.FetchUserDetail)

	// create user by superadmin
	userGroup.POST("", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.CreateUserBySuperadmin)

	// edit user by id user: Superadmin
	userGroup.PUT("/:id", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.EditUserBySuperadmin)

	// block user by id user: Superadmin
	userGroup.PATCH("/:id/block", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.BlockUserBySuperadmin)

	// disable user by id user: Superadmin
	userGroup.PATCH("/:id/disable", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.DisableUserBySuperadmin)

	// activate user by superadmin
	userGroup.PATCH("/:id/activate", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.ActivateUserBySuperadmin)

	// get profile picture by id
	userGroup.GET("/:id/profile-picture", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin"), h.FetchUserThumb)
}

// FetchUserThumb
// @Summary Fetch User Profile Picture
// @Description Fetch the profile picture URL of a specific user with _thumb suffix
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/{id}/profile-picture [get]
// @Security BearerAuth
func (h *routeHandler) FetchUserThumb(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.FetchUserThumb(c.Request.Context(), id)
	c.JSON(res.Status, res)
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

	// Extract admin ID from JWT claims for bookmark lookup
	var adminID *string
	if claims, ok := c.Get("token_data"); ok {
		tokenData := claims.(domain.JWTClaimUser)
		adminID = &tokenData.UserID
	}

	search := c.Query("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	res := h.Usecase.FetchUsers(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, searchPtr, &roleName, adminID)
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
// @Param role query string false "System role filtering (e.g., Admin, Employee, Candidate)"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users [get]
// @Security BearerAuth
func (h *routeHandler) FetchAllUsers(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	role := c.Query("role")

	var rolePtr *string
	if role != "" {
		rolePtr = &role
	}

	search := c.Query("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	res := h.Usecase.FetchUsers(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, searchPtr, rolePtr, nil)
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

// CreateUserBySuperadmin
// @Summary Create User By Superadmin
// @Description Create a new user by superadmin
// @Tags Users
// @Accept json
// @Produce json
// @Param user body request_model.CreateUserBySuperadminRequest true "User object"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users [post]
// @Security BearerAuth
func (h *routeHandler) CreateUserBySuperadmin(c *gin.Context) {
	var req request_model.CreateUserBySuperadminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := response.Error(http.StatusBadRequest, "Invalid request body")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.CreateBySuperadmin(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// EditUserBySuperadmin
// @Summary Edit User By Superadmin
// @Description Edit an existing user (Name, AssignedRoleId, AccountStatus) by superadmin
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body request_model.EditUserBySuperadminRequest true "User edit object"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/{id} [put]
// @Security BearerAuth
func (h *routeHandler) EditUserBySuperadmin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	var req request_model.EditUserBySuperadminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := response.Error(http.StatusBadRequest, "Invalid request body")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.EditUserBySuperadmin(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// BlockUserBySuperadmin
// @Summary Block User By Superadmin
// @Description Instantly block a user by setting AccountStatus to "Blocked"
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
// @Router /users/{id}/block [patch]
// @Security BearerAuth
func (h *routeHandler) BlockUserBySuperadmin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.BlockUserBySuperadmin(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// DisableUserBySuperadmin
// @Summary Disable User By Superadmin
// @Description Instantly disable a user by setting AccountStatus to "Inactive"
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
// @Router /users/{id}/disable [patch]
// @Security BearerAuth
func (h *routeHandler) DisableUserBySuperadmin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.DisableUserBySuperadmin(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// ActivateUserBySuperadmin
// @Summary Activate User By Superadmin
// @Description Instantly activate a user by setting AccountStatus to "Inactive"
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
// @Router /users/{id}/activate [patch]
// @Security BearerAuth
func (h *routeHandler) ActivateUserBySuperadmin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.ActivateUserBySuperadmin(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
