package config

import "os"

type Config struct {
	DBDriver string
	DBSource string
}

func LoadConfig() Config {
	return Config{
		DBDriver: "postgres",
		DBSource: os.Getenv("DB_SOURCE"),
	}
}
