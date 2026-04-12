package database

import (
	"core-banking/config"
	"core-banking/pkg/logging"
	"fmt"
	"time"

	"database/sql"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func NewPostgres(cfg *config.DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	driverName, err := otelsql.Register("pgx", otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		logging.Logger().Fatalw("Failed to register DB tracer", "error", err)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		logging.Logger().Fatalw("DB connection failed", "error", err)
	}

	if err = db.Ping(); err != nil {
		logging.Logger().Fatalw("DB ping failed", "error", err)
	}

	sqlxDB := sqlx.NewDb(db, "pgx")

	// Connection pool tuning
	sqlxDB.SetMaxOpenConns(25)
	sqlxDB.SetMaxIdleConns(10)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)

	logging.Logger().Infow("Connected to PostgreSQL", "host", cfg.Host, "db", cfg.Name)
	return sqlxDB
}
