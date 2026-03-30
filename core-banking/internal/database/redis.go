package database

import (
	"context"
	"core-banking/internal/config"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
	return client
}

// Close gracefully closes the Redis connection
func CloseRedis(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}
