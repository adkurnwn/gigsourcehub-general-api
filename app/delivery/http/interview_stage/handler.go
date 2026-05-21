package http_interview_stage

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.InterviewStageAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewInterviewStageHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.InterviewStageAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/interview-stages", mdl.Auth(), mdl.AuthRole("Admin", "Superadmin"))
	api.GET("", handler.FetchAll)
	api.GET("/:id", handler.FetchData)

	api.POST("", mdl.AuthRole("Superadmin"), handler.Create)
	api.PUT("/:id", mdl.AuthRole("Superadmin"), handler.Update)
	api.DELETE("/:id", mdl.AuthRole("Superadmin"), handler.Delete)
}

// Create Interview Stage
// @Security BearerAuth
// @Summary Create Interview Stage
// @Tags Interview Stage
// @Accept json
// @Produce json
// @Param request body request_model.CreateInterviewStageRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview-stages [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateInterviewStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Update Interview Stage
// @Security BearerAuth
// @Summary Update Interview Stage
// @Tags Interview Stage
// @Accept json
// @Produce json
// @Param id path string true "Interview Stage ID"
// @Param request body request_model.UpdateInterviewStageRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview-stages/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateInterviewStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// Delete Interview Stage
// @Security BearerAuth
// @Summary Delete Interview Stage
// @Tags Interview Stage
// @Produce json
// @Param id path string true "Interview Stage ID"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview-stages/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get All Interview Stages
// @Security BearerAuth
// @Summary Get All Interview Stages
// @Description Get All Interview Stages
// @Tags Interview Stage
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param search query string false "Search by name"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview-stages [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.InterviewStageFilter{}

	search := c.Query("search")
	if search != "" {
		filter.Search = &search
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Get Interview Stage By ID
// @Security BearerAuth
// @Summary Get Interview Stage By ID
// @Description Get Interview Stage By ID
// @Tags Interview Stage
// @Accept json
// @Produce json
// @Param id path string true "Interview Stage ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /interview-stages/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}