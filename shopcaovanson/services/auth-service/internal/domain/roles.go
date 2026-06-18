package domain

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleManager    = "manager"
	RoleStaff      = "staff"
	RoleSupport    = "support"
	RoleCustomer   = "customer"
)

var StaffRoles = []string{
	RoleSuperAdmin,
	RoleAdmin,
	RoleManager,
	RoleStaff,
	RoleSupport,
}

var roleLevel = map[string]int{
	RoleSuperAdmin: 100,
	RoleAdmin:      80,
	RoleManager:    60,
	RoleStaff:      40,
	RoleSupport:    20,
	RoleCustomer:   0,
}

func RoleLevel(role string) int {
	if v, ok := roleLevel[role]; ok {
		return v
	}
	return 0
}

func HasMinRole(role, minRole string) bool {
	return RoleLevel(role) >= RoleLevel(minRole)
}

func IsStaffRole(role string) bool {
	return RoleLevel(role) >= RoleLevel(RoleSupport)
}

func CanManageUsers(role string) bool {
	return RoleLevel(role) >= RoleLevel(RoleAdmin)
}

func CanAssignRole(actorRole, targetRole string) bool {
	return RoleLevel(actorRole) > RoleLevel(targetRole)
}

func ValidStaffRoles() []string {
	return StaffRoles
}
