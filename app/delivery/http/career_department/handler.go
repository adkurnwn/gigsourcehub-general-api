package http_career_department

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.CareerDepartmentAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewCareerDepartmentHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.CareerDepartmentAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	// CMS routes — Admin can CRUD, Superadmin can read
	cms := r.Group("/career-departments", mdl.Auth())
	cms.GET("", mdl.AuthRole("Admin", "Superadmin"), handler.FetchAll)
	cms.GET("/:id", mdl.AuthRole("Admin", "Superadmin"), handler.FetchData)
	cms.POST("", mdl.AuthRole("Admin"), handler.Create)
	cms.PUT("/:id", mdl.AuthRole("Admin"), handler.Update)
	cms.DELETE("/:id", mdl.AuthRole("Admin"), handler.Delete)
	cms.POST("/:id/image", mdl.AuthRole("Admin"), handler.UploadImage)

	// Superadmin approval routes
	approvals := r.Group("/career-departments/approvals", mdl.Auth(), mdl.AuthRole("Superadmin"))
	approvals.GET("", handler.FetchApprovals)
	approvals.POST("/:id/approve", handler.ApproveRequest)
	approvals.POST("/:id/reject", handler.RejectRequest)

	// Public routes — no auth required
	pub := r.Group("/public/career-departments")
	pub.GET("", handler.FetchPublic)
	pub.GET("/:id", handler.FetchPublicByID)
}

// Create Career Department
// @Security BearerAuth
// @Summary Create Career Department (Admin)
// @Description Admin creates a new Career Department approval request
// @Tags Career Department
// @Accept json
// @Produce json
// @Param request body request_model.CreateCareerDepartmentRequest true "Create Career Department Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateCareerDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.Create(c.Request.Context(), claims.UserID, req)
	c.JSON(res.Status, res)
}

// Update Career Department
// @Security BearerAuth
// @Summary Update Career Department (Admin)
// @Description Admin creates a Career Department update approval request
// @Tags Career Department
// @Accept json
// @Produce json
// @Param id path string true "Career Department ID"
// @Param request body request_model.UpdateCareerDepartmentRequest true "Update Career Department Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateCareerDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.Update(c.Request.Context(), claims.UserID, id, req)
	c.JSON(res.Status, res)
}

// Delete Career Department
// @Security BearerAuth
// @Summary Delete Career Department (Admin)
// @Description Admin directly deletes a Career Department (no approval needed)
// @Tags Career Department
// @Produce json
// @Param id path string true "Career Department ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Upload Career Department Image
// @Security BearerAuth
// @Summary Upload Career Department Image (Admin)
// @Description Admin uploads an image for a career department to S3
// @Tags Career Department
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Career Department ID"
// @Param image formData file true "Image file (JPEG, PNG, WebP)"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/{id}/image [post]
func (h *routeHandler) UploadImage(c *gin.Context) {
	id := c.Param("id")

	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, response.Error(400, "image file is required"))
		return
	}

	res := h.Usecase.UploadImage(c.Request.Context(), id, fileHeader)
	c.JSON(res.Status, res)
}

// Get All Career Departments (CMS)
// @Security BearerAuth
// @Summary Get All Career Departments (CMS)
// @Description Get all career departments with pagination (Admin & Superadmin)
// @Tags Career Department
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search by name"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)

	var search *string
	if v := c.Query("search"); v != "" {
		search = &v
	}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, search)
	c.JSON(res.Status, res)
}

// Get Career Department By ID (CMS)
// @Security BearerAuth
// @Summary Get Career Department By ID (CMS)
// @Description Get career department detail by ID (Admin & Superadmin)
// @Tags Career Department
// @Accept json
// @Produce json
// @Param id path string true "Career Department ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get Career Department Approval Requests (Superadmin)
// @Security BearerAuth
// @Summary Get Career Department Approval Requests
// @Description Get all career department approval requests with pagination (Superadmin)
// @Tags Career Department Approvals
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/approvals [get]
func (h *routeHandler) FetchApprovals(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	res := h.Usecase.FetchApprovals(c.Request.Context(), pagination.Page, pagination.Limit)
	c.JSON(res.Status, res)
}

// Approve Career Department Request (Superadmin)
// @Security BearerAuth
// @Summary Approve Career Department Approval Request
// @Description Superadmin approves a career department approval request
// @Tags Career Department Approvals
// @Accept json
// @Produce json
// @Param id path string true "Approval Request ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/approvals/{id}/approve [post]
func (h *routeHandler) ApproveRequest(c *gin.Context) {
	id := c.Param("id")
	claims := c.MustGet("token_data").(domain.JWTClaimUser)
	res := h.Usecase.ApproveRequest(c.Request.Context(), claims.UserID, id)
	c.JSON(res.Status, res)
}

// Reject Career Department Request (Superadmin)
// @Security BearerAuth
// @Summary Reject Career Department Approval Request
// @Description Superadmin rejects a career department approval request
// @Tags Career Department Approvals
// @Accept json
// @Produce json
// @Param id path string true "Approval Request ID"
// @Param request body request_model.ReviewApprovalRequest true "Rejection Reason"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /career-departments/approvals/{id}/reject [post]
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

// Get Public Career Departments (no auth)
// @Summary Get Public Career Departments
// @Description Get all approved career departments (no authentication required)
// @Tags Career Department Public
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search by name"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/career-departments [get]
func (h *routeHandler) FetchPublic(c *gin.Context) {
	pagination := helpers.GetPagination(c)

	var search *string
	if v := c.Query("search"); v != "" {
		search = &v
	}

	res := h.Usecase.FetchPublic(c.Request.Context(), pagination.Page, pagination.Limit, search)
	c.JSON(res.Status, res)
}

// Get Public Career Department By ID (no auth)
// @Summary Get Public Career Department By ID
// @Description Get a single approved career department by ID (no authentication required)
// @Tags Career Department Public
// @Accept json
// @Produce json
// @Param id path string true "Career Department ID"
// @Success 200 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/career-departments/{id} [get]
func (h *routeHandler) FetchPublicByID(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchPublicByID(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
