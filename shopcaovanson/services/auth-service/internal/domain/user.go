package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	Email                 string     `db:"email" json:"email"`
	PasswordHash          string     `db:"password_hash" json:"-"`
	FullName              string     `db:"full_name" json:"full_name"`
	Role                  string     `db:"role" json:"role"`
	IsActive              bool       `db:"is_active" json:"is_active"`
	EmailVerified         bool       `db:"email_verified" json:"email_verified"`
	VerificationToken     *string    `db:"verification_token" json:"-"`
	VerificationExpiresAt *time.Time `db:"verification_expires_at" json:"-"`
	PasswordResetOTPHash  *string    `db:"password_reset_otp_hash" json:"-"`
	PasswordResetExpiresAt *time.Time `db:"password_reset_expires_at" json:"-"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
	CreatedAt time.Time `db:"created_at"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserProfile struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

func (u *User) ToProfile() UserProfile {
	return UserProfile{
		ID:            u.ID,
		Email:         u.Email,
		FullName:      u.FullName,
		Role:          u.Role,
		IsActive:      u.IsActive,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}

type RegisterResponse struct {
	Message           string `json:"message"`
	Email             string `json:"email"`
	RequiresVerify    bool   `json:"requires_verification"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type UserListFilter struct {
	Page   int
	Limit  int
	Search string
	Role   string
}

type UserListResult struct {
	Items      []UserProfile `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

type AdminStats struct {
	TotalUsers    int            `json:"total_users"`
	UsersByRole   map[string]int `json:"users_by_role"`
	ActiveUsers   int            `json:"active_users"`
	InactiveUsers int            `json:"inactive_users"`
}

type UpdateUserRoleInput struct {
	Role string `json:"role"`
}

type UpdateUserStatusInput struct {
	IsActive bool `json:"is_active"`
}
