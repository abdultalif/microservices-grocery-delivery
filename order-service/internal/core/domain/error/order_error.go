package error

import "errors"

var (
	ErrNotFoundOrder = errors.New("Order not found")
)