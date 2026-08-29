// Package domain contains transport-independent application concepts.
package domain

import (
	"errors"
	"fmt"
)

// ErrorKind classifies an application failure without assigning an HTTP status.
type ErrorKind string

const (
	KindInvalid      ErrorKind = "invalid"
	KindUnauthorized ErrorKind = "unauthorized"
	KindForbidden    ErrorKind = "forbidden"
	KindNotFound     ErrorKind = "not_found"
	KindConflict     ErrorKind = "conflict"
	KindUnavailable  ErrorKind = "unavailable"
	KindInternal     ErrorKind = "internal"
)

// Error is an application error. HTTP adapters decide how a Kind is presented.
type Error struct {
	Kind     ErrorKind
	Message  string
	Resource string
	Err      error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	if e.Message == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

func Invalid(message string) error {
	return &Error{Kind: KindInvalid, Message: message}
}

func Unauthorized(message string) error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

func Forbidden(message string) error {
	return &Error{Kind: KindForbidden, Message: message}
}

func NotFound(resource string, cause error) error {
	return &Error{Kind: KindNotFound, Message: resource + " not found", Resource: resource, Err: cause}
}

func Conflict(message string, cause error) error {
	return &Error{Kind: KindConflict, Message: message, Err: cause}
}

func Unavailable(message string, cause error) error {
	return &Error{Kind: KindUnavailable, Message: message, Err: cause}
}

func Internal(message string, cause error) error {
	return &Error{Kind: KindInternal, Message: message, Err: cause}
}

func AsError(err error) (*Error, bool) {
	var appErr *Error
	ok := errors.As(err, &appErr)
	return appErr, ok
}

func IsKind(err error, kind ErrorKind) bool {
	appErr, ok := AsError(err)
	return ok && appErr.Kind == kind
}
