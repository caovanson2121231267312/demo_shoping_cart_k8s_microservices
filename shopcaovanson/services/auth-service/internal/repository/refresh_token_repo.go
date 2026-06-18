package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopcaovanson/auth-service/internal/domain"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.Revoked, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.GetContext(ctx, &token,
		`SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked = FALSE AND expires_at > NOW()`,
		tokenHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1 AND revoked = FALSE`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) cleanupExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < $1`,
		time.Now().Add(-24*time.Hour),
	)
	return err
}
