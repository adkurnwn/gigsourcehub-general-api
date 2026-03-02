package http_role_applied

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
	Usecase    domain.RoleAppliedAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewRoleAppliedHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.RoleAppliedAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/roles", mdl.Auth())
	api.GET("", handler.FetchAll)
	api.GET("/:id", handler.FetchData)

	// Note: Role Modifications should typically be protected by Admin/Superadmin layers.
	api.POST("", mdl.AuthRole("Superadmin"), handler.Create)
	api.PUT("/:id", mdl.AuthRole("Superadmin"), handler.Update)
	api.DELETE("/:id", mdl.AuthRole("Superadmin"), handler.Delete)
}

// Create Role Applied
// @Security BearerAuth
// @Summary Create Role Applied
// @Tags Role Applied
// @Accept json
// @Produce json
// @Param request body request_model.CreateRoleAppliedRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateRoleAppliedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Update Role Applied
// @Security BearerAuth
// @Summary Update Role Applied
// @Tags Role Applied
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body request_model.UpdateRoleAppliedRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateRoleAppliedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// Delete Role Applied
// @Security BearerAuth
// @Summary Delete Role Applied
// @Tags Role Applied
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

// Get All Role Applied
// @Security BearerAuth
// @Summary Get All Role Applied
// @Description Get All Role Applied
// @Tags Role Applied
// @Accept json
// @Produce json
// @Param sector_id query string false "Sector ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /roles [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.RoleAppliedFilter{}

	sectorID := c.Query("sector_id")
	if sectorID != "" {
		filter.SectorID = &sectorID
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Get Role Applied By ID
// @Security BearerAuth
// @Summary Get Role Applied By ID
// @Description Get Role Applied By ID
// @Tags Role Applied
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
