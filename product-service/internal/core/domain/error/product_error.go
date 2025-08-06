package error

import "errors"

var (
	ErrProductNotFound = errors.New("Product not found")
	ErrProductHasChildren = errors.New("Products Has Children")
	ErrProductAlreadyExists = errors.New("Product already exists")
)