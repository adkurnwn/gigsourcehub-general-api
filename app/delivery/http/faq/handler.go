package http_faq

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.FAQAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewFAQHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.FAQAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	// CMS routes — Admin can CRUD, Superadmin can read
	cms := r.Group("/faqs", mdl.Auth())
	cms.GET("", mdl.AuthRole("Admin", "Superadmin"), handler.FetchAll)
	cms.GET("/:id", mdl.AuthRole("Admin", "Superadmin"), handler.FetchData)
	cms.POST("", mdl.AuthRole("Admin"), handler.Create)
	cms.PUT("/:id", mdl.AuthRole("Admin"), handler.Update)
	cms.DELETE("/:id", mdl.AuthRole("Admin"), handler.Delete)

	// Superadmin approval routes
	approvals := r.Group("/faqs/approvals", mdl.Auth(), mdl.AuthRole("Superadmin"))
	approvals.GET("", handler.FetchApprovals)
	approvals.POST("/:id/approve", handler.ApproveRequest)
	approvals.POST("/:id/reject", handler.RejectRequest)

	// Public routes — no auth required
	pub := r.Group("/public/faqs")
	pub.GET("", handler.FetchPublic)
	pub.GET("/:id", handler.FetchPublicByID)
}

// Create FAQ
// @Security BearerAuth
// @Summary Create FAQ (Admin)
// @Description Admin creates a new FAQ approval request
// @Tags FAQ
// @Accept json
// @Produce json
// @Param request body request_model.CreateFAQRequest true "Create FAQ Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.Create(c.Request.Context(), claims.UserID, req)
	c.JSON(res.Status, res)
}

// Update FAQ
// @Security BearerAuth
// @Summary Update FAQ (Admin)
// @Description Admin creates a FAQ update approval request
// @Tags FAQ
// @Accept json
// @Produce json
// @Param id path string true "FAQ ID"
// @Param request body request_model.UpdateFAQRequest true "Update FAQ Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.Update(c.Request.Context(), claims.UserID, id, req)
	c.JSON(res.Status, res)
}

// Delete FAQ
// @Security BearerAuth
// @Summary Delete FAQ (Admin)
// @Description Admin directly deletes a FAQ (no approval needed)
// @Tags FAQ
// @Produce json
// @Param id path string true "FAQ ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get All FAQs (CMS)
// @Security BearerAuth
// @Summary Get All FAQs (CMS)
// @Description Get all FAQs with pagination (Admin & Superadmin)
// @Tags FAQ
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search by question"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)

	var search *string
	if v := c.Query("search"); v != "" {
		search = &v
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, search)
	c.JSON(res.Status, res)
}

// Get FAQ By ID (CMS)
// @Security BearerAuth
// @Summary Get FAQ By ID (CMS)
// @Description Get FAQ detail by ID (Admin & Superadmin)
// @Tags FAQ
// @Accept json
// @Produce json
// @Param id path string true "FAQ ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get FAQ Approval Requests (Superadmin)
// @Security BearerAuth
// @Summary Get FAQ Approval Requests
// @Description Get all FAQ approval requests with pagination (Superadmin)
// @Tags FAQ Approvals
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs/approvals [get]
func (h *routeHandler) FetchApprovals(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchApprovals(c.Request.Context(), pagination.Page, pagination.Limit)
	c.JSON(res.Status, res)
}

// Approve FAQ Request (Superadmin)
// @Security BearerAuth
// @Summary Approve FAQ Approval Request
// @Description Superadmin approves a FAQ approval request
// @Tags FAQ Approvals
// @Accept json
// @Produce json
// @Param id path string true "Approval Request ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs/approvals/{id}/approve [post]
func (h *routeHandler) ApproveRequest(c *gin.Context) {
	id := c.Param("id")
	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.ApproveRequest(c.Request.Context(), claims.UserID, id)
	c.JSON(res.Status, res)
}

// Reject FAQ Request (Superadmin)
// @Security BearerAuth
// @Summary Reject FAQ Approval Request
// @Description Superadmin rejects a FAQ approval request
// @Tags FAQ Approvals
// @Accept json
// @Produce json
// @Param id path string true "Approval Request ID"
// @Param request body request_model.ReviewApprovalRequest true "Rejection Reason"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /faqs/approvals/{id}/reject [post]
func (h *routeHandler) RejectRequest(c *gin.Context) {
	id := c.Param("id")

	var req request_model.ReviewApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.RejectRequest(c.Request.Context(), claims.UserID, id, req)
	c.JSON(res.Status, res)
}

// Get Public FAQs (no auth)
// @Summary Get Public FAQs
// @Description Get all approved FAQs (no authentication required)
// @Tags FAQ Public
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search by question"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/faqs [get]
func (h *routeHandler) FetchPublic(c *gin.Context) {
	pagination := helpers.GetPagination(c)

	var search *string
	if v := c.Query("search"); v != "" {
		search = &v
	}

	res := h.Usecase.FetchPublic(c.Request.Context(), pagination.Page, pagination.Limit, search)
	c.JSON(res.Status, res)
}

// Get Public FAQ By ID (no auth)
// @Summary Get Public FAQ By ID
// @Description Get a single approved FAQ by ID (no authentication required)
// @Tags FAQ Public
// @Accept json
// @Produce json
// @Param id path string true "FAQ ID"
// @Success 200 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/faqs/{id} [get]
func (h *routeHandler) FetchPublicByID(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchPublicByID(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
