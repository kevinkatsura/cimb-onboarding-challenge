package database

import (
	"database/sql"
	"fmt"

	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/config"
	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/pkg/logging"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func EnsureDatabase(cfg config.Config) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("failed to connect", "error", err)
	}
	defer db.Close()
	var exists bool
	_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", cfg.DBName).Scan(&exists)
	if !exists {
		_, _ = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		logging.Logger().Infow("Database created", "name", cfg.DBName)
	}
}
