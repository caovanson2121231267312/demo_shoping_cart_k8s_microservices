package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	Email                 string     `db:"email" json:"email"`
	PasswordHash          string     `db:"password_hash" json:"-"`
	FullName              string     `db:"full_name" json:"full_name"`
	AvatarKey             *string    `db:"avatar_key" json:"-"`
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
	HasAvatar     bool      `json:"has_avatar"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func AvatarURLForUser(id uuid.UUID) string {
	return fmt.Sprintf("/api/admin/users/%s/avatar", id.String())
}

func (u *User) ToProfile() UserProfile {
	profile := UserProfile{
		ID:            u.ID,
		Email:         u.Email,
		FullName:      u.FullName,
		Role:          u.Role,
		IsActive:      u.IsActive,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
	if u.AvatarKey != nil && *u.AvatarKey != "" {
		profile.HasAvatar = true
		profile.AvatarURL = AvatarURLForUser(u.ID)
	}
	return profile
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
	Page        int
	Limit       int
	Cursor      string
	Search      string
	Role        string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type UserListResult struct {
	Items      []UserProfile `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
	NextCursor string        `json:"next_cursor,omitempty"`
	HasMore    bool          `json:"has_more"`
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

type UpdateUserInput struct {
	FullName string `json:"full_name"`
}
