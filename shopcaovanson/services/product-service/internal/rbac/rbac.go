package rbac

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleManager    = "manager"
	RoleStaff      = "staff"
	RoleSupport    = "support"
	RoleCustomer   = "customer"
)

var roleLevel = map[string]int{
	RoleSuperAdmin: 100,
	RoleAdmin:      80,
	RoleManager:    60,
	RoleStaff:      40,
	RoleSupport:    20,
	RoleCustomer:   0,
}

func Level(role string) int {
	if v, ok := roleLevel[role]; ok {
		return v
	}
	if role == "admin" {
		return 80
	}
	return 0
}

func HasMinRole(role, minRole string) bool {
	return Level(role) >= Level(minRole)
}

func CanManageProducts(role string) bool {
	return Level(role) >= Level(RoleManager)
}

func CanManageCategories(role string) bool {
	return Level(role) >= Level(RoleManager)
}
