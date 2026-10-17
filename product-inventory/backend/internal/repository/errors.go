package repository

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrDuplicateSKU      = errors.New("sku already exists")
	ErrDuplicateCategory = errors.New("category already exists")
	ErrCategoryNotFound  = errors.New("category does not exist")
	ErrInsufficientStock = errors.New("insufficient stock")
)
