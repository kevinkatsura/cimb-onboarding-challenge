package database

import (
	"errors"
	"net/url"
	"notification-service/config"
	"notification-service/pkg/logging"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrateUp(cfg *config.DBConfig) {
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		logging.Logger().Fatalw("failed to resolve migrations path", "error", err)
	}
	sourceURL := (&url.URL{Scheme: "file", Path: absPath}).String()
	m, err := migrate.New(sourceURL, buildDSN(cfg))
	if err != nil {
		logging.Logger().Fatalw("migration init failed", "error", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if strings.Contains(err.Error(), "dirty") {
			logging.Logger().Warnw("database is dirty, trying to force version", "error", err)
			version, _, vErr := m.Version()
			if vErr == nil {
				if errForce := m.Force(int(version)); errForce == nil {
					logging.Logger().Infow("forced version, retrying migration", "version", version)
					if errUp := m.Up(); errUp == nil || errors.Is(errUp, migrate.ErrNoChange) {
						logging.Logger().Infow("Migrations UP applied after force")
						return
					}
				}
			}
		}
		logging.Logger().Fatalw("migration up failed", "error", err)
	}
	logging.Logger().Infow("Migrations UP applied")
}

func RunMigrateDown(cfg *config.DBConfig) {
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		logging.Logger().Fatalw("failed to resolve migrations path", "error", err)
	}
	sourceURL := (&url.URL{Scheme: "file", Path: absPath}).String()
	m, err := migrate.New(sourceURL, buildDSN(cfg))
	if err != nil {
		logging.Logger().Fatalw("migration init failed", "error", err)
	}
	if err = m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logging.Logger().Fatalw("migration down failed", "error", err)
	}
	logging.Logger().Infow("Migrations DOWN applied")
}

func buildDSN(cfg *config.DBConfig) string {
	return "pgx5://" + cfg.User + ":" + cfg.Password + "@" + cfg.Host + ":" + cfg.Port + "/" + cfg.Name + "?sslmode=" + cfg.SSLMode + "&search_path=notification,public"
}
