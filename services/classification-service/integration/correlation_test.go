//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrelation_WithUserToken(t *testing.T) {
	env := testEnv(t)
	body := mustJSON(t, map[string]any{
		"fractions": []map[string]any{
			{"name": "test", "object_ids": []int{1, 2, 3}},
		},
		"parameter_groups": []string{"all"},
	})

	resp := doPOST(
		t,
		apiV1(env.BaseURL, "correlation"),
		"application/json",
		body,
		bearer(env.UserToken),
	)

	var wrapper struct {
		Success bool `json:"success"`
		Result  []struct {
			Name string `json:"name"`
		} `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusOK, &wrapper)
	require.True(t, wrapper.Success)
	require.Len(t, wrapper.Result, 1)
	assert.Equal(t, "integration-mock", wrapper.Result[0].Name)
}
