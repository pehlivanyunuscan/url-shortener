package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type URL struct {
	OriginalURL string     `gorm:"not null;unique" json:"original_url"`
	ShortURL    string     `gorm:"not null;unique" json:"short_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsageCount  int        `json:"usage_count"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

var db *gorm.DB // Assume db is initialized and connected to a database
func initDatabase() {
	// Initialize the database connection here
	dsn := "host=localhost user=urluser dbname=urlshortener password=12345 port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}
	fmt.Println("Database connected successfully")
	database.AutoMigrate(&URL{}) // Auto migrate the URL struct to create the table
	db = database
}

var redisClient *redis.Client // Assume redisClient is initialized and connected to a Redis server
var ctx = context.Background()

func initRedis() {
	// Initialize the Redis connection here
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
	fmt.Println("Redis connected successfully")
}

func main() {
	app := fiber.New()
	initDatabase()
	initRedis()

	app.Post("/shorten", RateLimitMiddleware(5, time.Minute), func(c *fiber.Ctx) error {
		type Request struct {
			OriginalURL string `json:"original_url"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}
		if req.OriginalURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Original URL is required"})
		}
		var existing URL
		if err := db.Where("original_url = ?", req.OriginalURL).First(&existing).Error; err == nil {
			// URL already exists, return the existing short URL
			return c.JSON(existing)
		}

		shortCode := generateShortCode(6)

		url := URL{
			OriginalURL: req.OriginalURL,
			ShortURL:    shortCode,
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
			UsageCount:  0,
		}

		if err := db.Create(&url).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create URL"})
		}
		err := redisClient.Set(ctx, shortCode, url.OriginalURL, 24*time.Hour).Err()
		if err != nil {
			log.Println("Failed to cache URL in Redis:", err)
		}
		return c.JSON(url)
	})

	app.Get("/:shortCode", func(c *fiber.Ctx) error {
		shortCode := c.Params("shortCode")
		var url URL

		originalURL, err := redisClient.Get(ctx, shortCode).Result()
		if err == redis.Nil {
			// If not found in Redis, check the database
			if err := db.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
			}
			// Cache the URL in Redis for future requests
			redisClient.Set(ctx, shortCode, url.OriginalURL, time.Until(url.ExpiresAt))
			originalURL = url.OriginalURL
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch URL from Redis"})
		}
		if err := db.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
		}
		if time.Now().UTC().After(url.ExpiresAt) {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "URL has expired"})
		}

		db.Model(&url).Where("short_url = ?", shortCode).Update("usage_count", gorm.Expr("usage_count + ?", 1))
		return c.Redirect(originalURL, fiber.StatusFound)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		var urls []URL
		if err := db.Find(&urls).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch URLs"})
		}
		return c.JSON(urls)
	})

	app.Listen(":3000")
}

func generateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func RateLimitMiddleware(limit int, duration time.Duration) fiber.Handler { // RateLimitMiddleware limits the number of requests from a single IP address
	// limit: maximum number of requests allowed
	// duration: time window for the limit
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		// Use Redis to track the number of requests
		// for the given IP address within the specified duration
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check rate limit"})
		}
		if count >= limit {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Rate limit exceeded"})
		}
		// Increment the count and set the TTL
		pipe := redisClient.TxPipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, duration)
		_, _ = pipe.Exec(ctx)
		return c.Next()
	}
}
