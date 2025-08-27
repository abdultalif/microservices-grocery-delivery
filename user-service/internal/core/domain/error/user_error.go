package error

import "errors"

var (
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
)