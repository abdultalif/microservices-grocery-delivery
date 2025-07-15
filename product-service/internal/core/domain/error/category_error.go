package error

import "errors"

var (
	ErrCategoryNotFound     = errors.New("category not found")
	ErrCategoryHasProducts  = errors.New("category has products")
	ErrCategoryHasChildren  = errors.New("category has children")
)
