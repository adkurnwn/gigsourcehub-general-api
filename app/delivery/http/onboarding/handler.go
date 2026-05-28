package http_onboarding

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
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

	onboarding := r.Group("/onboarding", mdl.Auth(), mdl.AuthRole("Admin", "Employee"))
	onboarding.GET("/active-team", handler.FetchActiveTeam)
	onboarding.GET("/history", handler.FetchHistory)
	onboarding.GET("/:id", handler.FetchByCandidate)
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

	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchOnboardingByCandidate(c.Request.Context(), candidateID, pagination.Page, pagination.Limit, pagination.Cursor)
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
