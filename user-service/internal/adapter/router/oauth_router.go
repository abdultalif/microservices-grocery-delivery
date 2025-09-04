package router

import (
	"user-service/internal/adapter/handler"

	"github.com/labstack/echo/v4"
)

func OauthRouter(
	e *echo.Echo,
	oauthHandler handler.OAuthHandlerInterface,
	// rateLimiter middleware.RateLimiterMiddlewareInterface,
) {

	oauthGroup := e.Group("/api/v1/oauth")

	oauthGroup.GET("/google/login", oauthHandler.GoogleLoginAuth)
	oauthGroup.GET("/google/login/callback", oauthHandler.GoogleLoginCallback)

	// oauthGroup.GET("/google", oauthHandler.GoogleAuth)

	// oauthGroup.GET("/google/callback", oauthHandler.GoogleCallback)

	// oauthGroup.POST("/link", oauthHandler.LinkAccount)

}
