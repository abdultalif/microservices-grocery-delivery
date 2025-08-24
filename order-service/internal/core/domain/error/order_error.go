package error

import "errors"

var (
	ErrNotFoundOrder   = errors.New("Order Not Found")
	ErrNotFoundProduct = errors.New("Product Not Found")
	ErrNotFoundBuyer   = errors.New("Buyer Not Found")
	ErrInvalidStatus   = errors.New("Invalid status transaction")
)
