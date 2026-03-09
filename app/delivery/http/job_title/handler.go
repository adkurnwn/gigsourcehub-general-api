package http_job_title

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
	Usecase    domain.JobTitleAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewJobTitleHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.JobTitleAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/job-titles", mdl.Auth(), mdl.AuthRole("Superadmin"))
	api.GET("", handler.FetchAll)
	api.GET("/:id", handler.FetchData)

	api.POST("", handler.Create)
	api.PUT("/:id", handler.Update)
	api.DELETE("/:id", mdl.AuthRole("Superadmin"), handler.Delete)
}

// Create Job Title
// @Security BearerAuth
// @Summary Create Job Title
// @Tags Job Title (Jabatan)
// @Accept json
// @Produce json
// @Param request body request_model.CreateJobTitleRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-titles [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateJobTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Update Job Title
// @Security BearerAuth
// @Summary Update Job Title
// @Tags Job Title (Jabatan)
// @Accept json
// @Produce json
// @Param id path string true "Title ID"
// @Param request body request_model.UpdateJobTitleRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-titles/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateJobTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// Delete Job Title
// @Security BearerAuth
// @Summary Delete Job Title
// @Tags Job Title (Jabatan)
// @Produce json
// @Param id path string true "Title ID"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-titles/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get All Job Titles
// @Security BearerAuth
// @Summary Get All Job Titles
// @Description Get All Job Titles
// @Tags Job Title (Jabatan)
// @Accept json
// @Produce json
// @Param sector_id query string false "Sector ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-titles [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.JobTitleFilter{}

	sectorID := c.Query("sector_id")
	if sectorID != "" {
		filter.SectorID = &sectorID
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Get Job Title By ID
// @Security BearerAuth
// @Summary Get Job Title By ID
// @Description Get Job Title By ID
// @Tags Job Title (Jabatan)
// @Accept json
// @Produce json
// @Param id path string true "Title ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-titles/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
