package database

import (
	"errors"
	"log"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate"
)

// func RunMigrations(db *sqlx.DB, path string) {
// 	schema, err := os.ReadFile(path)
// 	if err != nil {
// 		log.Fatalf("Failed to read schema: %v", err)
// 	}

// 	_, err = db.Exec(string(schema))
// 	if err != nil {
// 		log.Fatalf("Migration failed: %v", err)
// 	}

// 	log.Println("Schema migrated successfully")
// }

func RunMigrateUp(dsn string) {
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

	m, err := migrate.New(sourceURL, dsn)
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
func RunMigrateDown(dsn string) {
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

	m, err := migrate.New(sourceURL, dsn)
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
