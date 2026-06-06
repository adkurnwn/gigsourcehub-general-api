package httpdelivery_request

import (
	"log"
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.RequestAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewRequestHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.RequestAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	reqRoute := r.Group("/requests", mdl.Auth())

	reqRoute.POST("", mdl.AuthEmployee(), handler.Create)
	reqRoute.GET("", mdl.AuthEmployee(), handler.Fetch)
	reqRoute.GET("/:id", mdl.AuthRole("Employee", "Admin"), handler.GetDetails)
	reqRoute.PUT("/:id", mdl.AuthEmployee(), handler.Update)
	reqRoute.PUT("/:id/subrequests/:sub_id", mdl.AuthEmployee(), handler.UpdateSubrequest)
	reqRoute.POST("/:id/subrequests", mdl.AuthEmployee(), handler.AddSubrequest)

	adminRoute := r.Group("/admin/requests")
	adminRoute.Use(mdl.Auth())
	adminRoute.Use(mdl.AuthAdmin())

	adminRoute.GET("", handler.FetchAllForAdmin)
	adminRoute.GET("/export", handler.ExportRequests)
	adminRoute.GET("/pending", handler.FetchPendingForAdmin)
	adminRoute.GET("/my-requests", handler.FetchMyRequestsForAdmin)
	adminRoute.PATCH("/:id/validate", handler.AssignPIC)
	adminRoute.PATCH("/:id/reject", handler.RejectRequest)
	adminRoute.POST("/:request_id/subrequests/:sub_id/assign", handler.AssignCandidateToSubrequest)
}

func parseRequestFilter(ctx *gin.Context) gorm_model.RequestFilter {
	status := ctx.Query("status")
	urgency := ctx.Query("urgency")
	search := ctx.Query("search")
	proposedBy := ctx.Query("proposed_by")
	adminName := ctx.Query("admin_name")

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}
	var urgencyPtr *string
	if urgency != "" {
		urgencyPtr = &urgency
	}
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}
	var proposedByPtr *string
	if proposedBy != "" {
		proposedByPtr = &proposedBy
	}
	var adminNamePtr *string
	if adminName != "" {
		adminNamePtr = &adminName
	}

	return gorm_model.RequestFilter{
		Status:     statusPtr,
		Urgency:    urgencyPtr,
		Search:     searchPtr,
		ProposedBy: proposedByPtr,
		AdminName:  adminNamePtr,
	}
}

// Create Request
// @Summary Create a Request and multiple Subrequests
// @Description Allows an Employee to submit a new Request containing multiple distinct Subrequest configurations
// @Tags Employee Request
// @Accept json
// @Produce json
// @Param req body request_model.CreateRequestRequest true "Request Data"
// @Success 200 {object} response.Base{data=nil}
// @Router /requests [post]
// @Security BearerAuth
func (h *routeHandler) Create(ctx *gin.Context) {
	var body request_model.CreateRequestRequest

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println("BindJSON error:", err)
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request payload"))
		return
	}

	// Extract UserID from JWT token data
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.CreateByEmployee(ctx.Request.Context(), userID, body)
	ctx.JSON(result.Status, result)
}

// Fetch Requests
// @Summary Fetch My Requests
// @Description Get paginated list of requests owned by the authenticated employee
// @Tags Employee Request
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /requests [get]
// @Security BearerAuth
func (h *routeHandler) Fetch(ctx *gin.Context) {
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID
	pagination := helpers.GetPagination(ctx)

	result := h.Usecase.FetchByEmployee(ctx.Request.Context(), userID, pagination.Page, pagination.Limit)
	ctx.JSON(result.Status, result)
}

// Fetch All Requests (Admin)
// @Summary Fetch all requests
// @Description Get paginated list of all requests for admin dashboard, with optional filters
// @Tags Admin Request
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param status query string false "Filter by request status"
// @Param urgency query string false "Filter by request urgency"
// @Param search query string false "Search by project name"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /admin/requests [get]
// @Security BearerAuth
func (h *routeHandler) FetchAllForAdmin(ctx *gin.Context) {
	pagination := helpers.GetPagination(ctx)
	filter := parseRequestFilter(ctx)

	result := h.Usecase.FetchByAdmin(ctx.Request.Context(), pagination.Page, pagination.Limit, filter)
	ctx.JSON(result.Status, result)
}

