package error

import "errors"

var (
	ErrNotFoundOrder   = errors.New("Order Not Found")
	ErrNotFoundProduct = errors.New("Product Not Found")
	ErrNotFoundBuyer   = errors.New("Buyer Not Found")
	ErrInvalidStatus   = errors.New("Invalid status transaction")
	ErrForbiddenOrder  = errors.New("you are not allowed to access this order")
	ErrStockNotEnough  = errors.New("Stok no enough")

	ErrBuyerNotSynced = errors.New("buyer data not synced yet, please try again in a moment")
	ErrLocationNotSet = errors.New("user location not found in local data, please provide lat and lng")
)
