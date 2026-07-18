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

type LoginHistoryRepository interface {
	Create(ctx context.Context, row *domain.LoginHistory) error
	List(ctx context.Context, filter domain.LoginHistoryFilter) (*domain.LoginHistoryResult, error)
	ListByDateRange(ctx context.Context, from, to time.Time) ([]domain.LoginHistory, error)
	StatsByDateRange(ctx context.Context, from, to time.Time) (total, success, failure, uniqueUsers int, err error)
}

type loginHistoryRepo struct {
	db *sqlx.DB
}

func NewLoginHistoryRepository(db *sqlx.DB) LoginHistoryRepository {
	return &loginHistoryRepo{db: db}
}

func (r *loginHistoryRepo) Create(ctx context.Context, row *domain.LoginHistory) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO login_history (
			id, user_id, email, success, failure_reason,
			ip_address, user_agent, device, browser, os, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, row.ID, row.UserID, row.Email, row.Success, row.FailureReason,
		row.IPAddress, row.UserAgent, row.Device, row.Browser, row.OS, row.CreatedAt)
	return err
}

func (r *loginHistoryRepo) List(ctx context.Context, filter domain.LoginHistoryFilter) (*domain.LoginHistoryResult, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if filter.Email != "" {
		where = append(where, fmt.Sprintf("h.email ILIKE $%d", idx))
		args = append(args, "%"+strings.TrimSpace(filter.Email)+"%")
		idx++
	}
	if filter.UserID != "" {
		if uid, err := uuid.Parse(filter.UserID); err == nil {
			where = append(where, fmt.Sprintf("h.user_id = $%d", idx))
			args = append(args, uid)
			idx++
		}
	}
	if filter.Success != nil {
		where = append(where, fmt.Sprintf("h.success = $%d", idx))
		args = append(args, *filter.Success)
		idx++
	}
	if filter.CreatedFrom != nil {
		where = append(where, fmt.Sprintf("h.created_at >= $%d", idx))
		args = append(args, *filter.CreatedFrom)
		idx++
	}
	if filter.CreatedTo != nil {
		where = append(where, fmt.Sprintf("h.created_at < $%d", idx))
		args = append(args, filter.CreatedTo.Add(24*time.Hour))
		idx++
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM login_history h WHERE "+whereSQL, args...); err != nil {
		return nil, err
	}

	useCursor := filter.Cursor != ""
	if useCursor {
		cur, err := pagination.Decode(filter.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		where = append(where, fmt.Sprintf("(h.created_at, h.id) < ($%d, $%d)", idx, idx+1))
		args = append(args, cur.CreatedAt, cur.ID)
		idx += 2
	}
	whereSQL = strings.Join(where, " AND ")

	fetchLimit := filter.Limit + 1
	listArgs := append(args, fetchLimit)
	query := fmt.Sprintf(`
		SELECT h.id, h.user_id, h.email, h.success, h.failure_reason,
		       h.ip_address, h.user_agent, h.device, h.browser, h.os, h.created_at,
		       u.full_name
		FROM login_history h
		LEFT JOIN users u ON u.id = h.user_id
		WHERE %s
		ORDER BY h.created_at DESC, h.id DESC
		LIMIT $%d
	`, whereSQL, idx)

	var rows []domain.LoginHistory
	if err := r.db.SelectContext(ctx, &rows, query, listArgs...); err != nil {
		return nil, err
	}

	hasMore := false
	if len(rows) > filter.Limit {
		hasMore = true
		rows = rows[:filter.Limit]
	}

	result := &domain.LoginHistoryResult{
		Items:   rows,
		Total:   total,
		Limit:   filter.Limit,
		HasMore: hasMore,
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		result.NextCursor = pagination.Encode(last.CreatedAt, last.ID)
	}
	return result, nil
}

func (r *loginHistoryRepo) ListByDateRange(ctx context.Context, from, to time.Time) ([]domain.LoginHistory, error) {
	var rows []domain.LoginHistory
	err := r.db.SelectContext(ctx, &rows, `
		SELECT h.id, h.user_id, h.email, h.success, h.failure_reason,
		       h.ip_address, h.user_agent, h.device, h.browser, h.os, h.created_at,
		       u.full_name
		FROM login_history h
		LEFT JOIN users u ON u.id = h.user_id
		WHERE h.created_at >= $1 AND h.created_at < $2
		ORDER BY h.created_at ASC, h.id ASC
	`, from, to)
	return rows, err
}

func (r *loginHistoryRepo) StatsByDateRange(ctx context.Context, from, to time.Time) (total, success, failure, uniqueUsers int, err error) {
	type stats struct {
		Total       int `db:"total"`
		Success     int `db:"success_count"`
		Failure     int `db:"failure_count"`
		UniqueUsers int `db:"unique_users"`
	}
	var s stats
	err = r.db.GetContext(ctx, &s, `
		SELECT
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE success)::int AS success_count,
			COUNT(*) FILTER (WHERE NOT success)::int AS failure_count,
			COUNT(DISTINCT user_id) FILTER (WHERE user_id IS NOT NULL)::int AS unique_users
		FROM login_history
		WHERE created_at >= $1 AND created_at < $2
	`, from, to)
	return s.Total, s.Success, s.Failure, s.UniqueUsers, err
}

type LoginReportRepository interface {
	Upsert(ctx context.Context, report *domain.LoginReport) error
	List(ctx context.Context, page, limit int, from, to *time.Time) (*domain.LoginReportListResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LoginReport, error)
	GetByDate(ctx context.Context, date time.Time) (*domain.LoginReport, error)
}

type loginReportRepo struct {
	db *sqlx.DB
}

func NewLoginReportRepository(db *sqlx.DB) LoginReportRepository {
	return &loginReportRepo{db: db}
}

func (r *loginReportRepo) Upsert(ctx context.Context, report *domain.LoginReport) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO login_reports (
			id, report_date, object_key, file_name, file_size,
			total_logins, success_count, failure_count, unique_users,
			status, error_message, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (report_date) DO UPDATE SET
			object_key = EXCLUDED.object_key,
			file_name = EXCLUDED.file_name,
			file_size = EXCLUDED.file_size,
			total_logins = EXCLUDED.total_logins,
			success_count = EXCLUDED.success_count,
			failure_count = EXCLUDED.failure_count,
			unique_users = EXCLUDED.unique_users,
			status = EXCLUDED.status,
			error_message = EXCLUDED.error_message,
			created_at = EXCLUDED.created_at,
			id = EXCLUDED.id
	`, report.ID, report.ReportDate.In(vietnamLoc()).Format("2006-01-02"), report.ObjectKey, report.FileName, report.FileSize,
		report.TotalLogins, report.SuccessCount, report.FailureCount, report.UniqueUsers,
		report.Status, report.ErrorMessage, report.CreatedAt)
	return err
}

func vietnamLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return time.FixedZone("ICT", 7*3600)
	}
	return loc
}

func (r *loginReportRepo) List(ctx context.Context, page, limit int, from, to *time.Time) (*domain.LoginReportListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1
	if from != nil {
		where = append(where, fmt.Sprintf("report_date >= $%d", idx))
		args = append(args, from.Format("2006-01-02"))
		idx++
	}
	if to != nil {
		where = append(where, fmt.Sprintf("report_date <= $%d", idx))
		args = append(args, to.Format("2006-01-02"))
		idx++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM login_reports WHERE "+whereSQL, args...); err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	listArgs := append(args, limit, offset)
	var items []domain.LoginReport
	err := r.db.SelectContext(ctx, &items, fmt.Sprintf(`
		SELECT id, report_date, object_key, file_name, file_size,
		       total_logins, success_count, failure_count, unique_users,
		       status, error_message, created_at
		FROM login_reports
		WHERE %s
		ORDER BY report_date DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1), listArgs...)
	if err != nil {
		return nil, err
	}
	return &domain.LoginReportListResult{Items: items, Total: total, Page: page, Limit: limit}, nil
}

func (r *loginReportRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoginReport, error) {
	var report domain.LoginReport
	err := r.db.GetContext(ctx, &report, `
		SELECT id, report_date, object_key, file_name, file_size,
		       total_logins, success_count, failure_count, unique_users,
		       status, error_message, created_at
		FROM login_reports WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *loginReportRepo) GetByDate(ctx context.Context, date time.Time) (*domain.LoginReport, error) {
	var report domain.LoginReport
	err := r.db.GetContext(ctx, &report, `
		SELECT id, report_date, object_key, file_name, file_size,
		       total_logins, success_count, failure_count, unique_users,
		       status, error_message, created_at
		FROM login_reports WHERE report_date = $1
	`, date.Format("2006-01-02"))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &report, nil
}
