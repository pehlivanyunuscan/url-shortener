package main

import (
	"time"
	"yunus/url-shortener/db"
	"yunus/url-shortener/handlers"
	"yunus/url-shortener/middleware"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	db.InitDatabase()
	db.InitRedis()

	app.Post("/shorten", middleware.RateLimitMiddleware(5, time.Minute), handlers.ShortenURL)
	app.Get("/:shortCode", handlers.RedirectURL)
	app.Get("/", handlers.ListURL)
	app.Delete("/:shortCode", handlers.DeleteURL)
	app.Get("/:shortCode/stats", handlers.StatsURL)

	app.Listen(":3000")
}
