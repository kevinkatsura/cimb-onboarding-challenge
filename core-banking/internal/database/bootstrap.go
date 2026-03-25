package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"core-banking/internal/config"
)

func EnsureDatabase(cfg *config.DBConfig) {
	// Connect to default postgres DB
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to postgres DB: %v", err)
	}
	defer db.Close()

	// Check existence
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM pg_database WHERE datname = $1
		)
	`
	err = db.QueryRow(query, cfg.Name).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check DB existence: %v", err)
	}

	if exists {
		log.Println("Database already exists:", cfg.Name)
		return
	}

	// Create DB
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Name))
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	log.Println("Database created:", cfg.Name)
}
