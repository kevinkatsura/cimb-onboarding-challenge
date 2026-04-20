package apperror

import "fmt"

type Category string

const (
	CategorySystem   Category = "System"
	CategoryBusiness Category = "Business"
	CategoryMessage  Category = "Message"
)

type ErrorType struct {
	HTTPCode int
	CaseCode string
	Category Category
	Message  string
}

var (
	ErrBadRequest         = ErrorType{HTTPCode: 400, CaseCode: "00", Category: CategorySystem, Message: "Bad Request"}
	ErrInvalidFieldFormat = ErrorType{HTTPCode: 400, CaseCode: "01", Category: CategoryMessage, Message: "Invalid Field Format"}
	ErrInvalidMandatory   = ErrorType{HTTPCode: 400, CaseCode: "02", Category: CategoryMessage, Message: "Invalid Mandatory Field"}
	ErrUnauthorized       = ErrorType{HTTPCode: 401, CaseCode: "00", Category: CategorySystem, Message: "Unauthorized"}
	ErrNotFound           = ErrorType{HTTPCode: 404, CaseCode: "01", Category: CategoryBusiness, Message: "Not Found"}
	ErrConflict           = ErrorType{HTTPCode: 409, CaseCode: "00", Category: CategorySystem, Message: "Conflict"}
	ErrDuplicate          = ErrorType{HTTPCode: 409, CaseCode: "01", Category: CategoryBusiness, Message: "Duplicate Record"}
	ErrGeneralError       = ErrorType{HTTPCode: 500, CaseCode: "00", Category: CategorySystem, Message: "General Error"}
	ErrInternalServer     = ErrorType{HTTPCode: 500, CaseCode: "01", Category: CategorySystem, Message: "Internal Server Error"}
)

type AppError struct {
	Type    ErrorType
	Err     error
	Message string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Type.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(t ErrorType, msg string) *AppError              { return &AppError{Type: t, Message: msg} }
func Wrap(t ErrorType, msg string, err error) *AppError   { return &AppError{Type: t, Message: msg, Err: err} }
func NewBadRequest(msg string) *AppError                  { return New(ErrBadRequest, msg) }
func NewNotFound(msg string) *AppError                    { return New(ErrNotFound, msg) }
func NewConflict(msg string) *AppError                    { return New(ErrConflict, msg) }
func NewInternal(msg string, err error) *AppError         { return Wrap(ErrInternalServer, msg, err) }
