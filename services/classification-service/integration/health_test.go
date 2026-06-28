//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	env := testEnv(t)

	resp := doGET(t, env.BaseURL+"/health", nil)

	var wrapper struct {
		Success bool `json:"success"`
		Result  struct {
			Status  string            `json:"status"`
			Service string            `json:"service"`
			Checks  map[string]string `json:"checks"`
		} `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusOK, &wrapper)
	require.True(t, wrapper.Success)
	require.Equal(t, "healthy", wrapper.Result.Status)
	require.Equal(t, "classification-service", wrapper.Result.Service)
	assert.Equal(t, "ok", wrapper.Result.Checks["database"])
	assert.Equal(t, "ok", wrapper.Result.Checks["identity_service"])
	assert.Equal(t, "ok", wrapper.Result.Checks["correlation_service"])
}
