package errx

import (
	"errors"
	"fmt"
)

// Code represents a machine-readable error code.
type Code string

const (
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeValidation      Code = "VALIDATION_ERROR"
	CodeNotFound        Code = "NOT_FOUND"
	CodeAlreadyExists   Code = "ALREADY_EXISTS"
	CodeUnauthenticated Code = "UNAUTHENTICATED"
	CodePermission      Code = "PERMISSION_DENIED"
	CodeConflict        Code = "CONFLICT"
	CodeRateLimited     Code = "RATE_LIMITED"
	CodePrecondition    Code = "PRECONDITION_FAILED"
	CodeInternal        Code = "INTERNAL_ERROR"
	CodeUnavailable     Code = "SERVICE_UNAVAILABLE"
	CodeTimeout         Code = "TIMEOUT"
	CodeUnauthorized    Code = "UNAUTHORIZED" // alias for permission problems from auth layer
)

// AppError is the core structured error used across the application.
// It wraps an underlying error and carries a stable code and optional metadata.
type AppError struct {
	Code      Code                   `json:"code"`
	Message   string                 `json:"message"`
	Op        string                 `json:"op,omitempty"`
	Err       error                  `json:"-"` // underlying error (not serialized)
	Meta      map[string]interface{} `json:"meta,omitempty"`
	Transient bool                   `json:"-"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Op != "" {
		return fmt.Sprintf("%s: %s: %s", e.Op, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// WithMeta returns a shallow clone with merged metadata.
func (e *AppError) WithMeta(k string, v interface{}) *AppError {
	if e.Meta == nil {
		e.Meta = map[string]interface{}{}
	}
	e.Meta[k] = v
	return e
}

// New creates a new AppError.
func New(code Code, msg string) *AppError { return &AppError{Code: code, Message: msg} }

// E builds an AppError using variadic arguments for ergonomics.
// Acceptable argument types: Code, string (message), error (wrapped), map[string]interface{} (meta), Op(op string), Transient(bool).
func E(args ...interface{}) *AppError {
	var (
		code      Code = CodeInternal
		msg       string
		op        string
		err       error
		meta      map[string]interface{}
		transient bool
	)
	for _, a := range args {
		switch v := a.(type) {
		case Code:
			code = v
		case string:
			if msg == "" {
				msg = v
			} else { // second string becomes op if op empty
				if op == "" {
					op = v
				} else {
					msg += " " + v
				}
			}
		case error:
			err = v
		case map[string]interface{}:
			if meta == nil {
				meta = map[string]interface{}{}
			}
			for mk, mv := range v {
				meta[mk] = mv
			}
		case Op:
			op = string(v)
		case Transient:
			transient = bool(v)
		}
	}
	if msg == "" {
		msg = string(code)
	}
	return &AppError{Code: code, Message: msg, Op: op, Err: err, Meta: meta, Transient: transient}
}

type Op string

type Transient bool

// IsCode checks if any error in the chain matches the given code.
func IsCode(err error, code Code) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		if ae.Code == code {
			return true
		}
	}
	if u := errors.Unwrap(err); u != nil {
		return IsCode(u, code)
	}
	return false
}

// ToAppError converts arbitrary error to *AppError (wrapping if needed)
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return &AppError{Code: CodeInternal, Message: err.Error(), Err: err}
}
