package apperror

import "fmt"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func NewInternal(msg string, err error) *AppError {
	return &AppError{Code: 500, Message: msg, Err: err}
}

func NewBadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

func NewNotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}
