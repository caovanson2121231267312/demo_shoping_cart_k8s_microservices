package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/domain"
)

func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) (*domain.User, error) {
	query := `
		UPDATE users SET role = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, email, password_hash, full_name, role, is_active, created_at, updated_at
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, role, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update role: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) (*domain.User, error) {
	query := `
		UPDATE users SET is_active = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, email, password_hash, full_name, role, is_active, created_at, updated_at
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, isActive, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update status: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`)
	return count, err
}

func (r *userRepository) List(ctx context.Context, filter domain.UserListFilter) (*domain.UserListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if filter.Role != "" {
		where = append(where, fmt.Sprintf("role = $%d", idx))
		args = append(args, filter.Role)
		idx++
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		where = append(where, fmt.Sprintf("(email ILIKE $%d OR full_name ILIKE $%d)", idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if filter.CreatedFrom != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *filter.CreatedFrom)
		idx++
	}
	if filter.CreatedTo != nil {
		where = append(where, fmt.Sprintf("created_at < $%d", idx))
		args = append(args, filter.CreatedTo.Add(24*time.Hour))
		idx++
	}

	whereSQL := strings.Join(where, " AND ")
	countQuery := "SELECT COUNT(*) FROM users WHERE " + whereSQL
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.Limit
	listArgs := append(args, filter.Limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)

	var users []domain.User
	if err := r.db.SelectContext(ctx, &users, listQuery, listArgs...); err != nil {
		return nil, err
	}

	items := make([]domain.UserProfile, len(users))
	for i, u := range users {
		items[i] = u.ToProfile()
	}

	return &domain.UserListResult{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages(total, filter.Limit),
	}, nil
}

func (r *userRepository) Stats(ctx context.Context) (*domain.AdminStats, error) {
	stats := &domain.AdminStats{UsersByRole: map[string]int{}}

	if err := r.db.GetContext(ctx, &stats.TotalUsers, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.ActiveUsers, `SELECT COUNT(*) FROM users WHERE is_active = TRUE`); err != nil {
		return nil, err
	}
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	type roleCount struct {
		Role  string `db:"role"`
		Count int    `db:"count"`
	}
	var rows []roleCount
	if err := r.db.SelectContext(ctx, &rows, `SELECT role, COUNT(*) AS count FROM users GROUP BY role`); err != nil {
		return nil, err
	}
	for _, row := range rows {
		stats.UsersByRole[row.Role] = row.Count
	}
	return stats, nil
}

func totalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return pages
}
