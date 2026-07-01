package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func parseAllowedOrigins(allowedOrigins string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, o := range strings.Split(allowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			out[o] = struct{}{}
		}
	}
	return out
}

func originAllowed(origin string, allowed map[string]struct{}) bool {
	if origin == "" {
		return false
	}
	_, ok := allowed[origin]
	return ok
}

func applyCORSHeaders(c *fiber.Ctx, origin string) {
	c.Set("Access-Control-Allow-Origin", origin)
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Set("Vary", "Origin")
	c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
	c.Set(
		"Access-Control-Allow-Headers",
		"Origin,Content-Type,Accept,Authorization,X-User-Id,X-User-Role,X-User-Email",
	)
	c.Set("Access-Control-Expose-Headers", "Content-Length,Content-Type")
	c.Set("Access-Control-Max-Age", "43200")
}

// CORS must be registered first (outermost) so headers are applied after proxy/errors.
func CORS(allowedOrigins string) fiber.Handler {
	allowed := parseAllowedOrigins(allowedOrigins)

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		if c.Method() == fiber.MethodOptions {
			if originAllowed(origin, allowed) {
				applyCORSHeaders(c, origin)
				return c.SendStatus(fiber.StatusNoContent)
			}
			return c.SendStatus(fiber.StatusForbidden)
		}

		err := c.Next()
		if originAllowed(origin, allowed) {
			applyCORSHeaders(c, origin)
		}
		return err
	}
}
