package apperror

import "net/http"

// Code is a stable, machine-readable error identifier.
type Code string

const (
	CodeInvalidArgument    Code = "invalid_argument"
	CodeUnauthenticated    Code = "unauthenticated"
	CodeForbidden          Code = "forbidden"
	CodeNotFound           Code = "not_found"
	CodeConflict           Code = "conflict"
	CodeInvalidTimezone    Code = "invalid_timezone"
	CodeActiveSleepExists  Code = "active_sleep_exists"
	CodeInvalidSleepInterval Code = "invalid_sleep_interval"
	CodeInvalidInviteLink  Code = "invalid_invite_link"
	CodeInternalError      Code = "internal_error"
)

// AppError is a structured application error with a stable code and a human-readable message.
type AppError struct {
	Code    Code
	Message string
}

func (e AppError) Error() string {
	return e.Message
}

// New creates an AppError with the given code and message.
func New(code Code, message string) AppError {
	return AppError{Code: code, Message: message}
}

// HTTPStatus maps an error code to the appropriate HTTP status code.
func HTTPStatus(code Code) int {
	switch code {
	case CodeInvalidArgument, CodeInvalidTimezone, CodeInvalidSleepInterval, CodeInvalidInviteLink:
		return http.StatusBadRequest
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeActiveSleepExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
