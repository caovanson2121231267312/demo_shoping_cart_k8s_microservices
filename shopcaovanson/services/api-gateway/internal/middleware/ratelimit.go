package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RateLimit(redisClient *redis.Client, limitPerMinute int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if limitPerMinute <= 0 {
			return c.Next()
		}

		ip := c.IP()
		minute := time.Now().UTC().Format("200601021504")
		key := fmt.Sprintf("ratelimit:%s:%s", ip, minute)

		ctx := context.Background()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "rate limiter unavailable",
			})
		}

		if count == 1 {
			redisClient.Expire(ctx, key, 70*time.Second)
		}

		if count > int64(limitPerMinute) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}

		return c.Next()
	}
}
