package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID  `json:"id"`
	CategoryID  uuid.UUID  `json:"category_id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	Price       float64    `json:"price"`
	SalePrice   *float64   `json:"sale_price,omitempty"`
	Stock       int        `json:"stock"`
	Images      []string   `json:"images"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Category    *Category  `json:"category,omitempty"`
	Details     *ProductDetail `json:"details,omitempty"`
}

type ProductDetail struct {
	ProductID        uuid.UUID         `json:"product_id" bson:"-"`
	Specifications   map[string]string `json:"specifications" bson:"specifications"`
	RichDescription  string            `json:"rich_description" bson:"rich_description"`
	UpdatedAt        time.Time         `json:"updated_at" bson:"updated_at"`
}

type ProductListFilter struct {
	Page       int
	Limit      int
	Category   string
	Search     string
	MinPrice   *float64
	MaxPrice   *float64
	Sort       string
}

type ProductListResult struct {
	Items      []Product `json:"items"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type CreateProductInput struct {
	CategoryID      uuid.UUID         `json:"category_id"`
	Name            string            `json:"name"`
	Slug            string            `json:"slug"`
	Description     *string           `json:"description"`
	Price           float64           `json:"price"`
	SalePrice       *float64          `json:"sale_price"`
	Stock           int               `json:"stock"`
	Images          []string          `json:"images"`
	IsActive        bool              `json:"is_active"`
	Specifications  map[string]string `json:"specifications"`
	RichDescription string            `json:"rich_description"`
}

type UpdateProductInput struct {
	CategoryID      *uuid.UUID        `json:"category_id"`
	Name            *string           `json:"name"`
	Slug            *string           `json:"slug"`
	Description     *string           `json:"description"`
	Price           *float64          `json:"price"`
	SalePrice       *float64          `json:"sale_price"`
	Stock           *int              `json:"stock"`
	Images          []string          `json:"images"`
	IsActive        *bool             `json:"is_active"`
	Specifications  map[string]string `json:"specifications"`
	RichDescription *string           `json:"rich_description"`
}

type ProductStockInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	SalePrice *float64  `json:"sale_price,omitempty"`
	Stock     int       `json:"stock"`
	Images    []string  `json:"images"`
	IsActive  bool      `json:"is_active"`
}
