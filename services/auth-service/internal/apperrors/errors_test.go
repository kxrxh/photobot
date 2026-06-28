package apperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(400, "Bad Request")
	requireAppError(t, err, 400, "Bad Request", nil)
}

func TestNewf(t *testing.T) {
	err := Newf(404, "user %d not found", 42)
	requireAppError(t, err, 404, "user 42 not found", nil)
}

func TestWrap(t *testing.T) {
	inner := errors.New("database connection failed")
	err := Wrap(inner, 500, "Internal Server Error")
	requireAppError(t, err, 500, "Internal Server Error", inner)
	assert.ErrorIs(t, err, inner)
}

func TestWrap_NilErrorReturnsNil(t *testing.T) {
	err := Wrap(nil, 500, "message")
	assert.Nil(t, err)
}

func TestWrapf(t *testing.T) {
	inner := errors.New("file not found")
	err := Wrapf(inner, 404, "resource %s not found", "document.pdf")
	requireAppError(t, err, 404, "resource document.pdf not found", inner)
	assert.ErrorIs(t, err, inner)
}

func TestWrapf_NilErrorReturnsNil(t *testing.T) {
	err := Wrapf(nil, 500, "format %s", "arg")
	assert.Nil(t, err)
}

func TestAppError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := New(400, "Invalid input")
		assert.Equal(t, "Invalid input", err.Error())
	})

	t.Run("with wrapped error", func(t *testing.T) {
		inner := errors.New("validation failed")
		err := Wrap(inner, 400, "Invalid input")
		assert.Contains(t, err.Error(), "Invalid input")
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestAppError_Unwrap(t *testing.T) {
	inner := errors.New("underlying")
	err := Wrap(inner, 500, "wrapped")
	httpErr, ok := err.(*AppError)
	requireAppError(t, err, 500, "wrapped", inner)
	assert.True(t, ok)
	assert.Same(t, inner, httpErr.Unwrap())
}

func TestFromError(t *testing.T) {
	t.Run("AppError returns true", func(t *testing.T) {
		orig := New(400, "Bad Request")
		httpErr, ok := FromError(orig)
		assert.True(t, ok)
		assert.NotNil(t, httpErr)
		assert.Equal(t, 400, httpErr.Code)
		assert.Equal(t, "Bad Request", httpErr.Message)
	})

	t.Run("wrapped AppError returns true", func(t *testing.T) {
		orig := Wrap(errors.New("inner"), 500, "Server Error")
		httpErr, ok := FromError(orig)
		assert.True(t, ok)
		assert.NotNil(t, httpErr)
		assert.Equal(t, 500, httpErr.Code)
	})

	t.Run("non-AppError returns false", func(t *testing.T) {
		plain := errors.New("plain error")
		httpErr, ok := FromError(plain)
		assert.False(t, ok)
		assert.Nil(t, httpErr)
	})

	t.Run("nil returns false", func(t *testing.T) {
		httpErr, ok := FromError(nil)
		assert.False(t, ok)
		assert.Nil(t, httpErr)
	})

	t.Run("double wrapped AppError", func(t *testing.T) {
		inner := New(404, "Not Found")
		wrapped := Wrap(inner, 502, "Upstream error")
		httpErr, ok := FromError(wrapped)
		assert.True(t, ok)
		assert.Equal(t, 502, httpErr.Code)
		var innerApp *AppError
		assert.True(t, errors.As(wrapped, &innerApp))
	})
}

func requireAppError(t *testing.T, err error, code int, message string, wrapped error) {
	t.Helper()
	assert.NotNil(t, err)
	httpErr, ok := err.(*AppError)
	assert.True(t, ok, "expected *AppError, got %T", err)
	assert.Equal(t, code, httpErr.Code)
	assert.Equal(t, message, httpErr.Message)
	if wrapped != nil {
		assert.NotNil(t, httpErr.Err)
		assert.ErrorIs(t, err, wrapped)
	} else {
		assert.Nil(t, httpErr.Err)
	}
}
