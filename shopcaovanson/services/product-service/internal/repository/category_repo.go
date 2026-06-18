package repository

import (
	"context"
	"database/sql"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository interface {
	ListAll(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, c *domain.Category) error
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateCategoryInput) (*domain.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepo struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) ListAll(ctx context.Context) ([]domain.Category, error) {
	var items []domain.Category
	err := r.db.SelectContext(ctx, &items, `
		SELECT id, name, slug, parent_id, image_url
		FROM categories
		ORDER BY name ASC
	`)
	return items, err
}

func (r *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var c domain.Category
	err := r.db.GetContext(ctx, &c, `
		SELECT id, name, slug, parent_id, image_url
		FROM categories WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var c domain.Category
	err := r.db.GetContext(ctx, &c, `
		SELECT id, name, slug, parent_id, image_url
		FROM categories WHERE slug = $1
	`, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM categories`)
	return count, err
}

func (r *categoryRepo) Create(ctx context.Context, c *domain.Category) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO categories (id, name, slug, parent_id, image_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (slug) DO NOTHING
	`, c.ID, c.Name, c.Slug, c.ParentID, c.ImageURL)
	return err
}
