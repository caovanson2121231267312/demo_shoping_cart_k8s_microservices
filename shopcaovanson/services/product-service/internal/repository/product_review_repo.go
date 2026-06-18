package repository

import (
	"context"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReviewRepository interface {
	ListByProduct(ctx context.Context, productID uuid.UUID, page, limit int) (*domain.ReviewListResult, error)
	Create(ctx context.Context, review *domain.ProductReview) error
	Count(ctx context.Context) (int, error)
	ExistsForUserProduct(ctx context.Context, userID, productID uuid.UUID) (bool, error)
}

type reviewRepo struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) ReviewRepository {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) ListByProduct(ctx context.Context, productID uuid.UUID, page, limit int) (*domain.ReviewListResult, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM product_reviews WHERE product_id = $1
	`, productID); err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	var items []domain.ProductReview
	err := r.db.SelectContext(ctx, &items, `
		SELECT id, product_id, user_id, rating, comment, created_at
		FROM product_reviews
		WHERE product_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, productID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.ReviewListResult{Items: items, Total: total}, nil
}

func (r *reviewRepo) Create(ctx context.Context, review *domain.ProductReview) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO product_reviews (id, product_id, user_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, review.ID, review.ProductID, review.UserID, review.Rating, review.Comment, review.CreatedAt)
	return err
}

func (r *reviewRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM product_reviews`)
	return count, err
}

func (r *reviewRepo) ExistsForUserProduct(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM product_reviews WHERE user_id = $1 AND product_id = $2
		)
	`, userID, productID)
	return exists, err
}
