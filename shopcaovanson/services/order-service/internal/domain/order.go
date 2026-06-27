package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipping  = "shipping"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

var ValidOrderStatuses = []string{
	OrderStatusPending,
	OrderStatusConfirmed,
	OrderStatusShipping,
	OrderStatusDelivered,
	OrderStatusCancelled,
}

type CartItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	Quantity     int       `json:"quantity"`
	UnitPrice    float64   `json:"unit_price"`
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image,omitempty"`
}

type Cart struct {
	UserID    uuid.UUID  `json:"user_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Order struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	OrderNumber     string      `json:"order_number" db:"order_number"`
	UserID          uuid.UUID   `json:"user_id" db:"user_id"`
	Status          string      `json:"status" db:"status"`
	SubtotalAmount  float64     `json:"subtotal_amount" db:"subtotal_amount"`
	DiscountAmount  float64     `json:"discount_amount" db:"discount_amount"`
	CouponCode      *string     `json:"coupon_code,omitempty" db:"coupon_code"`
	TotalAmount     float64     `json:"total_amount" db:"total_amount"`
	ShippingName    string      `json:"shipping_name" db:"shipping_name"`
	ShippingPhone   string      `json:"shipping_phone" db:"shipping_phone"`
	ShippingAddress string      `json:"shipping_address" db:"shipping_address"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	Items           []OrderItem `json:"items,omitempty" db:"-"`
}

type OrderItem struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	OrderID              uuid.UUID `json:"order_id" db:"order_id"`
	ProductID            uuid.UUID `json:"product_id" db:"product_id"`
	ProductNameSnapshot  string    `json:"product_name_snapshot" db:"product_name_snapshot"`
	ProductImageSnapshot *string   `json:"product_image_snapshot,omitempty" db:"product_image_snapshot"`
	UnitPrice            float64   `json:"unit_price" db:"unit_price"`
	Quantity             int       `json:"quantity" db:"quantity"`
}

type CheckoutInput struct {
	ShippingName    string  `json:"shipping_name"`
	ShippingPhone   string  `json:"shipping_phone"`
	ShippingAddress string  `json:"shipping_address"`
	CouponCode      string  `json:"coupon_code,omitempty"`
}

type AddCartItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type UpdateCartItemInput struct {
	Quantity int `json:"quantity"`
}

type UpdateOrderStatusInput struct {
	Status string `json:"status"`
}

type TrackOrderInput struct {
	OrderNumber   string `json:"order_number"`
	ShippingPhone string `json:"shipping_phone"`
}

type LookupOrdersInput struct {
	OrderNumber   string `json:"order_number,omitempty"`
	ShippingPhone string `json:"shipping_phone,omitempty"`
	Email         string `json:"email,omitempty"`
}

type OrderSearchFilter struct {
	Page          int
	Limit         int
	Status        string
	Search        string
	OrderNumber   string
	ShippingPhone string
	UserID        string
	CreatedFrom   *time.Time
	CreatedTo     *time.Time
}

type OrderStats struct {
	TotalOrders int     `json:"total_orders"`
	Revenue     float64 `json:"revenue"`
	ByStatus    map[string]int `json:"by_status"`
}

type OrderListResult struct {
	Items      []Order `json:"items"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}
