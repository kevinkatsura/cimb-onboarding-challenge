package config

import (
	"os"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	KafkaBrokers string
	GRPCPort     string
	HTTPPort     string
}

func LoadConfig() Config {
	return Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "pjp_db"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		GRPCPort:     getEnv("GRPC_PORT", ":50055"),
		HTTPPort:     getEnv("HTTP_PORT", ":8081"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
