package database

import (
	"payment-initiation-acquiring-service/config"
	"payment-initiation-acquiring-service/pkg/logging"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
	})
	logging.Logger().Infow("Connected to Redis", "host", cfg.Host)
	return client
}
