package db

import (
	"fmt"
	"log"
	"os"
	"yunus/url-shortener/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // Assume db is initialized and connected to a database

func InitDatabase() {
	// Initialize the database connection here
	host := os.Getenv("DB_HOST") // Get the database host from environment variables
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%s sslmode=disable", host, user, dbname, password, port)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}
	fmt.Println("Database connected successfully")
	database.AutoMigrate(&models.URL{}) // Auto migrate the URL struct to create the table
	DB = database
}
