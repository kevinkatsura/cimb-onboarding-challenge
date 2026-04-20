package database

import (
	"database/sql"
	"fmt"
	"time"

	"notification-service/config"
	"notification-service/pkg/logging"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgres(cfg *config.DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=notification,public",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("DB connection failed", "error", err)
	}
	if err = db.Ping(); err != nil {
		logging.Logger().Fatalw("DB ping failed", "error", err)
	}
	sqlxDB := sqlx.NewDb(db, "pgx")
	sqlxDB.SetMaxOpenConns(10)
	sqlxDB.SetMaxIdleConns(5)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)
	logging.Logger().Infow("Connected to PostgreSQL", "host", cfg.Host, "db", cfg.Name)
	return sqlxDB
}

func EnsureSchema(cfg *config.DBConfig) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("failed to connect for schema", "error", err)
	}
	defer db.Close()
	_, _ = db.Exec("CREATE SCHEMA IF NOT EXISTS notification")
	logging.Logger().Infow("Schema ensured", "schema", "notification")
}
