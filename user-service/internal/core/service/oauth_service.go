package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"user-service/config"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	errs "user-service/internal/core/domain/error"

	"github.com/labstack/gommon/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthServiceInterface interface {
	GetGoogleLoginURL(ctx context.Context, state, redirectPath string) string
	HandleGoogleLoginCallback(ctx context.Context, code, state string) (*entity.UserEntity, string, error)

	GetGoogleAuthURL(ctx context.Context, state string) string
	HandleGoogleCallback(ctx context.Context, code, state string) (*entity.UserEntity, string, error)
	GenerateState() (string, error)

	LinkOAuthAccount(ctx context.Context, userID int64, code, state, provider, userAgent, ipAddress string) error
}

type OAuthService struct {
	userRepo     repository.UserRepositoryInterface
	authRepo     repository.AuthRepositoryInterface
	oauthRepo    repository.OAuthRepositoryInterface
	cfg          *config.Config
	jwtService   JwtServiceInterface
	googleConfig *oauth2.Config
}

// HandleGoogleLoginCallback implements OAuthServiceInterface.
func (o *OAuthService) HandleGoogleLoginCallback(ctx context.Context, code string, state string) (*entity.UserEntity, string, error) {

	if !strings.HasSuffix(state, "_login") {
		return nil, "", fmt.Errorf("invalid state for login")
	}

	token, err := o.googleConfig.Exchange(ctx, code)
	if err != nil {
		log.Errorf("[OAuthService-LOGIN-1] HandleGoogleLoginCallback exchange token: %v", err)
		return nil, "", err
	}

	googleUser, err := o.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		log.Errorf("[OAuthService-LOGIN-2] HandleGoogleLoginCallback get user info: %v", err)
		return nil, "", err
	}

	// FIRST: Try to find existing OAuth connection
	existingOAuth, err := o.oauthRepo.GetOAuthProviderByProviderAndUserID(ctx, "google", googleUser.ID)
	if err != nil {
		log.Errorf("[OAuthService-LOGIN-3] HandleGoogleLoginCallback get oauth provider: %v", err)
		return nil, "", err
	}

	var user *entity.UserEntity

	if existingOAuth != nil {
		// OAuth connection exists - get the user
		user, err = o.userRepo.GetUserByID(ctx, existingOAuth.UserID)
		if err != nil {
			logOAuthActivity(ctx, existingOAuth.UserID, "google", "login", "failed", err.Error(), "", "")
			log.Errorf("[OAuthService-LOGIN-4] HandleGoogleLoginCallback get user by id: %v", err)
			return nil, "", err
		}

		// Update OAuth tokens
		oauthProvider := &entity.OAuthProviderEntity{
			ID:              existingOAuth.ID,
			UserID:          user.ID,
			Provider:        "google",
			ProviderUserID:  googleUser.ID,
			ProviderEmail:   googleUser.Email,
			ProviderName:    googleUser.Name,
			ProviderPicture: &googleUser.Picture,
			AccessToken:     &token.AccessToken,
			RefreshToken:    &token.RefreshToken,
		}

		if token.Expiry.After(time.Now()) {
			oauthProvider.TokenExpiresAt = &token.Expiry
		}

		err = o.oauthRepo.UpsertOAuthProvider(ctx, oauthProvider)
		if err != nil {
			logOAuthActivity(ctx, user.ID, "google", "login", "failed", err.Error(), "", "")
			log.Errorf("[OAuthService-LOGIN-5] HandleGoogleLoginCallback update oauth provider: %v", err)
			return nil, "", err
		}

		// Update user photo if empty
		if user.Photo == "" && googleUser.Picture != "" {
			user.Photo = googleUser.Picture
			err = o.userRepo.UpdateUser(ctx, user)
			if err != nil {
				log.Warnf("[OAuthService-LOGIN-6] HandleGoogleLoginCallback update user photo: %v", err)
			}
		}
	} else {
		// No OAuth connection - check if user exists by email
		existingUser, err := o.authRepo.GetUserByEmail(ctx, googleUser.Email)
		if err != nil && err != errs.ErrUserNotFound {
			log.Errorf("[OAuthService-LOGIN-7] HandleGoogleLoginCallback check existing user: %v", err)
			return nil, "", err
		}

		if existingUser == nil {
			// User doesn't exist at all - redirect to register
			logOAuthActivity(ctx, 0, "google", "login", "failed", "user not found, should register first", "", "")
			return nil, "", fmt.Errorf("no account found with this Google account. Please register first")
		}

		// User exists but no OAuth connection - link them
		user = existingUser

		// CREATE new OAuth provider record for existing user
		oauthProvider := &entity.OAuthProviderEntity{
			UserID:          user.ID,
			Provider:        "google",
			ProviderUserID:  googleUser.ID,
			ProviderEmail:   googleUser.Email,
			ProviderName:    googleUser.Name,
			ProviderPicture: &googleUser.Picture,
			AccessToken:     &token.AccessToken,
			RefreshToken:    &token.RefreshToken,
		}

		if token.Expiry.After(time.Now()) {
			oauthProvider.TokenExpiresAt = &token.Expiry
		}

		err = o.oauthRepo.CreateOAuthProvider(ctx, oauthProvider)
		if err != nil {
			logOAuthActivity(ctx, user.ID, "google", "login", "failed", err.Error(), "", "")
			log.Errorf("[OAuthService-LOGIN-8] HandleGoogleLoginCallback create oauth provider: %v", err)
			return nil, "", err
		}
	}

	// Log successful login
	logOAuthActivity(ctx, user.ID, "google", "login", "success", "", "", "")

	// Generate JWT token
	jwtToken, err := o.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Errorf("[OAuthService-LOGIN-9] HandleGoogleLoginCallback generate jwt token: %v", err)
		return nil, "", err
	}

	// Create session
	err = o.createUserSession(ctx, user, jwtToken, "google")
	if err != nil {
		log.Errorf("[OAuthService-LOGIN-10] HandleGoogleLoginCallback create session: %v", err)
		return nil, "", err
	}

	return user, jwtToken, nil

}

