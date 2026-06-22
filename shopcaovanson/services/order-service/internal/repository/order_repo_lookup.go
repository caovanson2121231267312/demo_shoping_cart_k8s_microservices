package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/google/uuid"
)

func (r *orderRepo) NextOrderNumber(ctx context.Context) (string, error) {
	var seq int64
	if err := r.db.GetContext(ctx, &seq, `SELECT nextval('order_number_seq')`); err != nil {
		return "", err
	}
	return fmt.Sprintf("ORD-%010d", seq), nil
}

func (r *orderRepo) GetByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.GetContext(ctx, &order, `
		SELECT `+orderSelectCols+` FROM orders WHERE order_number = $1
	`, strings.TrimSpace(orderNumber))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.loadItems(ctx, &order)
}

func (r *orderRepo) GetByOrderNumberAndPhone(ctx context.Context, orderNumber, phone string) (*domain.Order, error) {
	var order domain.Order
	normalizedPhone := normalizePhone(phone)
	err := r.db.GetContext(ctx, &order, `
		SELECT `+orderSelectCols+` FROM orders
		WHERE order_number = $1 AND REPLACE(shipping_phone, ' ', '') = $2
	`, strings.TrimSpace(orderNumber), normalizedPhone)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.loadItems(ctx, &order)
}

func (r *orderRepo) Search(ctx context.Context, filter domain.OrderSearchFilter) (*domain.OrderListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.OrderNumber != "" {
		where = append(where, fmt.Sprintf("order_number ILIKE $%d", idx))
		args = append(args, "%"+strings.TrimSpace(filter.OrderNumber)+"%")
		idx++
	}
	if filter.ShippingPhone != "" {
		where = append(where, fmt.Sprintf("REPLACE(shipping_phone, ' ', '') LIKE $%d", idx))
		args = append(args, "%"+normalizePhone(filter.ShippingPhone)+"%")
		idx++
	}
	if filter.UserID != "" {
		if uid, err := uuid.Parse(filter.UserID); err == nil {
			where = append(where, fmt.Sprintf("user_id = $%d", idx))
			args = append(args, uid)
			idx++
		}
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		where = append(where, fmt.Sprintf("(order_number ILIKE $%d OR shipping_name ILIKE $%d OR shipping_phone ILIKE $%d)", idx, idx, idx))
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
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM orders WHERE "+whereSQL, args...); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.Limit
	listArgs := append(args, filter.Limit, offset)
	query := fmt.Sprintf(`
		SELECT `+orderSelectCols+` FROM orders WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)

	var orders []domain.Order
	if err := r.db.SelectContext(ctx, &orders, query, listArgs...); err != nil {
		return nil, err
	}
	if err := r.attachItems(ctx, orders); err != nil {
		return nil, err
	}
	return &domain.OrderListResult{
		Items:      orders,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages(total, filter.Limit),
	}, nil
}

func (r *orderRepo) Stats(ctx context.Context) (*domain.OrderStats, error) {
	stats := &domain.OrderStats{ByStatus: map[string]int{}}
	if err := r.db.GetContext(ctx, &stats.TotalOrders, `SELECT COUNT(*) FROM orders`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.Revenue, `
		SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status != 'cancelled'
	`); err != nil {
		return nil, err
	}
	type row struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, `SELECT status, COUNT(*) AS count FROM orders GROUP BY status`); err != nil {
		return nil, err
	}
	for _, r := range rows {
		stats.ByStatus[r.Status] = r.Count
	}
	return stats, nil
}

func (r *orderRepo) loadItems(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	orders := []domain.Order{*order}
	if err := r.attachItems(ctx, orders); err != nil {
		return nil, err
	}
	return &orders[0], nil
}

func normalizePhone(phone string) string {
	return strings.NewReplacer(" ", "", "-", "", ".", "").Replace(strings.TrimSpace(phone))
}
