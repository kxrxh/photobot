//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"csort.ru/analysis-service/internal/api/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerRecalculateAgainstHTTPServer(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recalculate" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Response":"new-analysis-id"}`))
	}))
	t.Cleanup(srv.Close)

	c := worker.NewClient(worker.Config{
		BaseURL: srv.URL,
		Timeout: 5 * time.Second,
	})
	got, err := c.RecalculateAnalysis(ctx, "temp-xyz", []string{"obj-1"})
	require.NoError(t, err)
	assert.Equal(t, "new-analysis-id", got)
}
