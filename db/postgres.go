package db

import (
	"fmt"
	"log"
	"yunus/url-shortener/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // Assume db is initialized and connected to a database
func InitDatabase() {
	// Initialize the database connection here
	dsn := "host=localhost user=urluser dbname=urlshortener password=12345 port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}
	fmt.Println("Database connected successfully")
	database.AutoMigrate(&models.URL{}) // Auto migrate the URL struct to create the table
	DB = database
}
