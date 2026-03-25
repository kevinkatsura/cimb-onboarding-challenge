package database

import (
	"core-banking/internal/config"
	"errors"
	"log"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrateUp(cfg *config.DBConfig) {
	// Resolve absolute path dynamically
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("failed to resolve migrations path: %v", err)
	}

	// Build proper file:// URI (handles spaces safely)
	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   absPath,
	}).String()

	log.Println("Migration source:", sourceURL)

	m, err := migrate.New(sourceURL, buildMigrationDSN(cfg))
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	// Run migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration up failed: %v", err)
	}

	log.Println("Migrations UP applied")
}
func RunMigrateDown(cfg *config.DBConfig) {
	// Resolve absolute path dynamically
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("failed to resolve migrations path: %v", err)
	}

	// Build proper file:// URI (handles spaces safely)
	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   absPath,
	}).String()

	log.Println("Migration source:", sourceURL)

	m, err := migrate.New(sourceURL, buildMigrationDSN(cfg))
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	// Run migrations
	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration up failed: %v", err)
	}

	log.Println("Migrations UP applied")
}

func buildMigrationDSN(cfg *config.DBConfig) string {
	return "postgres://" +
		cfg.User + ":" + cfg.Password + "@" +
		cfg.Host + ":" + cfg.Port + "/" +
		cfg.Name + "?sslmode=" + cfg.SSLMode
}
