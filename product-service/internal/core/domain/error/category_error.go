package error

import "errors"

var (
	ErrCategoryNotFound     = errors.New("category not found")
	ErrCategoryBadRequest   = errors.New("category bad request")
	ErrCategoryHasProducts  = errors.New("category has products")
	ErrCategoryHasChildren  = errors.New("category has children")
	ErrCategoryConflict     = errors.New("category conflict")
	ErrParentCategoryNotFound = errors.New("parent category not found") 
)
