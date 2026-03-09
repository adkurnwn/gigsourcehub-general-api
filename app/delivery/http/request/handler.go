package httpdelivery_request

import (
	"log"
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
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

	reqRoute := r.Group("/requests")
	reqRoute.Use(mdl.Auth())
	reqRoute.Use(mdl.AuthEmployee())

	reqRoute.POST("", handler.Create)
	reqRoute.GET("", handler.Fetch)
	reqRoute.GET("/:id", handler.GetDetails)
	reqRoute.PUT("/:id", handler.Update)
	reqRoute.PUT("/:id/subrequests/:sub_id", handler.UpdateSubrequest)
	reqRoute.POST("/:id/subrequests", handler.AddSubrequest)
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

// Get Request Details
// @Summary Get specific Request Details
// @Description Fetch an employee's request and its associated subrequests
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
