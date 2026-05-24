package http_interview

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.InterviewAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewInterviewHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.InterviewAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/interview", mdl.Auth(), mdl.AuthRole("Admin"))
	api.GET("", handler.FetchAll)
	api.GET("/scheduled", handler.FetchScheduled)
	api.GET("/:id", handler.FetchData)
	api.POST("", handler.Create)
	api.PUT("/:id", handler.Update)
	api.PATCH("/stage", handler.PatchStage)
	api.PATCH("/status", handler.PatchStatus)
}

// Get All Interviews
// @Security BearerAuth
// @Summary Get Interviews (Admin)
// @Tags Interview
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param candidate_user_id query string false "Filter by candidate user id"
// @Param subrequest_id query string false "Filter by subrequest id"
// @Param stage_id query string false "Filter by stage id"
// @Param status query string false "Filter by status"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.InterviewFilter{}
	filter.ScheduledOnly = true

	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	adminID := claims.(domain.JWTClaimUser).UserID
	filter.AdminUserID = &adminID

	candidateID := c.Query("candidate_user_id")
	if candidateID != "" {
		filter.CandidateUserID = &candidateID
	}

	subrequestID := c.Query("subrequest_id")
	if subrequestID != "" {
		filter.SubrequestID = &subrequestID
	}

	stageID := c.Query("stage_id")
	if stageID != "" {
		filter.StageID = &stageID
	}

	status := c.Query("status")
	if status != "" {
		filter.Status = &status
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Get Scheduled Interviews
// @Security BearerAuth
// @Summary Get Scheduled Interviews (Admin)
// @Tags Interview
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview/scheduled [get]
func (h *routeHandler) FetchScheduled(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)
	pagination := helpers.GetPagination(c)

	res := h.Usecase.FetchScheduled(c.Request.Context(), tokenData.UserID, pagination.Page, pagination.Limit, pagination.Cursor)
	c.JSON(res.Status, res)
}

// Get Interview By ID
// @Security BearerAuth
// @Summary Get Interview By ID
// @Tags Interview
// @Accept json
// @Produce json
// @Param id path string true "Interview ID"
// @Success 200 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid id parameter"))
		return
	}

	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Create Interview
// @Security BearerAuth
// @Summary Create Interview
// @Tags Interview
// @Accept json
// @Produce json
// @Param request body request_model.CreateInterviewRequest true "Create Interview"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview [post]
func (h *routeHandler) Create(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)
	// server will generate interview ID

	var req request_model.CreateInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}

// Update Interview
// @Security BearerAuth
// @Summary Update Interview
// @Tags Interview
// @Accept json
// @Produce json
// @Param id path string true "Interview ID"
// @Param request body request_model.UpdateInterviewRequest true "Update Interview"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid id parameter"))
		return
	}

	var req request_model.UpdateInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), tokenData.UserID, id, req)
	c.JSON(res.Status, res)
}

// Patch Interview Stage
// @Security BearerAuth
// @Summary Patch Interview Stage
// @Tags Interview
// @Accept json
// @Produce json
// @Param request body request_model.PatchInterviewStageRequest true "Patch Stage"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview/stage [patch]
func (h *routeHandler) PatchStage(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)

	var req request_model.PatchInterviewStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.PatchStage(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}

// Patch Interview Status
// @Security BearerAuth
// @Summary Patch Interview Status
// @Tags Interview
// @Accept json
// @Produce json
// @Param request body request_model.PatchInterviewStatusRequest true "Patch Status"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview/status [patch]
func (h *routeHandler) PatchStatus(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)

	var req request_model.PatchInterviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.PatchStatus(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}