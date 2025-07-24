package error

import "errors"

var (
	ErrProductNotFound = errors.New("Product not found")
	ErrCategoryNotFound = errors.New("Category not found")
	ErrProductHasChildren = errors.New("Products Has Children")
)