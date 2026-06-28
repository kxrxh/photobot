package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_IncludesBodyAndPeer(t *testing.T) {
	err := NewHTTPError(
		"classification-service",
		"GET",
		"http://classification/api/v1/foo",
		502,
		[]byte(`{"error":"bad gateway"}`),
		"unexpected status",
		nil,
	)

	s := err.Error()
	assert.Contains(t, s, "peer=classification-service")
	assert.Contains(t, s, "status=502")
	assert.Contains(t, s, `{"error":"bad gateway"}`)
}

func TestExtract(t *testing.T) {
	inner := NewHTTPError("worker", "POST", "/recalculate", 500, []byte("oops"), "failed", nil)
	wrapped := fmt.Errorf("wrap: %w", inner)

	got := Extract(wrapped)
	require.NotNil(t, got)
	assert.Equal(t, 500, got.StatusCode)
	assert.Equal(t, "worker", got.PeerService)

	assert.Nil(t, Extract(errors.New("other")))
}

func TestBodyPreview_Truncates(t *testing.T) {
	long := make([]byte, 600)
	for i := range long {
		long[i] = 'x'
	}
	err := &Error{Body: string(long)}
	assert.Contains(t, err.BodyPreview(100), "...(truncated)")
}
