package http_kabupaten_kota

import (
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	kabupatenKotaAppUsecase domain.KabupatenKotaAppUsecase
	Route                   *gin.RouterGroup
}

func NewKabupatenKotaHandler(r *gin.RouterGroup, uc domain.KabupatenKotaAppUsecase) {
	handler := &routeHandler{
		kabupatenKotaAppUsecase: uc,
		Route:                   r,
	}

	api := r.Group("/kabupaten")
	api.GET("", handler.FetchAll)
}

// Get All Kabupaten Kota by Province ID
// @Summary Get All Kabupaten Kota by Province ID
// @Description Get All Kabupaten Kota by Province ID
// @Tags Domisili
// @Accept json
// @Produce json
// @Param provinsi_id query string true "Provinsi ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /kabupaten [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	filter := gorm_model.KabupatenKotaFilter{}

	provinsiID := c.Query("provinsi_id")
	if provinsiID == "" {
		c.JSON(400, response.ErrorValidation(map[string]string{"provinsi_id": "required"}, "provinsi_id is required"))
		return
	}
	filter.ProvinsiID = &provinsiID

	res := h.kabupatenKotaAppUsecase.FetchAll(c.Request.Context(), filter)
	c.JSON(res.Status, res)
}
