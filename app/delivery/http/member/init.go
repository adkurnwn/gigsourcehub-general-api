package http_member

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"

	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.MemberAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewRouteHandler(route *gin.RouterGroup, middleware middleware.Middleware, u domain.MemberAppUsecase) {
	handler := &routeHandler{
		Usecase:    u,
		Route:      route,
		Middleware: middleware,
	}

	handler.handleAuthRoute("/auth")
	handler.handleUserRoute("/users")
	route.GET("/profile", middleware.Auth(), handler.GetProfile)
	route.PUT("/profile", middleware.Auth(), handler.UpdateProfile)
	route.PUT("/profile/password", middleware.Auth(), handler.UpdatePassword)
	route.POST("/profile/picture", middleware.Auth(), handler.UploadProfilePicture)
}
