//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifications_WithUserToken(t *testing.T) {
	env := testEnv(t)
	resp := doGET(t, apiV1(env.BaseURL, "classifications"), bearer(env.UserToken))

	var wrapper struct {
		Success bool `json:"success"`
		Result  struct {
			Classifications      []any `json:"classifications"`
			ActiveClassification any   `json:"active_classification"`
		} `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusOK, &wrapper)
	require.True(t, wrapper.Success)
	assert.NotNil(t, wrapper.Result.Classifications)
}