// ExportRequests
// @Summary Export Admin Requests
// @Description Export admin requests to file
// @Tags Admin Request
// @Accept json
// @Produce octet-stream
// @Param format query string true "Format (pdf, csv, xlsx)"
// @Param status query string false "Filter by Status"
// @Param urgency query string false "Filter by Urgency"
// @Param search query string false "Filter by Keyword"
// @Param proposed_by query string false "Filter by Proposed By"
// @Param admin_name query string false "Filter by Admin Name"
// @Success 200 {file} file
// @Router /admin/requests/export [get]
// @Security BearerAuth
func (h *routeHandler) ExportRequests(ctx *gin.Context) {
	format := ctx.Query("format")
	if format == "" {
		format = "csv"
	}

	filter := parseRequestFilter(ctx)

	adminID := ""
	if claims, ok := ctx.Get("token_data"); ok {
		tokenData := claims.(domain.JWTClaimUser)
		adminID = tokenData.UserID
	}

	data, contentType, ext, err := h.Usecase.ExportRequests(ctx.Request.Context(), filter, format, adminID)
	if err != nil || data == nil {
		ctx.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export"))
		return
	}

	ctx.Header("Content-Disposition", "attachment; filename=requests_export."+ext)
	ctx.Data(http.StatusOK, contentType, data)
}

// Fetch Pending Requests (Admin)
// @Summary Fetch pending requests
// @Description Get paginated list of requests whose status is PENDING for admin dashboard
// @Tags Admin Request
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param status query string false "Filter by status"
// @Param urgency query string false "Filter by urgency"
// @Param proposed_by query string false "Filter by proposed_by"
// @Param admin_name query string false "Filter by admin_name"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /admin/requests/pending [get]
// @Security BearerAuth
func (h *routeHandler) FetchPendingForAdmin(ctx *gin.Context) {
	pagination := helpers.GetPagination(ctx)
	filter := parseRequestFilter(ctx)

	result := h.Usecase.FetchPendingForAdmin(ctx.Request.Context(), pagination.Page, pagination.Limit, filter)
	ctx.JSON(result.Status, result)
}

// Fetch My Requests (Admin)
// @Summary Fetch my assigned requests
// @Description Get paginated list of requests assigned to the authenticated admin
// @Tags Admin Request
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param status query string false "Filter by status"
// @Param urgency query string false "Filter by urgency"
// @Param proposed_by query string false "Filter by proposed_by"
// @Param admin_name query string false "Filter by admin_name"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /admin/requests/my-requests [get]
// @Security BearerAuth
func (h *routeHandler) FetchMyRequestsForAdmin(ctx *gin.Context) {
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	adminID := userClaim.(domain.JWTClaimUser).UserID
	pagination := helpers.GetPagination(ctx)
	filter := parseRequestFilter(ctx)

	result := h.Usecase.FetchMyRequestsForAdmin(ctx.Request.Context(), adminID, pagination.Page, pagination.Limit, filter)
	ctx.JSON(result.Status, result)
}

// Get Request Details
// @Summary Get specific Request Details
// @Description Fetch a request and its associated subrequests (employee or admin)
// @Tags Employee Request
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /requests/{id} [get]
// @Security BearerAuth
func (h *routeHandler) GetDetails(ctx *gin.Context) {
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID
	requestID := ctx.Param("id")

	if requestID == "" {
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid Request ID"))
		return
	}

	result := h.Usecase.GetByID(ctx.Request.Context(), userID, requestID)
	ctx.JSON(result.Status, result)
}

// Update Request
// @Summary Update an existing Request
// @Description Updates a request's core fields only (`project_name`, `due_date`, `urgency`) for an Employee.
// @Tags Employee Request
// @Accept json
// @Produce json
// @Param id path string true "Request ID"
// @Param req body request_model.UpdateRequestRequest true "Update Request Data"
// @Success 200 {object} response.Base{data=nil}
// @Router /requests/{id} [put]
// @Security BearerAuth
func (h *routeHandler) Update(ctx *gin.Context) {
	requestID := ctx.Param("id")

	var body request_model.UpdateRequestRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println("BindJSON error:", err)
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request payload"))
		return
	}

	// Extract UserID
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.UpdateByEmployee(ctx.Request.Context(), userID, requestID, body)
	ctx.JSON(result.Status, result)
}

