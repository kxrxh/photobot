package httperr

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	t.Run("with wrapped error includes both message and cause", func(t *testing.T) {
		e := &Error{Code: 400, Message: "bad request", Err: errors.New("invalid input")}
		got := e.Error()
		assert.Contains(t, got, "bad request")
	})
	t.Run("without wrapped error returns only message", func(t *testing.T) {
		e := &Error{Code: 404, Message: "not found"}
		assert.Equal(t, "not found", e.Error())
	})
	t.Run("nil receiver", func(t *testing.T) {
		var e *Error
		assert.Equal(t, "http error", e.Error())
	})
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	e := &Error{Code: 500, Message: "server error", Err: inner}
	assert.Equal(t, inner, e.Unwrap())

	e2 := &Error{Code: 404, Message: "not found"}
	assert.Nil(t, e2.Unwrap())
}

func TestNew(t *testing.T) {
	err := New(400, "bad request")
	require.NotNil(t, err)
	httpErr, ok := FromError(err)
	require.True(t, ok)
	assert.Equal(t, 400, httpErr.Code)
	assert.Equal(t, "bad request", httpErr.Message)
	assert.Nil(t, httpErr.Err)
}

func TestNewf(t *testing.T) {
	err := Newf(404, "resource %q not found", "user-123")
	require.NotNil(t, err)
	httpErr, ok := FromError(err)
	require.True(t, ok)
	assert.Equal(t, 404, httpErr.Code)
	assert.Equal(t, "resource \"user-123\" not found", httpErr.Message)
}

func TestWrap_NilError(t *testing.T) {
	err := Wrap(nil, 500, "internal error")
	assert.Nil(t, err)
}

func TestWrap_WithError(t *testing.T) {
	inner := errors.New("db connection failed")
	err := Wrap(inner, 503, "Service Unavailable")
	require.NotNil(t, err)
	httpErr, ok := FromError(err)
	require.True(t, ok)
	assert.Equal(t, 503, httpErr.Code)
	assert.Equal(t, "Service Unavailable", httpErr.Message)
	assert.Equal(t, inner, httpErr.Err)
	assert.True(t, errors.Is(err, inner))
}

func TestWrapf_NilError(t *testing.T) {
	err := Wrapf(nil, 400, "invalid %s", "id")
	assert.Nil(t, err)
}

func TestWrapf_WithError(t *testing.T) {
	inner := errors.New("validation failed")
	err := Wrapf(inner, 422, "validation error: %v", "email required")
	require.NotNil(t, err)
	httpErr, ok := FromError(err)
	require.True(t, ok)
	assert.Equal(t, 422, httpErr.Code)
	assert.Equal(t, "validation error: email required", httpErr.Message)
	assert.Equal(t, inner, httpErr.Err)
}

func TestFromError_Nil(t *testing.T) {
	httpErr, ok := FromError(nil)
	assert.Nil(t, httpErr)
	assert.False(t, ok)
}

func TestFromError_PlainError(t *testing.T) {
	plain := errors.New("plain error")
	httpErr, ok := FromError(plain)
	assert.Nil(t, httpErr)
	assert.False(t, ok)
}

func TestFromError_Error(t *testing.T) {
	he := &Error{Code: 500, Message: "internal"}
	httpErr, ok := FromError(he)
	require.True(t, ok)
	assert.Same(t, he, httpErr)
}

func TestFromError_WrappedError(t *testing.T) {
	he := &Error{Code: 404, Message: "not found"}
	wrapped := errors.Join(he, errors.New("extra"))
	httpErr, ok := FromError(wrapped)
	require.True(t, ok)
	assert.Same(t, he, httpErr)
}

func TestNew_CreatesErrorWithCodeAndMessage(t *testing.T) {
	err := New(fiber.StatusNotFound, "not found")
	httpErr, ok := FromError(err)
	assert.True(t, ok)
	assert.Equal(t, fiber.StatusNotFound, httpErr.Code)
	assert.Equal(t, "not found", httpErr.Message)
	assert.Nil(t, httpErr.Err)
}
