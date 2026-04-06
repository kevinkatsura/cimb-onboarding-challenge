package config

import (
	"os"

	"core-banking/pkg/logging"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadConfig() *DBConfig {
	err := godotenv.Load()
	if err != nil {
		logging.Logger().Warn("No .env file found, using system env for database config")
	}

	return &DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", "banking"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func LoadRedisConfig() *RedisConfig {
	err := godotenv.Load()
	if err != nil {
		logging.Logger().Warn("No .env file found, using system env for redis config")
	}

	return &RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
