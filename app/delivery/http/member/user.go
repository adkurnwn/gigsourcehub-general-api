package http_member

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
	"strings"
)

func (h *routeHandler) handleUserRoute(path string) {
	userGroup := h.Route.Group(path)

	// candidates list: Admin & Superadmin
	userGroup.GET("/candidates", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin", "Employee"), h.FetchCandidates)
	userGroup.GET("/candidate-recruitment", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.FetchCandidateRecruitment)
	userGroup.GET("/candidate-bookmarked", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Employee"), h.FetchCandidateBookmarked)

	userGroup.GET("/candidates/export", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin", "Employee"), h.ExportCandidates)
	userGroup.GET("/candidate-recruitment/export", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.ExportCandidateRecruitment)
	userGroup.GET("/candidate-bookmarked/export", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Employee"), h.ExportCandidateBookmarked)

	// users list: Superadmin
	userGroup.GET("", h.Middleware.Auth(), h.Middleware.AuthSuperadmin(), h.FetchAllUsers)

	// user detail: Admin & Superadmin
	userGroup.GET("/:id", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin", "Employee"), h.FetchUserDetail)

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

	// patch user recruitment status and candidate level by admin
	userGroup.PATCH("/:id/recruitment-status", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.PatchUserRecruitmentStatus)

	// candidate declines recruitment for self
	userGroup.PATCH("/:id/decline", h.Middleware.Auth(), h.Middleware.AuthCandidate(), h.DeclineRecruitment)

	// admin confirms decline and clears candidate recruitment data
	userGroup.PATCH("/:id/decline-confirmation", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.ConfirmDeclineRecruitment)

	// admin stops onboarding, clears status, and marks the candidate pivot as stopped
	userGroup.PATCH("/:id/stop-onboarding", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.StopOnboarding)

	// cancel recruitment process for candidate by admin
	userGroup.PATCH("/:id/cancel-recruitment", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.CancelRecruitment)

	// finalize recruitment for candidate by admin
	userGroup.POST("/finalize-recruitment", h.Middleware.Auth(), h.Middleware.AuthAdmin(), h.FinalizeRecruitment)

	// get active subrequest for candidate by admin/superadmin
	userGroup.GET("/:id/active-subrequest", h.Middleware.Auth(), h.Middleware.AuthRole("Admin", "Superadmin", "Employee", "Candidate"), h.GetActiveSubrequest)

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

	filter := gorm_model.CandidateFilter{}
	if bidang := c.Query("bidang"); bidang != "" {
		filter.Bidang = strings.Split(bidang, ",")
	}
	if jobRoles := c.Query("job_roles"); jobRoles != "" {
		filter.JobRoles = strings.Split(jobRoles, ",")
	}
	if candidateLevel := c.Query("candidate_level"); candidateLevel != "" {
		filter.CandidateLevel = strings.Split(candidateLevel, ",")
	}

	res := h.Usecase.FetchUsers(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, searchPtr, &roleName, adminID, filter)
	c.JSON(res.Status, res)
}

// ExportCandidates
// @Summary Export Candidates
// @Description Export a list of candidate users
// @Tags Users
// @Accept json
// @Produce octet-stream
// @Param format query string true "Format (pdf, csv, xlsx)"
// @Param bidang query string false "Filter by Bidang (comma separated)"
// @Param job_roles query string false "Filter by Job Roles (comma separated)"
// @Param candidate_level query string false "Filter by Candidate Level (comma separated)"
// @Success 200 {file} file
// @Router /users/candidates/export [get]
// @Security BearerAuth
func (h *routeHandler) ExportCandidates(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	filter := gorm_model.CandidateFilter{}
	if bidang := c.Query("bidang"); bidang != "" {
		filter.Bidang = strings.Split(bidang, ",")
	}
	if jobRoles := c.Query("job_roles"); jobRoles != "" {
		filter.JobRoles = strings.Split(jobRoles, ",")
	}
	if candidateLevel := c.Query("candidate_level"); candidateLevel != "" {
		filter.CandidateLevel = strings.Split(candidateLevel, ",")
	}

	adminID := ""
	if claims, ok := c.Get("token_data"); ok {
		tokenData := claims.(domain.JWTClaimUser)
		adminID = tokenData.UserID
	}

	data, contentType, ext, err := h.Usecase.ExportCandidates(c.Request.Context(), filter, format, adminID)
	if err != nil || data == nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=candidates_export."+ext)
	c.Data(http.StatusOK, contentType, data)
}

// FetchCandidateRecruitment
// @Summary Fetch Candidate Recruitment
// @Description Fetch a paginated list of candidate users whose recruitment status is set
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/candidate-recruitment [get]
// @Security BearerAuth
func (h *routeHandler) FetchCandidateRecruitment(c *gin.Context) {
	pagination := helpers.GetPagination(c)

	filter := gorm_model.CandidateRecruitmentFilter{}
	if jobRoleName := c.Query("job_role_name"); jobRoleName != "" {
		filter.JobRoleName = strings.Split(jobRoleName, ",")
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter.ProjectName = strings.Split(projectName, ",")
	}
	if candidateLevel := c.Query("candidate_level"); candidateLevel != "" {
		filter.CandidateLevel = strings.Split(candidateLevel, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	res := h.Usecase.FetchCandidateRecruitment(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// ExportCandidateRecruitment
// @Summary Export Candidate Recruitment
// @Description Export a list of candidate recruitment
// @Tags Users
// @Accept json
// @Produce octet-stream
// @Param format query string true "Format (pdf, csv, xlsx)"
// @Param job_role_name query string false "Filter by Job Role Name (comma separated)"
// @Param project_name query string false "Filter by Project Name (comma separated)"
// @Param candidate_level query string false "Filter by Candidate Level (comma separated)"
// @Success 200 {file} file
// @Router /users/candidate-recruitment/export [get]
// @Security BearerAuth
func (h *routeHandler) ExportCandidateRecruitment(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	filter := gorm_model.CandidateRecruitmentFilter{}
	if jobRoleName := c.Query("job_role_name"); jobRoleName != "" {
		filter.JobRoleName = strings.Split(jobRoleName, ",")
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter.ProjectName = strings.Split(projectName, ",")
	}
	if candidateLevel := c.Query("candidate_level"); candidateLevel != "" {
		filter.CandidateLevel = strings.Split(candidateLevel, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	adminID := ""
	if claims, ok := c.Get("token_data"); ok {
		tokenData := claims.(domain.JWTClaimUser)
		adminID = tokenData.UserID
	}

	data, contentType, ext, err := h.Usecase.ExportCandidateRecruitment(c.Request.Context(), filter, format, adminID)
	if err != nil || data == nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=candidate_recruitment_export."+ext)
	c.Data(http.StatusOK, contentType, data)
}

// FetchCandidateBookmarked
// @Summary Fetch Candidate Bookmarked
// @Description Fetch a paginated list of candidates bookmarked by the authenticated admin
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/candidate-bookmarked [get]
// @Security BearerAuth
func (h *routeHandler) FetchCandidateBookmarked(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	claims, ok := c.Get("token_data")
	if !ok {
		res := response.Error(http.StatusUnauthorized, "Unauthorized")
		c.JSON(res.Status, res)
		return
	}

	tokenData := claims.(domain.JWTClaimUser)

	filter := gorm_model.CandidateFilter{}
	if bidang := c.Query("bidang"); bidang != "" {
		filter.Bidang = strings.Split(bidang, ",")
	}
	if jobRoles := c.Query("job_roles"); jobRoles != "" {
		filter.JobRoles = strings.Split(jobRoles, ",")
	}
	if candidateLevel := c.Query("candidate_level"); candidateLevel != "" {
		filter.CandidateLevel = strings.Split(candidateLevel, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	res := h.Usecase.FetchCandidateBookmarked(c.Request.Context(), tokenData.UserID, pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// ExportCandidateBookmarked
// @Summary Export Candidate Bookmarked
// @Description Export a list of candidate bookmarked
// @Tags Users
// @Accept json
// @Produce octet-stream
// @Param format query string true "Format (pdf, csv, xlsx)"
// @Param bidang query string false "Filter by Bidang (comma separated)"
// @Param job_roles query string false "Filter by Job Roles (comma separated)"
// @Param candidate_level query string false "Filter by Candidate Level (comma separated)"
// @Success 200 {file} file
// @Router /users/candidate-bookmarked/export [get]
// @Security BearerAuth
func (h *routeHandler) ExportCandidateBookmarked(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	claims, ok := c.Get("token_data")
	if !ok {
		res := response.Error(http.StatusUnauthorized, "Unauthorized")
		c.JSON(res.Status, res)
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	filter := gorm_model.CandidateFilter{}
	if bidang := c.Query("bidang"); bidang != "" {
		filter.Bidang = strings.Split(bidang, ",")
	}
	if jobRoles := c.Query("job_roles"); jobRoles != "" {
		filter.JobRoles = strings.Split(jobRoles, ",")
	}
	if candidateLevel := c.Query("candidate_level"); candidateLevel != "" {
		filter.CandidateLevel = strings.Split(candidateLevel, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	data, contentType, ext, err := h.Usecase.ExportCandidateBookmarked(c.Request.Context(), tokenData.UserID, filter, format)
	if err != nil || data == nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=candidate_bookmarked_export."+ext)
	c.Data(http.StatusOK, contentType, data)
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

	res := h.Usecase.FetchUsers(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, searchPtr, rolePtr, nil, gorm_model.CandidateFilter{})
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

// PatchUserRecruitmentStatus
// @Summary Patch User Recruitment Status and Candidate Level
// @Description Update a user's recruitment status and candidate level by superadmin
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param req body request_model.PatchUserRecruitmentStatusRequest true "Recruitment status patch object"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/{id}/recruitment-status [patch]
// @Security BearerAuth
func (h *routeHandler) PatchUserRecruitmentStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	var req request_model.PatchUserRecruitmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := response.Error(http.StatusBadRequest, "Invalid request body")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.PatchUserRecruitmentStatus(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// DeclineRecruitment
// @Summary Decline Recruitment
// @Description Mark the authenticated candidate's recruitment status as Decline
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
// @Router /users/{id}/decline [patch]
// @Security BearerAuth
func (h *routeHandler) DeclineRecruitment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	claims, ok := c.Get("token_data")
	if !ok {
		res := response.Error(http.StatusUnauthorized, "Unauthorized")
		c.JSON(res.Status, res)
		return
	}

	tokenData := claims.(domain.JWTClaimUser)
	if tokenData.UserID != id {
		res := response.Error(http.StatusForbidden, "You can only decline your own recruitment status")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.DeclineRecruitment(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// ConfirmDeclineRecruitment
// @Summary Confirm Decline Recruitment
// @Description Reset a declined candidate recruitment status to null and soft delete subrequest candidate rows
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
// @Router /users/{id}/decline-confirmation [patch]
// @Security BearerAuth
func (h *routeHandler) ConfirmDeclineRecruitment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.ConfirmDeclineRecruitment(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// StopOnboarding
// @Summary Stop Onboarding
// @Description Stop an accepted candidate onboarding, clear recruitment status, soft delete the pivot row, and send an apology email
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
// @Router /users/{id}/stop-onboarding [patch]
// @Security BearerAuth
func (h *routeHandler) StopOnboarding(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.StopOnboarding(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// CancelRecruitment
// @Summary Cancel Recruitment Process
// @Description Cancel a candidate recruitment process and set status to Null
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
// @Router /users/{id}/cancel-recruitment [patch]
// @Security BearerAuth
func (h *routeHandler) CancelRecruitment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.CancelRecruitment(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// FinalizeRecruitment
// @Summary Finalize Recruitment
// @Description Finalize a candidate recruitment and create onboarding history
// @Tags Users
// @Accept json
// @Produce json
// @Param request body request_model.FinalizeRecruitmentRequest true "Finalize Recruitment"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /users/finalize-recruitment [post]
// @Security BearerAuth
func (h *routeHandler) FinalizeRecruitment(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	var req request_model.FinalizeRecruitmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := response.Error(http.StatusBadRequest, "Invalid request body")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.FinalizeRecruitment(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}

// GetActiveSubrequest
// @Summary Get Active Subrequest
// @Description Fetch the active subrequest (project name and job role) for a candidate
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
// @Router /users/{id}/active-subrequest [get]
// @Security BearerAuth
func (h *routeHandler) GetActiveSubrequest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		res := response.Error(http.StatusBadRequest, "Invalid ID parameter")
		c.JSON(res.Status, res)
		return
	}

	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	// Fetch caller's details to verify candidate restriction
	callerBase := h.Usecase.FetchUserDetail(c.Request.Context(), tokenData.UserID)
	if callerBase.Status != http.StatusOK {
		c.JSON(callerBase.Status, callerBase)
		return
	}

	var roleName string
	if detail, ok := callerBase.Data.(gorm_model.UserDetailResp); ok && detail.SystemRoleName != nil {
		roleName = *detail.SystemRoleName
	} else if detailPtr, ok := callerBase.Data.(*gorm_model.UserDetailResp); ok && detailPtr.SystemRoleName != nil {
		roleName = *detailPtr.SystemRoleName
	}

	if roleName == "Candidate" && tokenData.UserID != id {
		res := response.Error(http.StatusForbidden, "You can only view your own active subrequest")
		c.JSON(res.Status, res)
		return
	}

	res := h.Usecase.GetActiveSubrequest(c.Request.Context(), id)

	// Clear project details if user is Candidate
	if res.Status == http.StatusOK && roleName == "Candidate" {
		if info, ok := res.Data.(*gorm_model.ActiveSubrequestInfo); ok && info != nil {
			// Copy to a new struct so we don't modify repository caches/shared pointers directly if any
			infoCopy := *info
			infoCopy.ProjectName = ""
			res.Data = &infoCopy
		}
	}

	c.JSON(res.Status, res)
}
