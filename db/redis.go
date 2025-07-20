package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client // Assume redisClient is initialized and connected to a Redis server
var Ctx = context.Background()

func InitRedis() {
	// Read Redis host and port from environment variables
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Set password here if needed
		DB:       0,
	})
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
	fmt.Println("Redis connected successfully")
}
