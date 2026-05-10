package errors

import (
	"fmt"
	"net/http"
)

var (
	ErrNotFound           = NewAppError(http.StatusNotFound, "RESOURCE_NOT_FOUND", "resource not found")
	ErrUnauthorized       = NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	ErrForbidden          = NewAppError(http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
	ErrBadRequest         = NewAppError(http.StatusBadRequest, "BAD_REQUEST", "invalid request")
	ErrInternal           = NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	ErrServiceUnavailable = NewAppError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "service temporarily unavailable")
	ErrConflict           = NewAppError(http.StatusConflict, "CONFLICT", "resource conflict")
	ErrRateLimited        = NewAppError(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
	ErrTimeout            = NewAppError(http.StatusGatewayTimeout, "TIMEOUT", "request timed out")
)

type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

func NewAppError(statusCode int, code, message string) *AppError {
	return &AppError{StatusCode: statusCode, Code: code, Message: message}
}

func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{StatusCode: e.StatusCode, Code: e.Code, Message: e.Message, Details: details, Err: e.Err}
}

func (e *AppError) WithError(err error) *AppError {
	return &AppError{StatusCode: e.StatusCode, Code: e.Code, Message: e.Message, Details: e.Details, Err: err}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve *ValidationErrors) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(ve.Errors))
}

func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{Field: field, Message: message})
}

func (ve *ValidationErrors) HasErrors() bool { return len(ve.Errors) > 0 }

type TelemetryError struct {
	Source string
	Metric string
	Reason string
	Err    error
}

func (te *TelemetryError) Error() string {
	return fmt.Sprintf("telemetry error [%s/%s]: %s", te.Source, te.Metric, te.Reason)
}

func (te *TelemetryError) Unwrap() error { return te.Err }

func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.StatusCode {
		case http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusTooManyRequests:
			return true
		}
	}
	return false
}
