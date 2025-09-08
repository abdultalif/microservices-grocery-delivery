package error

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserExist       = errors.New("user already exists")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidToken    = errors.New("invalid token")

	ErrLastAuthMethod = errors.New("cannot unlink the last authentication method. Please set a password first")
	ErrGoogleUnlinked = errors.New("this Google account has been unlinked. Please relink manually")
	ErrOAuthNotFound  = errors.New("OAuth provider not found")
	ErrUnauthorized   = errors.New("unauthorized to unlink this account")
	ErrInvalidState   = errors.New("invalid state for registration")
	ErrGoogleLinked   = errors.New("this Google account is already linked to another user")
)
