//go:build integration

package integration

import (
	"net/http"
	"strconv"
	"testing"

	"csort.ru/auth-service/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersCRUD(t *testing.T) {
	env := testEnv(t)
	base := apiV1(env.BaseURL)
	token := getAdminToken(t, env)
	authHeader := bearer(token)

	t.Run("list success", func(t *testing.T) {
		resp := doGET(t, base+"/users/", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result dto.Response[[]any]
		parseJSON(t, resp, &result)
		require.True(t, result.Success)
		assert.NotNil(t, result.Result)
	})

	t.Run("get by id success", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDList)

		resp := doGET(t, base+"/users/"+strconv.Itoa(int(userID)), authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			ID int32 `json:"id"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, userID, r.Result.ID)
	})

	t.Run("get by id not found returns 404", func(t *testing.T) {
		resp := doGET(t, base+"/users/"+strconv.Itoa(TestUserNotFoundID), authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("get by id invalid format returns 400", func(t *testing.T) {
		resp := doGET(t, base+"/users/abc", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get by messenger id (max) success", func(t *testing.T) {
		maxID := int64(TestUserIDMaxMessenger)
		userID := createTestUser(t, env, maxID)

		resp := doGET(
			t,
			base+"/users/by-messenger-id/"+strconv.FormatInt(maxID, 10)+"?platform=max",
			authHeader,
		)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			ID int32 `json:"id"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, userID, r.Result.ID)
	})

	t.Run("get by messenger id (telegram) success", func(t *testing.T) {
		telegramID := int64(TestUserIDTelegramMessenger)
		userID := createTestUserWithTelegramId(t, env, telegramID)

		resp := doGET(
			t,
			base+"/users/by-messenger-id/"+strconv.FormatInt(telegramID, 10)+"?platform=telegram",
			authHeader,
		)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			ID int32 `json:"id"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, userID, r.Result.ID)
	})

	t.Run("get by messenger id with service token success", func(t *testing.T) {
		telegramID := int64(TestUserIDServiceToken)
		userID := createTestUserWithTelegramId(t, env, telegramID)

		serviceID := "users-test-service"
		serviceSecret := "users-test-secret-12345"

		svcBody := mustJSON(t, map[string]string{
			"service_id":     serviceID,
			"service_secret": serviceSecret,
		})
		createResp := doPOST(t, base+"/services/", "application/json", svcBody, authHeader)
		createResp.Body.Close()
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		loginResp := doPOST(t, base+"/auth/login", "application/json", svcBody, map[string]string{
			"X-Grant-Type": "client_credentials",
		})
		defer loginResp.Body.Close()
		require.Equal(t, http.StatusOK, loginResp.StatusCode)

		var login ServiceLoginResponse
		parseJSON(t, loginResp, &login)
		require.True(t, login.Success)
		require.NotEmpty(t, login.Result.AccessToken)

		svcHeader := bearer(login.Result.AccessToken)
		resp := doGET(
			t,
			base+"/users/by-messenger-id/"+strconv.FormatInt(telegramID, 10)+"?platform=telegram",
			svcHeader,
		)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			ID int32 `json:"id"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, userID, r.Result.ID)
	})

	t.Run("get by messenger id not found returns 404", func(t *testing.T) {
		resp := doGET(t, base+"/users/by-messenger-id/999999?platform=telegram", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("get by messenger id missing platform returns 400", func(t *testing.T) {
		resp := doGET(t, base+"/users/by-messenger-id/999999", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get by messenger id invalid platform returns 400", func(t *testing.T) {
		resp := doGET(t, base+"/users/by-messenger-id/999999?platform=vk", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("legacy max route removed returns 404", func(t *testing.T) {
		resp := doGET(t, base+"/users/max/999999", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("legacy telegram route removed returns 404", func(t *testing.T) {
		resp := doGET(t, base+"/users/telegram/999999", authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("update user success", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDUpdate)

		body := mustJSON(t, map[string]string{
			"full_name":         "Updated Name",
			"organization_name": "Test Org",
		})
		resp := doPUT(
			t,
			base+"/users/"+strconv.Itoa(int(userID)),
			"application/json",
			body,
			authHeader,
		)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			Data struct {
				FullName         string `json:"full_name"`
				OrganizationName string `json:"organization_name"`
			} `json:"data"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, "Updated Name", r.Result.Data.FullName)
		assert.Equal(t, "Test Org", r.Result.Data.OrganizationName)
	})

	t.Run("update user not found returns 404", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"full_name": "x"})

		resp := doPUT(t, base+"/users/"+strconv.Itoa(TestUserNotFoundID), "application/json", body, authHeader)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("get me success", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDGetMe)
		userToken := getUserToken(t, userID)
		userAuth := bearer(userToken)

		resp := doGET(t, base+"/users/me", userAuth)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			ID int32 `json:"id"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, userID, r.Result.ID)
	})

	t.Run("update me success", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDUpdateMe)
		userToken := getUserToken(t, userID)
		userAuth := bearer(userToken)

		body := mustJSON(t, map[string]string{"full_name": "My Updated Name"})
		resp := doPUT(t, base+"/users/me", "application/json", body, userAuth)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[struct {
			Data struct {
				FullName string `json:"full_name"`
			} `json:"data"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, "My Updated Name", r.Result.Data.FullName)
	})

	t.Run("get my roles success", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDServiceToken)
		userToken := getUserToken(t, userID)
		userAuth := bearer(userToken)

		resp := doGET(t, base+"/users/me/roles", userAuth)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var r dto.Response[[]struct {
			ID int `json:"id"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.NotNil(t, r.Result)
	})
}
