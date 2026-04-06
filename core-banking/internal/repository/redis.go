package repository

import (
	"context"
	"core-banking/config"
	"core-banking/pkg/logging"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
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
		logging.Logger().Fatalw("Failed to connect to Redis", "error", err)
	}

	if err := redisotel.InstrumentTracing(client); err != nil {
		logging.Logger().Warnw("Failed to instrument Redis", "error", err)
	}

	logging.Logger().Infow("Connected to Redis")
	return client
}

// Close gracefully closes the Redis connection
func CloseRedis(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}
