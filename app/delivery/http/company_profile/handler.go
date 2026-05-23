package http_company_profile

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.CompanyProfileAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewCompanyProfileHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.CompanyProfileAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	// CMS routes — Admin & Superadmin can view, Admin can update
	cms := r.Group("/company-profile", mdl.Auth())
	cms.GET("", mdl.AuthRole("Admin", "Superadmin"), handler.Get)
	cms.PUT("", mdl.AuthRole("Admin"), handler.AdminUpdate)

	// Admin cancel pending approval
	cms.DELETE("/approvals/:id", mdl.AuthRole("Admin"), handler.CancelPendingApproval)

	// Superadmin approval routes
	approvals := r.Group("/company-profile/approvals", mdl.Auth(), mdl.AuthRole("Superadmin"))
	approvals.GET("", handler.FetchPendingApprovals)
	approvals.POST("/:id/approve", handler.ApproveRequest)
	approvals.POST("/:id/reject", handler.RejectRequest)

	// Public routes — no auth required
	pub := r.Group("/public/company-profile")
	pub.GET("", handler.GetPublic)
}

// Get Company Profile (CMS)
// @Security BearerAuth
// @Summary Get Company Profile
// @Description Get company profile data (Admin & Superadmin)
// @Tags Company Profile
// @Accept json
// @Produce json
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /company-profile [get]
func (h *routeHandler) Get(c *gin.Context) {
	res := h.Usecase.Get(c.Request.Context())
	c.JSON(res.Status, res)
}

// Update Company Profile (Admin)
// @Security BearerAuth
// @Summary Update Company Profile (Admin)
// @Description Admin creates a company profile update approval request
// @Tags Company Profile
// @Accept json
// @Produce json
// @Param request body request_model.UpdateCompanyProfileRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /company-profile [put]
func (h *routeHandler) AdminUpdate(c *gin.Context) {
	var req request_model.UpdateCompanyProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.AdminUpdate(c.Request.Context(), claims.UserID, req)
	c.JSON(res.Status, res)
}

// Cancel Pending Approval (Admin)
// @Security BearerAuth
// @Summary Cancel Pending Company Profile Approval
// @Description Admin cancels a pending company profile update approval
// @Tags Company Profile
// @Produce json
// @Param id path string true "Approval Request ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /company-profile/approvals/{id} [delete]
func (h *routeHandler) CancelPendingApproval(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.CancelPendingApproval(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get Pending Approvals (Superadmin)
// @Security BearerAuth
// @Summary Get Company Profile Pending Approvals
// @Description Get all pending company profile approval requests (Superadmin)
// @Tags Company Profile Approvals
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /company-profile/approvals [get]
func (h *routeHandler) FetchPendingApprovals(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchPendingApprovals(c.Request.Context(), pagination.Page, pagination.Limit)
	c.JSON(res.Status, res)
}

// Approve Request (Superadmin)
// @Security BearerAuth
// @Summary Approve Company Profile Update
// @Description Superadmin approves a company profile update request
// @Tags Company Profile Approvals
// @Accept json
// @Produce json
// @Param id path string true "Approval Request ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /company-profile/approvals/{id}/approve [post]
func (h *routeHandler) ApproveRequest(c *gin.Context) {
	id := c.Param("id")
	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.ApproveRequest(c.Request.Context(), claims.UserID, id)
	c.JSON(res.Status, res)
}

// Reject Request (Superadmin)
// @Security BearerAuth
// @Summary Reject Company Profile Update
// @Description Superadmin rejects a company profile update request
// @Tags Company Profile Approvals
// @Accept json
// @Produce json
// @Param id path string true "Approval Request ID"
// @Param request body request_model.ReviewApprovalRequest true "Rejection Reason"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /company-profile/approvals/{id}/reject [post]
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

// Get Public Company Profile (no auth)
// @Summary Get Public Company Profile
// @Description Get company profile data (no authentication required)
// @Tags Company Profile Public
// @Accept json
// @Produce json
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/company-profile [get]
func (h *routeHandler) GetPublic(c *gin.Context) {
	res := h.Usecase.GetPublic(c.Request.Context())
	c.JSON(res.Status, res)
}
