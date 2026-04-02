package repository

import (
	"core-banking/config"
	"core-banking/pkg/logging"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgres(cfg *config.DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		logging.Logger().Fatalw("DB connection failed", "error", err)
	}

	// Connection pool tuning
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	logging.Logger().Infow("Connected to PostgreSQL", "host", cfg.Host, "db", cfg.Name)
	return db
}
