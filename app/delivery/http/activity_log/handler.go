package http_activity_log

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.ActivityLogAppUsecase
	Middleware middleware.Middleware
}

func NewActivityLogHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.ActivityLogAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Middleware: mdl,
	}

	// Protected by Superadmin middleware as per request
	api := r.Group("/activity-logs", mdl.Auth(), mdl.AuthSuperadmin())
	api.GET("", handler.FetchAll)
	api.GET("/export", handler.Export)
}

func (h *routeHandler) getFilter(c *gin.Context) gorm_model.LogActivityFilter {
	filter := gorm_model.LogActivityFilter{}

	if actorID := c.Query("actor_id"); actorID != "" {
		filter.ActorID = &actorID
	}
	if actionType := c.Query("action_type"); actionType != "" {
		filter.ActionType = &actionType
	}
	if module := c.Query("module"); module != "" {
		filter.Module = &module
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := helpers.ParseDatetime(startDateStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := helpers.ParseDatetime(endDateStr); err == nil {
			filter.EndDate = &t
		}
	}
	if role := c.Query("role"); role != "" {
		filter.RoleName = &role
	}
	if status := c.Query("status"); status != "" {
		isSuccess := status == "success"
		filter.IsSuccess = &isSuccess
	}
	return filter
}

// FetchAll Activity Logs
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := h.getFilter(c)

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, pagination.Cursor, filter)
	c.JSON(res.Status, res)
}

// Export Activity Logs
func (h *routeHandler) Export(c *gin.Context) {
	filter := h.getFilter(c)
	format := c.DefaultQuery("format", "csv")

	tokenData, exists := c.Get("token_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized: Token data missing"))
		return
	}
	userClaims := tokenData.(domain.JWTClaimUser)

	data, contentType, ext, err := h.Usecase.ExportData(c.Request.Context(), filter, format, userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, "failed to export data"))
		return
	}

	filename := "audit_log_" + helpers.GetCurrentTime().Format("20060102150405") + "." + ext
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}
