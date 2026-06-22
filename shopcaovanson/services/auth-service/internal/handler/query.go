package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func parseCreatedRange(c *fiber.Ctx) (from, to *time.Time) {
	if v := c.Query("created_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = &t
		}
	}
	if v := c.Query("created_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = &t
		}
	}
	return from, to
}
