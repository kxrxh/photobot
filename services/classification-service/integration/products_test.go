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

func TestProducts_ListWithUserToken(t *testing.T) {
	env := testEnv(t)
	resp := doGET(t, apiV1(env.BaseURL, "products"), bearer(env.UserToken))

	var wrapper struct {
		Success bool  `json:"success"`
		Result  []any `json:"result"`
	}
	readResponseJSON(t, resp, http.StatusOK, &wrapper)
	require.True(t, wrapper.Success)
	assert.NotNil(t, wrapper.Result)
}

func TestProducts_CreateDeleteFlow(t *testing.T) {
	env := testEnv(t)
	v1 := func(parts ...string) string { return apiV1(env.BaseURL, parts...) }

	adminHdr := bearer(env.AdminToken)
	userHdr := bearer(env.UserToken)

	t.Run("create requires admin", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "Test Product"})
		resp := doPOST(t, v1("products"), "application/json", body, userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("create success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "Integration Test Product"})
		resp := doPOST(t, v1("products"), "application/json", body, adminHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusCreated, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Integration Test Product", wrapper.Result.Name)
		_, err := uuid.Parse(wrapper.Result.ID)
		require.NoError(t, err)
	})

	productID := harness.SeedProduct(t, t.Context(), env.DBPool, "Seed Product For Flow")

	t.Run("get by id", func(t *testing.T) {
		resp := doGET(t, v1("products", productID.String()), userHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Seed Product For Flow", wrapper.Result.Name)
	})

	t.Run("update requires admin", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "Updated Name"})
		resp := doPUT(t, v1("products", productID.String()), "application/json", body, userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("update success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "Updated Product Name"})
		resp := doPUT(t, v1("products", productID.String()), "application/json", body, adminHdr)

		var wrapper struct {
			Success bool `json:"success"`
			Result  struct {
				Name string `json:"name"`
			} `json:"result"`
		}
		readResponseJSON(t, resp, http.StatusOK, &wrapper)
		require.True(t, wrapper.Success)
		assert.Equal(t, "Updated Product Name", wrapper.Result.Name)
	})

	t.Run("delete requires admin", func(t *testing.T) {
		resp := doDELETE(t, v1("products", productID.String()), userHdr)
		requireHTTPStatus(t, resp, http.StatusForbidden)
	})

	t.Run("delete success", func(t *testing.T) {
		resp := doDELETE(t, v1("products", productID.String()), adminHdr)

		var out successMsgBody
		readResponseJSON(t, resp, http.StatusOK, &out)
		require.True(t, out.Success)
	})

	t.Run("get deleted returns 404", func(t *testing.T) {
		resp := doGET(t, v1("products", productID.String()), userHdr)
		requireHTTPStatus(t, resp, http.StatusNotFound)
	})
}
