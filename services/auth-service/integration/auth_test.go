//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminLogin(t *testing.T) {
	env := testEnv(t)
	t.Run("success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"login":    env.Config.Security.AdminLogin,
			"password": "password",
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "password",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result AuthResponse
		parseJSON(t, resp, &result)
		require.True(t, result.Success)
		assert.NotEmpty(t, result.Result.AccessToken)
		assert.NotEmpty(t, result.Result.RefreshToken)
	})

	t.Run("wrong password", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"login":    env.Config.Security.AdminLogin,
			"password": "wrong",
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "password",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("wrong login", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"login":    "wrong-admin",
			"password": "password",
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "password",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid grant type returns 400", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"login": "admin", "password": "x"})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "invalid",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing body returns 400", func(t *testing.T) {
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", []byte("{}"), map[string]string{
			"X-Grant-Type": "password",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestServiceLogin(t *testing.T) {
	env := testEnv(t)
	authHeader := bearer(getAdminToken(t, env))

	serviceID := "auth-test-service"
	serviceSecret := "auth-test-secret-12345"

	createBody := mustJSON(t, map[string]string{
		"service_id":     serviceID,
		"service_secret": serviceSecret,
	})
	createResp := doPOST(
		t,
		apiV1(env.BaseURL, "services")+"/",
		"application/json",
		createBody,
		authHeader,
	)
	createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	t.Run("success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     serviceID,
			"service_secret": serviceSecret,
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result AuthResponse
		parseJSON(t, resp, &result)
		require.True(t, result.Success)
		assert.NotEmpty(t, result.Result.AccessToken)
		assert.NotEmpty(t, result.Result.RefreshToken)
		assert.Contains(t, result.Result.Roles, "service")
	})

	t.Run("wrong secret", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     serviceID,
			"service_secret": "wrong-secret",
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid service", func(t *testing.T) {
		body := mustJSON(t, map[string]string{
			"service_id":     "non-existent-service",
			"service_secret": "any-secret",
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing credentials returns 400", func(t *testing.T) {
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", []byte("{}"), map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestRefresh(t *testing.T) {
	env := testEnv(t)
	t.Run("invalid token returns 401", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"refresh_token": "invalid.refresh.token"})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "refresh"), "application/json", body, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing refresh token returns 400", func(t *testing.T) {
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "refresh"), "application/json", []byte("{}"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		loginBody := mustJSON(t, map[string]string{
			"login":    env.Config.Security.AdminLogin,
			"password": "password",
		})
		loginResp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", loginBody, map[string]string{
			"X-Grant-Type": "password",
		})
		require.Equal(t, http.StatusOK, loginResp.StatusCode)
		var loginResult RefreshTokenResponse
		parseJSON(t, loginResp, &loginResult)
		require.NotEmpty(t, loginResult.Result.RefreshToken)

		refreshBody := mustJSON(t, map[string]string{
			"refresh_token": loginResult.Result.RefreshToken,
		})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "refresh"), "application/json", refreshBody, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var refreshResult RefreshResponse
		parseJSON(t, resp, &refreshResult)
		require.True(t, refreshResult.Success)
		assert.NotEmpty(t, refreshResult.Result.AccessToken)
	})
}

func TestRegister(t *testing.T) {
	env := testEnv(t)
	t.Run("missing X-Bot-Name returns 400", func(t *testing.T) {
		body := mustJSON(t, map[string]string{})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "register"), "application/json", body, map[string]string{
			"X-Init-Data": "x",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing X-Init-Data returns 400", func(t *testing.T) {
		body := mustJSON(t, map[string]string{})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "register"), "application/json", body, map[string]string{
			"X-Bot-Name": "test-bot",
		})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing both returns 400", func(t *testing.T) {
		body := mustJSON(t, map[string]string{})
		resp := doPOST(t, apiV1(env.BaseURL, "auth", "register"), "application/json", body, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestLinkCode(t *testing.T) {
	env := testEnv(t)
	t.Run("request link code success", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDLinkCode1)
		userToken := getUserToken(t, userID)
		userAuth := map[string]string{"Authorization": "Bearer " + userToken}

		resp := doPOST(t, apiV1(env.BaseURL, "auth", "link-code"), "application/json", nil, userAuth)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r LinkCodeResponse
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		require.Len(t, r.Result.Code, 6)
		assert.Greater(t, r.Result.ExpiresInSeconds, 0)
	})
}

func TestLinkWithCode(t *testing.T) {
	env := testEnv(t)
	t.Run("missing X-Bot-Name returns 400", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDLinkCode2)
		userToken := getUserToken(t, userID)
		userAuth := map[string]string{
			"Authorization":        "Bearer " + userToken,
			"X-Init-Data":          "x",
			"X-Messenger-Platform": "telegram",
		}
		body := mustJSON(t, map[string]string{"code": "123456"})

		resp := doPOST(t, apiV1(env.BaseURL, "auth", "link-with-code"), "application/json", body, userAuth)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid code length returns 400", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDLinkCode3)
		userToken := getUserToken(t, userID)
		userAuth := map[string]string{
			"Authorization":        "Bearer " + userToken,
			"X-Bot-Name":           "test-bot",
			"X-Init-Data":          "x",
			"X-Messenger-Platform": "telegram",
		}
		body := mustJSON(t, map[string]string{"code": "12345"})

		resp := doPOST(t, apiV1(env.BaseURL, "auth", "link-with-code"), "application/json", body, userAuth)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestJWKS(t *testing.T) {
	env := testEnv(t)

	resp := doGET(t, apiV1(env.BaseURL, "auth", ".well-known", "jwks.json"), nil)

	var result struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
		} `json:"keys"`
	}
	readResponseJSON(t, resp, http.StatusOK, &result)
	require.Len(t, result.Keys, 1)
	assert.Equal(t, "RSA", result.Keys[0].Kty)
	assert.Equal(t, "RS256", result.Keys[0].Alg)
	assert.NotEmpty(t, result.Keys[0].Kid)
}
