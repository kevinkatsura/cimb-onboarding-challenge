package database

import (
	"context"
	"core-banking/internal/dberror"
	"time"
)

const maxRetries = 5

func WithSerializableRetry(ctx context.Context, fn func() error) error {
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
