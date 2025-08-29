package error

import "errors"

var (

	// Category Errors
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryBadRequest     = errors.New("category bad request")
	ErrCategoryHasProducts    = errors.New("category has products")
	ErrCategoryHasChildren    = errors.New("category has children")
	ErrCategoryConflict       = errors.New("category conflict")
	ErrParentCategoryNotFound = errors.New("parent category not found")

	// Product Errors
	ErrProductNotFound      = errors.New("Product not found")
	ErrProductHasChildren   = errors.New("Products Has Children")
	ErrProductAlreadyExists = errors.New("Product already exists")

	// Cart Errors
	ErrCartNotFound    = errors.New("Cart not found")
	ErrInvalidQuantity = errors.New("invalid quantity, must be greater than 0")
)
