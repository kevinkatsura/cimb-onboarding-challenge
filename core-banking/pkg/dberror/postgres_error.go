package dberror

import "github.com/jackc/pgx/v5/pgconn"

func IsSerializationError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "40001"
	}
	return false
}
