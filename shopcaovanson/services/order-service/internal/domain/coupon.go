package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	CouponTypePercent = "percent"
	CouponTypeFixed   = "fixed"
)

type Coupon struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Code        string     `json:"code" db:"code"`
	Type        string     `json:"type" db:"type"`
	Value       float64    `json:"value" db:"value"`
	MinOrder    float64    `json:"min_order" db:"min_order"`
	MaxDiscount *float64   `json:"max_discount,omitempty" db:"max_discount"`
	UsageLimit  *int       `json:"usage_limit,omitempty" db:"usage_limit"`
	UsedCount   int        `json:"used_count" db:"used_count"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateCouponInput struct {
	Code        string   `json:"code"`
	Type        string   `json:"type"`
	Value       float64  `json:"value"`
	MinOrder    float64  `json:"min_order"`
	MaxDiscount *float64 `json:"max_discount"`
	UsageLimit  *int     `json:"usage_limit"`
	IsActive    bool     `json:"is_active"`
	ExpiresAt   *string  `json:"expires_at"`
}

type UpdateCouponInput struct {
	Type        *string  `json:"type"`
	Value       *float64 `json:"value"`
	MinOrder    *float64 `json:"min_order"`
	MaxDiscount *float64 `json:"max_discount"`
	UsageLimit  *int     `json:"usage_limit"`
	IsActive    *bool    `json:"is_active"`
	ExpiresAt   *string  `json:"expires_at"`
}

type ValidateCouponInput struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}

type ValidateCouponResult struct {
	Valid          bool    `json:"valid"`
	Code           string  `json:"code,omitempty"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
	Message        string  `json:"message,omitempty"`
}

type CouponListResult struct {
	Items      []Coupon `json:"items"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}
