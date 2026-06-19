package repository

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ReviewRepository interface {
	ListByProduct(ctx context.Context, productID uuid.UUID, page, limit int) (*domain.ReviewListResult, error)
	SummarizeByProducts(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]domain.ReviewSummary, error)
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
	var rows []reviewRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, product_id, user_id, rating, comment, created_at
		FROM product_reviews
		WHERE product_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, productID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.ReviewListResult{Items: rowsToReviews(rows), Total: total}, nil
}

func (r *reviewRepo) SummarizeByProducts(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]domain.ReviewSummary, error) {
	out := make(map[uuid.UUID]domain.ReviewSummary, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}

	type summaryRow struct {
		ProductID uuid.UUID `db:"product_id"`
		Total     int       `db:"total"`
		Average   float64   `db:"average"`
	}
	var rows []summaryRow
	idStrings := make([]string, len(productIDs))
	for i, id := range productIDs {
		idStrings[i] = id.String()
	}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT product_id,
		       COUNT(*)::int AS total,
		       COALESCE(AVG(rating)::float8, 0) AS average
		FROM product_reviews
		WHERE product_id = ANY($1::uuid[])
		GROUP BY product_id
	`, pq.Array(idStrings))
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		avg := math.Round(row.Average*10) / 10
		out[row.ProductID] = domain.ReviewSummary{Average: avg, Total: row.Total}
	}
	return out, nil
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

type reviewRow struct {
	ID        uuid.UUID      `db:"id"`
	ProductID uuid.UUID      `db:"product_id"`
	UserID    uuid.UUID      `db:"user_id"`
	Rating    int            `db:"rating"`
	Comment   sql.NullString `db:"comment"`
	CreatedAt time.Time      `db:"created_at"`
}

func (row reviewRow) toReview() domain.ProductReview {
	review := domain.ProductReview{
		ID:        row.ID,
		ProductID: row.ProductID,
		UserID:    row.UserID,
		Rating:    row.Rating,
		CreatedAt: row.CreatedAt,
	}
	if row.Comment.Valid {
		review.Comment = &row.Comment.String
	}
	return review
}

func rowsToReviews(rows []reviewRow) []domain.ProductReview {
	items := make([]domain.ProductReview, len(rows))
	for i, row := range rows {
		items[i] = row.toReview()
	}
	return items
}
