package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopcaovanson/auth-service/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByVerificationToken(ctx context.Context, token string) (*domain.User, error)
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
	SetVerificationToken(ctx context.Context, id uuid.UUID, token string, expiresAt time.Time) error
	SetPasswordResetOTP(ctx context.Context, id uuid.UUID, otpHash string, expiresAt time.Time) error
	ClearPasswordResetOTP(ctx context.Context, id uuid.UUID) error
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateFullName(ctx context.Context, id uuid.UUID, fullName string) (*domain.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, fullName string) (*domain.User, error)
	UpdateAvatarKey(ctx context.Context, id uuid.UUID, avatarKey *string) (*domain.User, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role string) (*domain.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, filter domain.UserListFilter) (*domain.UserListResult, error)
	Stats(ctx context.Context) (*domain.AdminStats, error)
}

type userRepository struct {
	db *sqlx.DB

	statsMu      sync.Mutex
	statsCache   *domain.AdminStats
	statsCacheAt time.Time
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, is_active, email_verified, verification_token, verification_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.IsActive,
		user.EmailVerified, user.VerificationToken, user.VerificationExpiresAt,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = $1`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateFullName(ctx context.Context, id uuid.UUID, fullName string) (*domain.User, error) {
	return r.UpdateUser(ctx, id, fullName)
}

func (r *userRepository) UpdateUser(ctx context.Context, id uuid.UUID, fullName string) (*domain.User, error) {
	query := `
		UPDATE users SET full_name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING *
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, fullName, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateAvatarKey(ctx context.Context, id uuid.UUID, avatarKey *string) (*domain.User, error) {
	query := `
		UPDATE users SET avatar_key = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING *
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, avatarKey, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update avatar: %w", err)
	}
	return &user, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email)
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}

func (r *userRepository) GetByVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE verification_token = $1`, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by verification token: %w", err)
	}
	return &user, nil
}

func (r *userRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET email_verified = TRUE, verification_token = NULL, verification_expires_at = NULL, updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) SetVerificationToken(ctx context.Context, id uuid.UUID, token string, expiresAt time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET verification_token = $1, verification_expires_at = $2, updated_at = NOW()
		WHERE id = $3
	`, token, expiresAt, id)
	if err != nil {
		return fmt.Errorf("set verification token: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) SetPasswordResetOTP(ctx context.Context, id uuid.UUID, otpHash string, expiresAt time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_reset_otp_hash = $1, password_reset_expires_at = $2, updated_at = NOW()
		WHERE id = $3
	`, otpHash, expiresAt, id)
	if err != nil {
		return fmt.Errorf("set password reset otp: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) ClearPasswordResetOTP(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_reset_otp_hash = NULL, password_reset_expires_at = NULL, updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("clear password reset otp: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $1, password_reset_otp_hash = NULL, password_reset_expires_at = NULL, updated_at = NOW()
		WHERE id = $2
	`, passwordHash, id)
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}
