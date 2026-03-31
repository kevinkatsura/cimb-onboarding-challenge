package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Tx interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
	QueryRowx(query string, args ...interface{}) *sqlx.Row
}

type sqlxTxWrapper struct {
	tx *sqlx.Tx
}

func (s *sqlxTxWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	return s.tx.Get(dest, query, args...)
}

func (s *sqlxTxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.tx.Exec(query, args...)
}

func (s *sqlxTxWrapper) Commit() error {
	return s.tx.Commit()
}

func (s *sqlxTxWrapper) Rollback() error {
	return s.tx.Rollback()
}

func (s *sqlxTxWrapper) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return s.tx.QueryRowx(query, args...)
}
