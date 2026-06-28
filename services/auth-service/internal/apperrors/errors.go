package apperrors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
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

// New creates a new AppError with the given status code and message.
func New(code int, message string) error {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new AppError with the given status code and formatted message.
func Newf(code int, format string, args ...interface{}) error {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error into an AppError with the given status code and message.
func Wrap(err error, code int, message string) error {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf wraps an existing error into an AppError with the given status code and formatted message.
func Wrapf(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// FromError attempts to extract an AppError from the given error chain.
func FromError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	var httpErr *AppError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}
