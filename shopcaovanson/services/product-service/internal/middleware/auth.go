package middleware

import (
	"github.com/caovanson/shopcaovanson/product-service/internal/rbac"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	ContextUserIDKey   = "user_id"
	ContextUserRoleKey = "user_role"
)

func GatewayAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-Id")
		role := c.Get("X-User-Role")
		if userID != "" {
			if id, err := uuid.Parse(userID); err == nil {
				c.Locals(ContextUserIDKey, id)
			}
		}
		if role != "" {
			c.Locals(ContextUserRoleKey, role)
		}
		return c.Next()
	}
}

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, ok := c.Locals(ContextUserIDKey).(uuid.UUID); !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		return c.Next()
	}
}

func RequireMinRole(minRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(ContextUserRoleKey).(string)
		if !rbac.HasMinRole(role, minRole) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}
		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return RequireMinRole(rbac.RoleManager)
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(ContextUserIDKey).(uuid.UUID)
	return id, ok
}

func GetUserRole(c *fiber.Ctx) string {
	role, _ := c.Locals(ContextUserRoleKey).(string)
	return role
}
