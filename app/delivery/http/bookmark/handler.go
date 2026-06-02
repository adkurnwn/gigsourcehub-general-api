package http_bookmark

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.BookmarkAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewBookmarkHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.BookmarkAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/bookmarks", mdl.Auth(), mdl.AuthRole("Admin", "Employee"))
	api.POST("", handler.Create)
	api.DELETE("/:candidate_id", handler.Delete)
	api.GET("", handler.FetchByAdmin)
}

// Create Bookmark
// @Security BearerAuth
// @Summary Create Bookmark
// @Description Admin bookmarks a candidate user
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Param request body request_model.CreateBookmarkRequest true "Create Bookmark Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /bookmarks [post]
func (h *routeHandler) Create(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)

	var req request_model.CreateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}

// Delete Bookmark
// @Security BearerAuth
// @Summary Delete Bookmark
// @Description Admin removes a bookmarked candidate
// @Tags Bookmarks
// @Produce json
// @Param candidate_id path string true "Candidate User ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /bookmarks/{candidate_id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)

	candidateID := c.Param("candidate_id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid candidate_id parameter"))
		return
	}

	res := h.Usecase.Delete(c.Request.Context(), tokenData.UserID, candidateID)
	c.JSON(res.Status, res)
}

// Fetch Bookmarks
// @Security BearerAuth
// @Summary Fetch Admin Bookmarks (note: endpoint /users/candidates bisa digunakan juga, hasil nempel ditiap kandidat)
// @Description Get paginated list of bookmarks for the authenticated admin
// @Tags Bookmarks
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /bookmarks [get]
func (h *routeHandler) FetchByAdmin(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)
	pagination := helpers.GetPagination(c)

	res := h.Usecase.FetchByAdmin(c.Request.Context(), tokenData.UserID, pagination.Page, pagination.Limit)
	c.JSON(res.Status, res)
}
