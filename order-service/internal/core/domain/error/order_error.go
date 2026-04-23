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

	ErrUserServiceUnavailable    = errors.New("user service is currently unavailable, please try again later")
	ErrProductServiceUnavailable = errors.New("product service is currently unavailable, please try again later")
	ErrDependencyTimeout         = errors.New("request to dependent service timed out, please retry")
	ErrDependencyFailed          = errors.New("failed to retrieve data from dependent service")
)
