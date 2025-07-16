package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
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

func main() {
	app := fiber.New()
	initDatabase()

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
		return c.JSON(url)
	})

	app.Get("/:shortCode", func(c *fiber.Ctx) error {
		shortCode := c.Params("shortCode")
		var url URL
		if err := db.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
		}
		if time.Now().UTC().After(url.ExpiresAt) {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "URL has expired"})
		}

		url.UsageCount++
		if err := db.Save(&url).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update usage count"})
		}
		return c.Redirect(url.OriginalURL, fiber.StatusTemporaryRedirect)
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
