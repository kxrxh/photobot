//go:build integration

package integration

import (
	"net/http"
	"testing"

	"csort.ru/classification-service/integration/harness"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParams_List(t *testing.T) {
	env := testEnv(t)
	resp := doGET(t, apiV1(env.BaseURL, "params"), bearer(env.UserToken))

	var wrapper struct {
		Success bool  `json:"success"`
		Result  []any `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusOK, &wrapper)
	require.True(t, wrapper.Success)
	assert.NotNil(t, wrapper.Result)
}

func TestParams_CreateListDelete(t *testing.T) {
	env := testEnv(t)
	v1 := func(parts ...string) string { return apiV1(env.BaseURL, parts...) }

	adminHdr := bearer(env.AdminToken)
	userHdr := bearer(env.UserToken)

	t.Run("create requires admin", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "Test Param"})
		resp := doPOST(t, v1("params"), "application/json", body, userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("create success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "Integration Test Param"})
		resp := doPOST(t, v1("params"), "application/json", body, adminHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusCreated, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Integration Test Param", wrapper.Result.Name)
		_, err := uuid.Parse(wrapper.Result.ID)
		require.NoError(t, err)
	})

	paramID := harness.SeedParam(t, t.Context(), env.DBPool, "Seed Param For Delete")

	t.Run("delete requires admin", func(t *testing.T) {
		resp := doDELETE(t, v1("params", paramID.String()), userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("delete success", func(t *testing.T) {
		resp := doDELETE(t, v1("params", paramID.String()), adminHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})
}
