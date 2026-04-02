package repository

import (
	"fmt"

	"core-banking/config"
	"core-banking/pkg/logging"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func EnsureDatabase(cfg *config.DBConfig) {
	// Connect to default postgres DB
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("Failed to connect to postgres DB", "error", err)
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
		logging.Logger().Fatalw("Failed to check DB existence", "error", err)
	}

	if exists {
		logging.Logger().Infow("Database already exists", "name", cfg.Name)
		return
	}

	// Create DB
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Name))
	if err != nil {
		logging.Logger().Fatalw("Failed to create database", "error", err)
	}

	logging.Logger().Infow("Database created", "name", cfg.Name)
}
