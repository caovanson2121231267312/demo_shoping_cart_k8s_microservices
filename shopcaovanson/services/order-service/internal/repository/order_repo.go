package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order, items []domain.OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error)
	GetByOrderNumberAndPhone(ctx context.Context, orderNumber, phone string) (*domain.Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.OrderListResult, error)
	ListAll(ctx context.Context, status string, page, limit int) (*domain.OrderListResult, error)
	Search(ctx context.Context, filter domain.OrderSearchFilter) (*domain.OrderListResult, error)
	Stats(ctx context.Context) (*domain.OrderStats, error)
	NextOrderNumber(ctx context.Context) (string, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Count(ctx context.Context) (int, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type orderRepo struct {
	db *sqlx.DB

	statsMu      sync.Mutex
	statsCache   *domain.OrderStats
	statsCacheAt time.Time
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepo{db: db}
}

const orderSelectCols = `id, order_number, user_id, status, subtotal_amount, discount_amount, coupon_code, total_amount, shipping_name, shipping_phone, shipping_address, created_at, updated_at`

func (r *orderRepo) Create(ctx context.Context, order *domain.Order, items []domain.OrderItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (id, order_number, user_id, status, subtotal_amount, discount_amount, coupon_code, total_amount, shipping_name, shipping_phone, shipping_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, order.ID, order.OrderNumber, order.UserID, order.Status, order.SubtotalAmount, order.DiscountAmount, order.CouponCode, order.TotalAmount, order.ShippingName, order.ShippingPhone, order.ShippingAddress, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (id, order_id, product_id, product_name_snapshot, product_image_snapshot, unit_price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.OrderID, item.ProductID, item.ProductNameSnapshot, item.ProductImageSnapshot, item.UnitPrice, item.Quantity)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.GetContext(ctx, &order, `
		SELECT `+orderSelectCols+` FROM orders WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var items []domain.OrderItem
	if err := r.db.SelectContext(ctx, &items, `
		SELECT id, order_id, product_id, product_name_snapshot, product_image_snapshot, unit_price, quantity
		FROM order_items WHERE order_id = $1
	`, id); err != nil {
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *orderRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.OrderListResult, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID); err != nil {
		return nil, err
	}
	offset := (page - 1) * limit
	var orders []domain.Order
	if err := r.db.SelectContext(ctx, &orders, `
		SELECT `+orderSelectCols+` FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset); err != nil {
		return nil, err
	}
	if err := r.attachItems(ctx, orders); err != nil {
		return nil, err
	}
	return &domain.OrderListResult{
		Items:      orders,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages(total, limit),
	}, nil
}

func (r *orderRepo) ListAll(ctx context.Context, status string, page, limit int) (*domain.OrderListResult, error) {
	where := ""
	args := []interface{}{}
	idx := 1
	if status != "" {
		where = fmt.Sprintf("WHERE status = $%d", idx)
		args = append(args, status)
		idx++
	}

	countQuery := "SELECT COUNT(*) FROM orders " + where
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	listArgs := append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT `+orderSelectCols+` FROM orders %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, idx, idx+1)

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
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages(total, limit),
	}, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3
	`, status, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *orderRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM orders`)
	return count, err
}

func (r *orderRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`, id)
	return exists, err
}

func (r *orderRepo) attachItems(ctx context.Context, orders []domain.Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]string, len(orders))
	args := make([]interface{}, len(orders))
	for i, o := range orders {
		ids[i] = fmt.Sprintf("$%d", i+1)
		args[i] = o.ID
	}
	query := fmt.Sprintf(`
		SELECT id, order_id, product_id, product_name_snapshot, product_image_snapshot, unit_price, quantity
		FROM order_items WHERE order_id IN (%s)
	`, strings.Join(ids, ","))

	var items []domain.OrderItem
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return err
	}
	byOrder := make(map[uuid.UUID][]domain.OrderItem)
	for _, item := range items {
		byOrder[item.OrderID] = append(byOrder[item.OrderID], item)
	}
	for i := range orders {
		orders[i].Items = byOrder[orders[i].ID]
	}
	return nil
}

func totalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}
