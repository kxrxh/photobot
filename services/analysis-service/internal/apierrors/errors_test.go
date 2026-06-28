package apierrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndFrom(t *testing.T) {
	t.Run("New creates error with code and message", func(t *testing.T) {
		err := New(http.StatusNotFound, "not found")
		ae, ok := From(err)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, ae.Code)
		assert.Equal(t, "not found", ae.Message)
		assert.Nil(t, ae.Err)
	})

	t.Run("Wrap wraps underlying error", func(t *testing.T) {
		inner := errors.New("db error")
		err := Wrap(inner, http.StatusInternalServerError, "failed")
		ae, ok := From(err)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, ae.Code)
		assert.Equal(t, "failed", ae.Message)
		assert.Equal(t, inner, ae.Err)
		assert.True(t, errors.Is(err, inner))
	})

	t.Run("Wrap returns nil when err is nil", func(t *testing.T) {
		assert.Nil(t, Wrap(nil, http.StatusBadRequest, "msg"))
	})

	t.Run("From returns false for non-Error", func(t *testing.T) {
		_, ok := From(errors.New("plain"))
		assert.False(t, ok)
	})

	t.Run("FromError alias", func(t *testing.T) {
		err := New(http.StatusBadGateway, "x")
		a, ok := FromError(err)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadGateway, a.Code)
	})
}

func TestWithDetails(t *testing.T) {
	base := BadRequest("x")
	with := WithDetails(base, []string{"a"})
	ae, ok := From(with)
	require.True(t, ok)
	assert.Equal(t, "x", ae.Message)
	require.Equal(t, []string{"a"}, ae.Details)
}

func TestHelpers(t *testing.T) {
	t.Run("BadRequest 400", func(t *testing.T) {
		err := BadRequest("x")
		var ae *Error
		require.True(t, errors.As(err, &ae))
		assert.Equal(t, http.StatusBadRequest, ae.Code)
	})

	t.Run("UpstreamFailed 502", func(t *testing.T) {
		err := UpstreamFailed(errors.New("worker down"))
		var ae *Error
		require.True(t, errors.As(err, &ae))
		assert.Equal(t, http.StatusBadGateway, ae.Code)
		assert.Contains(t, ae.Message, "recalculate")
	})

	t.Run("UpstreamFailed nil", func(t *testing.T) {
		assert.Nil(t, UpstreamFailed(nil))
	})
}
