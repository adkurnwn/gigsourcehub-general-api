package http_role_applied

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
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
