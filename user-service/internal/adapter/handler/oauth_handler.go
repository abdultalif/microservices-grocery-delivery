package handler

import (
	"errors"
	"net/http"
	"strconv"
	"user-service/config"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"
	"user-service/internal/core/service"
	v "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type OAuthHandlerInterface interface {
	GoogleLoginAuth(c echo.Context) error
	GoogleLoginCallback(c echo.Context) error

	GoogleRegisterAuth(c echo.Context) error
	GoogleRegisterCallback(c echo.Context) error

	UnlinkAccount(c echo.Context) error

	// ini oauth yg register dan login di gabung dalam satu
	// GoogleAuth(c echo.Context) error
	// GoogleCallback(c echo.Context) error

	LinkAccount(c echo.Context) error
}

type oauthHandler struct {
	oauthService service.OAuthServiceInterface
	cfg          *config.Config
}

// UnlinkAccount implements OAuthHandlerInterface.
func (h *oauthHandler) UnlinkAccount(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] ChangePassword: user data not found in context")
		return c.JSON(http.StatusUnauthorized, response.ResponseAPI(false, http.StatusUnauthorized, "Unauthorized", nil))
	}

	// Get provider ID from path parameter
	providerIDStr := c.Param("provider_id")
	providerID, err := strconv.ParseInt(providerIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Invalid provider ID", nil))
	}

	err = h.oauthService.UnlinkOAuthAccount(ctx, user.UserID, providerID)
	if err != nil {
		log.Errorf("[OAuthHandler-UNLINK] UnlinkAccount: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
	}

	return c.JSON(http.StatusOK,
		response.ResponseAPI(true, http.StatusOK, "Account unlinked successfully", nil))
}