// GetGoogleLoginURL implements OAuthServiceInterface.
func (o *OAuthService) GetGoogleLoginURL(ctx context.Context, state, redirectPath string) string {
	cfg := *o.googleConfig
	cfg.RedirectURL = "http://localhost:" + o.cfg.App.AppPort + redirectPath
	return cfg.AuthCodeURL(state+"_login", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// LinkOAuthAccount implements OAuthServiceInterface.
func (o *OAuthService) LinkOAuthAccount(ctx context.Context, userID int64, code, state, provider, userAgent, ipAddress string) error {
	if provider != "google" {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Exchange code for token
	token, err := o.googleConfig.Exchange(ctx, code)
	if err != nil {
		o.logOAuthActivity(ctx, userID, provider, "link", "failed", err.Error(), userAgent, ipAddress)
		log.Errorf("[OAuthService-13] LinkOAuthAccount exchange token: %v", err)
		return err
	}

	// Get user info from provider
	googleUser, err := o.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		o.logOAuthActivity(ctx, userID, provider, "link", "failed", err.Error(), userAgent, ipAddress)
		log.Errorf("[OAuthService-14] LinkOAuthAccount get user info: %v", err)
		return err
	}

	// Check if this OAuth account is already linked to another user
	existingOAuth, err := o.oauthRepo.GetOAuthProviderByProviderAndUserID(ctx, provider, googleUser.ID)
	if err != nil {
		log.Errorf("[OAuthService-15] LinkOAuthAccount check existing oauth: %v", err)
		return err
	}

	if existingOAuth != nil && existingOAuth.UserID != userID {
		o.logOAuthActivity(ctx, userID, provider, "link", "failed", "account already linked to another user", userAgent, ipAddress)
		return fmt.Errorf("this %s account is already linked to another user", provider)
	}

	// Create or update OAuth provider
	oauthProvider := &entity.OAuthProviderEntity{
		UserID:          userID,
		Provider:        provider,
		ProviderUserID:  googleUser.ID,
		ProviderEmail:   googleUser.Email,
		ProviderName:    googleUser.Name,
		ProviderPicture: &googleUser.Picture,
		AccessToken:     &token.AccessToken,
		RefreshToken:    &token.RefreshToken,
	}

	if token.Expiry.After(time.Now()) {
		oauthProvider.TokenExpiresAt = &token.Expiry
	}

	err = o.oauthRepo.UpsertOAuthProvider(ctx, oauthProvider)
	if err != nil {
		o.logOAuthActivity(ctx, userID, provider, "link", "failed", err.Error(), userAgent, ipAddress)
		log.Errorf("[OAuthService-16] LinkOAuthAccount upsert oauth provider: %v", err)
		return err
	}

	o.logOAuthActivity(ctx, userID, provider, "link", "success", "", userAgent, ipAddress)
	return nil

}

// GenerateState implements OAuthServiceInterface.
func (o *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleAuthURL implements OAuthServiceInterface.
func (o *OAuthService) GetGoogleAuthURL(ctx context.Context, state string) string {
	return o.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// HandleGoogleCallback implements OAuthServiceInterface.
func (o *OAuthService) HandleGoogleCallback(ctx context.Context, code, state string) (*entity.UserEntity, string, error) {
	token, err := o.googleConfig.Exchange(ctx, code)
	if err != nil {
		log.Errorf("[OAuthService-1] HandleGoogleCallback exchange token: %v", err)
		return nil, "", err
	}

	googleUser, err := o.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		log.Errorf("[OAuthService-2] HandleGoogleCallback get user info: %v", err)
		return nil, "", err
	}

	existingUser, err := o.authRepo.GetUserByEmail(ctx, googleUser.Email)
	if err != nil && err != errs.ErrUserNotFound {
		log.Errorf("[OAuthService-3] HandleGoogleCallback get user by email: %v", err)
		return nil, "", err
	}

	var user *entity.UserEntity
	var jwtToken string

	if existingUser != nil {
		oauthProvider := &entity.OAuthProviderEntity{
			UserID:          existingUser.ID,
			Provider:        "google",
			ProviderUserID:  googleUser.ID,
			ProviderEmail:   googleUser.Email,
			ProviderName:    googleUser.Name,
			ProviderPicture: &googleUser.Picture,
			AccessToken:     &token.AccessToken,
			RefreshToken:    &token.RefreshToken,
		}

		if token.Expiry.After(time.Now()) {
			oauthProvider.TokenExpiresAt = &token.Expiry
		}

		err = o.oauthRepo.UpsertOAuthProvider(ctx, oauthProvider)
		if err != nil {
			log.Errorf("[OAuthService-4] HandleGoogleCallback upsert oauth provider: %v", err)
			return nil, "", err
		}

		if existingUser.Photo == "" {
			existingUser.Photo = googleUser.Picture
			err = o.userRepo.UpdateUser(ctx, existingUser)
			if err != nil {
				log.Warnf("[OAuthService-5] HandleGoogleCallback update user photo: %v", err)
			}
		}

		user = existingUser
	} else {
		// Create new user
		newUser := &entity.UserEntity{
			Name:       googleUser.Name,
			Email:      googleUser.Email,
			Photo:      googleUser.Picture,
			IsVerified: true,
			Password:   "",
		}

		createdUser, err := o.userRepo.CreateUser(ctx, newUser)
		if err != nil {
			log.Errorf("[OAuthService-6] HandleGoogleCallback create user: %v", err)
			return nil, "", err
		}

		err = o.oauthRepo.AssignRoleToUser(ctx, createdUser.ID, 2)
		if err != nil {
			log.Errorf("[OAuthService-6b] HandleGoogleCallback assign role: %v", err)
			return nil, "", err
		}

		oauthProvider := &entity.OAuthProviderEntity{
			UserID:          createdUser.ID,
			Provider:        "google",
			ProviderUserID:  googleUser.ID,
			ProviderEmail:   googleUser.Email,
			ProviderName:    googleUser.Name,
			ProviderPicture: &googleUser.Picture,
			AccessToken:     &token.AccessToken,
			RefreshToken:    &token.RefreshToken,
		}

		if token.Expiry.After(time.Now()) {
			oauthProvider.TokenExpiresAt = &token.Expiry
		}

		err = o.oauthRepo.CreateOAuthProvider(ctx, oauthProvider)
		if err != nil {
			log.Errorf("[OAuthService-7] HandleGoogleCallback create oauth provider: %v", err)
			return nil, "", err
		}

		user = createdUser
	}

	jwtToken, err = o.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Errorf("[OAuthService-8] HandleGoogleCallback generate jwt token: %v", err)
		return nil, "", err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      jwtToken,
		"role":       user.RoleName,
		"oauth":      true,
		"provider":   "google",
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		log.Errorf("[OAuthService-9] HandleGoogleCallback marshal session data: %v", err)
		return nil, "", err
	}

	redisConn, err := o.cfg.NewRedisClient()
	if err != nil {
		log.Errorf("[OAuthService-10] HandleGoogleCallback connect redis: %v", err)
		return nil, "", err
	}

	err = redisConn.Set(ctx, jwtToken, jsonData, 23*time.Hour).Err()
	if err != nil {
		log.Errorf("[OAuthService-11] HandleGoogleCallback set redis: %v", err)
		return nil, "", err
	}

	err = redisConn.Expire(ctx, jwtToken, 24*time.Hour).Err()
	if err != nil {
		log.Errorf("[OAuthService-12] HandleGoogleCallback set redis expiry: %v", err)
		return nil, "", err
	}

	return user, jwtToken, nil
}

func (o *OAuthService) getGoogleUserInfo(ctx context.Context, accessToken string) (*entity.GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var googleUser entity.GoogleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&googleUser)
	if err != nil {
		return nil, err
	}

	return &googleUser, nil
}

func (o *OAuthService) logOAuthActivity(
	ctx context.Context,
	userID int64,
	provider, action, status, errorMsg, userAgent, ipAddress string,
) {
	activity := &entity.OAuthActivityLog{
		UserID:    userID,
		Provider:  provider,
		Action:    action,
		Status:    status,
		ErrorMsg:  errorMsg,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}

	if err := o.oauthRepo.LogOAuthActivity(ctx, activity); err != nil {
		log.Errorf("[OAuthService-logOAuthActivity] Failed to log OAuth activity: %v", err)
	}
}

func (o *OAuthService) createUserSession(ctx context.Context, user *entity.UserEntity, jwtToken, provider string) error {
	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      jwtToken,
		"role":       user.RoleName,
		"oauth":      true,
		"provider":   provider,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	redisConn, err := o.cfg.NewRedisClient()
	if err != nil {
		return err
	}

	err = redisConn.Set(ctx, jwtToken, jsonData, 23*time.Hour).Err()
	if err != nil {
		return err
	}

	err = redisConn.Expire(ctx, jwtToken, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}

func NewOAuthService(
	userRepo repository.UserRepositoryInterface,
	oauthRepo repository.OAuthRepositoryInterface,
	cfg *config.Config,
	jwtService JwtServiceInterface,
	authRepo repository.AuthRepositoryInterface,
) OAuthServiceInterface {
	googleConfig := &oauth2.Config{
		ClientID:     cfg.Oauth.GoogleOauthClientID,
		ClientSecret: cfg.Oauth.GoogleOauthClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthService{
		userRepo:     userRepo,
		authRepo:     authRepo,
		oauthRepo:    oauthRepo,
		cfg:          cfg,
		jwtService:   jwtService,
		googleConfig: googleConfig,
	}
}
