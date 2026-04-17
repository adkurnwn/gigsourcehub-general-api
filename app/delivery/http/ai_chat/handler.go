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
