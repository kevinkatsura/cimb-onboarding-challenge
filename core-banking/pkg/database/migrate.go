package database

import (
	"core-banking/config"
	"core-banking/pkg/logging"
	"errors"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrateUp(cfg *config.DBConfig) {
	// Resolve absolute path dynamically
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		logging.Logger().Fatalw("failed to resolve migrations path", "error", err)
	}

	// Build proper file:// URI (handles spaces safely)
	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   absPath,
	}).String()

	logging.Logger().Infow("Migration source", "sourceURL", sourceURL)

	m, err := migrate.New(sourceURL, buildMigrationDSN(cfg))
	if err != nil {
		logging.Logger().Fatalw("migration init failed", "error", err)
	}

	// Run migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logging.Logger().Fatalw("migration up failed", "error", err)
	}

	logging.Logger().Infow("Migrations UP applied")
}
func RunMigrateDown(cfg *config.DBConfig) {
	// Resolve absolute path dynamically
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		logging.Logger().Fatalw("failed to resolve migrations path", "error", err)
	}

	// Build proper file:// URI (handles spaces safely)
	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   absPath,
	}).String()

	logging.Logger().Infow("Migration source", "sourceURL", sourceURL)

	m, err := migrate.New(sourceURL, buildMigrationDSN(cfg))
	if err != nil {
		logging.Logger().Fatalw("migration init failed", "error", err)
	}

	// Run migrations
	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logging.Logger().Fatalw("migration down failed", "error", err)
	}

	logging.Logger().Infow("Migrations DOWN applied")
}

func buildMigrationDSN(cfg *config.DBConfig) string {
	return "pgx5://" +
		cfg.User + ":" + cfg.Password + "@" +
		cfg.Host + ":" + cfg.Port + "/" +
		cfg.Name + "?sslmode=" + cfg.SSLMode
}
