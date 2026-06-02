package http_chat

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	jwt_helper "github.com/adkurnwn/gigsourcehub-general-api/helpers/jsonwebtoken"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type routeHandler struct {
	Usecase    domain.ChatAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
	Hub        *Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins; tighten in production
	},
}

func NewChatHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.ChatAppUsecase, hub *Hub) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
		Hub:        hub,
	}

	r.POST("/chats/upload-offering", mdl.Auth(), mdl.AuthRole("Admin"), handler.UploadOffering)

	// Admin-only: create conversation
	adminAPI := r.Group("/chats", mdl.Auth(), mdl.AuthAdmin())
	adminAPI.POST("", handler.CreateConversation)
	adminAPI.POST("/start", handler.StartConversation)

	// Admin + Candidate: shared endpoints
	chatAPI := r.Group("/chats", mdl.Auth(), mdl.AuthRole("Admin", "Candidate"))
	chatAPI.GET("", handler.FetchMyConversations)
	chatAPI.GET("/:id", handler.GetConversation)
	chatAPI.POST("/:id/messages", handler.SendMessage)
	chatAPI.GET("/:id/messages", handler.FetchMessages)
	chatAPI.PUT("/:id/read", handler.MarkAsRead)

	// WebSocket endpoint (auth via query param)
	r.GET("/ws/chat", handler.HandleWebSocket)
}

// CreateConversation godoc
// @Security BearerAuth
// @Summary Create a conversation with a candidate
// @Description Admin creates a new chat conversation scoped to a subrequest
// @Tags Chat
// @Accept json
// @Produce json
// @Param request body request_model.CreateConversationRequest true "Create Conversation"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /chats [post]
func (h *routeHandler) CreateConversation(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	var req request_model.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.CreateConversation(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}

// StartConversation godoc
// @Security BearerAuth
// @Summary Start a chat with an assigned candidate
// @Description Admin creates or reuses a conversation and marks the candidate recruitment status as Contacted
// @Tags Chat
// @Accept json
// @Produce json
// @Param request body request_model.CreateConversationRequest true "Create Conversation"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /chats/start [post]
func (h *routeHandler) StartConversation(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	var req request_model.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.StartConversation(c.Request.Context(), tokenData.UserID, req)
	c.JSON(res.Status, res)
}

// UploadOffering godoc
// @Security BearerAuth
// @Summary Upload offering file
// @Description Upload a PDF offering file and store it as a chat message content payload
// @Tags Chat
// @Accept multipart/form-data
// @Produce json
// @Param conversation_id formData string true "Conversation ID"
// @Param file formData file true "Offering PDF file"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /chats/upload-offering [post]
func (h *routeHandler) UploadOffering(c *gin.Context) {
	conversationID := c.PostForm("conversation_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "conversation_id is required"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "file is required"))
		return
	}

	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)

	res := h.Usecase.UploadOffering(c.Request.Context(), tokenData.UserID, conversationID, file)
	c.JSON(res.Status, res)
}

// FetchMyConversations godoc
// @Security BearerAuth
// @Summary List my conversations
// @Description Get paginated list of conversations for the authenticated user
// @Tags Chat
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Success 200 {object} response.Base
// @Router /chats [get]
func (h *routeHandler) FetchMyConversations(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)
	pagination := helpers.GetPagination(c)

	res := h.Usecase.FetchMyConversations(c.Request.Context(), tokenData.UserID, pagination.Page, pagination.Limit)
	c.JSON(res.Status, res)
}

// GetConversation godoc
// @Security BearerAuth
// @Summary Get conversation detail
// @Description Get a specific conversation by ID
// @Tags Chat
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /chats/{id} [get]
func (h *routeHandler) GetConversation(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)
	conversationID := c.Param("id")

	res := h.Usecase.GetConversation(c.Request.Context(), tokenData.UserID, conversationID)
	c.JSON(res.Status, res)
}

// SendMessage godoc
// @Security BearerAuth
// @Summary Send a message
// @Description Send a message in a conversation
// @Tags Chat
// @Accept json
// @Produce json
// @Param id path string true "Conversation ID"
// @Param request body request_model.SendMessageRequest true "Send Message"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /chats/{id}/messages [post]
func (h *routeHandler) SendMessage(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)
	conversationID := c.Param("id")

	var req request_model.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	res := h.Usecase.SendMessage(c.Request.Context(), tokenData.UserID, conversationID, req)
	c.JSON(res.Status, res)
}

// FetchMessages godoc
// @Security BearerAuth
// @Summary List messages in a conversation
// @Description Get paginated messages for a conversation (newest first, default 20)
// @Tags Chat
// @Produce json
// @Param id path string true "Conversation ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(20)
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /chats/{id}/messages [get]
func (h *routeHandler) FetchMessages(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)
	conversationID := c.Param("id")
	pagination := helpers.GetPagination(c)

	// Default limit to 20 for messages
	limit := pagination.Limit
	if c.Query("limit") == "" {
		limit = 20
	}

	res := h.Usecase.FetchMessages(c.Request.Context(), tokenData.UserID, conversationID, pagination.Page, limit)
	c.JSON(res.Status, res)
}

// MarkAsRead godoc
// @Security BearerAuth
// @Summary Mark messages as read
// @Description Mark all unread messages from the other participant as read
// @Tags Chat
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /chats/{id}/read [put]
func (h *routeHandler) MarkAsRead(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	tokenData := claims.(domain.JWTClaimUser)
	conversationID := c.Param("id")

	res := h.Usecase.MarkAsRead(c.Request.Context(), tokenData.UserID, conversationID)
	c.JSON(res.Status, res)
}

// HandleWebSocket godoc
// @Summary WebSocket connection for real-time chat
// @Description Connect via WebSocket. Pass JWT token as query param: /ws/chat?token=<jwt>
// @Tags Chat
// @Param token query string true "JWT token"
// @Success 101 "Switching Protocols"
// @Router /ws/chat [get]
func (h *routeHandler) HandleWebSocket(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Token is required"))
		return
	}

	// Validate JWT
	secret := jwt_helper.GetJwtCredential().Member.Secret
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaimUser{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Invalid token"))
		return
	}

	jwtClaims, ok := token.Claims.(*domain.JWTClaimUser)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Invalid token claims"))
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:    h.Hub,
		conn:   conn,
		userID: jwtClaims.UserID,
		send:   make(chan []byte, 256),
	}
	h.Hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
