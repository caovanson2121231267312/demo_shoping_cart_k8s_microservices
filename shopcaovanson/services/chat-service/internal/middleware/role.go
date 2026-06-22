package middleware

import (
	"github.com/caovanson/shopcaovanson/chat-service/internal/staff"
	"github.com/gofiber/fiber/v2"
)

func RequireStaff() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(UserRoleKey).(string)
		if !staff.IsStaffRole(role) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "staff access required"})
		}
		return c.Next()
	}
}

func GetUserRole(c *fiber.Ctx) string {
	if v, ok := c.Locals(UserRoleKey).(string); ok {
		return v
	}
	return ""
}
