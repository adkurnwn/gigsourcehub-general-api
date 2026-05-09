package http_admin_note

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.AdminNoteAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewAdminNoteHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.AdminNoteAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/notes", mdl.Auth(), mdl.AuthRole("Admin", "Employee"))
	api.POST("/:id", handler.Create)
	api.GET("/:id", handler.FetchByCandidate)
}

// Create Admin Note
// @Security BearerAuth
// @Summary Create Admin Note
// @Description Add a note to a candidate user (Admin/Employee only)
// @Tags Notes
// @Accept json
// @Produce json
// @Param id path string true "Candidate User ID"
// @Param request body request_model.CreateAdminNoteRequest true "Create Note Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /notes/{id} [post]
func (h *routeHandler) Create(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	tokenData := claims.(domain.JWTClaimUser)
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid id parameter"))
		return
	}

	var req request_model.CreateAdminNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), tokenData.UserID, candidateID, req)
	c.JSON(res.Status, res)
}

// Fetch Admin Notes
// @Security BearerAuth
// @Summary Fetch Admin Notes
// @Description Get all notes for a candidate user (Admin/Employee only)
// @Tags Notes
// @Produce json
// @Param id path string true "Candidate User ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /notes/{id} [get]
func (h *routeHandler) FetchByCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	if candidateID == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid id parameter"))
		return
	}

	res := h.Usecase.FetchByCandidate(c.Request.Context(), candidateID)
	c.JSON(res.Status, res)
}
