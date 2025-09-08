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
	rateLimiter middleware.RateLimiterMiddlewareInterface,
) {

	oauthGroup := e.Group("/api/v1/oauth")

	oauthGroup.GET("/google/login", oauthHandler.GoogleLoginAuth,
		rateLimiter.RateLimiter(RateLimitOauthLogin, RateLimitWindowOneMinute))

	oauthGroup.GET("/google/login/callback", oauthHandler.GoogleLoginCallback,
		rateLimiter.RateLimiter(RateLimitOauthLoginCallback, RateLimitWindowOneMinute))

	oauthGroup.GET("/google/register", oauthHandler.GoogleRegisterAuth,
		rateLimiter.RateLimiter(RateLimitOauthRegister, RateLimitWindowOneMinute))

	oauthGroup.GET("/google/register/callback", oauthHandler.GoogleRegisterCallback,
		rateLimiter.RateLimiter(RateLimitOauthRegisterCallback, RateLimitWindowOneMinute))

	protectedOAuth := e.Group("/api/v1/oauth", mid.CheckToken(), mid.CheckRole("Customer", "Super Admin"))

	protectedOAuth.POST("/link", oauthHandler.LinkAccount,
		rateLimiter.RateLimiter(RateLimitOauthLink, RateLimitWindowOneMinute))

	protectedOAuth.PATCH("/set-password", oauthHandler.SetPassword,
		rateLimiter.RateLimiter(RateLimitOauthSetPassword, RateLimitWindowOneMinute))

	protectedOAuth.DELETE("/unlink/:provider_id", oauthHandler.UnlinkAccount,
		rateLimiter.RateLimiter(RateLimitOauthUnlink, RateLimitWindowOneMinute))

	protectedOAuth.GET("/logout", oauthHandler.OAuthLogout,
		rateLimiter.RateLimiter(RateLimitOauthLogout, RateLimitWindowOneMinute))

	// oauthGroup.GET("/google", oauthHandler.GoogleAuth)
	// oauthGroup.GET("/google/callback", oauthHandler.GoogleCallback)

}
