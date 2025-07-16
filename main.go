package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

var urlStore = make(map[string]URL) // URL is a struct that holds the original URL and its shortened version

type URL struct {
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsageCount  int        `json:"usage_count"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func main() {
	app := fiber.New()

	app.Post("/shorten", func(c *fiber.Ctx) error {
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

		shortCode := generateShortCode(6)
		shortURL := fmt.Sprintf("http://localhost:3000/%s", shortCode)

		createdAt := time.Now().UTC()
		expiresAt := createdAt.Add(24 * time.Hour) // Set expiration to 24 hours

		urlStore[shortCode] = URL{
			OriginalURL: req.OriginalURL,
			ShortURL:    shortURL,
			CreatedAt:   createdAt,
			ExpiresAt:   expiresAt,
			UsageCount:  0,
		}
		return c.JSON(urlStore[shortCode])
	})

	app.Get("/:shortCode", func(c *fiber.Ctx) error {
		shortCode := c.Params("shortCode")
		url, exists := urlStore[shortCode]
		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
		}
		if url.DeletedAt != nil {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "URL has been deleted"})
		}
		if time.Now().UTC().After(url.ExpiresAt) {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "URL has expired"})
		}

		url.UsageCount++
		urlStore[shortCode] = url
		return c.Redirect(url.OriginalURL, fiber.StatusTemporaryRedirect)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(urlStore)
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
