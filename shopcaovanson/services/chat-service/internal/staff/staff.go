package staff

import (
	"os"
	"strings"

	"github.com/google/uuid"
)

// DefaultSupportStaffID matches auth-service seed for support@shop.com
const DefaultSupportStaffID = "d0773246-879b-522b-b5f9-35e4722abc4e"

var staffRoles = map[string]bool{
	"support":     true,
	"staff":       true,
	"manager":     true,
	"admin":       true,
	"super_admin": true,
}

func SupportUserID() string {
	if v := strings.TrimSpace(os.Getenv("SUPPORT_STAFF_USER_ID")); v != "" {
		return v
	}
	return DefaultSupportStaffID
}

func IsStaffRole(role string) bool {
	return staffRoles[strings.ToLower(strings.TrimSpace(role))]
}

func AuthSeedStaffID(email string) string {
	seedNS := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	return uuid.NewSHA1(seedNS, []byte("staff-"+strings.ToLower(email))).String()
}