// GoogleRegisterAuth implements OAuthHandlerInterface.
func (h *oauthHandler) GoogleRegisterAuth(c echo.Context) error {
	ctx := c.Request().Context()

	state, err := h.oauthService.GenerateState()
	if err != nil {
		log.Errorf("[OAuthHandler-REG-1] GoogleRegisterAuth generate state: %v", err)
		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to generate auth state", nil))
	}

	redirectPath := "/api/v1/oauth/google/register/callback"

	authURL := h.oauthService.GetGoogleRegisterURL(ctx, state, redirectPath)

	if c.Request().Header.Get("Accept") == "application/json" || c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, response.ResponseAPI(true, http.StatusOK, "Google registration URL generated", map[string]string{
			"auth_url": authURL,
			"state":    state, // Kirim state asli tanpa suffix, suffix sudah ditambah di service
		}))
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleRegisterCallback implements OAuthHandlerInterface.
func (h *oauthHandler) GoogleRegisterCallback(c echo.Context) error {
	ctx := c.Request().Context()

	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")

	if errorParam != "" {
		log.Errorf("[OAuthHandler-REG-2] GoogleRegisterCallback OAuth error: %s", errorParam)

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/register/error?error=oauth_cancelled")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "OAuth registration cancelled", nil))
	}

	if code == "" {
		log.Errorf("[OAuthHandler-REG-3] GoogleRegisterCallback missing authorization code")

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/register/error?error=missing_code")
		}

		return c.JSON(http.StatusBadRequest,
			response.ResponseAPI(false, http.StatusBadRequest, "Missing authorization code", nil))
	}

	user, token, err := h.oauthService.HandleGoogleRegisterCallback(ctx, code, state)
	if err != nil {
		log.Errorf("[OAuthHandler-REG-4] GoogleRegisterCallback handle callback: %v", err)

		// Special handling for "user already exists" error
		if errors.Is(err, errs.ErrUserExist) {
			if h.isWebRequest(c) {
				return c.Redirect(http.StatusTemporaryRedirect,
					h.cfg.App.UrlFrontend+"/auth/login?message=user_exists")
			}
			return c.JSON(http.StatusConflict,
				response.ResponseAPI(false, http.StatusConflict, "User already exists. Please use login instead.", nil))
		}

		if h.isWebRequest(c) {
			return c.Redirect(http.StatusTemporaryRedirect,
				h.cfg.App.UrlFrontend+"/auth/register/error?error=registration_failed")
		}

		return c.JSON(http.StatusInternalServerError,
			response.ResponseAPI(false, http.StatusInternalServerError, "OAuth registration failed", nil))
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
		"is_new_user":  true,
	}

	if h.isWebRequest(c) {
		redirectURL := h.cfg.App.UrlFrontend + "/auth/register/success?token=" + token
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	return c.JSON(http.StatusCreated,
		response.ResponseAPI(true, http.StatusCreated, "Google OAuth registration successful", resData))

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
		if errors.Is(err, errs.ErrUserNotFound) {
			if h.isWebRequest(c) {
				return c.Redirect(http.StatusTemporaryRedirect,
					h.cfg.App.UrlFrontend+"/auth/register?message=account_not_found")
			}
			return c.JSON(http.StatusNotFound,
				response.ResponseAPI(false, http.StatusNotFound, "No account found. Please register first.", nil))
		} else if errors.Is(err, errs.ErrGoogleUnlinked) {

			if h.isWebRequest(c) {
				return c.Redirect(http.StatusTemporaryRedirect,
					h.cfg.App.UrlFrontend+"/auth/login/error?error=google_unlinked")
			}
			return c.JSON(http.StatusUnauthorized,
				response.ResponseAPI(false, http.StatusUnauthorized, "Google account not linked", nil))

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

	authURL := h.oauthService.GetGoogleLoginURL(ctx, state, redirectPath)

	if c.Request().Header.Get("Accept") == "application/json" || c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, response.ResponseAPI(true, http.StatusOK, "Google login URL generated", map[string]string{
			"auth_url": authURL,
			"state":    state, // Kirim state asli tanpa suffix
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
// func (h *oauthHandler) GoogleAuth(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	state, err := h.oauthService.GenerateState()
// 	if err != nil {
// 		log.Errorf("[OAuthHandler-1] GoogleAuth generate state: %v", err)
// 		return c.JSON(http.StatusInternalServerError,
// 			response.ResponseAPI(false, http.StatusInternalServerError, "Failed to generate auth state", nil))
// 	}

// 	authURL := h.oauthService.GetGoogleAuthURL(ctx, state)

// 	if c.Request().Header.Get("Accept") == "application/json" || c.QueryParam("format") == "json" {
// 		return c.JSON(http.StatusOK, response.ResponseAPI(true, http.StatusOK, "Google auth URL generated", map[string]string{
// 			"auth_url": authURL,
// 			"state":    state,
// 		}))
// 	}

// 	return c.Redirect(http.StatusTemporaryRedirect, authURL)
// }

// GoogleCallback implements OAuthHandlerInterface.
// func (h *oauthHandler) GoogleCallback(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	code := c.QueryParam("code")
// 	state := c.QueryParam("state")
// 	errorParam := c.QueryParam("error")

// 	if errorParam != "" {
// 		log.Errorf("[OAuthHandler-2] GoogleCallback OAuth error: %s", errorParam)

// 		if h.isWebRequest(c) {
// 			return c.Redirect(http.StatusTemporaryRedirect,
// 				h.cfg.App.UrlFrontend+"/auth/error?error=oauth_cancelled")
// 		}

// 		return c.JSON(http.StatusBadRequest,
// 			response.ResponseAPI(false, http.StatusBadRequest, "OAuth authentication cancelled", nil))
// 	}

// 	if code == "" {
// 		log.Errorf("[OAuthHandler-3] GoogleCallback missing authorization code")

// 		if h.isWebRequest(c) {
// 			return c.Redirect(http.StatusTemporaryRedirect,
// 				h.cfg.App.UrlFrontend+"/auth/error?error=missing_code")
// 		}

// 		return c.JSON(http.StatusBadRequest,
// 			response.ResponseAPI(false, http.StatusBadRequest, "Missing authorization code", nil))
// 	}

// 	if state == "" {
// 		log.Errorf("[OAuthHandler-4] GoogleCallback missing state parameter")

// 		if h.isWebRequest(c) {
// 			return c.Redirect(http.StatusTemporaryRedirect,
// 				h.cfg.App.UrlFrontend+"/auth/error?error=missing_state")
// 		}

// 		return c.JSON(http.StatusBadRequest,
// 			response.ResponseAPI(false, http.StatusBadRequest, "Missing state parameter", nil))
// 	}

// 	user, token, err := h.oauthService.HandleGoogleCallback(ctx, code, state)
// 	if err != nil {
// 		log.Errorf("[OAuthHandler-5] GoogleCallback handle callback: %v", err)

// 		return c.JSON(http.StatusInternalServerError,
// 			response.ResponseAPI(false, http.StatusInternalServerError, "OAuth authentication failed", nil))
// 	}

// 	resData := map[string]interface{}{
// 		"id":           user.ID,
// 		"name":         user.Name,
// 		"email":        user.Email,
// 		"role":         user.RoleName,
// 		"lat":          user.Lat,
// 		"lng":          user.Lng,
// 		"phone":        user.Phone,
// 		"photo":        user.Photo,
// 		"access_token": token,
// 		"oauth":        true,
// 		"provider":     "google",
// 	}

// 	if h.isWebRequest(c) {
// 		redirectURL := h.cfg.App.UrlFrontend + "/auth/success?token=" + token
// 		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
// 	}

// 	return c.JSON(http.StatusOK,
// 		response.ResponseAPI(true, http.StatusOK, "Google OAuth authentication successful", resData))
// }

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
