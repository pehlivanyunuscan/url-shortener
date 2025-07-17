package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"yunus/url-shortener/db"
	"yunus/url-shortener/models"
	"yunus/url-shortener/utils"
)

func ShortenURL(c *fiber.Ctx) error {
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
	var existing models.URL
	if err := db.DB.Unscoped().Where("original_url = ?", req.OriginalURL).First(&existing).Error; err == nil {
		// URL already exists, return the existing short URL
		return c.JSON(existing)
	}

	shortCode := utils.GenerateShortCode(6)

	url := models.URL{
		OriginalURL: req.OriginalURL,
		ShortURL:    shortCode,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		UsageCount:  0,
	}

	if err := db.DB.Create(&url).Error; err != nil {
		log.Println("DB Create Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create URL"})
	}
	err := db.RedisClient.Set(db.Ctx, shortCode, url.OriginalURL, 24*time.Hour).Err()
	if err != nil {
		log.Println("Failed to cache URL in Redis:", err)
	}

	response := fiber.Map{
		"created_at":   url.CreatedAt,
		"deleted_at":   nil,
		"original_url": url.OriginalURL,
		"short_url":    "http://localhost:3000/" + url.ShortURL,
		"expires_at":   url.ExpiresAt,
		"usage_count":  url.UsageCount,
	}
	return c.JSON(response)
}

func RedirectURL(c *fiber.Ctx) error {
	shortCode := c.Params("shortCode")
	var url models.URL

	originalURL, err := db.RedisClient.Get(db.Ctx, shortCode).Result()
	if err == redis.Nil {
		// If not found in Redis, check the database
		if err := db.DB.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
		}
		// Cache the URL in Redis for future requests
		db.RedisClient.Set(db.Ctx, shortCode, url.OriginalURL, time.Until(url.ExpiresAt))
		originalURL = url.OriginalURL
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch URL from Redis"})
	}
	if err := db.DB.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
	}
	if time.Now().UTC().After(url.ExpiresAt) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "URL has expired"})
	}

	db.DB.Model(&url).Where("short_url = ?", shortCode).Update("usage_count", gorm.Expr("usage_count + ?", 1))
	return c.Redirect(originalURL, fiber.StatusFound)
}

func ListURL(c *fiber.Ctx) error {
	var urls []models.URL
	if err := db.DB.Find(&urls).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch URLs"})
	}
	return c.JSON(urls)
}

func DeleteURL(c *fiber.Ctx) error {
	shortCode := c.Params("shortCode")
	var url models.URL
	if err := db.DB.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
	}

	if err := db.DB.Delete(&url).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete URL"})
	}

	// Remove from Redis cache
	if err := db.RedisClient.Del(db.Ctx, shortCode).Err(); err != nil {
		log.Println("Failed to delete URL from Redis:", err)
	}
	return c.JSON(fiber.Map{"message": "URL deleted successfully"})
}

func StatsURL(c *fiber.Ctx) error {
	shortCode := c.Params("shortCode")
	var url models.URL
	if err := db.DB.Where("short_url = ?", shortCode).First(&url).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
	}
	return c.JSON(fiber.Map{
		"short_url":    url.ShortURL,
		"original_url": url.OriginalURL,
		"usage_count":  url.UsageCount,
		"created_at":   url.CreatedAt,
		"expires_at":   url.ExpiresAt,
	})
}
