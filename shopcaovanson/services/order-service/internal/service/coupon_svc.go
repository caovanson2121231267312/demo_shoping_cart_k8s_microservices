package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/caovanson/shopcaovanson/order-service/internal/repository"
	"github.com/google/uuid"
)

var ErrInvalidCoupon = errors.New("invalid coupon")

type CouponService struct {
	repo repository.CouponRepository
}

func NewCouponService(repo repository.CouponRepository) *CouponService {
	return &CouponService{repo: repo}
}

func (s *CouponService) Validate(ctx context.Context, code string, amount float64) (*domain.ValidateCouponResult, error) {
	discount, coupon, err := s.calculateDiscount(ctx, code, amount, false)
	if err != nil {
		return &domain.ValidateCouponResult{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}
	return &domain.ValidateCouponResult{
		Valid:          true,
		Code:           coupon.Code,
		DiscountAmount: discount,
		FinalAmount:    amount - discount,
	}, nil
}

func (s *CouponService) Apply(ctx context.Context, code string, amount float64) (float64, *domain.Coupon, error) {
	return s.calculateDiscount(ctx, code, amount, true)
}

func (s *CouponService) calculateDiscount(ctx context.Context, code string, amount float64, consume bool) (float64, *domain.Coupon, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil, fmt.Errorf("%w: coupon code is required", ErrInvalidCoupon)
	}

	coupon, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return 0, nil, err
	}
	if coupon == nil {
		return 0, nil, fmt.Errorf("%w: coupon not found", ErrInvalidCoupon)
	}
	if !coupon.IsActive {
		return 0, nil, fmt.Errorf("%w: coupon is inactive", ErrInvalidCoupon)
	}
	if coupon.ExpiresAt != nil && time.Now().UTC().After(*coupon.ExpiresAt) {
		return 0, nil, fmt.Errorf("%w: coupon expired", ErrInvalidCoupon)
	}
	if coupon.UsageLimit != nil && coupon.UsedCount >= *coupon.UsageLimit {
		return 0, nil, fmt.Errorf("%w: coupon usage limit reached", ErrInvalidCoupon)
	}
	if amount < coupon.MinOrder {
		return 0, nil, fmt.Errorf("%w: minimum order is %.0f", ErrInvalidCoupon, coupon.MinOrder)
	}

	var discount float64
	switch coupon.Type {
	case domain.CouponTypePercent:
		discount = amount * coupon.Value / 100
		if coupon.MaxDiscount != nil && discount > *coupon.MaxDiscount {
			discount = *coupon.MaxDiscount
		}
	case domain.CouponTypeFixed:
		discount = coupon.Value
	default:
		return 0, nil, fmt.Errorf("%w: invalid coupon type", ErrInvalidCoupon)
	}
	if discount > amount {
		discount = amount
	}
	if discount < 0 {
		discount = 0
	}

	if consume {
		if err := s.repo.IncrementUsedCount(ctx, coupon.ID); err != nil {
			return 0, nil, err
		}
	}
	return discount, coupon, nil
}

func (s *CouponService) Create(ctx context.Context, input domain.CreateCouponInput) (*domain.Coupon, error) {
	code := strings.TrimSpace(strings.ToUpper(input.Code))
	if code == "" {
		return nil, fmt.Errorf("%w: code is required", ErrInvalidInput)
	}
	if input.Type != domain.CouponTypePercent && input.Type != domain.CouponTypeFixed {
		return nil, fmt.Errorf("%w: type must be percent or fixed", ErrInvalidInput)
	}
	if input.Value <= 0 {
		return nil, fmt.Errorf("%w: value must be positive", ErrInvalidInput)
	}

	var expiresAt *time.Time
	if input.ExpiresAt != nil && *input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid expires_at", ErrInvalidInput)
		}
		expiresAt = &t
	}

	now := time.Now().UTC()
	coupon := &domain.Coupon{
		ID:          uuid.New(),
		Code:        code,
		Type:        input.Type,
		Value:       input.Value,
		MinOrder:    input.MinOrder,
		MaxDiscount: input.MaxDiscount,
		UsageLimit:  input.UsageLimit,
		UsedCount:   0,
		IsActive:    input.IsActive,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

func (s *CouponService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	coupon, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, ErrNotFound
	}
	return coupon, nil
}

func (s *CouponService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateCouponInput) (*domain.Coupon, error) {
	coupon, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, ErrNotFound
	}
	return coupon, nil
}

func (s *CouponService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return ErrNotFound
	}
	return nil
}

func (s *CouponService) List(ctx context.Context, filter domain.CouponListFilter) (*domain.CouponListResult, error) {
	return s.repo.List(ctx, filter)
}
