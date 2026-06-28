package worker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecalculateAnalysis_RetriesOn5xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Response":"analysis-id"}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(Config{
		BaseURL:    srv.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
	})

	got, err := c.RecalculateAnalysis(context.Background(), "temp-1", nil)
	require.NoError(t, err)
	assert.Equal(t, "analysis-id", got)
	assert.Equal(t, int32(3), calls.Load())
}

func TestRecalculateAnalysis_FailsFastOn4xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(Config{
		BaseURL:    srv.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
	})

	_, err := c.RecalculateAnalysis(context.Background(), "temp-1", nil)
	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load())
}
