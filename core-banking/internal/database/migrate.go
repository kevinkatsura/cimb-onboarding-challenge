package database

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB, path string) {
	schema, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Schema migrated successfully")
}
