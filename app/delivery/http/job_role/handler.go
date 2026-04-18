package http_job_role

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
	Usecase    domain.JobRoleAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewJobRoleHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.JobRoleAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/roles", mdl.Auth())
	api.GET("/system", handler.FetchSystemRoles)
	api.GET("", handler.FetchAll)
	api.GET("/:id", handler.FetchData)

	// Note: Role Modifications should typically be protected by Admin/Superadmin layers.
	api.POST("", mdl.AuthRole("Superadmin"), handler.Create)
	api.PUT("/:id", mdl.AuthRole("Superadmin"), handler.Update)
	api.DELETE("/:id", mdl.AuthRole("Superadmin"), handler.Delete)
}

// Create Job Role
// @Security BearerAuth
// @Summary Create Job Role
// @Tags Job Role
// @Accept json
// @Produce json
// @Param request body request_model.CreateJobRoleRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateJobRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Update Job Role
// @Security BearerAuth
// @Summary Update Job Role
// @Tags Job Role
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body request_model.UpdateJobRoleRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateJobRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// Delete Job Role
// @Security BearerAuth
// @Summary Delete Job Role
// @Tags Job Role
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get All Job Roles
// @Security BearerAuth
// @Summary Get All Job Roles
// @Description Get All Job Roles
// @Tags Job Role
// @Accept json
// @Produce json
// @Param sector_id query string false "Sector ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.JobRoleFilter{}

	sectorID := c.Query("sector_id")
	if sectorID != "" {
		filter.SectorID = &sectorID
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Get Job Role By ID
// @Security BearerAuth
// @Summary Get Job Role By ID
// @Description Get Job Role By ID
// @Tags Job Role
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get System Job Roles
// @Security BearerAuth
// @Summary Get System Job Roles
// @Description Get System Job Roles
// @Tags Job Role
// @Accept json
// @Produce json
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles/system [get]
func (h *routeHandler) FetchSystemRoles(c *gin.Context) {
	res := h.Usecase.FetchSystemRoles(c.Request.Context())
	c.JSON(res.Status, res)
}
