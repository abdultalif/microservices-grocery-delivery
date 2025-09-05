package error

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserExist       = errors.New("user already exists")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidToken    = errors.New("invalid token")
	ErrGoogleUnlinked  = errors.New("this Google account has been unlinked. Please relink manually")
)
