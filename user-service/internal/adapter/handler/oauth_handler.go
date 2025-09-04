package handler

import (
	"net/http"
	"user-service/config"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	v "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OAuthHandlerInterface interface {
	GoogleLoginAuth(c echo.Context) error
	GoogleLoginCallback(c echo.Context) error

	GoogleAuth(c echo.Context) error
	GoogleCallback(c echo.Context) error

	LinkAccount(c echo.Context) error
}

type oauthHandler struct {
	oauthService service.OAuthServiceInterface
	cfg          *config.Config
}

// GoogleLoginCallback implements OAuthHandlerInterface.
func (h *oauthHandler) GoogleLoginCallback(c echo.Context) error {

	ctx := c.Request().Context()

	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")

	if errorParam != "" {
		log.Errorf("[OAuthHandler-LOGIN-2] GoogleLoginCallback OAuth error: %s", errorParam)

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/login/error?error=oauth_cancelled")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "OAuth login cancelled", nil))
	}

	if code == "" {
		log.Errorf("[OAuthHandler-LOGIN-3] GoogleLoginCallback missing authorization code")

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/login/error?error=missing_code")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Missing authorization code", nil))
	}

	user, token, err := h.oauthService.HandleGoogleLoginCallback(ctx, code, state)
	if err != nil {
		log.Errorf("[OAuthHandler-LOGIN-4] GoogleLoginCallback handle callback: %v", err)

		// Special handling for "user not found" error
		if err.Error() == "user not found" {
			if h.isWebRequest(c) {
				return c.Redirect(http.StatusTemporaryRedirect,
					h.cfg.App.UrlFrontend+"/auth/register?message=account_not_found")
			}
			return c.JSON(http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "No account found. Please register first.", nil))
		}

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/login/error?error=login_failed")
		}

		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "OAuth login failed", nil))
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
		"is_new_user":  false,
	}

	if h.isWebRequest(c) {
		redirectURL := h.cfg.App.UrlFrontend + "/auth/login/success?token=" + token
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Google OAuth login successful", resData))

}

// GoogleLoginAuth implements OAuthHandlerInterface.
func (h *oauthHandler) GoogleLoginAuth(c echo.Context) error {

	ctx := c.Request().Context()

	state, err := h.oauthService.GenerateState()
	if err != nil {
		log.Errorf("[OAuthHandler-LOGIN-1] GoogleLoginAuth generate state: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to generate auth state", nil))
	}

	redirectPath := "/api/v1/oauth/google/login/callback"
	if c.QueryParam("type") == "register" {
		redirectPath = "/api/v1/oauth/google/register/callback"
	}

	authURL := h.oauthService.GetGoogleLoginURL(ctx, state, redirectPath)

	if c.Request().Header.Get("Accept") == "application/json" || c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, response.ResponseAPI(true, http.StatusOK, "Google login URL generated", map[string]string{
			"auth_url": authURL,
			"state":    state + "_login",
		}))
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)

}

// LinkAccount implements OAuthHandlerInterface.
func (h *oauthHandler) LinkAccount(c echo.Context) error {

	ctx := c.Request().Context()
	req := request.LinkAccountRequest{}

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		return c.JSON(http.StatusUnauthorized,
			response.ResponseAPI(false, http.StatusUnauthorized, "Unauthorized", nil))
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[AuthHandler-1] SignIn: %v", err)
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid request body format", nil))
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[AuthHandler-2] SignIn: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			return c.JSON(http.StatusUnprocessableEntity,
				response.ResponseAPI(false, http.StatusUnprocessableEntity, ve.Errors, nil))
		}

		return c.JSON(http.StatusUnprocessableEntity,
			response.ResponseAPI(false, http.StatusUnprocessableEntity, err.Error(), nil))
	}

	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()

	err := h.oauthService.LinkOAuthAccount(ctx, user.UserID, req.Code, req.State, req.Provider, userAgent, ipAddress)
	if err != nil {
		log.Errorf("[OAuthHandler-6] LinkAccount: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Account linked successfully", nil))

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

func NewOAuthHandler(oauthService service.OAuthServiceInterface, cfg *config.Config) OAuthHandlerInterface {
	return &oauthHandler{
		oauthService: oauthService,
		cfg:          cfg,
	}
}
