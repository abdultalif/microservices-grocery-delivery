package error

import "errors"

var (
	ErrUserNotFound   	= errors.New("user not found")
	ErrInvalidPassword	= errors.New("invalid password")
	ErrUserExist       	= errors.New("Email already exists")
	ErrTokenExpired   = errors.New("token expired")
	ErrInvalidToken    = errors.New("invalid token")
)