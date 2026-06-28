package apierrors

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return "api error"
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

func From(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func FromError(err error) (*Error, bool) {
	return From(err)
}

func WithDetails(err error, details any) error {
	ae, ok := From(err)
	if !ok {
		return err
	}
	return &Error{
		Code:    ae.Code,
		Message: ae.Message,
		Details: details,
		Err:     ae.Err,
	}
}
