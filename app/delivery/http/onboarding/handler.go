package http_onboarding

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
	"strings"
)

type routeHandler struct {
	Usecase    domain.MemberAppUsecase
	Middleware middleware.Middleware
}

func NewOnboardingHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.MemberAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Middleware: mdl,
	}

	onboarding := r.Group("/onboarding", mdl.Auth())
	onboarding.GET("/active-team", mdl.AuthRole("Admin", "Employee"), handler.FetchActiveTeam)
	onboarding.GET("/history", mdl.AuthRole("Admin", "Employee"), handler.FetchHistory)
	onboarding.GET("/:id", mdl.AuthRole("Admin", "Employee", "Candidate"), handler.FetchByCandidate)

	adminOnboarding := r.Group("/onboarding", mdl.Auth(), mdl.AuthAdmin())
	adminOnboarding.GET("/active", handler.FetchActive)
	adminOnboarding.GET("/archive", handler.FetchArchive)
	adminOnboarding.GET("/active/export", handler.ExportActive)
	adminOnboarding.GET("/archive/export", handler.ExportArchive)
}

// FetchByCandidate
// @Summary Get Onboarding History By Candidate
// @Description Fetch onboarding history for a specific candidate
// @Tags Onboarding
// @Accept json
// @Produce json
// @Param id path string true "Candidate User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /onboarding/{id} [get]
// @Security BearerAuth
func (h *routeHandler) FetchByCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
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

	if roleName == "Candidate" && tokenData.UserID != candidateID {
		res := response.Error(http.StatusForbidden, "You can only view your own onboarding history")
		c.JSON(res.Status, res)
		return
	}

	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchOnboardingByCandidate(c.Request.Context(), candidateID, pagination.Page, pagination.Limit, pagination.Cursor)

	// Filter out project details if user is Candidate
	if res.Status == http.StatusOK && roleName == "Candidate" {
		if listData, ok := res.Data.(response.List); ok {
			for i, item := range listData.List {
				if ohResp, ok := item.(gorm_model.OnboardHistoryResp); ok {
					ohResp.ProjectName = nil
					ohResp.Snapshot = nil
					listData.List[i] = ohResp
				}
			}
			res.Data = listData
		}
	}

	c.JSON(res.Status, res)
}

// FetchActiveTeam
// @Summary Get Active Team Onboarding
// @Description Fetch onboarding history for candidates submitted by the authenticated employee
// @Tags Onboarding
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /onboarding/active-team [get]
// @Security BearerAuth
func (h *routeHandler) FetchActiveTeam(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchOnboardingActiveTeam(c.Request.Context(), tokenData.UserID, pagination.Page, pagination.Limit, pagination.Cursor)
	c.JSON(res.Status, res)
}

// FetchHistory
// @Summary Get Onboarding History
// @Description Fetch onboarding history for candidates whose end date has passed for the authenticated employee
// @Tags Onboarding
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /onboarding/history [get]
// @Security BearerAuth
func (h *routeHandler) FetchHistory(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchOnboardingHistory(c.Request.Context(), tokenData.UserID, pagination.Page, pagination.Limit, pagination.Cursor)
	c.JSON(res.Status, res)
}

