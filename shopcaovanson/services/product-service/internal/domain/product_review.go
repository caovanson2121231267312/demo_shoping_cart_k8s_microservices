package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductReview struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReviewInput struct {
	Rating  int     `json:"rating"`
	Comment *string `json:"comment"`
}

type ReviewListResult struct {
	Items []ProductReview `json:"items"`
	Total int             `json:"total"`
}
