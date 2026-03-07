package httpdelivery_request

import (
	"log"
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
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
