package handler

import (
	"net/http"
	"user-service/config"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OAuthHandlerInterface interface {
	GoogleAuth(c echo.Context) error
	GoogleCallback(c echo.Context) error
}

type oauthHandler struct {
	oauthService service.OAuthServiceInterface
	cfg          *config.Config
}

func NewOAuthHandler(oauthService service.OAuthServiceInterface, cfg *config.Config) OAuthHandlerInterface {
	return &oauthHandler{
		oauthService: oauthService,
		cfg:          cfg,
	}
}

// GoogleAuth implements OAuthHandlerInterface.
func (h *oauthHandler) GoogleAuth(c echo.Context) error {
	ctx := c.Request().Context()

	state, err := h.oauthService.GenerateState()
	if err != nil {
		log.Errorf("[OAuthHandler-1] GoogleAuth generate state: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to generate auth state", nil))
	}

	authURL := h.oauthService.GetGoogleAuthURL(ctx, state)

	if c.Request().Header.Get("Accept") == "application/json" || c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, response.ResponseAPI(true, http.StatusOK, "Google auth URL generated", map[string]string{
			"auth_url": authURL,
			"state":    state,
		}))
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback implements OAuthHandlerInterface.
func (h *oauthHandler) GoogleCallback(c echo.Context) error {
	ctx := c.Request().Context()

	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")

	if errorParam != "" {
		log.Errorf("[OAuthHandler-2] GoogleCallback OAuth error: %s", errorParam)

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/error?error=oauth_cancelled")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "OAuth authentication cancelled", nil))
	}

	if code == "" {
		log.Errorf("[OAuthHandler-3] GoogleCallback missing authorization code")

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/error?error=missing_code")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Missing authorization code", nil))
	}

	if state == "" {
		log.Errorf("[OAuthHandler-4] GoogleCallback missing state parameter")

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/error?error=missing_state")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Missing state parameter", nil))
	}

	user, token, err := h.oauthService.HandleGoogleCallback(ctx, code, state)
	if err != nil {
		log.Errorf("[OAuthHandler-5] GoogleCallback handle callback: %v", err)

		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "OAuth authentication failed", nil))
	}

	resData := map[string]interface{}{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"role":         user.RoleName,
		"lat":          user.Lat,
		"lng":          user.Lng,
		"phone":        user.Phone,
		"photo":        user.Photo,
		"access_token": token,
		"oauth":        true,
		"provider":     "google",
	}

	if h.isWebRequest(c) {
		redirectURL := h.cfg.App.UrlFrontend + "/auth/success?token=" + token
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Google OAuth authentication successful", resData))
}

// isWebRequest determines if the request is coming from a web browser
func (h *oauthHandler) isWebRequest(c echo.Context) bool {
	userAgent := c.Request().Header.Get("User-Agent")
	accept := c.Request().Header.Get("Accept")

	return (accept != "application/json" &&
		userAgent != "" &&
		(contains(userAgent, "Mozilla") || contains(userAgent, "Chrome") || contains(userAgent, "Safari")))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
