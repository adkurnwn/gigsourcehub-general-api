package http_provinsi

import (
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	provinsiAppUsecase domain.ProvinsiAppUsecase
	Route                   *gin.RouterGroup
}

func NewProvinsiHandler(r *gin.RouterGroup, uc domain.ProvinsiAppUsecase) {
	handler := &routeHandler{
		provinsiAppUsecase: uc,
		Route:                   r,
	}

	api := r.Group("/provinsi")
	api.GET("", handler.FetchAll)
}

// Get All Provinsi
// @Summary Get All Provinsi
// @Description Get All Provinsi
// @Tags Domisili
// @Accept json
// @Produce json
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /provinsi [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	filter := gorm_model.ProvinsiFilter{}

	res := h.provinsiAppUsecase.FetchAll(c.Request.Context(), filter)
	c.JSON(res.Status, res)
}
