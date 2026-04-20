package database

import (
	"errors"
	"net/url"
	"path/filepath"

	"notification-service/config"
	"notification-service/pkg/logging"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrateUp(cfg *config.DBConfig) {
	absPath, _ := filepath.Abs("migrations")
	sourceURL := (&url.URL{Scheme: "file", Path: absPath}).String()
	m, err := migrate.New(sourceURL, buildDSN(cfg))
	if err != nil {
		logging.Logger().Fatalw("migration init failed", "error", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logging.Logger().Fatalw("migration up failed", "error", err)
	}
	logging.Logger().Infow("Migrations applied")
}

func RunMigrateDown(cfg *config.DBConfig) {
	absPath, _ := filepath.Abs("migrations")
	sourceURL := (&url.URL{Scheme: "file", Path: absPath}).String()
	m, err := migrate.New(sourceURL, buildDSN(cfg))
	if err != nil {
		logging.Logger().Fatalw("migration init failed", "error", err)
	}
	if err = m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logging.Logger().Fatalw("migration down failed", "error", err)
	}
	logging.Logger().Infow("Migrations reverted")
}

func buildDSN(cfg *config.DBConfig) string {
	return "pgx5://" + cfg.User + ":" + cfg.Password + "@" + cfg.Host + ":" + cfg.Port + "/" + cfg.Name + "?sslmode=" + cfg.SSLMode + "&search_path=notification,public"
}
