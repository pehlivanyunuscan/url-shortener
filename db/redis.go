package db

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client // Assume redisClient is initialized and connected to a Redis server
var Ctx = context.Background()

func InitRedis() {
	// Initialize the Redis connection here
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
	fmt.Println("Redis connected successfully")
}
