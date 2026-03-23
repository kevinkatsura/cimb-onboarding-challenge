package dberror

import "github.com/lib/pq"

func IsSerializationError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "40001"
	}
	return false
}
