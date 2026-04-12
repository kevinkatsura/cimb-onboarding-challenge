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

// Predefined error types based on Indonesia SNAP BI guidelines
// Pattern: http_status_code + service_code + case_code
// Here we define only the CaseCode, the builder will assemble the full code.
var (
	ErrSuccess             = ErrorType{HTTPCode: 200, CaseCode: "00", Category: CategorySystem, Message: "Successful"}
	ErrBadRequest          = ErrorType{HTTPCode: 400, CaseCode: "00", Category: CategorySystem, Message: "Bad Request"}
	ErrInvalidFieldFormat  = ErrorType{HTTPCode: 400, CaseCode: "01", Category: CategoryMessage, Message: "Invalid Field Format"}
	ErrInvalidMandatory    = ErrorType{HTTPCode: 400, CaseCode: "02", Category: CategoryMessage, Message: "Invalid Mandatory Field"}
	ErrUnauthorized        = ErrorType{HTTPCode: 401, CaseCode: "00", Category: CategorySystem, Message: "Unauthorized"}
	ErrInvalidSignature    = ErrorType{HTTPCode: 401, CaseCode: "01", Category: CategorySystem, Message: "Invalid Signature"}
	ErrInvalidToken        = ErrorType{HTTPCode: 401, CaseCode: "03", Category: CategorySystem, Message: "Invalid Token (B2B)"}
	ErrTransactionExpired  = ErrorType{HTTPCode: 403, CaseCode: "00", Category: CategoryBusiness, Message: "Transaction Expired"}
	ErrFeatureNotAllowed   = ErrorType{HTTPCode: 403, CaseCode: "01", Category: CategorySystem, Message: "Feature Not Allowed"}
	ErrExceedsAmountLimit  = ErrorType{HTTPCode: 403, CaseCode: "02", Category: CategoryBusiness, Message: "Exceeds Transaction Amount Limit"}
	ErrInsufficientFunds   = ErrorType{HTTPCode: 403, CaseCode: "14", Category: CategoryBusiness, Message: "Insufficient Funds"}
	ErrTransactionNotFound = ErrorType{HTTPCode: 404, CaseCode: "01", Category: CategoryBusiness, Message: "Transaction Not Found"}
	ErrConflict            = ErrorType{HTTPCode: 409, CaseCode: "00", Category: CategorySystem, Message: "Conflict"}
	ErrDuplicateTransfer   = ErrorType{HTTPCode: 409, CaseCode: "01", Category: CategoryBusiness, Message: "Duplicate Partner Reference Number"}
	ErrGeneralError        = ErrorType{HTTPCode: 500, CaseCode: "00", Category: CategorySystem, Message: "General Error"}
	ErrInternalServerError = ErrorType{HTTPCode: 500, CaseCode: "01", Category: CategorySystem, Message: "Internal Server Error"}
	ErrGatewayTimeout      = ErrorType{HTTPCode: 504, CaseCode: "00", Category: CategorySystem, Message: "Timeout"}
)

type AppError struct {
	Type          ErrorType
	Err           error
	CustomMessage string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		msg := e.Type.Message
		if e.CustomMessage != "" {
			msg = e.CustomMessage
		}
		return fmt.Sprintf("%s: %v", msg, e.Err)
	}
	if e.CustomMessage != "" {
		return e.CustomMessage
	}
	return e.Type.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new application error representing standard SNAP topologies
func New(t ErrorType, customMessage string) *AppError {
	return &AppError{Type: t, CustomMessage: customMessage}
}

// Wrap nests an error within a SNAP taxonomy explicitly
func Wrap(t ErrorType, customMessage string, err error) *AppError {
	return &AppError{Type: t, CustomMessage: customMessage, Err: err}
}

// Helper initializers mapping seamlessly to backward-compatible service uses.
func NewBadRequest(msg string) *AppError          { return New(ErrBadRequest, msg) }
func NewNotFound(msg string) *AppError            { return New(ErrTransactionNotFound, msg) }
func NewConflict(msg string) *AppError            { return New(ErrConflict, msg) }
func NewInternal(msg string, err error) *AppError { return Wrap(ErrInternalServerError, msg, err) }
func NewUnavailable(msg string) *AppError         { return New(ErrGatewayTimeout, msg) }
