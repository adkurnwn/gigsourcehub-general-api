package http_search

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	usecase_search "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/search"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	Usecase usecase_search.SearchUsecase
}

func NewSearchHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc usecase_search.SearchUsecase) {
	handler := &SearchHandler{Usecase: uc}

	api := r.Group("/search")
	api.POST("", mdl.Auth(), mdl.AuthRole("Admin", "Superadmin"), handler.Search)
}

// Search
// @Summary Search
// @Description Search for items
// @Tags Search
// @Accept json
// @Produce json
// @Param request body request_model.SearchRequest true "Search Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /search [post]
// @Security BearerAuth
func (h *SearchHandler) Search(c *gin.Context) {
	var req request_model.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid request body"))
		return
	}

	resp := h.Usecase.Search(c.Request.Context(), req.Query)
	c.JSON(resp.Status, resp)
}
