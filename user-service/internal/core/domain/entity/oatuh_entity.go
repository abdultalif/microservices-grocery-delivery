package entity

import "time"

type OAuthProviderEntity struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	Provider        string     `json:"provider"`
	ProviderUserID  string     `json:"provider_user_id"`
	ProviderEmail   string     `json:"provider_email"`
	ProviderName    string     `json:"provider_name"`
	ProviderPicture *string    `json:"provider_picture"`
	AccessToken     *string    `json:"access_token"`
	RefreshToken    *string    `json:"refresh_token"`
	IsRevoked       bool       `json:"is_revoked"`
	TokenExpiresAt  *time.Time `json:"token_expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}
