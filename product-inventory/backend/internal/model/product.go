package model

import "time"

const (
	StatusActive       = "active"
	StatusInactive     = "inactive"
	StatusDiscontinued = "discontinued"
)

func ValidStatus(s string) bool {
	return s == StatusActive || s == StatusInactive || s == StatusDiscontinued
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku"`
	CategoryID int64     `json:"category_id"`
	Category   string    `json:"category"`
	Price      float64   `json:"price"`
	Quantity   int       `json:"quantity"`
	Status     string    `json:"status"`
	LowStock   bool      `json:"low_stock"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ProductInput is the payload for create and update.
type ProductInput struct {
	Name       string  `json:"name"`
	SKU        string  `json:"sku"`
	CategoryID int64   `json:"category_id"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
	Status     string  `json:"status"`
}

// StockUpdate adjusts stock. Mode is "set", "add" or "remove".
// Quantity is a pointer so a missing or null value can be told apart from 0.
type StockUpdate struct {
	Mode     string `json:"mode"`
	Quantity *int   `json:"quantity"`
}

type ListFilter struct {
	Search     string
	CategoryID int64
	Status     string
}
