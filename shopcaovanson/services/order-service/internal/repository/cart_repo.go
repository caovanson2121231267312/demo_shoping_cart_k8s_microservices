package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CartRepository interface {
	Get(ctx context.Context, userID uuid.UUID) (*domain.Cart, error)
	Save(ctx context.Context, cart *domain.Cart, ttl time.Duration) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type cartRepo struct {
	client *redis.Client
}

func NewCartRepository(client *redis.Client) CartRepository {
	return &cartRepo{client: client}
}

func cartKey(userID uuid.UUID) string {
	return fmt.Sprintf("cart:%s", userID.String())
}

func (r *cartRepo) Get(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	val, err := r.client.Get(ctx, cartKey(userID)).Result()
	if err == redis.Nil {
		return &domain.Cart{
			UserID:    userID,
			Items:     []domain.CartItem{},
			UpdatedAt: time.Now().UTC(),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var items []domain.CartItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		return nil, err
	}
	return &domain.Cart{
		UserID:    userID,
		Items:     items,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (r *cartRepo) Save(ctx context.Context, cart *domain.Cart, ttl time.Duration) error {
	cart.UpdatedAt = time.Now().UTC()
	body, err := json.Marshal(cart.Items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, cartKey(cart.UserID), body, ttl).Err()
}

func (r *cartRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.client.Del(ctx, cartKey(userID)).Err()
}
