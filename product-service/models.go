package main

import (
	"time"
)

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	SKU           string    `json:"sku"`
	CategoryID    *int      `json:"category_id"`
	Category      *Category `json:"category,omitempty"`
	Price         float64   `json:"price"`
	Cost          float64   `json:"cost"`
	StockQuantity int       `json:"stock_quantity"`
	MinStock      int       `json:"min_stock"`
	Description   *string   `json:"description"`
	ImageURL      *string   `json:"image_url"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	CategoryID    *int    `json:"category_id"`
	Price         float64 `json:"price"`
	Cost          float64 `json:"cost"`
	StockQuantity int     `json:"stock_quantity"`
	MinStock      int     `json:"min_stock"`
	Description   *string `json:"description"`
	ImageURL      *string `json:"image_url"`
}

type UpdateProductRequest struct {
	Name          *string  `json:"name"`
	CategoryID    *int     `json:"category_id"`
	Price         *float64 `json:"price"`
	Cost          *float64 `json:"cost"`
	StockQuantity *int     `json:"stock_quantity"`
	MinStock      *int     `json:"min_stock"`
	Description   *string  `json:"description"`
	ImageURL      *string  `json:"image_url"`
	IsActive      *bool    `json:"is_active"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type StockUpdateRequest struct {
	Quantity int    `json:"quantity"`
	Type     string `json:"type"` //jenisnya "add" atau "subtract"
	Notes    string `json:"notes"`
}

