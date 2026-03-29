package account

import "errors"

const (
	StatusActive = "active"
	StatusFrozen = "frozen"
	StatusClosed = "closed"
)

var (
	ErrInvalidStatus    = errors.New("invalid status")
	ErrNonZeroBalance   = errors.New("cannot delete account with non-zero balance")
	ErrAccountNotClosed = errors.New("account must be closed before deletion")
)
