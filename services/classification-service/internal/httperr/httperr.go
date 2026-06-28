package httperr

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return "http error"
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(code int, message string) error {
	return Newf(code, "%s", message)
}

func Newf(code int, format string, args ...any) error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func Wrap(err error, code int, message string) error {
	return Wrapf(err, code, "%s", message)
}

func Wrapf(err error, code int, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

func FromError(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var httpErr *Error
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}
