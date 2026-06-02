package http_onboarding

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
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
func (h *routeHandler) FetchActive(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchOnboardingActive(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor)
	c.JSON(res.Status, res)
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
func (h *routeHandler) FetchArchive(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchOnboardingArchive(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor)
	c.JSON(res.Status, res)
}
