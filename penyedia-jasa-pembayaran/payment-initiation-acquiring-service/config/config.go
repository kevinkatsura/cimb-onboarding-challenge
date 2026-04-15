package config

import (
	"os"
	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host, Port, User, Password, Name, SSLMode string
}

type RedisConfig struct {
	Host, Port string
}

func LoadConfig() *DBConfig {
	_ = godotenv.Load()
	return &DBConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "postgres"), Password: getEnv("DB_PASSWORD", "postgres"),
		Name: getEnv("DB_NAME", "pjp"), SSLMode: getEnv("DB_SSLMODE", "disable"),
	}
}

func LoadRedisConfig() *RedisConfig {
	_ = godotenv.Load()
	return &RedisConfig{Host: getEnv("REDIS_HOST", "localhost"), Port: getEnv("REDIS_PORT", "6379")}
}

func getEnv(key, fb string) string {
	if v, ok := os.LookupEnv(key); ok { return v }
	return fb
}
