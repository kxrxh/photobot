//go:build integration

package integration

import (
	"net/http"
	"testing"

	"csort.ru/auth-service/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicesCRUD(t *testing.T) {
	env := testEnv(t)
	base := apiV1(env.BaseURL)
	authHeader := bearer(getAdminToken(t, env))

	serviceID := "integration-test-service"
	serviceSecret := "secret12345"

	t.Run("create success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     serviceID,
			"service_secret": serviceSecret,
		})
		resp := doPOST(t, base+"/services/", "application/json", body, authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("create duplicate returns 409", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     serviceID,
			"service_secret": serviceSecret,
		})
		resp := doPOST(t, base+"/services/", "application/json", body, authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("create missing fields returns 400", func(t *testing.T) {
		resp := doPOST(t, base+"/services/", "application/json", []byte("{}"), authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create without auth returns 401", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     "unauth-service",
			"service_secret": "secret",
		})
		resp := doPOST(t, base+"/services/", "application/json", body, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("list success", func(t *testing.T) {
		resp := doGET(t, base+"/services/", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result dto.Response[[]struct {
			ServiceID string `json:"service_id"`
		}]
		parseJSON(t, resp, &result)
		require.True(t, result.Success)
		var found bool
		for _, s := range result.Result {
			if s.ServiceID == serviceID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("service login", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     serviceID,
			"service_secret": serviceSecret,
		})
		resp := doPOST(t, base+"/auth/login", "application/json", body, map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result ServiceLoginResponse
		parseJSON(t, resp, &result)
		require.True(t, result.Success)
		assert.NotEmpty(t, result.Result.AccessToken)
	})

	t.Run("delete success", func(t *testing.T) {
		resp := doDELETE(t, base+"/services/"+serviceID, authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("delete nonexistent succeeds idempotently", func(t *testing.T) {
		resp := doDELETE(t, base+"/services/nonexistent-service-id", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("delete without auth returns 401", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     "cleanup-service",
			"service_secret": "secret",
		})
		createResp := doPOST(t, base+"/services/", "application/json", body, authHeader)
		createResp.Body.Close()
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		resp := doDELETE(t, base+"/services/cleanup-service", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
