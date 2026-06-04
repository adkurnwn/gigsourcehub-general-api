package http_notification

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.NotificationAppUsecase
	Middleware middleware.Middleware
}

func NewNotificationHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.NotificationAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Middleware: mdl,
	}

	notifRoute := r.Group("/notifications", mdl.Auth())
	notifRoute.GET("", handler.FetchMyNotifications)
	notifRoute.GET("/unread-count", handler.GetUnreadCount)
	notifRoute.PATCH("/:id/read", handler.MarkAsRead)
	notifRoute.PATCH("/read-all", handler.MarkAllAsRead)
}

// FetchMyNotifications fetches paginated notifications for the authenticated user.
// @Summary Fetch my notifications
// @Description Get paginated list of in-app notifications for the authenticated user
// @Tags Notification
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(20)
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /notifications [get]
// @Security BearerAuth
func (h *routeHandler) FetchMyNotifications(ctx *gin.Context) {
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID
	pagination := helpers.GetPagination(ctx)

	result := h.Usecase.FetchMyNotifications(ctx.Request.Context(), userID, pagination.Page, pagination.Limit)
	ctx.JSON(result.Status, result)
}

// GetUnreadCount returns the count of unread notifications for the authenticated user.
// @Summary Get unread notification count
// @Description Returns the number of unread in-app notifications for the authenticated user
// @Tags Notification
// @Produce json
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /notifications/unread-count [get]
// @Security BearerAuth
func (h *routeHandler) GetUnreadCount(ctx *gin.Context) {
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.GetUnreadCount(ctx.Request.Context(), userID)
	ctx.JSON(result.Status, result)
}

// MarkAsRead marks a single notification as read.
// @Summary Mark notification as read
// @Description Mark a specific notification as read for the authenticated user
// @Tags Notification
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /notifications/{id}/read [patch]
// @Security BearerAuth
func (h *routeHandler) MarkAsRead(ctx *gin.Context) {
	notifID := ctx.Param("id")
	if notifID == "" {
		ctx.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Notification ID is required"))
		return
	}

	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.MarkAsRead(ctx.Request.Context(), userID, notifID)
	ctx.JSON(result.Status, result)
}

// MarkAllAsRead marks all notifications as read for the authenticated user.
// @Summary Mark all notifications as read
// @Description Mark all in-app notifications as read for the authenticated user
// @Tags Notification
// @Produce json
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /notifications/read-all [patch]
// @Security BearerAuth
func (h *routeHandler) MarkAllAsRead(ctx *gin.Context) {
	userClaim, exists := ctx.Get("token_data")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "User ID not found in context"))
		return
	}
	userID := userClaim.(domain.JWTClaimUser).UserID

	result := h.Usecase.MarkAllAsRead(ctx.Request.Context(), userID)
	ctx.JSON(result.Status, result)
}
