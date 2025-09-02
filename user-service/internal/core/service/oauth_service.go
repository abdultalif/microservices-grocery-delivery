package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
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
	GetGoogleAuthURL(ctx context.Context, state string) string
	HandleGoogleCallback(ctx context.Context, code, state string) (*entity.UserEntity, string, error)
	GenerateState() (string, error)
}

type OAuthService struct {
	userRepo     repository.UserRepositoryInterface
	authRepo     repository.AuthRepositoryInterface
	oauthRepo    repository.OAuthRepositoryInterface
	cfg          *config.Config
	jwtService   JwtServiceInterface
	googleConfig *oauth2.Config
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
		RedirectURL:  cfg.Oauth.GoogleRedirectUrl,
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
