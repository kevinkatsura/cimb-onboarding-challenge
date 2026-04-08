package apperror

import "fmt"

type Type int

const (
	NotFound    Type = iota // 404
	BadRequest              // 400
	Conflict                // 409
	Forbidden               // 403
	Internal                // 500
	Unavailable             // 503
)

type AppError struct {
	Type    Type
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(t Type, message string) *AppError {
	return &AppError{Type: t, Message: message}
}

func Wrap(t Type, message string, err error) *AppError {
	return &AppError{Type: t, Message: message, Err: err}
}

func NewNotFound(msg string) *AppError            { return New(NotFound, msg) }
func NewBadRequest(msg string) *AppError          { return New(BadRequest, msg) }
func NewConflict(msg string) *AppError            { return New(Conflict, msg) }
func NewForbidden(msg string) *AppError           { return New(Forbidden, msg) }
func NewInternal(msg string, err error) *AppError { return Wrap(Internal, msg, err) }
func NewUnavailable(msg string) *AppError         { return New(Unavailable, msg) }
