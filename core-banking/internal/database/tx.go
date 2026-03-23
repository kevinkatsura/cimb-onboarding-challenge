package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func BeginSerializableTx(ctx context.Context, db *sqlx.DB) (*sqlx.Tx, error) {
	return db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
}
