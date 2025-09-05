package router

import (
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
)

func OauthRouter(
	e *echo.Echo,
	oauthHandler handler.OAuthHandlerInterface,
	mid middleware.AuthMiddlewareInterface,
	// rateLimiter middleware.RateLimiterMiddlewareInterface,
) {

	oauthGroup := e.Group("/api/v1/oauth")

	oauthGroup.GET("/google/login", oauthHandler.GoogleLoginAuth)
	oauthGroup.GET("/google/login/callback", oauthHandler.GoogleLoginCallback)

	oauthGroup.GET("/google/register", oauthHandler.GoogleRegisterAuth)
	oauthGroup.GET("/google/register/callback", oauthHandler.GoogleRegisterCallback)

	protectedOAuth := e.Group("/api/v1/oauth", mid.CheckToken(), mid.CheckRole("Customer", "Super Admin"))
	protectedOAuth.DELETE("/unlink/:provider_id", oauthHandler.UnlinkAccount)

	// oauthGroup.GET("/google", oauthHandler.GoogleAuth)

	// oauthGroup.GET("/google/callback", oauthHandler.GoogleCallback)

	// oauthGroup.POST("/link", oauthHandler.LinkAccount)

}
