package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/google/uuid"
)

func (r *categoryRepo) Update(ctx context.Context, id uuid.UUID, input domain.UpdateCategoryInput) (*domain.Category, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Slug != nil {
		sets = append(sets, fmt.Sprintf("slug = $%d", idx))
		args = append(args, *input.Slug)
		idx++
	}
	if input.ParentID != nil {
		sets = append(sets, fmt.Sprintf("parent_id = $%d", idx))
		args = append(args, *input.ParentID)
		idx++
	}
	if input.ImageURL != nil {
		sets = append(sets, fmt.Sprintf("image_url = $%d", idx))
		args = append(args, *input.ImageURL)
		idx++
	}
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE categories SET %s WHERE id = $%d
		RETURNING id, name, slug, parent_id, image_url
	`, strings.Join(sets, ", "), idx)

	var c domain.Category
	if err := r.db.GetContext(ctx, &c, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
