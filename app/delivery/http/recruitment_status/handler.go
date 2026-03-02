package http_recruitment_status

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
	Usecase    domain.RecruitmentStatusAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewRecruitmentStatusHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.RecruitmentStatusAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/recruitment-statuses", mdl.Auth(), mdl.AuthRole("Admin", "Superadmin"))
	api.GET("", handler.FetchAll)
	api.GET("/:id", handler.FetchData)

	// Note: Recruitment Status Modifications should typically be protected by Admin/Superadmin layers.
	api.POST("", mdl.AuthRole("Superadmin"), handler.Create)
	api.PUT("/:id", mdl.AuthRole("Superadmin"), handler.Update)
	api.DELETE("/:id", mdl.AuthRole("Superadmin"), handler.Delete)
}

// Create Recruitment Status
// @Security BearerAuth
// @Summary Create Recruitment Status
// @Tags Recruitment Status
// @Accept json
// @Produce json
// @Param request body request_model.CreateRecruitmentStatusRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /recruitment-statuses [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateRecruitmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Update Recruitment Status
// @Security BearerAuth
// @Summary Update Recruitment Status
// @Tags Recruitment Status
// @Accept json
// @Produce json
// @Param id path string true "Recruitment Status ID"
// @Param request body request_model.UpdateRecruitmentStatusRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /recruitment-statuses/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateRecruitmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// Delete Recruitment Status
// @Security BearerAuth
// @Summary Delete Recruitment Status
// @Tags Recruitment Status
// @Produce json
// @Param id path string true "Recruitment Status ID"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /recruitment-statuses/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get All Recruitment Status
// @Security BearerAuth
// @Summary Get All Recruitment Status
// @Description Get All Recruitment Status
// @Tags Recruitment Status
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param recruitment_status_id query string false "Recruitment Status ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /recruitment-statuses [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.RecruitmentStatusFilter{}

	recruitmentStatusID := c.Query("recruitment_status_id")
	if recruitmentStatusID != "" {
		filter.RecruitmentStatusID = &recruitmentStatusID
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Get Recruitment Status By ID
// @Security BearerAuth
// @Summary Get Recruitment Status By ID
// @Description Get Recruitment Status By ID
// @Tags Recruitment Status
// @Accept json
// @Produce json
// @Param id path string true "Recruitment Status ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /recruitment-statuses/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}