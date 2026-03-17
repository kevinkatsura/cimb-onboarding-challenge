package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewSQLXConnection(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "host=localhost user=katsuke dbname=go_db_exercise sslmode=disable")
	return db, err
}

func NewGormConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open("host=localhost user=katsuke dbname=go_db_exercise sslmode=disable"), &gorm.Config{})
	return db, err
}
