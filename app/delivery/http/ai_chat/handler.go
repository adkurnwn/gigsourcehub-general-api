package http_aichat

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase domain.AIChatAppUsecase
}

func NewAIChatHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.AIChatAppUsecase) {
	handler := &routeHandler{
		Usecase: uc,
	}

	api := r.Group("/ai-chat", mdl.Auth(), mdl.AuthAdmin())
	{
		api.GET("", handler.FetchMyChats)
		api.POST("", handler.CreateChat)
		api.DELETE("/:id", handler.DeleteChat)
		api.GET("/:id/messages", handler.FetchChatMessages)
		api.POST("/:id/messages", handler.StoreChatMessage)
	}
}

// @Security BearerAuth
// @Summary Fetch My Chats
// @Description Fetch My Chats
// @Tags AI Chat
// @Accept json
// @Produce json
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /ai-chat [get]
func (h *routeHandler) FetchMyChats(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	res := h.Usecase.FetchMyChats(c.Request.Context(), tokenData.UserID)
	c.JSON(res.Status, res)
}

// @Security BearerAuth
// @Summary Create Chat
// @Description Create Chat
// @Tags AI Chat
// @Accept json
// @Produce json
// @Param request body object true "Create Chat Request (first_query)"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /ai-chat [post]
func (h *routeHandler) CreateChat(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	var req struct {
		FirstQuery string `json:"first_query"`
	}
	_ = c.ShouldBindJSON(&req)

	res := h.Usecase.CreateChat(c.Request.Context(), tokenData.UserID, req.FirstQuery)
	c.JSON(res.Status, res)
}

// @Security BearerAuth
// @Summary Delete Chat
// @Description Delete Chat
// @Tags AI Chat
// @Accept json
// @Produce json
// @Param id path string true "Chat ID"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /ai-chat/{id} [delete]
func (h *routeHandler) DeleteChat(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	chatID := c.Param("id")
	res := h.Usecase.DeleteChat(c.Request.Context(), tokenData.UserID, chatID)
	c.JSON(res.Status, res)
}

// @Security BearerAuth
// @Summary Fetch Chat Messages
// @Description Fetch Chat Messages
// @Tags AI Chat
// @Accept json
// @Produce json
// @Param id path string true "Chat ID"
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /ai-chat/{id}/messages [get]
func (h *routeHandler) FetchChatMessages(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	chatID := c.Param("id")
	res := h.Usecase.FetchChatMessages(c.Request.Context(), tokenData.UserID, chatID)
	c.JSON(res.Status, res)
}

// @Security BearerAuth
// @Summary Store Chat Message
// @Description Store Chat Message
// @Tags AI Chat
// @Accept json
// @Produce json
// @Param id path string true "Chat ID"
// @Param request body object true "Store Message Request (role, content, is_last)"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /ai-chat/{id}/messages [post]
func (h *routeHandler) StoreChatMessage(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	chatID := c.Param("id")

	var req struct {
		Role    string `json:"role" binding:"required"`
		Content string `json:"content" binding:"required"`
		IsLast  bool   `json:"is_last"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "role and content are required"))
		return
	}

	res := h.Usecase.StoreChatMessage(c.Request.Context(), tokenData.UserID, chatID, req.Role, req.Content, req.IsLast)
	c.JSON(res.Status, res)
}
