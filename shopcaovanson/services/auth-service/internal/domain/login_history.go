package domain

import (
	"time"

	"github.com/google/uuid"
)

type LoginHistory struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Email         string     `json:"email" db:"email"`
	Success       bool       `json:"success" db:"success"`
	FailureReason *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	IPAddress     *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string    `json:"user_agent,omitempty" db:"user_agent"`
	Device        *string    `json:"device,omitempty" db:"device"`
	Browser       *string    `json:"browser,omitempty" db:"browser"`
	OS            *string    `json:"os,omitempty" db:"os"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	FullName      *string    `json:"full_name,omitempty" db:"full_name"`
}

type LoginHistoryFilter struct {
	Page        int
	Limit       int
	Cursor      string
	Email       string
	UserID      string
	Success     *bool
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type LoginHistoryResult struct {
	Items      []LoginHistory `json:"items"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type LoginReport struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ReportDate   time.Time `json:"report_date" db:"report_date"`
	ObjectKey    string    `json:"object_key" db:"object_key"`
	FileName     string    `json:"file_name" db:"file_name"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	TotalLogins  int       `json:"total_logins" db:"total_logins"`
	SuccessCount int       `json:"success_count" db:"success_count"`
	FailureCount int       `json:"failure_count" db:"failure_count"`
	UniqueUsers  int       `json:"unique_users" db:"unique_users"`
	Status       string    `json:"status" db:"status"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type LoginReportListResult struct {
	Items []LoginReport `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

type LoginMeta struct {
	IPAddress string
	UserAgent string
}