// FetchActive
// @Summary Get Active Onboarding
// @Description Fetch onboarding history for candidates currently within their onboarding period
// @Tags Onboarding
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /onboarding/active [get]
// @Security BearerAuth
// @Param job_role_name query string false "Filter by Job Role Name (comma separated)"
// @Param project_name query string false "Filter by Project Name (comma separated)"
// @Param employee_user query string false "Filter by Employee User ID or Name (comma separated)"
func (h *routeHandler) FetchActive(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.OnboardingFilter{}
	if jobRoleName := c.Query("job_role_name"); jobRoleName != "" {
		filter.JobRoleName = strings.Split(jobRoleName, ",")
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter.ProjectName = strings.Split(projectName, ",")
	}
	employeeUser := c.Query("employee_user")
	if employeeUser == "" {
		employeeUser = c.Query("employee_user_name")
	}
	if employeeUser != "" {
		filter.EmployeeUser = strings.Split(employeeUser, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	res := h.Usecase.FetchOnboardingActive(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// ExportActive
// @Summary Export Active Onboarding
// @Description Export active onboarding history
// @Tags Onboarding
// @Accept json
// @Produce octet-stream
// @Param format query string true "Format (pdf, csv, xlsx)"
// @Param job_role_name query string false "Filter by Job Role Name (comma separated)"
// @Param project_name query string false "Filter by Project Name (comma separated)"
// @Param employee_user query string false "Filter by Employee User ID or Name (comma separated)"
// @Success 200 {file} file
// @Router /onboarding/active/export [get]
// @Security BearerAuth
func (h *routeHandler) ExportActive(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	filter := gorm_model.OnboardingFilter{}
	if jobRoleName := c.Query("job_role_name"); jobRoleName != "" {
		filter.JobRoleName = strings.Split(jobRoleName, ",")
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter.ProjectName = strings.Split(projectName, ",")
	}
	employeeUser := c.Query("employee_user")
	if employeeUser == "" {
		employeeUser = c.Query("employee_user_name")
	}
	if employeeUser != "" {
		filter.EmployeeUser = strings.Split(employeeUser, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	adminID := ""
	if claims, ok := c.Get("token_data"); ok {
		tokenData := claims.(domain.JWTClaimUser)
		adminID = tokenData.UserID
	}

	data, contentType, ext, err := h.Usecase.ExportOnboardingActive(c.Request.Context(), filter, format, adminID)
	if err != nil || data == nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=active_onboarding_export."+ext)
	c.Data(http.StatusOK, contentType, data)
}

// FetchArchive
// @Summary Get Archived Onboarding
// @Description Fetch onboarding history for candidates whose onboarding period has ended
// @Tags Onboarding
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /onboarding/archive [get]
// @Security BearerAuth
// @Param job_role_name query string false "Filter by Job Role Name (comma separated)"
// @Param project_name query string false "Filter by Project Name (comma separated)"
// @Param employee_user query string false "Filter by Employee User ID or Name (comma separated)"
func (h *routeHandler) FetchArchive(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.OnboardingFilter{}
	if jobRoleName := c.Query("job_role_name"); jobRoleName != "" {
		filter.JobRoleName = strings.Split(jobRoleName, ",")
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter.ProjectName = strings.Split(projectName, ",")
	}
	employeeUser := c.Query("employee_user")
	if employeeUser == "" {
		employeeUser = c.Query("employee_user_name")
	}
	if employeeUser != "" {
		filter.EmployeeUser = strings.Split(employeeUser, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	res := h.Usecase.FetchOnboardingArchive(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// ExportArchive
// @Summary Export Archived Onboarding
// @Description Export archived onboarding history
// @Tags Onboarding
// @Accept json
// @Produce octet-stream
// @Param format query string true "Format (pdf, csv, xlsx)"
// @Param job_role_name query string false "Filter by Job Role Name (comma separated)"
// @Param project_name query string false "Filter by Project Name (comma separated)"
// @Param employee_user query string false "Filter by Employee User ID or Name (comma separated)"
// @Success 200 {file} file
// @Router /onboarding/archive/export [get]
// @Security BearerAuth
func (h *routeHandler) ExportArchive(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	filter := gorm_model.OnboardingFilter{}
	if jobRoleName := c.Query("job_role_name"); jobRoleName != "" {
		filter.JobRoleName = strings.Split(jobRoleName, ",")
	}
	if projectName := c.Query("project_name"); projectName != "" {
		filter.ProjectName = strings.Split(projectName, ",")
	}
	employeeUser := c.Query("employee_user")
	if employeeUser == "" {
		employeeUser = c.Query("employee_user_name")
	}
	if employeeUser != "" {
		filter.EmployeeUser = strings.Split(employeeUser, ",")
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	adminID := ""
	if claims, ok := c.Get("token_data"); ok {
		tokenData := claims.(domain.JWTClaimUser)
		adminID = tokenData.UserID
	}

	data, contentType, ext, err := h.Usecase.ExportOnboardingArchive(c.Request.Context(), filter, format, adminID)
	if err != nil || data == nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=archived_onboarding_export."+ext)
	c.Data(http.StatusOK, contentType, data)
}
