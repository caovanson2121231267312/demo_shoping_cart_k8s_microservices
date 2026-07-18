package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/pagination"
)

func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) (*domain.User, error) {
	query := `
		UPDATE users SET role = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING *
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
		RETURNING *
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
	var total int
	var err error
	if isBroadCount(where, args) {
		total, err = approximateTableCount(ctx, r.db, "users")
	} else {
		err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM users WHERE "+whereSQL, args...)
	}
	if err != nil {
		return nil, err
	}

	useCursor := filter.Cursor != ""
	if useCursor {
		cur, err := pagination.Decode(filter.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		where = append(where, fmt.Sprintf("(created_at, id) < ($%d, $%d)", idx, idx+1))
		args = append(args, cur.CreatedAt, cur.ID)
		idx += 2
	}

	whereSQL = strings.Join(where, " AND ")
	useKeyset := useCursor || filter.Page <= 1
	fetchLimit := filter.Limit
	if useKeyset {
		fetchLimit = filter.Limit + 1
	}
	listArgs := append(args, fetchLimit)
	listQuery := fmt.Sprintf(`
		SELECT id, email, full_name, role, is_active, email_verified, avatar_key, created_at, updated_at
		FROM users WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, whereSQL, idx)

	if !useKeyset {
		offset := (filter.Page - 1) * filter.Limit
		listArgs = append(listArgs, offset)
		listQuery = fmt.Sprintf(`
			SELECT id, email, full_name, role, is_active, email_verified, avatar_key, created_at, updated_at
			FROM users WHERE %s
			ORDER BY created_at DESC, id DESC
			LIMIT $%d OFFSET $%d
		`, whereSQL, idx, idx+1)
	}

	var users []domain.User
	if err := r.db.SelectContext(ctx, &users, listQuery, listArgs...); err != nil {
		return nil, err
	}

	hasMore := false
	if useKeyset && len(users) > filter.Limit {
		hasMore = true
		users = users[:filter.Limit]
	}

	items := make([]domain.UserProfile, len(users))
	for i, u := range users {
		items[i] = u.ToProfile()
	}

	result := &domain.UserListResult{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages(total, filter.Limit),
		HasMore:    hasMore,
	}
	if hasMore && len(users) > 0 {
		last := users[len(users)-1]
		result.NextCursor = pagination.Encode(last.CreatedAt, last.ID)
	}
	return result, nil
}

func (r *userRepository) Stats(ctx context.Context) (*domain.AdminStats, error) {
	const statsCacheTTL = 60 * time.Second

	r.statsMu.Lock()
	if r.statsCache != nil && time.Since(r.statsCacheAt) < statsCacheTTL {
		cached := cloneAdminStats(r.statsCache)
		r.statsMu.Unlock()
		return cached, nil
	}
	r.statsMu.Unlock()

	stats := &domain.AdminStats{UsersByRole: map[string]int{}}
	total, err := approximateTableCount(ctx, r.db, "users")
	if err != nil {
		if err := r.db.GetContext(ctx, &stats.TotalUsers, `SELECT COUNT(*) FROM users`); err != nil {
			return nil, err
		}
	} else {
		stats.TotalUsers = total
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

	r.statsMu.Lock()
	r.statsCache = cloneAdminStats(stats)
	r.statsCacheAt = time.Now()
	r.statsMu.Unlock()
	return stats, nil
}

func cloneAdminStats(s *domain.AdminStats) *domain.AdminStats {
	if s == nil {
		return nil
	}
	c := *s
	c.UsersByRole = make(map[string]int, len(s.UsersByRole))
	for k, v := range s.UsersByRole {
		c.UsersByRole[k] = v
	}
	return &c
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

func isBroadCount(where []string, args []interface{}) bool {
	return len(where) == 1 && where[0] == "1=1" && len(args) == 0
}

func approximateTableCount(ctx context.Context, db *sqlx.DB, table string) (int, error) {
	var total int
	err := db.GetContext(ctx, &total, `
		SELECT COALESCE(GREATEST(reltuples::bigint, 0), 0)::int
		FROM pg_class
		WHERE relname = $1
	`, table)
	return total, err
}
