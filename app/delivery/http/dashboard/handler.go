package http_dashboard

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.DashboardAppUsecase
	Middleware middleware.Middleware
}

func NewDashboardHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.DashboardAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Middleware: mdl,
	}

	// Dashboard routes grouping
	api := r.Group("/dashboard", mdl.Auth())
	
	// Admin specific dashboard endpoints
	api.GET("/admin/summary", mdl.AuthAdmin(), handler.GetAdminSummary)
	api.GET("/admin/analytics", mdl.AuthAdmin(), handler.GetAdminAnalytics)

	// Superadmin specific dashboard endpoints
	api.GET("/superadmin/summary", mdl.AuthSuperadmin(), handler.GetSuperadminSummary)
}

func (h *routeHandler) GetAdminSummary(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	res := h.Usecase.GetAdminDashboardSummary(c.Request.Context(), tokenData.UserID)
	c.JSON(res.Status, res)
}

func (h *routeHandler) GetAdminAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "6months")

	res := h.Usecase.GetAdminDashboardAnalytics(c.Request.Context(), period)
	c.JSON(res.Status, res)
}

func (h *routeHandler) GetSuperadminSummary(c *gin.Context) {
	res := h.Usecase.GetSuperadminDashboardSummary(c.Request.Context())
	c.JSON(res.Status, res)
}