// Assign PIC (Admin)
// @Summary Assign request PIC
// @Description Assign the authenticated admin as PIC for a request
// @Tags Admin Request
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} response.Base{data=nil}
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /admin/requests/{id}/validate [patch]
// @Security BearerAuth
func (h *routeHandler) AssignPIC(ctx *gin.Context) {
	requestID := ctx.Param("id")
	if requestID == "" {
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid Request ID"))
		return
	}

	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	adminID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.AssignPIC(ctx.Request.Context(), adminID, requestID)
	ctx.JSON(result.Status, result)
}

// Reject Request (Admin)
// @Summary Reject a Request
// @Description Reject the request and record a reason for the employee
// @Tags Admin Request
// @Accept json
// @Produce json
// @Param id path string true "Request ID"
// @Param req body request_model.RejectRequestRequest true "Reject payload"
// @Success 200 {object} response.Base{data=nil}
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /admin/requests/{id}/reject [patch]
// @Security BearerAuth
func (h *routeHandler) RejectRequest(ctx *gin.Context) {
	requestID := ctx.Param("id")
	if requestID == "" {
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid Request ID"))
		return
	}

	var body request_model.RejectRequestRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println("BindJSON error:", err)
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid payload"))
		return
	}

	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	adminID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.RejectRequest(ctx.Request.Context(), adminID, requestID, body.RejectedReason)
	ctx.JSON(result.Status, result)
}

// Update Subrequest
// @Summary Update an existing Subrequest specifically
// @Description Updates the isolated fields of a subrequest, maintaining data relationships natively.
// @Tags Employee Request
// @Accept json
// @Produce json
// @Param id path string true "Request ID"
// @Param sub_id path string true "Subrequest ID"
// @Param req body request_model.UpdateSubrequestRequest true "Update Subrequest Data"
// @Success 200 {object} response.Base{data=nil}
// @Router /requests/{id}/subrequests/{sub_id} [put]
// @Security BearerAuth
func (h *routeHandler) UpdateSubrequest(ctx *gin.Context) {
	requestID := ctx.Param("id")
	subrequestID := ctx.Param("sub_id")

	var body request_model.UpdateSubrequestRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println("BindJSON error:", err)
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid subrequest payload"))
		return
	}

	// Extract UserID
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.UpdateSubrequestByEmployee(ctx.Request.Context(), userID, requestID, subrequestID, body)
	ctx.JSON(result.Status, result)
}

// Add Subrequest
// @Summary Add a new Subrequest to an existing Request
// @Description Appends a new subrequest and automatically increments the parent request's required headcount.
// @Tags Employee Request
// @Accept json
// @Produce json
// @Param id path string true "Request ID"
// @Param req body request_model.CreateSubrequestRequest true "New Subrequest Data"
// @Success 200 {object} response.Base{data=nil}
// @Router /requests/{id}/subrequests [post]
// @Security BearerAuth
func (h *routeHandler) AddSubrequest(ctx *gin.Context) {
	requestID := ctx.Param("id")

	var body request_model.CreateSubrequestRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println("BindJSON error:", err)
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid subrequest payload"))
		return
	}

	// Extract UserID
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.AddSubrequestByEmployee(ctx.Request.Context(), userID, requestID, body)
	ctx.JSON(result.Status, result)
}

// Assign Candidate To Subrequest
// @Summary Assign a candidate to a subrequest
// @Description Creates a subrequest-candidate pivot row and marks the candidate's recruitment status as Assigned
// @Tags Admin Request
// @Accept json
// @Produce json
// @Param request_id path string true "Request ID"
// @Param sub_id path string true "Subrequest ID"
// @Param req body request_model.AssignCandidateToSubrequestRequest true "Assignment Data"
// @Success 200 {object} response.Base{data=nil}
// @Router /admin/requests/{request_id}/subrequests/{sub_id}/assign [post]
// @Security BearerAuth
func (h *routeHandler) AssignCandidateToSubrequest(ctx *gin.Context) {
	requestID := ctx.Param("request_id")
	subrequestID := ctx.Param("sub_id")

	var body request_model.AssignCandidateToSubrequestRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println("BindJSON error:", err)
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid assignment payload"))
		return
	}

	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	adminID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.AssignCandidateToSubrequest(ctx.Request.Context(), adminID, requestID, subrequestID, body)
	ctx.JSON(result.Status, result)
}
