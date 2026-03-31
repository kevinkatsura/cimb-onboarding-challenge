package repository

import (
	"context"
	"core-banking/pkg/dberror"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type TxManager interface {
	WithSerializableRetry(ctx context.Context, fn func() error) error
	BeginSerializableTx(ctx context.Context) (*sqlxTxWrapper, error)
}

type DefaultTxManager struct {
	DB *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *DefaultTxManager {
	return &DefaultTxManager{DB: db}
}

func (d *DefaultTxManager) WithSerializableRetry(ctx context.Context, fn func() error) error {
	const maxRetries = 5
	var err error

	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !dberror.IsSerializationError(err) {
			return err
		}

		time.Sleep(time.Duration(50*(1<<i)) * time.Millisecond)
	}

	return err
}

func (d *DefaultTxManager) BeginSerializableTx(ctx context.Context) (Tx, error) {
	tx, err := d.DB.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	return &sqlxTxWrapper{tx: tx}, nil
}
