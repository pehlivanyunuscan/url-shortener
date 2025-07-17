package middleware

import (
	"fmt"
	"time"
	"yunus/url-shortener/db"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(limit int, duration time.Duration) fiber.Handler { // RateLimitMiddleware limits the number of requests from a single IP address
	// limit: maximum number of requests allowed
	// duration: time window for the limit
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		// Use Redis to track the number of requests
		// for the given IP address within the specified duration
		count, err := db.RedisClient.Get(db.Ctx, key).Int()
		if err != nil && err != redis.Nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check rate limit"})
		}
		if count >= limit {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Rate limit exceeded"})
		}
		// Increment the count and set the TTL
		pipe := db.RedisClient.TxPipeline()
		pipe.Incr(db.Ctx, key)
		pipe.Expire(db.Ctx, key, duration)
		_, _ = pipe.Exec(db.Ctx)
		return c.Next()
	}
}
