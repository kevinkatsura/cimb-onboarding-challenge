package database

import (
	"payment-initiation-acquiring-service/config"
	"payment-initiation-acquiring-service/pkg/logging"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgres(cfg *config.DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("DB connection failed", "error", err)
	}
	if err = db.Ping(); err != nil {
		logging.Logger().Fatalw("DB ping failed", "error", err)
	}
	sqlxDB := sqlx.NewDb(db, "pgx")
	sqlxDB.SetMaxOpenConns(25)
	sqlxDB.SetMaxIdleConns(10)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)
	logging.Logger().Infow("Connected to PostgreSQL", "host", cfg.Host, "db", cfg.Name)
	return sqlxDB
}

func EnsureDatabase(cfg *config.DBConfig) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("failed to connect", "error", err)
	}
	defer db.Close()
	var exists bool
	_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", cfg.Name).Scan(&exists)
	if !exists {
		_, _ = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Name))
		logging.Logger().Infow("Database created", "name", cfg.Name)
	}
}
