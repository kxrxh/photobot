//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

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

type successMsgBody struct {
	Success bool `json:"success"`
	Result  struct {
		Message string `json:"message"`
	} `json:"result"`
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
