//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"csort.ru/auth-service/integration/harness"
	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/dto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

var apiClient = &http.Client{Transport: &http.Transport{}}

func bearer(token string) map[string]string {
	if token == "" {
		return nil
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func apiV1(base string, pathParts ...string) string {
	u := base + "/api/v1"
	for _, p := range pathParts {
		u += "/" + p
	}
	return u
}

func doGET(t *testing.T, url string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := apiClient.Do(req)
	require.NoError(t, err)
	return resp
}

func doPOST(
	t *testing.T,
	url, contentType string,
	body []byte,
	headers map[string]string,
) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := apiClient.Do(req)
	require.NoError(t, err)
	return resp
}

func doPUT(
	t *testing.T,
	url, contentType string,
	body []byte,
	headers map[string]string,
) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPut, url, bodyReader)
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := apiClient.Do(req)
	require.NoError(t, err)
	return resp
}

func doDELETE(t *testing.T, url string, headers map[string]string, body ...[]byte) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if len(body) > 0 && body[0] != nil {
		bodyReader = bytes.NewReader(body[0])
	}
	req, err := http.NewRequest(http.MethodDelete, url, bodyReader)
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := apiClient.Do(req)
	require.NoError(t, err)
	return resp
}

func requireHTTPStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	defer resp.Body.Close()
	require.Equal(t, want, resp.StatusCode)
}

func readResponseJSON(t *testing.T, resp *http.Response, wantStatus int, v any) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, wantStatus, resp.StatusCode, "body: %s", string(body))
	if v != nil && resp.StatusCode == wantStatus {
		require.NoError(t, json.Unmarshal(body, v))
	}
}

func parseJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, v))
}

type AuthResponse = dto.Response[struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Roles        []string `json:"roles"`
}]

type RefreshTokenResponse = dto.Response[struct {
	RefreshToken string `json:"refresh_token"`
}]

type RefreshResponse = dto.Response[struct {
	AccessToken string `json:"access_token"`
}]

type LinkCodeResponse = dto.Response[struct {
	Code             string `json:"code"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}]

type CreateIDResponse = dto.Response[struct {
	ID int32 `json:"id"`
}]

type BotListResponse = dto.Response[[]struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}]

type BotResponse = dto.Response[struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}]

type BotTokenResponse = dto.Response[struct {
	Token string `json:"token"`
}]

type RoleResponse = dto.Response[struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}]

type RoleListResponse = dto.Response[[]struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}]

type UserRoleListResponse = dto.Response[[]struct {
	ID int32 `json:"id"`
}]

type ServiceLoginResponse = dto.Response[struct {
	AccessToken string `json:"access_token"`
}]

const (
	TestUserIDLinkCode1         = 300001
	TestUserIDLinkCode2         = 300002
	TestUserIDLinkCode3         = 300003
	TestUserIDRoleAssign        = 100001
	TestUserIDRoleDelete        = 100002
	TestUserIDList              = 200001
	TestUserIDMaxMessenger      = 200002
	TestUserIDTelegramMessenger = 200004
	TestUserIDUpdate            = 200005
	TestUserIDGetMe             = 200007
	TestUserIDUpdateMe          = 200008
	TestUserIDServiceToken      = 200009
	TestUserNotFoundID          = 99999
)

func getAdminToken(t *testing.T, env *harness.TestEnv) string {
	t.Helper()
	body := mustJSON(t, map[string]string{
		"login":    env.Config.Security.AdminLogin,
		"password": "password",
	})
	resp := doPOST(t, apiV1(env.BaseURL, "auth", "login"), "application/json", body, map[string]string{
		"X-Grant-Type": "password",
	})
	var result ServiceLoginResponse
	readResponseJSON(t, resp, http.StatusOK, &result)
	return result.Result.AccessToken
}

func createTestUser(t *testing.T, env *harness.TestEnv, maxID int64) int32 {
	t.Helper()
	queries := database.New(env.DBPool)
	user, err := queries.CreateUserWithMaxId(
		context.Background(),
		database.CreateUserWithMaxIdParams{
			MaxID: pgtype.Int8{Int64: maxID, Valid: true},
		},
	)
	require.NoError(t, err)
	return user.ID
}

func createTestUserWithTelegramId(t *testing.T, env *harness.TestEnv, telegramID int64) int32 {
	t.Helper()
	queries := database.New(env.DBPool)
	user, err := queries.CreateUser(
		context.Background(),
		database.CreateUserParams{
			TelegramID: pgtype.Int8{Int64: telegramID, Valid: true},
		},
	)
	require.NoError(t, err)
	return user.ID
}

func getUserToken(t *testing.T, userID int32) string {
	t.Helper()
	token, err := auth.GenerateJWT(
		&auth.GenerationParams{
			UserID: &userID,
			Roles:  []string{},
			GTY:    auth.GrantTypeInitData,
		},
		auth.AccessToken,
		15*time.Minute,
	)
	require.NoError(t, err)
	return token
}
