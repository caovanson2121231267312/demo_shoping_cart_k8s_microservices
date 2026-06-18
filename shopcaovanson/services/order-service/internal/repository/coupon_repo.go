package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CouponRepository interface {
	Create(ctx context.Context, coupon *domain.Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error)
	GetByCode(ctx context.Context, code string) (*domain.Coupon, error)
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateCouponInput) (*domain.Coupon, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, limit int, search string) (*domain.CouponListResult, error)
	IncrementUsedCount(ctx context.Context, id uuid.UUID) error
}

type couponRepo struct {
	db *sqlx.DB
}

func NewCouponRepository(db *sqlx.DB) CouponRepository {
	return &couponRepo{db: db}
}

const couponCols = `id, code, type, value, min_order, max_discount, usage_limit, used_count, is_active, expires_at, created_at, updated_at`

func (r *couponRepo) Create(ctx context.Context, coupon *domain.Coupon) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO coupons (id, code, type, value, min_order, max_discount, usage_limit, used_count, is_active, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, coupon.ID, coupon.Code, coupon.Type, coupon.Value, coupon.MinOrder, coupon.MaxDiscount,
		coupon.UsageLimit, coupon.UsedCount, coupon.IsActive, coupon.ExpiresAt, coupon.CreatedAt, coupon.UpdatedAt)
	return err
}

func (r *couponRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	var coupon domain.Coupon
	err := r.db.GetContext(ctx, &coupon, `SELECT `+couponCols+` FROM coupons WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepo) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	err := r.db.GetContext(ctx, &coupon, `SELECT `+couponCols+` FROM coupons WHERE UPPER(code) = UPPER($1)`, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepo) Update(ctx context.Context, id uuid.UUID, input domain.UpdateCouponInput) (*domain.Coupon, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil || existing == nil {
		return nil, sql.ErrNoRows
	}

	if input.Type != nil {
		existing.Type = *input.Type
	}
	if input.Value != nil {
		existing.Value = *input.Value
	}
	if input.MinOrder != nil {
		existing.MinOrder = *input.MinOrder
	}
	if input.MaxDiscount != nil {
		existing.MaxDiscount = input.MaxDiscount
	}
	if input.UsageLimit != nil {
		existing.UsageLimit = input.UsageLimit
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}
	if input.ExpiresAt != nil {
		if *input.ExpiresAt == "" {
			existing.ExpiresAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("invalid expires_at")
			}
			existing.ExpiresAt = &t
		}
	}
	existing.UpdatedAt = time.Now().UTC()

	var updated domain.Coupon
	err = r.db.GetContext(ctx, &updated, `
		UPDATE coupons SET type=$1, value=$2, min_order=$3, max_discount=$4, usage_limit=$5, is_active=$6, expires_at=$7, updated_at=$8
		WHERE id=$9
		RETURNING `+couponCols,
		existing.Type, existing.Value, existing.MinOrder, existing.MaxDiscount, existing.UsageLimit,
		existing.IsActive, existing.ExpiresAt, existing.UpdatedAt, id,
	)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *couponRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM coupons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *couponRepo) List(ctx context.Context, page, limit int, search string) (*domain.CouponListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	where := "1=1"
	args := []interface{}{}
	idx := 1
	if s := strings.TrimSpace(search); s != "" {
		where += fmt.Sprintf(" AND code ILIKE $%d", idx)
		args = append(args, "%"+s+"%")
		idx++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM coupons WHERE `+where, args...); err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	listArgs := append(args, limit, offset)
	query := fmt.Sprintf(`SELECT %s FROM coupons WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, couponCols, where, idx, idx+1)

	var items []domain.Coupon
	if err := r.db.SelectContext(ctx, &items, query, listArgs...); err != nil {
		return nil, err
	}
	return &domain.CouponListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages(total, limit),
	}, nil
}

func (r *couponRepo) IncrementUsedCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE coupons SET used_count = used_count + 1, updated_at = NOW() WHERE id = $1`, id)
	return err
}
